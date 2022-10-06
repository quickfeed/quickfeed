//go:build darwin

package cert

import (
	"fmt"
	"log"
	"os"

	"github.com/quickfeed/quickfeed/kit/sh"
)

// AddTrustedCert adds given the certificate the user's keychain.
func AddTrustedCert(certFile string) error {
	keychain := fmt.Sprintf("/Users/%s/Library/Keychains/login.keychain", os.Getenv("USER"))
	out, err := sh.OutputA("sudo", "security", "add-trusted-cert", "-d", "-r", "trustAsRoot", "-k", keychain, certFile)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return err
	}
	return nil
}
