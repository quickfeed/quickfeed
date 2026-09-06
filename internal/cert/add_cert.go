//go:build !(darwin || linux || windows)

package cert

import (
	"log"
	"runtime"
)

// AddTrustedCert prints not supported message for unsupported OS.
func AddTrustedCert(_ string) error {
	log.Printf("Adding self-signed certificate to the system trust store on %s currently not supported", runtime.GOOS)
	return nil
}

// RemoveTrustedCert prints not supported message for unsupported OS.
func RemoveTrustedCert(_ string) error {
	log.Printf("Removing self-signed certificate from the system trust store on %s currently not supported", runtime.GOOS)
	return nil
}
