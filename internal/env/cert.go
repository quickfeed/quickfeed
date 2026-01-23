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
	defaultDomain        = "localhost"
	defaultPort          = "443"
	defaultFullchainFile = "fullchain.pem"
	defaultCAFile        = "cert.pem"
	defaultPrivKeyFile   = "privkey.pem"
	defaultConfigDir     = ".config/quickfeed"
	defaultCertDir       = "certs"
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

func HttpAddr() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	return fmt.Sprintf(":%s", port)
}

func DomainWithPort() string {
	return Domain() + HttpAddr()
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
	for domain := range strings.SplitSeq(strings.ReplaceAll(domains, " ", ""), ",") {
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

// FullchainFile returns the full path to the certificate file containing the full certificate chain.
// To specify a different file, use the QUICKFEED_FULLCHAIN_FILE environment variable.
func FullchainFile() string {
	certFile := os.Getenv("QUICKFEED_FULLCHAIN_FILE")
	if certFile == "" {
		// If cert file is not specified, use the default cert file.
		certFile = filepath.Join(CertPath(), defaultFullchainFile)
	}
	return certFile
}

// CAFile returns the full path to the CA certificate file.
// To specify a different file, use the QUICKFEED_CA_FILE environment variable.
func CAFile() string {
	caFile := os.Getenv("QUICKFEED_CA_FILE")
	if caFile == "" {
		// If CA file is not specified, use the default CA file.
		caFile = filepath.Join(CertPath(), defaultCAFile)
	}
	return caFile
}

// PrivKeyFile returns the full path to the private key file.
// To specify a different key, use the QUICKFEED_PRIVKEY_FILE environment variable.
func PrivKeyFile() string {
	privKeyFile := os.Getenv("QUICKFEED_PRIVKEY_FILE")
	if privKeyFile == "" {
		// If cert key is not specified, use the default cert key.
		privKeyFile = filepath.Join(CertPath(), defaultPrivKeyFile)
	}
	return privKeyFile
}

// CertPath returns the full path to the directory containing the certificates.
// If QUICKFEED_CERT_PATH is not set, the default path $HOME/.config/quickfeed/certs is used.
func CertPath() string {
	certPath := os.Getenv("QUICKFEED_CERT_PATH")
	if certPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			// Fallback to source tree if home directory is not available
			home = Root()
		}
		certPath = filepath.Join(home, defaultConfigDir, defaultCertDir)
	}
	return certPath
}

func IsLocal(domain string) bool {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return false
	}
	for _, ip := range ips {
		if !ip.IsLoopback() && !ip.IsPrivate() {
			return false
		}
	}
	return true
}

func IsDomainLocal() bool {
	return IsLocal(Domain())
}
