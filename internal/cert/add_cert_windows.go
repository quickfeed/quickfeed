//go:build windows

package cert

import (
	"crypto/sha1"
	"encoding/pem"
	"fmt"
	"log"
	"os"

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
func RemoveTrustedCert(caFile string) error {
	thumbprint, err := caCertThumbprint(caFile)
	if err != nil {
		return err
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

// caCertThumbprint returns the SHA-1 thumbprint of the certificate in caFile.
// certutil identifies a certificate in the store by this thumbprint; SHA-1 is
// used here only as the store's identifier, not as a security primitive.
func caCertThumbprint(caFile string) (string, error) {
	pemBytes, err := os.ReadFile(caFile)
	if err != nil {
		return "", fmt.Errorf("reading CA certificate: %w", err)
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != "CERTIFICATE" {
		return "", fmt.Errorf("no certificate found in %s", caFile)
	}
	return fmt.Sprintf("%x", sha1.Sum(block.Bytes)), nil
}
