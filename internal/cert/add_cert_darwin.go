//go:build darwin

package cert

import (
	"fmt"
	"log"
	"strings"

	"github.com/quickfeed/quickfeed/kit/sh"
)

const keychain = "/Library/Keychains/System.keychain"

// AddTrustedCert adds the CA certificate to the system trust store.
func AddTrustedCert(caFile string) error {
	// trustRoot (not trustAsRoot) is required for a self-signed root certificate;
	// trustAsRoot is only valid for non-root certificates and fails with errSecParam.
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
// It is a no-op if the certificate is not installed.
func RemoveTrustedCert(caFile string) error {
	thumbprint, err := caCertThumbprint(caFile)
	if err != nil {
		return err
	}
	if !isTrustedCertInstalled(thumbprint) {
		log.Printf("No QuickFeed CA certificate found in %s", keychain)
		return nil
	}
	// Removal takes two steps. remove-trusted-cert only drops the trust settings;
	// it leaves the certificate itself in the keychain, so delete-certificate must
	// remove the keychain item as well. Without that, every regenerated CA leaves
	// another untrusted QuickFeed certificate behind.
	//
	// remove-trusted-cert takes the certificate itself, not a keychain reference;
	// -d selects the admin domain that add-trusted-cert -d wrote to. It fails if the
	// certificate has no trust settings for that domain. This is not an error since
	// the certificate is already untrusted, and deleting it is all that remains.
	if _, err := sh.OutputA("sudo", "security", "remove-trusted-cert", "-d", caFile); err != nil {
		log.Print("The QuickFeed CA certificate has no trust settings to remove; deleting it from the keychain")
	}
	// delete-certificate -t would only remove trust settings from the user domain,
	// so it is no substitute for remove-trusted-cert -d above.
	out, err := sh.OutputA("sudo", "security", "delete-certificate", "-Z", thumbprint, keychain)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return fmt.Errorf("removing CA certificate from %s: %w", keychain, err)
	}
	return nil
}

// isTrustedCertInstalled reports whether the certificate with the given SHA-1
// thumbprint is present in the system keychain. Given -Z, security(1) prints a
// "SHA-1 hash:" line for each certificate, so the certificate is matched by
// thumbprint rather than by subject name, which is not unique.
func isTrustedCertInstalled(thumbprint string) bool {
	// find-certificate exits non-zero when the keychain holds no certificates,
	// which is not an error here; an empty listing simply matches nothing.
	out, _ := sh.OutputA("security", "find-certificate", "-a", "-Z", keychain)
	return strings.Contains(strings.ToUpper(out), thumbprint)
}
