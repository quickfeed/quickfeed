package env

import (
	"os"
	"path/filepath"
)

const (
	defaultCertPath = "/etc/letsencrypt/live"
	defaultDomain   = "localhost"
	defaultCertFile = "fullchain.pem"
	defaultKeyFile  = "privkey.pem"
)

// Domain returns the domain name where quickfeed will be served.
// Domain should not include the server name.
func Domain() string {
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = defaultDomain
	}
	return domain
}

// CertFile returns the full path to the certificate file.
// To specify a different file, use the QUICKFEED_CERT_FILE environment variable.
func CertFile() string {
	certFile := os.Getenv("QUICKFEED_CERT_FILE")
	if certFile == "" {
		// If cert file is not specified, use the default cert file.
		certFile = filepath.Join(certPath(), Domain(), defaultCertFile)
	}
	return certFile
}

// KeyFile returns the full path to the certificate key file.
// To specify a different key, use the QUICKFEED_KEY_FILE environment variable.
func KeyFile() string {
	keyFile := os.Getenv("QUICKFEED_KEY_FILE")
	if keyFile == "" {
		// If cert key is not specified, use the default cert key.
		keyFile = filepath.Join(certPath(), Domain(), defaultKeyFile)
	}
	return keyFile
}

func certPath() string {
	certPath := os.Getenv("QUICKFEED_CERT_PATH")
	if certPath == "" {
		certPath = defaultCertPath
	}
	return certPath
}
