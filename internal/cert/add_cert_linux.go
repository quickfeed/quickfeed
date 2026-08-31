//go:build linux

package cert

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/quickfeed/quickfeed/kit/sh"
)

const (
	// certPath is the directory update-ca-certificates scans for locally added CA certificates.
	certPath = "/usr/local/share/ca-certificates/"
	// caCertFile is the name QuickFeed's CA certificate is installed under. The .crt
	// extension is required; update-ca-certificates ignores files with any other extension.
	caCertFile = "quickfeed-ca.crt"
	// caCertPath is the full path to QuickFeed's CA certificate in the system trust store.
	caCertPath = certPath + caCertFile
)

// AddTrustedCert adds the CA certificate to the system trust store.
// The certFile is expected to be a fullchain containing both server cert and CA cert.
// This function extracts the CA certificate (the last one) and adds it to the trust store.
func AddTrustedCert(certFile string) error {
	// Extract only the CA certificate from the fullchain; update-ca-certificates
	// requires one certificate per file.
	tmpFile, err := caCertTempFile(certFile)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	// Install into the system certificate directory with sudo. The mode is set
	// explicitly since the temporary file is only readable by the current user.
	out, err := sh.OutputA("sudo", "install", "-m", "0644", tmpFile, caCertPath)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return fmt.Errorf("failed to copy CA certificate to system trust store: %w", err)
	}
	return updateCACertificates()
}

// RemoveTrustedCert removes QuickFeed's CA certificate from the system trust store.
// It is a no-op if the certificate is not installed.
func RemoveTrustedCert(_ string) error {
	if _, err := os.Stat(caCertPath); errors.Is(err, os.ErrNotExist) {
		log.Printf("No QuickFeed CA certificate found at %s", caCertPath)
		return nil
	}
	out, err := sh.OutputA("sudo", "rm", "-f", caCertPath)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return fmt.Errorf("failed to remove CA certificate from system trust store: %w", err)
	}
	// --fresh is required to drop the certificate from the generated bundle;
	// a plain update only adds newly found certificates.
	return updateCACertificates("--fresh")
}

func updateCACertificates(args ...string) error {
	out, err := sh.OutputA("sudo", append([]string{"update-ca-certificates"}, args...)...)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return fmt.Errorf("failed to update system trust store: %w", err)
	}
	return nil
}
