package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/quickfeed/quickfeed/internal/cert"
	"github.com/quickfeed/quickfeed/internal/env"
)

var (
	_, pwd, _, _    = runtime.Caller(0)
	codePath        = path.Join(path.Dir(pwd), "../..")
	dotEnv          = filepath.Join(codePath, ".env")
	envoyDockerRoot = filepath.Join(codePath, "ci/docker/envoy")
	certsDir        = filepath.Join("internal", "cert", "certs")
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
	envoyConfigFile := path.Join(envoyDockerRoot, fmt.Sprintf("envoy-%s.yaml", config.Domain))

	if err := os.MkdirAll(path.Dir(envoyConfigFile), 0o600); err != nil {
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
	log.Println("envoy config file created at:", envoyConfigFile)
	return nil
}

// TODO: improve parameter handling when generating certificates (keyType, etc).
// TODO: save certs at: /etc/ssl/certs and private keys at: /etc/ssl/private by default.
func main() {
	// Load environment variables from the .env file.
	// It will not override a variable that already exists in the environment.
	if err := env.Load(dotEnv); err != nil {
		log.Fatal(err)
	}
	var (
		withTLS = flag.Bool("tls", false, "enable TLS configuration")
		// Defaults are from the environment.
		certFile = flag.String("cert", env.CertFile(), "certificate file path")
		keyFile  = flag.String("key", env.CertKey(), "private key file path") // TODO: rename CertKey to KeyFile
	)
	flag.Parse()

	// TODO replace with env.Domain() and env.ServerHost()
	serverHost := os.Getenv("SERVER_HOST")
	domain := strings.Trim(os.Getenv("DOMAIN"), `"`)

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
	// TODO check if credential files can be loaded
	// cred, err := credentials.NewServerTLSFromFile(*certFile, *certKey)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	if *withTLS {
		if err := cert.GenerateSelfSignedCert(cert.Options{
			Path:  certsDir,
			Hosts: fmt.Sprintf("%s,%s", serverHost, domain),
		}); err != nil {
			log.Fatal(err)
		}
		log.Printf("certificates successfully generated at: %s", certsDir)
	}

	if err := createEnvoyConfigFile(config); err != nil {
		log.Fatal(err)
	}
}
