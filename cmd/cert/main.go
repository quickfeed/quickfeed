// Command cert generates self-signed certificates for local development and
// manages the CA certificate in the system trust store.
//
// Exactly one action is required; there is no default, because every action either
// writes a new private key or changes the system trust store:
//
//	cert -gen         generate certificates and trust the CA
//	cert -gen -force  replace existing certificates, untrusting the CA they replace
//	cert -add         trust the existing CA, without regenerating
//	cert -remove      stop trusting the CA
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/quickfeed/quickfeed/internal/cert"
	"github.com/quickfeed/quickfeed/internal/env"
)

// errUsage marks an invalid flag combination, which is reported with the usage
// message rather than as a plain failure.
var errUsage = errors.New("invalid flags")

func main() {
	var (
		gen    = flag.Bool("gen", false, "generate certificates and add the CA certificate to the system trust store")
		add    = flag.Bool("add", false, "add the existing CA certificate to the system trust store")
		remove = flag.Bool("remove", false, "remove the CA certificate from the system trust store")
		force  = flag.Bool("force", false, "with -gen: replace existing certificates, untrusting the CA they replace")
	)
	flag.Parse()

	if err := run(*gen, *add, *remove, *force); err != nil {
		if errors.Is(err, errUsage) {
			fmt.Fprintln(os.Stderr, "cert:", err)
			flag.Usage()
			os.Exit(2)
		}
		log.Fatal(err)
	}
}

func run(gen, add, remove, force bool) error {
	switch {
	case actions(gen, add, remove) != 1:
		return fmt.Errorf("%w: exactly one of -gen, -add or -remove is required", errUsage)
	case force && !gen:
		return fmt.Errorf("%w: -force can only be used together with -gen", errUsage)
	}

	// Load environment variables from $QUICKFEED/.env.
	const envFile = ".env"
	if err := env.Load(env.RootEnv(envFile)); err != nil {
		return err
	}

	switch {
	case gen:
		if err := generateCerts(force); err != nil {
			return err
		}
		return addTrustedCert()
	case add:
		return addTrustedCert()
	default:
		return removeTrustedCert()
	}
}

// actions returns the number of the given action flags that are set.
func actions(flags ...bool) int {
	n := 0
	for _, f := range flags {
		if f {
			n++
		}
	}
	return n
}

// generateCerts generates a new certificate chain. Existing certificates are only
// replaced when force is set, since generating always mints a new private key and
// CA certificate; anything already trusting the old CA stops working.
func generateCerts(force bool) error {
	caFile := env.CAFile()
	if existing := existingFiles(env.FullchainFile(), caFile, env.PrivKeyFile()); len(existing) > 0 {
		if !force {
			return fmt.Errorf("certificates already exist at %s (%s); use -gen -force to replace them, or -add to trust the existing CA",
				env.CertPath(), strings.Join(existing, ", "))
		}
		// Untrust the CA that is about to be overwritten while its file is still on
		// disk. macOS and Windows identify a trusted certificate by the file's
		// contents, so once it is replaced the old CA can no longer be removed and
		// would stay trusted forever.
		if fileExists(caFile) {
			log.Println("Removing the CA certificate being replaced from the system trust store...")
			if err := cert.RemoveTrustedCert(caFile); err != nil {
				return fmt.Errorf("removing the CA certificate being replaced: %w", err)
			}
		}
	}

	log.Printf("Generating self-signed certificates for %s...", env.Domain())
	if err := cert.GenerateSelfSignedCert(cert.Options{
		FullchainFile: env.FullchainFile(),
		CAFile:        caFile,
		PrivKeyFile:   env.PrivKeyFile(),
		Hosts:         env.Domain(),
	}); err != nil {
		return fmt.Errorf("generating certificates: %w", err)
	}
	log.Printf("Certificates successfully generated at: %s", env.CertPath())
	return nil
}

func addTrustedCert() error {
	caFile := env.CAFile()
	if !fileExists(caFile) {
		return fmt.Errorf("no CA certificate found at %s; run cert -gen to generate one", caFile)
	}
	log.Println("Adding certificate to system trust store (requires sudo access)...")
	if err := cert.AddTrustedCert(caFile); err != nil {
		return fmt.Errorf("adding certificate to trust store: %w", err)
	}
	log.Println("Certificate successfully added to system trust store")
	return nil
}

// removeTrustedCert does not require the CA certificate to still exist on disk;
// on Linux the trust store entry is identified by its installed path, so it can be
// removed even after the generated certificates have been deleted.
func removeTrustedCert() error {
	log.Println("Removing certificate from system trust store (requires sudo access)...")
	if err := cert.RemoveTrustedCert(env.CAFile()); err != nil {
		return fmt.Errorf("removing certificate from trust store: %w", err)
	}
	log.Println("Certificate successfully removed from system trust store")
	return nil
}

// existingFiles returns the base names of those paths that already exist.
func existingFiles(paths ...string) []string {
	var found []string
	for _, path := range paths {
		if fileExists(path) {
			found = append(found, filepath.Base(path))
		}
	}
	return found
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
