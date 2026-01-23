//go:build darwin

package cert

import (
	"log"

	"github.com/quickfeed/quickfeed/kit/sh"
)

const keychain = "/Library/Keychains/System.keychain"

// AddTrustedCert adds given the certificate the user's keychain.
func AddTrustedCert(caFile string) error {
	out, err := sh.OutputA("sudo", "security", "add-trusted-cert", "-d", "-r", "trustRoot", "-k", keychain, caFile)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return err
	}
	return nil
}
