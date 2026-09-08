//go:build windows

package cert

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"slices"
	"strings"

	"github.com/quickfeed/quickfeed/kit/sh"
)

// notFoundCodes are the HRESULTs certutil exits with when the store holds no
// certificate matching the requested thumbprint. Which one is returned depends on
// the lookup path certutil takes, so both must be recognized as "not installed".
var notFoundCodes = []uint32{
	0x80092004, // CRYPT_E_NOT_FOUND
	0x80090011, // NTE_NOT_FOUND
}

// isNotFound reports whether err is certutil exiting with a not-found HRESULT.
func isNotFound(err error) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	return slices.Contains(notFoundCodes, uint32(exitErr.ExitCode()))
}

// hasThumbprint reports whether a certutil store dump describes the certificate
// with the given thumbprint. certutil matches its CertId argument loosely: an
// all-digit token is read as a store index, and other tokens match substrings of
// several certificate fields, so a successful exit says only that certutil found
// something, not that it found ours. The dump prints the SHA-1 hash of every
// certificate it lists, so require that our thumbprint is among them.
func hasThumbprint(storeDump, thumbprint string) bool {
	normalize := func(s string) string {
		return strings.ToLower(strings.NewReplacer(" ", "", ":", "").Replace(s))
	}
	return strings.Contains(normalize(storeDump), normalize(thumbprint))
}

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
// It is a no-op if the certificate is not installed.
func RemoveTrustedCert(caFile string) error {
	thumbprint, err := caCertThumbprint(caFile)
	if err != nil {
		return err
	}
	// certutil -store exits with a not-found HRESULT when the store holds no
	// certificate matching this thumbprint, which is not an error here. Any other
	// failure means we could not determine whether the certificate is installed,
	// and must be propagated: reporting a still-installed CA as absent would let
	// the caller replace the CA file while the old root remains trusted.
	// -verifystore is not used for this check: it also validates the certificate,
	// so an expired CA that is present and still needs removing would be reported
	// as absent.
	out, err := sh.OutputA("certutil", "-store", "ROOT", thumbprint)
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("checking ROOT store for CA certificate: %w", err)
	}
	// Deleting on anything less than a thumbprint match in the dump risks removing
	// a certificate certutil matched by some other rule, such as a store index.
	if err != nil || !hasThumbprint(out, thumbprint) {
		log.Print("No QuickFeed CA certificate found in the ROOT store")
		return nil
	}
	out, err = sh.OutputA("certutil", "-delstore", "ROOT", thumbprint)
	if out != "" {
		log.Print(out)
	}
	if err != nil {
		return fmt.Errorf("removing CA certificate from ROOT store: %w", err)
	}
	return nil
}
