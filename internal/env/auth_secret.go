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
