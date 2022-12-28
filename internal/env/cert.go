package env

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	defaultCertPath = "internal/config/certs"
	defaultDomain   = "127.0.0.1"
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

// WhiteList returns a list of domains that the server will create certificates for.
func Whitelist() ([]string, error) {
	domains := os.Getenv("QUICKFEED_WHITELIST")
	if domains == "" {
		return nil, errors.New("required whitelist is undefined")
	}
	if regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`).MatchString(domains) {
		return nil, errors.New("whitelist contains IP addresses")
	}
	// Split domains by comma and remove whitespace and empty entries
	domainList := make([]string, 0)
	for _, domain := range strings.Split(strings.ReplaceAll(domains, " ", ""), ",") {
		if domain == "" {
			continue
		}
		if IsLocal(domain) {
			return nil, fmt.Errorf("whitelist contains local/private domain: %s", domain)
		}
		domainList = append(domainList, domain)
	}
	if len(domainList) == 0 {
		return nil, errors.New("required whitelist is undefined")
	}
	return domainList, nil
}

// CertFile returns the full path to the certificate file.
// To specify a different file, use the QUICKFEED_CERT_FILE environment variable.
func CertFile() string {
	certFile := os.Getenv("QUICKFEED_CERT_FILE")
	if certFile == "" {
		// If cert file is not specified, use the default cert file.
		certFile = filepath.Join(CertPath(), Domain(), defaultCertFile)
	}
	return certFile
}

// KeyFile returns the full path to the certificate key file.
// To specify a different key, use the QUICKFEED_KEY_FILE environment variable.
func KeyFile() string {
	keyFile := os.Getenv("QUICKFEED_KEY_FILE")
	if keyFile == "" {
		// If cert key is not specified, use the default cert key.
		keyFile = filepath.Join(CertPath(), Domain(), defaultKeyFile)
	}
	return keyFile
}

func CertPath() string {
	certPath := os.Getenv("QUICKFEED_CERT_PATH")
	if certPath == "" {
		certPath = defaultCertPath
	}
	return certPath
}

func IsLocal(domain string) bool {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return false
	}
	for _, ip := range ips {
		if !(ip.IsLoopback() || ip.IsPrivate()) {
			return false
		}
	}
	return true
}
