package env

import (
	"errors"
	"fmt"
	"log"

	"github.com/quickfeed/quickfeed/internal/cert"
)

// EnsureReady checks and prepares the environment for running QuickFeed.
// It auto-generates missing configuration where appropriate:
//   - Auth secret: generates if missing (both dev and production)
//   - Certificates: generates self-signed if missing (dev mode only)
//   - Domain: validates based on mode (localhost default in dev, required in production)
//
// Returns an error only for unrecoverable conditions.
// App data readiness should be checked separately after calling this function.
func EnsureReady(dev bool, envFile string, forceNewSecret bool) error {
	// Step 1: Ensure auth secret exists (or force regeneration)
	if generated, err := EnsureAuthSecret(envFile, forceNewSecret); err != nil {
		return fmt.Errorf("failed to ensure auth secret: %w", err)
	} else if generated {
		if forceNewSecret {
			log.Println("Generated new JWT signing secret (forced rotation)")
		} else {
			log.Println("Generated new JWT signing secret")
		}
	}

	// Step 2: Validate domain based on mode
	if err := validateDomain(dev); err != nil {
		return err
	}

	// Step 3: Ensure certificates exist (dev mode only)
	if dev {
		if err := ensureCertificates(); err != nil {
			return err
		}
	}

	return nil
}

// validateDomain checks that the domain is appropriate for the given mode.
// In dev mode, any domain is allowed (defaults to localhost).
// In production mode, the domain must be public (not local/private).
func validateDomain(dev bool) error {
	if dev {
		// Dev mode: any domain is fine, localhost is the default
		return nil
	}
	// Production mode: domain must be set and public
	if IsDomainLocal() {
		return fmt.Errorf("domain %q is local/private; production mode requires a public domain", Domain())
	}
	return nil
}

// ensureCertificates checks if self-signed certificates exist, and generates them if not.
// This is only used in dev mode.
func ensureCertificates() error {
	// Check if all certificate files exist
	if exists(FullchainFile()) && exists(PrivKeyFile()) && exists(CAFile()) {
		return nil
	}

	log.Printf("Generating self-signed certificates for domain: %s", Domain())
	if err := cert.GenerateSelfSignedCert(cert.Options{
		FullchainFile: FullchainFile(),
		PrivKeyFile:   PrivKeyFile(),
		CAFile:        CAFile(),
		Hosts:         Domain(),
	}); err != nil {
		return fmt.Errorf("failed to generate self-signed certificates: %w", err)
	}
	log.Printf("Certificates successfully generated at: %s", CertPath())

	log.Print("Adding certificate to local trust store (may require elevated privileges)")
	if err := cert.AddTrustedCert(CAFile()); err != nil {
		return fmt.Errorf("failed to install self-signed certificate: %w", err)
	}
	log.Print("Certificate successfully added to trust store")

	return nil
}

// NeedsAppCreation returns true if any app data is missing and the app creation flow should run.
// This checks for: ClientID, ClientSecret, AppID, AppURL, and AppKey file.
func NeedsAppCreation() bool {
	if _, err := ClientID(); err != nil {
		return true
	}
	if _, err := ClientSecret(); err != nil {
		return true
	}
	if !HasAppID() {
		return true
	}
	if GetAppURL() == "" {
		return true
	}
	if !exists(AppPrivKeyFile()) {
		return true
	}
	return false
}

// CheckAppData returns an error describing which app data is missing, or nil if all data exists.
func CheckAppData() (errs error) {
	if _, err := ClientID(); err != nil {
		errs = errors.Join(errs, errors.New("QUICKFEED_CLIENT_ID is not set"))
	}
	if _, err := ClientSecret(); err != nil {
		errs = errors.Join(errs, errors.New("QUICKFEED_CLIENT_SECRET is not set"))
	}
	if !HasAppID() {
		errs = errors.Join(errs, errors.New("QUICKFEED_APP_ID is not set"))
	}
	if GetAppURL() == "" {
		errs = errors.Join(errs, errors.New("QUICKFEED_APP_URL is not set"))
	}
	if !exists(AppPrivKeyFile()) {
		errs = errors.Join(errs, fmt.Errorf("QuickFeed App private key file not found: %s", AppPrivKeyFile()))
	}
	return
}
