package env

import (
	"os"
)

const (
	defaultCertFile = "cert.pem"
	defaultCertKey  = "key.pem"
)

var (
	certFile string
	certKey  string
)

func init() {
	certFile = os.Getenv("QUICKFEED_CERT_FILE")
	if certFile == "" {
		certFile = defaultCertFile
	}
	certKey = os.Getenv("QUICKFEED_CERT_KEY")
	if certKey == "" {
		certKey = defaultCertKey
	}
}

func CertFile() string {
	return certFile
}

func CertKey() string {
	return certKey
}
