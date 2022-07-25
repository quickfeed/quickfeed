package env

import (
	"os"
	"path/filepath"
)

const (
	defaultCertPath = "/etc/letsencrypt/live"
	defaultDomain   = "localhost"
	defaultCertFile = "fullchain.pem"
	defaultCertKey  = "privkey.pem"
)

var (
	domain   string
	certPath string
	certFile string
	certKey  string
)

func init() {
	// Domain should not include the server name.
	domain = os.Getenv("DOMAIN")
	if domain == "" {
		domain = defaultDomain
	}
	certPath = os.Getenv("QUICKFEED_CERT_PATH")
	if certPath == "" {
		certPath = defaultCertPath
	}
	certFile = os.Getenv("QUICKFEED_CERT_FILE")
	if certFile == "" {
		// If cert file is not specified, use the default cert file.
		certFile = filepath.Join(defaultCertPath, domain, defaultCertFile)
	}
	certKey = os.Getenv("QUICKFEED_CERT_KEY")
	if certKey == "" {
		// If cert key is not specified, use the default cert key.
		certFile = filepath.Join(defaultCertPath, domain, defaultCertKey)
	}
}

// CertFile returns the full path to the certificate file.
// To specify a different file, use the QUICKFEED_CERT_FILE environment variable.
func CertFile() string {
	return certFile
}

// CertKey returns the full path to the certificate key file.
// To specify a different key, use the QUICKFEED_CERT_KEY environment variable.
func CertKey() string {
	return certKey
}
