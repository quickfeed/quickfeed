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
	"io/fs"
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
		err := generateSelfSignedCert(path.Join(path.Dir(pwd), "certs"), certOptions{
			hosts: fmt.Sprintf("%s,%s", "127.0.0.1", domain),
		})
		if err != nil {
			return nil, err
		}
		config.CertConfig = &CertificateConfig{
			CertFile: "fullchain.pem",
			KeyFile:  "key.pem",
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

func CARootTemplate(serialNumber *big.Int, subject *pkix.Name, notBefore, notAfter time.Time) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               *subject,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
	}
}

func newCertificateTemplate(privKey interface{}, hostList string, notBefore, notAfter time.Time, isCA bool) (*x509.Certificate, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("serial number generation failed: %v", err)
	}

	var template *x509.Certificate
	if isCA {
		caSubject := &pkix.Name{
			Country:      []string{"NO"},
			Organization: []string{"Example Co."},
			CommonName:   "Example CA",
		}
		template = CARootTemplate(serialNumber, caSubject, notBefore, notAfter)
	} else {
		keyUsage := x509.KeyUsageDigitalSignature
		// https://go-review.googlesource.com/c/go/+/214337/
		// If is RSA set KeyEncipherment KeyUsage bits.
		if _, isRSA := privKey.(*rsa.PrivateKey); isRSA {
			keyUsage |= x509.KeyUsageKeyEncipherment
		}

		template = &x509.Certificate{
			SerialNumber:          serialNumber,
			NotBefore:             notBefore,
			NotAfter:              notAfter,
			KeyUsage:              keyUsage,
			IsCA:                  false,
			BasicConstraintsValid: true,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		}
	}

	hosts := strings.Split(hostList, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	return template, nil
}

func makeCertificate(template, parent *x509.Certificate, publicKey interface{}, privateKey interface{}) (*x509.Certificate, []byte, error) {
	derCertBytes, err := x509.CreateCertificate(rand.Reader, template, parent, publicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(derCertBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %v", err)
	}
	return cert, derCertBytes, nil
}

func savePEM(certPath, filename string, block []*pem.Block, flag int, perm fs.FileMode) error {
	out, err := os.OpenFile(path.Join(certPath, filename), flag, perm)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %v", filename, err)
	}

	for _, b := range block {
		if err := pem.Encode(out, b); err != nil {
			return fmt.Errorf("failed to write data to %s: %v", filename, err)
		}
	}

	if err := out.Close(); err != nil {
		return fmt.Errorf("error closing %s: %v", filename, err)
	}
	return nil
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
	hosts     string        // comma-separated hostnames and IPs to generate a certificate for.
	validFrom time.Time     // creation date (default duration is 1 year: 365*24*time.Hour)
	validFor  time.Duration // for how long the certificate is valid.
	keyType   string
}

const defaultFileFlags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC

// generateSelfSignedCert generates a self-signed X.509 certificate for testing purposes.
// It supports ECDSA curve P256 or RSA 2048 bits to generate the key.
// based on: https://golang.org/src/crypto/tls/generate_cert.go
func generateSelfSignedCert(certsDir string, opts certOptions) (err error) {
	if len(opts.hosts) == 0 {
		return errors.New("at least one hostname must be specified")
	}

	err = os.MkdirAll(certsDir, 0755)
	if err != nil {
		return err
	}

	var caKey, serverKey interface{}
	switch opts.keyType {
	case "rsa":
		caKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}
		serverKey, err = rsa.GenerateKey(rand.Reader, 2048)
	default:
		caKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return err
		}
		serverKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	}
	if err != nil {
		return err
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

	caTemplate, err := newCertificateTemplate(caKey, opts.hosts, notBefore, notAfter, true)
	if err != nil {
		return err
	}

	caCert, caCertBytes, err := makeCertificate(caTemplate, caTemplate, publicKey(caKey), caKey)
	if err != nil {
		return err
	}

	serverTemplate, err := newCertificateTemplate(serverKey, opts.hosts, notBefore, notAfter, false)
	if err != nil {
		return err
	}

	_, serverCertBytes, err := makeCertificate(serverTemplate, caCert, publicKey(serverKey), caKey)
	if err != nil {
		return err
	}

	// save ca certificate
	err = savePEM(certsDir, "cacert.pem", []*pem.Block{{Type: "CERTIFICATE", Bytes: caCertBytes}}, defaultFileFlags, 0600)
	if err != nil {
		return err
	}

	caKeyBytes, err := x509.MarshalPKCS8PrivateKey(caKey)
	if err != nil {
		return fmt.Errorf("unable to marshal ca private key: %v", err)
	}
	err = savePEM(certsDir, "cakey.pem", []*pem.Block{{Type: "PRIVATE KEY", Bytes: caKeyBytes}}, defaultFileFlags, 0600)
	if err != nil {
		return err
	}

	// save server certificate
	err = savePEM(certsDir, "cert.pem", []*pem.Block{{Type: "CERTIFICATE", Bytes: serverCertBytes}}, defaultFileFlags, 0600)
	if err != nil {
		return err
	}

	serverKeyByes, err := x509.MarshalPKCS8PrivateKey(serverKey)
	if err != nil {
		return fmt.Errorf("unable to marshal server private key: %v", err)
	}

	err = savePEM(certsDir, "key.pem", []*pem.Block{{Type: "PRIVATE KEY", Bytes: serverKeyByes}}, defaultFileFlags, 0600)
	if err != nil {
		return err
	}

	// save fullchain
	err = savePEM(certsDir, "fullchain.pem", []*pem.Block{
		{Type: "CERTIFICATE", Bytes: serverCertBytes},
		{Type: "CERTIFICATE", Bytes: caCertBytes},
	}, defaultFileFlags, 0600)
	if err != nil {
		return err
	}

	log.Printf("certificates successfully generated at: %s", certsDir)
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
