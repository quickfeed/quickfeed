//go:build windows

package cert

import (
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
		return err
	}
	return nil
}
