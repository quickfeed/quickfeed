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
	withTLS         bool
	certFile        string
	keyFile         string
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

func newEnvoyConfig(domain, serverHost, GRPCPort, HTTPPort string, withTLS bool, certConfig *CertificateConfig) (*EnvoyConfig, error) {
	config := &EnvoyConfig{
		Domain:     strings.Trim(domain, "\""),
		ServerHost: serverHost,
		GRPCPort:   GRPCPort,
		HTTPPort:   HTTPPort,
		TLSEnabled: withTLS,
	}

	if withTLS {
		if certConfig.CertFile != "" && certConfig.KeyFile != "" {
			config.CertConfig = certConfig
			return config, nil
		}
		if err := cert.GenerateSelfSignedCert(cert.Options{
			Path:  certsDir,
			Hosts: fmt.Sprintf("%s,%s", config.ServerHost, domain),
		}); err != nil {
			return nil, err
		}
		log.Printf("certificates successfully generated at: %s", certsDir)

		config.CertConfig = &CertificateConfig{
			CertFile: "fullchain.pem",
			KeyFile:  "privkey.pem",
		}
	}

	return config, nil
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

// loadConfigEnv loads the  envoy config from the environment variables.
// It will not override a variable that already exists.
// Consider the .env file to set development vars or defaults.
func loadConfigEnv(withTLS bool, config *CertificateConfig) (*EnvoyConfig, error) {
	if err := env.Load(dotEnv); err != nil {
		return nil, err
	}
	return newEnvoyConfig(
		os.Getenv("DOMAIN"),
		os.Getenv("SERVER_HOST"),
		os.Getenv("GRPC_PORT"),
		os.Getenv("HTTP_PORT"),
		withTLS,
		config,
	)
}

// TODO: improve parameter handling when generating certificates (keyType, etc).
// TODO: save certs at: /etc/ssl/certs and private keys at: /etc/ssl/private by default.
func main() {
	flag.BoolVar(&withTLS, "tls", false, "enable TLS configuration")
	flag.StringVar(&certFile, "cert", "", "certificate file")
	flag.StringVar(&keyFile, "key", "", "private key file")
	flag.Parse()

	config, err := loadConfigEnv(withTLS, &CertificateConfig{
		CertFile: certFile,
		KeyFile:  keyFile,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = createEnvoyConfigFile(config)
	if err != nil {
		log.Fatal(err)
	}
}
