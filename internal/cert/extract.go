package cert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// extractCACert reads a fullchain certificate file and extracts only the CA certificate.
// The CA certificate is expected to be the last certificate in the chain.
func extractCACert(fullchainFile string) ([]byte, error) {
	// Read the fullchain certificate file
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

	// The CA certificate is the last one in the chain
	caCert := certs[len(certs)-1]

	// Encode the CA certificate to PEM format
	caCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCert.Raw,
	})

	return caCertPEM, nil
}
