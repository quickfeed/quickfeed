package env

import (
	"os"

	"github.com/quickfeed/quickfeed/internal/rand"
)

// AuthSecret returns the secret used to sign JWT tokens.
// If QUICKFEED_AUTH_SECRET is not set, a random secret is generated.
// Allows for a custom secret to be set.
func AuthSecret() string {
	authSecret := os.Getenv("QUICKFEED_AUTH_SECRET")
	if authSecret == "" {
		return rand.String()
	}
	return authSecret
}
