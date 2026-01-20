//go:build linux

package cert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/quickfeed/quickfeed/kit/sh"
)

// AddTrustedCert adds the CA certificate to the system trust store.
// The certFile is expected to be a fullchain containing both server cert and CA cert.
// This function extracts the CA certificate (the last one) and adds it to the trust store.
func AddTrustedCert(certFile string) error {
	const certPath = "/usr/local/share/ca-certificates/"
	const caCertFile = "quickfeed-ca.crt"

	// Read the fullchain certificate file
	fullchainBytes, err := os.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Parse all certificates from the fullchain
	var certs []*x509.Certificate
	for block, rest := pem.Decode(fullchainBytes); block != nil; block, rest = pem.Decode(rest) {
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return fmt.Errorf("failed to parse certificate: %w", err)
			}
			certs = append(certs, cert)
		}
	}

	if len(certs) == 0 {
		return fmt.Errorf("no certificates found in %s", certFile)
	}

	// The CA certificate is the last one in the chain
	caCert := certs[len(certs)-1]

	// Encode the CA certificate to PEM format
	caCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCert.Raw,
	})

	// Write the CA certificate to a temporary file first (in user's home directory)
	tmpFile := filepath.Join(os.TempDir(), caCertFile)
	if err := os.WriteFile(tmpFile, caCertPEM, 0o644); err != nil {
		return fmt.Errorf("failed to write temporary CA certificate: %w", err)
	}
	defer os.Remove(tmpFile)

	// Copy to system certificate directory with sudo
	destFile := filepath.Join(certPath, caCertFile)
	out, err := sh.OutputA("sudo", "cp", tmpFile, destFile)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return fmt.Errorf("failed to copy CA certificate to system trust store: %w", err)
	}

	// Update the certificate trust store
	out, err = sh.Output("sudo update-ca-certificates")
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return err
	}
	return nil
}
