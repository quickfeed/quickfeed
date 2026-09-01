package cert

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

// certFiles holds the paths written by GenerateSelfSignedCert.
type certFiles struct {
	fullchain string
	caCert    string
	privKey   string
}

func generateCerts(t *testing.T) certFiles {
	t.Helper()
	dir := t.TempDir()
	files := certFiles{
		fullchain: filepath.Join(dir, "fullchain.pem"),
		caCert:    filepath.Join(dir, "quickfeed-ca.crt"),
		privKey:   filepath.Join(dir, "privkey.pem"),
	}
	if err := GenerateSelfSignedCert(Options{
		FullchainFile: files.fullchain,
		CAFile:        files.caCert,
		PrivKeyFile:   files.privKey,
		Hosts:         "127.0.0.1",
	}); err != nil {
		t.Fatalf("GenerateSelfSignedCert() = %v, want nil", err)
	}
	return files
}

// TestCAFileHoldsSingleCACert checks the invariant AddTrustedCert relies on: the
// CA file holds exactly one certificate, and it is a self-signed CA.
// update-ca-certificates(8) requires one certificate per file, so installing the
// fullchain instead would be silently ignored.
func TestCAFileHoldsSingleCACert(t *testing.T) {
	caPEM, err := os.ReadFile(generateCerts(t).caCert)
	if err != nil {
		t.Fatal(err)
	}
	block, rest := pem.Decode(caPEM)
	if block == nil {
		t.Fatal("CA file does not contain a PEM encoded certificate")
	}
	if len(rest) != 0 {
		t.Errorf("CA file has %d trailing bytes, want exactly one certificate", len(rest))
	}
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if !caCert.IsCA {
		t.Error("CA file holds a certificate with IsCA = false, want true")
	}
	if caCert.Subject.String() != caCert.Issuer.String() {
		t.Errorf("CA certificate is not self-signed: subject %q, issuer %q", caCert.Subject, caCert.Issuer)
	}
}

// TestServerCertVerifiesAgainstCA checks that trusting the CA file is enough to
// validate the server certificate; that is what installing it buys.
func TestServerCertVerifiesAgainstCA(t *testing.T) {
	files := generateCerts(t)
	caPEM, err := os.ReadFile(files.caCert)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(caPEM) {
		t.Fatal("CA file is not a usable trust anchor")
	}

	fullchain, err := os.ReadFile(files.fullchain)
	if err != nil {
		t.Fatal(err)
	}
	// The server certificate is the first entry in the fullchain.
	serverBlock, _ := pem.Decode(fullchain)
	if serverBlock == nil {
		t.Fatal("fullchain does not contain a PEM encoded certificate")
	}
	serverCert, err := x509.ParseCertificate(serverBlock.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := serverCert.Verify(x509.VerifyOptions{Roots: roots, DNSName: "127.0.0.1"}); err != nil {
		t.Errorf("server certificate does not verify against the CA: %v", err)
	}
}

// TestTrustStoreRoundTrip installs the CA certificate into the system trust store
// and removes it again. It requires elevated privileges and modifies the machine it
// runs on, so it is skipped unless QUICKFEED_TRUST_STORE_TEST is set.
//
// Note: this deliberately does not assert via x509.SystemCertPool; Go caches the
// system pool for the lifetime of the process, so a read after AddTrustedCert would
// not reflect the change. Verify manually with openssl instead.
func TestTrustStoreRoundTrip(t *testing.T) {
	if os.Getenv("QUICKFEED_TRUST_STORE_TEST") == "" {
		t.Skip("skipping; set QUICKFEED_TRUST_STORE_TEST=1 to run (requires sudo, modifies the system trust store)")
	}
	caFile := generateCerts(t).caCert
	if err := AddTrustedCert(caFile); err != nil {
		t.Fatalf("AddTrustedCert() = %v, want nil", err)
	}
	if err := RemoveTrustedCert(caFile); err != nil {
		t.Fatalf("RemoveTrustedCert() = %v, want nil", err)
	}
}
