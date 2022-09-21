//go:build !(darwin || linux || windows)

package cert

import (
	"log"
	"os"
)

// AddTrustedCert adds given the certificate the user's keychain.
func AddTrustedCert(_ string) error {
	log.Printf("Adding self-signed certificate to keychain on %s currently not supported", os.Getenv("OS"))
	return nil
}
