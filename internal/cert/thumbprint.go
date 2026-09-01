//go:build darwin || windows

package cert

import (
	"crypto/sha1"
	"encoding/pem"
	"fmt"
	"os"
)

// caCertThumbprint returns the SHA-1 thumbprint of the certificate in caFile as
// an uppercase hex string. Both security(1) and certutil identify a certificate
// already in the store by this thumbprint; SHA-1 is used here only as the store's
// identifier, not as a security primitive.
func caCertThumbprint(caFile string) (string, error) {
	pemBytes, err := os.ReadFile(caFile)
	if err != nil {
		return "", fmt.Errorf("reading CA certificate: %w", err)
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != "CERTIFICATE" {
		return "", fmt.Errorf("no certificate found in %s", caFile)
	}
	return fmt.Sprintf("%X", sha1.Sum(block.Bytes)), nil
}
