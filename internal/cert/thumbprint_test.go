//go:build darwin || windows

package cert

import (
	"crypto/sha1"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestCACertThumbprint checks that the thumbprint identifies the CA certificate
// the way security(1) and certutil do: the SHA-1 digest of the DER bytes, as
// uppercase hex without separators.
func TestCACertThumbprint(t *testing.T) {
	caFile := generateCerts(t).caCert
	got, err := caCertThumbprint(caFile)
	if err != nil {
		t.Fatalf("caCertThumbprint(%q) = %v, want nil", caFile, err)
	}
	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(caPEM)
	if block == nil {
		t.Fatal("CA file does not contain a PEM encoded certificate")
	}
	if want := fmt.Sprintf("%X", sha1.Sum(block.Bytes)); got != want {
		t.Errorf("caCertThumbprint(%q) = %q, want %q", caFile, got, want)
	}
	if len(got) != 40 {
		t.Errorf("caCertThumbprint(%q) has length %d, want 40 hex characters", caFile, len(got))
	}
}

func TestCACertThumbprintErrors(t *testing.T) {
	dir := t.TempDir()
	notPEM := filepath.Join(dir, "not-a-cert.crt")
	if err := os.WriteFile(notPEM, []byte("this is not a certificate"), 0o600); err != nil {
		t.Fatal(err)
	}
	tests := map[string]string{
		"missing file": filepath.Join(dir, "missing.crt"),
		"not PEM":      notPEM,
	}
	for name, caFile := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := caCertThumbprint(caFile); err == nil {
				t.Errorf("caCertThumbprint(%q) = nil, want error", caFile)
			}
		})
	}
}
