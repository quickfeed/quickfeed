//go:build windows

package cert

import (
	"fmt"
	"log"

	"github.com/quickfeed/quickfeed/kit/sh"
)

// AddTrustedCert adds the CA certificate to the system trust store.
func AddTrustedCert(caFile string) error {
	out, err := sh.OutputA("certutil", "-addstore", "-f", "ROOT", caFile)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return fmt.Errorf("adding CA certificate to ROOT store: %w", err)
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
	// certutil -verifystore exits non-zero when the store holds no certificate
	// with this thumbprint, which is not an error here.
	if _, err := sh.OutputA("certutil", "-verifystore", "ROOT", thumbprint); err != nil {
		log.Print("No QuickFeed CA certificate found in the ROOT store")
		return nil
	}
	out, err := sh.OutputA("certutil", "-delstore", "ROOT", thumbprint)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return fmt.Errorf("removing CA certificate from ROOT store: %w", err)
	}
	return nil
}
