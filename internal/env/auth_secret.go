package env

import "os"

// AuthSecret returns the JWT signing secret obtained from
// the QUICKFEED_AUTH_SECRET environment variable.
func AuthSecret() string {
	return os.Getenv("QUICKFEED_AUTH_SECRET")
}
