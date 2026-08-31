package cert

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

// generateChain writes a fullchain/key pair into a temporary directory and
// returns the path to the fullchain file.
func generateChain(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	certFile := filepath.Join(dir, "fullchain.pem")
	if err := GenerateSelfSignedCert(Options{
		KeyFile:  filepath.Join(dir, "privkey.pem"),
		CertFile: certFile,
		Hosts:    "127.0.0.1",
	}); err != nil {
		t.Fatalf("GenerateSelfSignedCert() = %v, want nil", err)
	}
	return certFile
}

func TestCACertificate(t *testing.T) {
	caCert, err := caCertificate(generateChain(t))
	if err != nil {
		t.Fatalf("caCertificate() = %v, want nil", err)
	}
	if !caCert.IsCA {
		t.Error("caCertificate() returned a certificate with IsCA = false, want true")
	}
	if caCert.Subject.String() != caCert.Issuer.String() {
		t.Errorf("caCertificate() returned a non-self-signed certificate: subject %q, issuer %q",
			caCert.Subject, caCert.Issuer)
	}
}

// TestExtractCACert checks that exactly one certificate is extracted, and that
// it is the CA that signed the server certificate. update-ca-certificates(8)
// requires one certificate per file, so a fullchain must not be installed as is.
func TestExtractCACert(t *testing.T) {
	certFile := generateChain(t)
	caCertPEM, err := extractCACert(certFile)
	if err != nil {
		t.Fatalf("extractCACert() = %v, want nil", err)
	}

	block, rest := pem.Decode(caCertPEM)
	if block == nil {
		t.Fatal("extractCACert() did not return a PEM encoded certificate")
	}
	if len(rest) != 0 {
		t.Errorf("extractCACert() returned %d trailing bytes, want a single certificate", len(rest))
	}

	// The extracted CA must verify the server certificate in the fullchain.
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(caCertPEM) {
		t.Fatal("extractCACert() did not return a usable CA certificate")
	}
	fullchain, err := os.ReadFile(certFile)
	if err != nil {
		t.Fatal(err)
	}
	serverBlock, _ := pem.Decode(fullchain)
	serverCert, err := x509.ParseCertificate(serverBlock.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := serverCert.Verify(x509.VerifyOptions{Roots: roots, DNSName: "127.0.0.1"}); err != nil {
		t.Errorf("server certificate does not verify against extracted CA: %v", err)
	}
}

func TestExtractCACertErrors(t *testing.T) {
	if _, err := extractCACert(filepath.Join(t.TempDir(), "missing.pem")); err == nil {
		t.Error("extractCACert() = nil, want error for missing file")
	}
	empty := filepath.Join(t.TempDir(), "empty.pem")
	if err := os.WriteFile(empty, []byte("not a certificate"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := extractCACert(empty); err == nil {
		t.Error("extractCACert() = nil, want error for file without certificates")
	}
}

// TestCACertTempFile checks that the temporary file is private to the current
// user. The trust store commands run under sudo, so a file another user could
// create or modify would let them choose which CA gets installed.
func TestCACertTempFile(t *testing.T) {
	certFile := generateChain(t)
	tmpFile, err := caCertTempFile(certFile)
	if err != nil {
		t.Fatalf("caCertTempFile() = %v, want nil", err)
	}
	defer os.Remove(tmpFile)

	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("caCertTempFile() created file with mode %#o, want 0600", perm)
	}

	want, err := extractCACert(certFile)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Error("caCertTempFile() wrote contents that differ from the extracted CA certificate")
	}

	// A second call must not reuse the same path; a predictable name in a
	// world-writable temp directory can be pre-created by another user.
	other, err := caCertTempFile(certFile)
	if err != nil {
		t.Fatalf("caCertTempFile() = %v, want nil", err)
	}
	defer os.Remove(other)
	if other == tmpFile {
		t.Errorf("caCertTempFile() reused path %s, want a unique path per call", tmpFile)
	}
}

// TestTrustStoreRoundTrip installs QuickFeed's CA certificate into the system
// trust store and removes it again. It requires sudo and modifies the machine
// it runs on, so it is skipped unless QUICKFEED_TRUST_STORE_TEST is set.
//
// Note: this deliberately does not assert via x509.SystemCertPool; Go caches the
// system pool for the lifetime of the process, so a read after AddTrustedCert
// would not reflect the change. Verify manually with openssl instead.
func TestTrustStoreRoundTrip(t *testing.T) {
	if os.Getenv("QUICKFEED_TRUST_STORE_TEST") == "" {
		t.Skip("skipping; set QUICKFEED_TRUST_STORE_TEST=1 to run (requires sudo, modifies the system trust store)")
	}
	certFile := generateChain(t)
	if err := AddTrustedCert(certFile); err != nil {
		t.Fatalf("AddTrustedCert() = %v, want nil", err)
	}
	if err := RemoveTrustedCert(certFile); err != nil {
		t.Fatalf("RemoveTrustedCert() = %v, want nil", err)
	}
}
