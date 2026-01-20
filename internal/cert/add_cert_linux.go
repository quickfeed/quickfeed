//go:build linux

package cert

import (
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

	// Extract only the CA certificate from the fullchain
	caCertPEM, err := extractCACert(certFile)
	if err != nil {
		return err
	}

	// Write the CA certificate to a temporary file first
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
