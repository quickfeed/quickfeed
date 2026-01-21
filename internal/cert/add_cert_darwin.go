//go:build darwin

package cert

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/quickfeed/quickfeed/kit/sh"
)

// AddTrustedCert adds the CA certificate to the user's keychain.
// The certFile is expected to be a fullchain containing both server cert and CA cert.
// This function extracts the CA certificate (the last one) and adds it to the keychain.
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

	// Add to keychain
	keychain := fmt.Sprintf("/Users/%s/Library/Keychains/login.keychain", os.Getenv("USER"))
	out, err := sh.OutputA("sudo", "security", "add-trusted-cert", "-d", "-r", "trustAsRoot", "-k", keychain, tmpFile)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return fmt.Errorf("failed to add certificate to keychain: %w", err)
	}
	return nil
}
