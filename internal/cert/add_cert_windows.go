//go:build windows

package cert

import (
	"log"

	"github.com/quickfeed/quickfeed/kit/sh"
)

// AddTrustedCert adds given the certificate the user's keychain.
func AddTrustedCert(certFile string) error {
	out, err := sh.OutputA("certutil", "-addstore", "-f", "ROOT", certFile)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return err
	}
	return nil
}
