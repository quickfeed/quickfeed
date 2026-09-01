// Command gencert generates self-signed certificates for local development and
// manages the CA certificate in the system trust store.
//
// Exactly one action is required; there is no default, because every action either
// writes a new private key or changes the system trust store:
//
//	gencert -gencert         generate certificates and trust the CA
//	gencert -gencert -force  replace existing certificates, untrusting the CA they replace
//	gencert -addcert         trust the existing CA, without regenerating
//	gencert -removecert      stop trusting the CA
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/quickfeed/quickfeed/internal/cert"
	"github.com/quickfeed/quickfeed/internal/env"
)

func main() {
	var (
		genCert    = flag.Bool("gencert", false, "generate certificates and add the CA certificate to the system trust store")
		addCert    = flag.Bool("addcert", false, "add the existing CA certificate to the system trust store")
		removeCert = flag.Bool("removecert", false, "remove the CA certificate from the system trust store")
		force      = flag.Bool("force", false, "with -gencert: replace existing certificates, untrusting the CA they replace")
	)
	flag.Parse()

	if actions(*genCert, *addCert, *removeCert) != 1 {
		fmt.Fprintln(os.Stderr, "gencert: exactly one of -gencert, -addcert or -removecert is required")
		flag.Usage()
		os.Exit(2)
	}
	if *force && !*genCert {
		log.Fatal("-force can only be used together with -gencert")
	}

	// Load environment variables from $QUICKFEED/.env.
	const envFile = ".env"
	if err := env.Load(env.RootEnv(envFile)); err != nil {
		log.Fatal(err)
	}

	switch {
	case *genCert:
		generateCerts(*force)
		addTrustedCert()
	case *addCert:
		addTrustedCert()
	case *removeCert:
		removeTrustedCert()
	}
}

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
func generateCerts(force bool) {
	caFile := env.CAFile()
	if existing := existingFiles(env.FullchainFile(), caFile, env.PrivKeyFile()); len(existing) > 0 {
		if !force {
			log.Fatalf("Certificates already exist at %s (%s); use -gencert -force to replace them, or -addcert to trust the existing CA",
				env.CertPath(), strings.Join(existing, ", "))
		}
		// Untrust the CA that is about to be overwritten while its file is still on
		// disk. macOS and Windows identify a trusted certificate by the file's
		// contents, so once it is replaced the old CA can no longer be removed and
		// would stay trusted forever.
		if fileExists(caFile) {
			log.Println("Removing the CA certificate being replaced from the system trust store...")
			if err := cert.RemoveTrustedCert(caFile); err != nil {
				log.Fatalf("Failed to remove the CA certificate being replaced: %v", err)
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
		log.Fatalf("Failed to generate certificates: %v", err)
	}
	log.Printf("Certificates successfully generated at: %s", env.CertPath())
}

func addTrustedCert() {
	caFile := env.CAFile()
	if !fileExists(caFile) {
		log.Fatalf("No CA certificate found at %s; run gencert -gencert to generate one", caFile)
	}
	log.Println("Adding certificate to system trust store (requires sudo access)...")
	if err := cert.AddTrustedCert(caFile); err != nil {
		log.Fatalf("Failed to add certificate to trust store: %v", err)
	}
	log.Println("Certificate successfully added to system trust store")
}

// removeTrustedCert does not require the CA certificate to still exist on disk;
// on Linux the trust store entry is identified by its installed path, so it can be
// removed even after the generated certificates have been deleted.
func removeTrustedCert() {
	log.Println("Removing certificate from system trust store (requires sudo access)...")
	if err := cert.RemoveTrustedCert(env.CAFile()); err != nil {
		log.Fatalf("Failed to remove certificate from trust store: %v", err)
	}
	log.Println("Certificate successfully removed from system trust store")
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
