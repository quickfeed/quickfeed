//go:build darwin

package cert

import (
	"fmt"
	"log"

	"github.com/quickfeed/quickfeed/kit/sh"
)

const keychain = "/Library/Keychains/System.keychain"

// AddTrustedCert adds the CA certificate to the system trust store.
func AddTrustedCert(caFile string) error {
	// trustRoot (not trustAsRoot) is required for a self-signed root certificate;
	// trustAsRoot is for non-root certificates and fails with errSecParam here.
	out, err := sh.OutputA("sudo", "security", "add-trusted-cert", "-d", "-r", "trustRoot", "-k", keychain, caFile)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return fmt.Errorf("adding CA certificate to trust store: %w", err)
	}
	return nil
}

// RemoveTrustedCert removes the CA certificate from the system trust store.
func RemoveTrustedCert(caFile string) error {
	// remove-trusted-cert takes the certificate itself, not a keychain reference;
	// -d selects the admin domain that add-trusted-cert -d wrote to.
	out, err := sh.OutputA("sudo", "security", "remove-trusted-cert", "-d", caFile)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return fmt.Errorf("removing CA certificate from trust store: %w", err)
	}
	return nil
}
