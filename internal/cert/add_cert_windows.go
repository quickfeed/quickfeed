//go:build windows

package cert

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/quickfeed/quickfeed/kit/sh"
)

// AddTrustedCert adds the CA certificate to the Windows certificate store.
// The certFile is expected to be a fullchain containing both server cert and CA cert.
// This function extracts the CA certificate (the last one) and adds it to the ROOT store.
func AddTrustedCert(certFile string) error {
	// Extract only the CA certificate from the fullchain
	caCertPEM, err := extractCACert(certFile)
	if err != nil {
		return err
	}

	// Write the CA certificate to a temporary file
	tmpFile := filepath.Join(os.TempDir(), "quickfeed-ca.crt")
	if err := os.WriteFile(tmpFile, caCertPEM, 0o644); err != nil {
		return fmt.Errorf("failed to write temporary CA certificate: %w", err)
	}
	defer os.Remove(tmpFile)

	// Add to Windows ROOT certificate store
	out, err := sh.OutputA("certutil", "-addstore", "-f", "ROOT", tmpFile)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return fmt.Errorf("failed to add certificate to ROOT store: %w", err)
	}
	return nil
}
