//go:build windows

package cert

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"

	"github.com/quickfeed/quickfeed/kit/sh"
)

// AddTrustedCert adds the CA certificate to the Windows certificate store.
// The certFile is expected to be a fullchain containing both server cert and CA cert.
// This function extracts the CA certificate (the last one) and adds it to the ROOT store.
func AddTrustedCert(certFile string) error {
	tmpFile, err := caCertTempFile(certFile)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	out, err := sh.OutputA("certutil", "-addstore", "-f", "ROOT", tmpFile)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return fmt.Errorf("failed to add certificate to ROOT store: %w", err)
	}
	return nil
}

// RemoveTrustedCert removes QuickFeed's CA certificate from the Windows ROOT store.
// The certFile is expected to be the same fullchain passed to AddTrustedCert.
func RemoveTrustedCert(certFile string) error {
	caCert, err := caCertificate(certFile)
	if err != nil {
		return err
	}
	// certutil identifies a certificate in the store by its SHA-1 thumbprint.
	// SHA-1 is used here only as the store's identifier, not as a security primitive.
	thumbprint := fmt.Sprintf("%x", sha1.Sum(caCert.Raw))

	out, err := sh.OutputA("certutil", "-delstore", "-f", "ROOT", thumbprint)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return fmt.Errorf("failed to remove certificate from ROOT store: %w", err)
	}
	return nil
}
