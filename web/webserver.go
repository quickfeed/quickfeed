package web

import (
	"log"
	"net/http"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/hooks"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// timeouts for http server
var (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 5 * time.Minute
)

type GrpcMultiplexer struct {
	*grpcweb.WrappedGrpcServer
}

// ServerWithCredentials starts a new gRPC server with credentials
// generated from TLS certificates.
func ServerWithCredentials(logger *zap.Logger, certFile, certKey string) (*grpc.Server, error) {
	// Generate TLS credentials from certificates

	cred, err := credentials.NewServerTLSFromFile(certFile, certKey)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer(
		grpc.Creds(cred),
		grpc.ChainUnaryInterceptor(
			auth.UserVerifier(),
			qf.Interceptor(logger),
		),
	)
	return s, nil
}

// MuxHandler routes HTTP and gRPC requests.
func (m *GrpcMultiplexer) MuxHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.IsGrpcWebRequest(r) {
			log.Printf("MUX: got GRPC request: %v", r)
			m.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RegisterRouter registers http endpoints for authentication API and scm provider webhooks.
func (s *QuickFeedService) RegisterRouter(authConfig *auth.AuthConfig, scms *auth.Scms, mux GrpcMultiplexer, public, secret string) *http.ServeMux {
	// Serve static files.
	router := http.NewServeMux()
	assets := http.FileServer(http.Dir(public + "/assets"))
	dist := http.FileServer(http.Dir(public + "/dist"))

	router.Handle("/", mux.MuxHandler(http.StripPrefix("/", assets)))
	router.Handle("/assets/", mux.MuxHandler(http.StripPrefix("/assets/", assets)))
	router.Handle("/static/", mux.MuxHandler(http.StripPrefix("/static/", dist)))

	// Register auth endpoints.
	router.HandleFunc("/auth/", auth.OAuth2Login(s.logger, s.db, authConfig, secret))
	router.HandleFunc("/auth/callback/", auth.OAuth2Callback(s.logger, s.db, authConfig, scms, secret))
	router.HandleFunc("/logout", auth.OAuth2Logout(s.logger))

	// Register hooks.
	ghHook := hooks.NewGitHubWebHook(s.logger, s.db, s.runner, s.bh.Secret)
	router.HandleFunc("/hook/", ghHook.Handle())

	return router
}
