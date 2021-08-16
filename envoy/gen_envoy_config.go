package main

import (
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
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/joho/godotenv"
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
		err := generateSelfSignedCert(certOptions{
			hosts: fmt.Sprintf("%s,%s", config.ServerHost, domain),
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

// createEnvoyConfigFile creates the envoy.yaml config file.
func createEnvoyConfigFile(config *EnvoyConfig) error {
	envoyConfigFile := path.Join(path.Dir(pwd), fmt.Sprintf("envoy-%s.yaml", config.Domain))

	err := os.MkdirAll(path.Dir(envoyConfigFile), 0755)
	if err != nil {
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
	validFrom time.Time     // creation date (default duration is 1 year)
	validFor  time.Duration // for how long the certificate is valid.
	keyType   string        // default ECDSA curve P256
}

const defaultFileFlags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC

var (
	withTLS      bool
	certFile     string
	keyFile      string
	_, pwd, _, _ = runtime.Caller(0)
	codePath     = path.Join(path.Dir(pwd), "..")
	env          = filepath.Join(codePath, ".env")
	certsDir     = path.Join(path.Dir(pwd), "certs")
)

// generateSelfSignedCert generates a self-signed X.509 certificate for testing purposes.
// It supports ECDSA curve P256 or RSA 2048 bits to generate the key.
// based on: https://golang.org/src/crypto/tls/generate_cert.go
func generateSelfSignedCert(opts certOptions) (err error) {
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

// loadConfigEnv loads the  envoy config from the environment variables.
// It will not override a variable that already exists.
// Consider the .env file to set development vars or defaults.
func loadConfigEnv(withTLS bool, config *CertificateConfig) (*EnvoyConfig, error) {
	err := godotenv.Load(env)
	if err != nil {
		return nil, err
	}

	return newEnvoyConfig(os.Getenv("DOMAIN"), os.Getenv("SERVER_HOST"), os.Getenv("GRPC_PORT"), os.Getenv("HTTP_PORT"), withTLS, config)
}

// TODO: improve parameter handling when generating certificates (keyType, etc).
// TODO: save certs at: /etc/ssl/certs and private keys at: /etc/ssl/private by default.
func main() {
	flag.BoolVar(&withTLS, "withTLS", false, "enable TLS configuration")
	flag.StringVar(&certFile, "certFile", "", "certificate file")
	flag.StringVar(&keyFile, "keyFile", "", "private key file")
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
