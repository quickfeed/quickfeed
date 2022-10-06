package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/quickfeed/quickfeed/internal/cert"
	"github.com/quickfeed/quickfeed/internal/env"
	"google.golang.org/grpc/credentials"
)

// CertificateConfig holds certificate information
type CertificateConfig struct {
	CertFile string // The certificate file name.
	KeyFile  string // The key file name.
}

// EnvoyConfig defines the required configurations used by the template when generating
// the envoy config file.
type EnvoyConfig struct {
	Domain     string             // The domain name where envoy will be served.
	ServerHost string             // The container name, ip or domain name where quickfeed will run.
	GRPCPort   string             // The grpc port listened by quickfeed.
	HTTPPort   string             // The http port listened by quickfeed.
	TLSEnabled bool               // Whether TLS should be configured.
	CertConfig *CertificateConfig // The certificate to be used when using TLS.
}

//go:embed envoy.tmpl
var envoyTmpl embed.FS

// createEnvoyConfigFile creates the envoy.yaml config file.
func createEnvoyConfigFile(config *EnvoyConfig) error {
	envoyConfigFile := os.Getenv("ENVOY_CONFIG")
	if err := os.MkdirAll(path.Dir(envoyConfigFile), 0o700); err != nil {
		return err
	}
	f, err := os.Create(envoyConfigFile)
	if err != nil {
		return err
	}
	tmpl, err := template.ParseFS(envoyTmpl, "envoy.tmpl")
	if err != nil {
		return err
	}
	if err = tmpl.ExecuteTemplate(f, "envoy", config); err != nil {
		return err
	}
	log.Println("Envoy config file created at:", envoyConfigFile)
	return nil
}

func main() {
	// Load environment variables from $QUICKFEED/.env.
	// It will not override a variable that already exists in the environment.
	if err := env.Load(""); err != nil {
		log.Fatal(err)
	}

	// Default certificate directory; used when generating certificates and -cert-dir not specified.
	defaultCertDir := filepath.Join(os.Getenv("QUICKFEED"), "internal", "config", "certs")
	var (
		withTLS = flag.Bool("tls", false, "enable TLS configuration")
		// Defaults are from the environment.
		certFile = flag.String("cert", env.CertFile(), "certificate file path")
		keyFile  = flag.String("key", env.KeyFile(), "private key file path")
		certDir  = flag.String("cert-dir", defaultCertDir, "certificate directory path")
	)
	flag.Parse()

	serverHost := os.Getenv("SERVER_HOST")
	domain := env.Domain()

	config := &EnvoyConfig{
		Domain:     domain,
		ServerHost: serverHost,
		GRPCPort:   os.Getenv("GRPC_PORT"),
		HTTPPort:   os.Getenv("HTTP_PORT"),
		TLSEnabled: *withTLS,
		CertConfig: &CertificateConfig{
			CertFile: *certFile,
			KeyFile:  *keyFile,
		},
	}

	if *withTLS {
		// Check if the certificate files exist.
		if _, err := credentials.NewServerTLSFromFile(*certFile, *keyFile); err != nil {
			// Couldn't load credentials; generate self-signed certificates.
			log.Println("Generating self-signed certificates.")
			if err := cert.GenerateSelfSignedCert(cert.Options{
				CertFile: *certFile,
				KeyFile:  *keyFile,
				Hosts:    fmt.Sprintf("%s,%s", serverHost, domain),
			}); err != nil {
				log.Fatal(err)
			}
			log.Printf("Certificates successfully generated at: %s", *certDir)
		} else {
			log.Println("Existing credentials successfully loaded.")
		}
	}

	if err := createEnvoyConfigFile(config); err != nil {
		log.Fatal(err)
	}
}
