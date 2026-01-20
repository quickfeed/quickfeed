package main

import (
	"log"

	"github.com/quickfeed/quickfeed/internal/cert"
	"github.com/quickfeed/quickfeed/internal/env"
)

func main() {
	// Load environment variables from $QUICKFEED/.env.
	const envFile = ".env"
	if err := env.Load(env.RootEnv(envFile)); err != nil {
		log.Fatal(err)
	}

	log.Printf("Generating self-signed certificates for %s...", env.Domain())
	if err := cert.GenerateSelfSignedCert(cert.Options{
		KeyFile:  env.KeyFile(),
		CertFile: env.CertFile(),
		Hosts:    env.Domain(),
	}); err != nil {
		log.Fatalf("Failed to generate certificates: %v", err)
	}
	log.Printf("Certificates successfully generated at: %s", env.CertPath())

	log.Println("Adding certificate to system trust store (requires sudo access)...")
	if err := cert.AddTrustedCert(env.CertFile()); err != nil {
		log.Fatalf("Failed to add certificate to trust store: %v", err)
	}
	log.Println("Certificate successfully added to system trust store")
}
