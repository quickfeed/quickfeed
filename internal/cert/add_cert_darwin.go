//go:build darwin

package cert

import (
	"fmt"
	"log"
	"os"

	"github.com/quickfeed/quickfeed/kit/sh"
)

// AddTrustedCert adds the CA certificate to the user's keychain.
// The certFile is expected to be a fullchain containing both server cert and CA cert.
// This function extracts the CA certificate (the last one) and adds it to the keychain.
func AddTrustedCert(certFile string) error {
	tmpFile, err := caCertTempFile(certFile)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)

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

// RemoveTrustedCert removes QuickFeed's CA certificate from the user's keychain.
// The certFile is expected to be the same fullchain passed to AddTrustedCert.
func RemoveTrustedCert(certFile string) error {
	tmpFile, err := caCertTempFile(certFile)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	// remove-trusted-cert takes the certificate itself, not a keychain reference;
	// -d selects the admin domain that add-trusted-cert -d wrote to.
	out, err := sh.OutputA("sudo", "security", "remove-trusted-cert", "-d", tmpFile)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return fmt.Errorf("failed to remove certificate from keychain: %w", err)
	}
	return nil
}
