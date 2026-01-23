//go:build linux

package cert

import (
	"log"

	"github.com/quickfeed/quickfeed/kit/sh"
)

// AddTrustedCert adds given the certificate the user's keychain.
func AddTrustedCert(caFile string) error {
	const certPath = "/usr/local/share/ca-certificates/"
	out, err := sh.OutputA("sudo", "cp", caFile, certPath)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return err
	}
	out, err = sh.Output("sudo update-ca-certificates")
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return err
	}
	return nil
}
