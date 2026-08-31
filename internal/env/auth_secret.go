package env

import (
	"os"

	"github.com/quickfeed/quickfeed/internal/rand"
)

const authSecret = "QUICKFEED_AUTH_SECRET" // skipcq: SCT-A000

// AuthSecret returns the JWT signing secret obtained from
// the QUICKFEED_AUTH_SECRET environment variable.
func AuthSecret() string {
	return os.Getenv(authSecret)
}

// NewAuthSecret generates a new random JWT signing secret,
// and saves it to the given environment file for future use.
// It returns an error if the environment file cannot be updated.
func NewAuthSecret(envFile string) error {
	return Save(RootEnv(envFile), map[string]string{
		authSecret: rand.String(),
	})
}

// EnsureAuthSecret generates and saves a new auth secret if one doesn't exist.
// If force is true, it will regenerate even if one already exists.
// Returns true if a new secret was generated, false otherwise.
func EnsureAuthSecret(envFile string, force bool) (bool, error) {
	if !force && AuthSecret() != "" {
		return false, nil
	}
	if err := NewAuthSecret(envFile); err != nil {
		return false, err
	}
	// Reload environment to pick up the new secret
	if err := Load(RootEnv(envFile)); err != nil {
		return false, err
	}
	return true, nil
}
