package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"embed"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type CertificateConfig struct {
	CertFile string
	KeyFile  string
}

type EnvoyConfig struct {
	Domain     string
	GRPCPort   string
	HTTPPort   string
	TLSEnabled bool
	CertConfig *CertificateConfig
}

func newEnvoyConfig(domain, GRPCPort, HTTPPort string, withTLS bool, certConfig *CertificateConfig) (*EnvoyConfig, error) {
	config := &EnvoyConfig{
		Domain:     strings.Trim(domain, "\""),
		GRPCPort:   GRPCPort,
		HTTPPort:   HTTPPort,
		TLSEnabled: withTLS,
	}

	if withTLS {
		if certConfig != nil {
			config.CertConfig = certConfig
			return config, nil
		}
		err := generateSelfSignedCert(path.Join(path.Dir(pwd), "certs"), "cert", certOptions{
			org:   domain,
			hosts: "localhost",
			isCA:  true,
		})
		if err != nil {
			return nil, err
		}
		config.CertConfig = &CertificateConfig{
			CertFile: "cert.pem",
			KeyFile:  "cert.key",
		}
	}

	return config, nil
}

//go:embed envoy.tmpl
var envoyTmpl embed.FS

// createEnvoyConfig creates the envoy.yaml config file.
func createEnvoyConfig(envoyPath string, data *EnvoyConfig) error {
	sanitizedPath := strings.Trim(envoyPath, "\"")

	err := os.MkdirAll(path.Dir(sanitizedPath), 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(sanitizedPath)
	if err != nil {
		return err
	}

	tmpl, err := template.ParseFS(envoyTmpl, "envoy.tmpl")
	if err != nil {
		return err
	}

	if err = tmpl.ExecuteTemplate(f, "envoy", data); err != nil {
		return err
	}
	return nil
}

var (
	genConfig           bool
	withTLS             bool
	runEnvoy            bool
	envoyConfigFilePath string
	_, pwd, _, _        = runtime.Caller(0)
	codePath            = path.Join(path.Dir(pwd), "..")
	env                 = filepath.Join(codePath, ".env")
	defaultEnvoyConfig  = filepath.Join(path.Dir(pwd), "envoy.yaml")
)

// loadConfigEnv loads the  envoy config from the environment variables.
// It will not override a variable that already exists.
// Consider the .env file to set development vars or defaults.
func loadConfigEnv(withTLS bool) (*EnvoyConfig, error) {
	err := godotenv.Load(env)
	if err != nil {
		return nil, err
	}
	return newEnvoyConfig(os.Getenv("DOMAIN"), os.Getenv("GRPC_PORT"), os.Getenv("HTTP_PORT"), withTLS, nil)
}

func main() {
	flag.BoolVar(&genConfig, "genconfig", false, "generate envoy config")
	flag.BoolVar(&withTLS, "withTLS", false, "enable TLS configuration")
	flag.StringVar(&envoyConfigFilePath, "config", defaultEnvoyConfig, "filepath where the envoy configuration should be created")
	flag.BoolVar(&runEnvoy, "run", false, "run envoy container")
	flag.Parse()

	// TODO: receive config
	config, err := loadConfigEnv(withTLS)
	if err != nil {
		log.Fatal(err)
	}

	switch {
	case genConfig:
		err := createEnvoyConfig(envoyConfigFilePath, config)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("envoy config file created at", envoyConfigFilePath)
	case runEnvoy:
		// TODO: refactor startEnvoy or delete it.
	default:
		fmt.Println("unknown command.")
	}
}

func newCertificateTemplate(privKey interface{}, hostList, organization string, notBefore, notAfter time.Time, isCA bool) (*x509.Certificate, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("serial number generation failed: %v", err)
	}

	keyUsage := x509.KeyUsageDigitalSignature
	// https://go-review.googlesource.com/c/go/+/214337/
	// If is RSA set KeyEncipherment KeyUsage bits.
	if _, isRSA := privKey.(*rsa.PrivateKey); isRSA {
		keyUsage |= x509.KeyUsageKeyEncipherment
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{organization},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(hostList, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	return template, nil
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

type certOptions struct {
	org       string        // organization name.
	hosts     string        // comma-separated hostnames and IPs to generate a certificate for.
	validFrom time.Time     // creation date (default duration is 1 year: 365*24*time.Hour)
	validFor  time.Duration // for how long the certificate is valid.
	isCA      bool          // whether the certificate should be its own Certificate Authority
	keyType   string
}

// generateSelfSignedCert generates a self-signed X.509 certificate for testing purposes.
// It supports ECDSA curve P256 or RSA 2048 bits to generate the key.
// based on: https://golang.org/src/crypto/tls/generate_cert.go
func generateSelfSignedCert(certsDir string, certName string, opts certOptions) (err error) {
	if len(opts.hosts) == 0 {
		return errors.New("at least one hostname must be specified")
	}

	err = os.MkdirAll(certsDir, 0755)
	if err != nil {
		return err
	}

	var privKey interface{}
	switch opts.keyType {
	case "rsa":
		privKey, err = rsa.GenerateKey(rand.Reader, 2048)
	default:
		privKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	}
	if err != nil {
		return fmt.Errorf("key generation failed: %v", err)
	}

	var notBefore, notAfter time.Time
	if opts.validFrom.IsZero() {
		notBefore = time.Now()
	} else {
		notBefore = opts.validFrom
	}

	if opts.validFor == 0 {
		notAfter = notBefore.Add(365 * 24 * time.Hour)
	} else {
		notAfter = notBefore.Add(opts.validFor)
	}

	if notBefore.After(notAfter) {
		return errors.New("wrong certificate validity")
	}

	template, err := newCertificateTemplate(privKey, opts.hosts, opts.org, notBefore, notAfter, opts.isCA)
	if err != nil {
		return err
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, publicKey(privKey), privKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	certFile := fmt.Sprintf("%s.pem", certName)
	certOut, err := os.Create(path.Join(certsDir, certFile))
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %v", certFile, err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write data to %s: %v", certFile, err)
	}
	if err := certOut.Close(); err != nil {
		return fmt.Errorf("error closing %s: %v", certFile, err)
	}

	certKey := fmt.Sprintf("%s.key", certName)
	keyOut, err := os.OpenFile(path.Join(certsDir, certKey), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %v", certKey, err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("failed to write data to %s: %v", certKey, err)
	}
	if err := keyOut.Close(); err != nil {
		return fmt.Errorf("error closing %s: %v", certKey, err)
	}

	log.Printf("certificate: %s and key: %s successfully generated at: %s", certFile, certKey, certsDir)
	return nil
}

// StartEnvoy creates a Docker API client. If an envoy container is not running,
// it will be started from an image. If no image exists, it will pull an Envoy
// image from docker and build it with options from envoy.yaml.
// TODO(meling) since this runs in a separate goroutine it is actually bad practice
// to Panic or Fatal on error, since other goroutines may not exit cleanly.
// Instead it would be better to return an error and run synchronously.
func StartEnvoy(l *zap.Logger) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		l.Fatal("failed to start docker client", zap.Error(err))
	}

	// removes all stopped containers
	_, err = cli.ContainersPrune(ctx, filters.Args{})
	if err != nil {
		//
		l.Info("failed to prune unused containers", zap.Error(err))
	}

	// check for existing Envoy containers
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		l.Fatal("failed to retrieve docker container list", zap.Error(err))
	}
	for _, container := range containers {
		if container.Names[0] == "/envoy" {
			l.Info("Envoy container is already running")
			return
		}
	}

	if !hasEnvoyImage(ctx, l, cli) {
		// if there is no active Envoy image, we build it
		l.Info("building Envoy image...")
		// TODO(meling) use docker api to build image: "docker build -t ag_envoy -f ./envoy/Dockerfile ."
		out, err := exec.Command("/bin/sh", "./envoy/envoy.sh", "build").Output()
		if err != nil {
			l.Fatal("failed to execute bash script", zap.Error(err))
		}
		l.Debug("envoy.sh build", zap.String("output", string(out)))
	}
	l.Info("starting Envoy container...")
	// TODO(meling) use docker api to run image: "docker run --name=envoy -p 8080:8080 --net=host ag_envoy"
	out, err := exec.Command("/bin/sh", "./envoy/envoy.sh").Output()
	if err != nil {
		l.Fatal("failed to execute bash script", zap.Error(err))
	}
	l.Debug("envoy.sh", zap.String("output", string(out)))
}

// hasEnvoyImage returns true if the docker client has the latest Envoy image.
func hasEnvoyImage(ctx context.Context, l *zap.Logger, cli *client.Client) bool {
	l.Debug("no running Envoy container found")
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		l.Fatal("failed to retrieve docker image list", zap.Error(err))
	}
	l.Debug("checking for Autograder's Envoy image")
	for _, img := range images {
		l.Debug("found image", zap.Strings("repo", img.RepoTags))
		if len(img.RepoTags) > 0 && img.RepoTags[0] == "ag_envoy:latest" {
			l.Debug("found Envoy image")
			return true
		}
	}
	return false
}
