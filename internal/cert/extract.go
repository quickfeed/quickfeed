package cert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// caCertificate reads a fullchain certificate file and returns the CA certificate.
// The CA certificate is expected to be the last certificate in the chain.
func caCertificate(fullchainFile string) (*x509.Certificate, error) {
	fullchainBytes, err := os.ReadFile(fullchainFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Parse all certificates from the fullchain
	var certs []*x509.Certificate
	for block, rest := pem.Decode(fullchainBytes); block != nil; block, rest = pem.Decode(rest) {
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate: %w", err)
			}
			certs = append(certs, cert)
		}
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found in %s", fullchainFile)
	}
	return certs[len(certs)-1], nil
}

// extractCACert reads a fullchain certificate file and extracts only the CA
// certificate, encoded in PEM format.
func extractCACert(fullchainFile string) ([]byte, error) {
	caCert, err := caCertificate(fullchainFile)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCert.Raw,
	}), nil
}

// caCertTempFile extracts the CA certificate from the given fullchain file and
// writes it to a temporary file, returning the path to that file. The caller is
// responsible for removing the file.
//
// The file is created with a random name and mode 0600 so that it cannot be
// pre-created or modified by another user on a shared machine; the trust store
// commands run under sudo and would otherwise install whatever the file holds.
func caCertTempFile(fullchainFile string) (string, error) {
	caCertPEM, err := extractCACert(fullchainFile)
	if err != nil {
		return "", err
	}
	f, err := os.CreateTemp("", "quickfeed-ca-*.crt")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary CA certificate: %w", err)
	}
	if _, err := f.Write(caCertPEM); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", fmt.Errorf("failed to write temporary CA certificate: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("failed to close temporary CA certificate: %w", err)
	}
	return f.Name(), nil
}
