package web

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/quickfeed/quickfeed/internal/cert"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/reload"
	"github.com/quickfeed/quickfeed/metrics"
	"golang.org/x/crypto/acme/autocert"
)

// hardcoded metrics server address
const metricsServerAddr = "127.0.0.1:9097"

type Server struct {
	httpServer     *http.Server
	redirectServer *http.Server
	metricsServer  *http.Server
	keyFile        string
	certFile       string
}

type ServerType func(addr string, handler http.Handler) (*Server, error)

func NewProductionServer(addr string, handler http.Handler) (*Server, error) {
	whitelist, err := env.Whitelist()
	if err != nil {
		return nil, fmt.Errorf("failed to get whitelist: %w", err)
	}
	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(env.CertPath()),
		HostPolicy: autocert.HostWhitelist(
			whitelist...,
		),
	}

	httpServer := &http.Server{
		Handler:           handler,
		Addr:              addr,
		ReadHeaderTimeout: 3 * time.Second, // to prevent Slowloris (CWE-400)
		WriteTimeout:      2 * time.Minute,
		ReadTimeout:       2 * time.Minute,
		TLSConfig:         certManager.TLSConfig(),
	}

	redirectServer := &http.Server{
		Handler:           certManager.HTTPHandler(nil),
		Addr:              ":http",
		ReadHeaderTimeout: 3 * time.Second, // to prevent Slowloris (CWE-400)
	}

	return &Server{
		httpServer:     httpServer,
		redirectServer: redirectServer,
		metricsServer:  metricsServer(),
	}, nil
}

func NewDevelopmentServer(addr string, handler http.Handler) (*Server, error) {
	certificate, err := tls.LoadX509KeyPair(env.CertFile(), env.KeyFile())
	if err != nil {
		// Couldn't load credentials; generate self-signed certificates.
		log.Println("Generating self-signed certificates.")
		if err := cert.GenerateSelfSignedCert(cert.Options{
			KeyFile:  env.KeyFile(),
			CertFile: env.CertFile(),
			Hosts:    env.Domain(),
		}); err != nil {
			return nil, fmt.Errorf("failed to generate self-signed certificates: %w", err)
		}
		log.Printf("Certificates successfully generated at: %s", env.CertPath())
		log.Print("Adding certificate to local keychain (requires sudo access)")
		if err := cert.AddTrustedCert(env.CertFile()); err != nil {
			return nil, fmt.Errorf("failed to install self-signed certificate: %w", err)
		}
	} else {
		log.Println("Existing credentials successfully loaded.")
	}

	httpServer := &http.Server{
		Handler:           handler,
		Addr:              addr,
		ReadHeaderTimeout: 3 * time.Second, // to prevent Slowloris (CWE-400)
		WriteTimeout:      2 * time.Minute,
		ReadTimeout:       2 * time.Minute,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{certificate},
			MinVersion:   tls.VersionTLS13,
			MaxVersion:   tls.VersionTLS13,
		},
	}

	return &Server{
		httpServer:    httpServer,
		metricsServer: metricsServer(),
		keyFile:       env.KeyFile(),
		certFile:      env.CertFile(),
	}, nil
}

func WatchHandler(ctx context.Context, handler http.Handler) http.Handler {
	watcher, err := reload.NewWatcher(ctx, filepath.Join(env.PublicDir(), "dist"))
	if err != nil {
		log.Printf("Failed to create watcher: %v", err)
		return handler
	}
	mux := http.NewServeMux()
	mux.Handle("/", handler)
	mux.HandleFunc("/watch", watcher.Handler)
	return mux
}

func metricsServer() *http.Server {
	return &http.Server{
		Handler:           metrics.Handler(),
		Addr:              metricsServerAddr,
		ReadHeaderTimeout: 3 * time.Second, // to prevent Slowloris (CWE-400)
	}
}

// Serve starts the underlying http server and redirect server, if any.
// This is a blocking call and must be called last.
func (srv *Server) Serve() error {
	if srv.redirectServer != nil {
		// Redirect all HTTP traffic to HTTPS.
		go func() {
			if err := srv.redirectServer.ListenAndServe(); err != nil {
				if !errors.Is(err, http.ErrServerClosed) {
					log.Printf("Redirect server exited with unexpected error: %v", err)
				}
			}
		}()
	}
	if srv.metricsServer != nil {
		// Start HTTP server for Prometheus metrics collection.
		go func() {
			if err := srv.metricsServer.ListenAndServe(); err != nil {
				if !errors.Is(err, http.ErrServerClosed) {
					log.Printf("Metrics server exited with unexpected error: %v", err)
				}
			}
		}()
	}
	// Start the HTTPS server.
	// For production, the certFile and keyFile are empty and managed by autocert.
	if err := srv.httpServer.ListenAndServeTLS(srv.certFile, srv.keyFile); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server exited with unexpected error: %w", err)
		}
	}
	// Exit with nil means graceful shutdown
	return nil
}

// Shutdown gracefully shuts down the server.
func (srv *Server) Shutdown(ctx context.Context) error {
	var redirectShutdownErr, metricsShutdownErr error
	if srv.redirectServer != nil {
		redirectShutdownErr = srv.redirectServer.Shutdown(ctx)
	}
	if srv.metricsServer != nil {
		metricsShutdownErr = srv.metricsServer.Shutdown(ctx)
	}
	srvShutdownErr := srv.httpServer.Shutdown(ctx)
	return errors.Join(redirectShutdownErr, metricsShutdownErr, srvShutdownErr)
}
