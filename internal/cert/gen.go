package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var certsDir = filepath.Join("internal", "cert", "certs")

// Options for generating a self-signed certificate.
type Options struct {
	Hosts     string        // comma-separated hostnames and IPs to generate a certificate for.
	ValidFrom time.Time     // creation date (default duration is 1 year)
	ValidFor  time.Duration // for how long the certificate is valid.
	KeyType   string        // default ECDSA curve P256
}

// GenerateSelfSignedCert generates a self-signed X.509 certificate for testing purposes.
// It supports ECDSA curve P256 or RSA 2048 bits to generate the key.
// based on: https://golang.org/src/crypto/tls/generate_cert.go
func GenerateSelfSignedCert(opts Options) (err error) {
	if opts.Hosts == "" {
		return errors.New("at least one hostname must be specified")
	}
	if err = os.MkdirAll(certsDir, 0o700); err != nil {
		return err
	}

	var caKey, serverKey any
	switch opts.KeyType {
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
	if opts.ValidFrom.IsZero() {
		notBefore = time.Now()
	} else {
		notBefore = opts.ValidFrom
	}

	if opts.ValidFor == 0 {
		notAfter = notBefore.Add(365 * 24 * time.Hour)
	} else {
		notAfter = notBefore.Add(opts.ValidFor)
	}

	if notBefore.After(notAfter) {
		return errors.New("wrong certificate validity")
	}

	caTemplate, err := newCertificateTemplate(caKey, opts.Hosts, notBefore, notAfter, true)
	if err != nil {
		return err
	}
	caCert, caCertBytes, err := makeCertificate(caTemplate, caTemplate, publicKey(caKey), caKey)
	if err != nil {
		return err
	}

	serverTemplate, err := newCertificateTemplate(serverKey, opts.Hosts, notBefore, notAfter, false)
	if err != nil {
		return err
	}
	_, serverCertBytes, err := makeCertificate(serverTemplate, caCert, publicKey(serverKey), caKey)
	if err != nil {
		return err
	}

	serverKeyBytes, err := x509.MarshalPKCS8PrivateKey(serverKey)
	if err != nil {
		return fmt.Errorf("unable to marshal server private key: %w", err)
	}

	// save server private key
	if err = savePEM(certsDir, "privkey.pem", []*pem.Block{
		{Type: "PRIVATE KEY", Bytes: serverKeyBytes},
	}); err != nil {
		return err
	}

	// save fullchain (server certificate and CA certificate)
	if err = savePEM(certsDir, "fullchain.pem", []*pem.Block{
		{Type: "CERTIFICATE", Bytes: serverCertBytes},
		{Type: "CERTIFICATE", Bytes: caCertBytes},
	}); err != nil {
		return err
	}
	return nil
}

func newCertificateTemplate(privKey any, hostList string, notBefore, notAfter time.Time, isCA bool) (*x509.Certificate, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("serial number generation failed: %w", err)
	}

	var template *x509.Certificate
	if isCA {
		caSubject := &pkix.Name{
			Country:      []string{"NO"},
			Organization: []string{"QuickFeed Corp."},
			CommonName:   "QuickFeed CA",
		}
		template = &x509.Certificate{
			SerialNumber:          serialNumber,
			Subject:               *caSubject,
			NotBefore:             notBefore,
			NotAfter:              notAfter,
			KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
			IsCA:                  true,
			MaxPathLenZero:        true,
		}
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

	for _, host := range strings.Split(hostList, ",") {
		if ip := net.ParseIP(host); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, host)
		}
	}
	return template, nil
}

func makeCertificate(template, parent *x509.Certificate, publicKey any, privateKey any) (*x509.Certificate, []byte, error) {
	derCertBytes, err := x509.CreateCertificate(rand.Reader, template, parent, publicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}
	cert, err := x509.ParseCertificate(derCertBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}
	return cert, derCertBytes, nil
}

const defaultFileFlags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC

func savePEM(certPath, filename string, block []*pem.Block) error {
	out, err := os.OpenFile(filepath.Join(certPath, filename), defaultFileFlags, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %w", filename, err)
	}

	for _, b := range block {
		if err := pem.Encode(out, b); err != nil {
			return fmt.Errorf("failed to write data to %s: %w", filename, err)
		}
	}

	if err := out.Close(); err != nil {
		return fmt.Errorf("error closing %s: %w", filename, err)
	}
	return nil
}

func publicKey(priv any) any {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}
