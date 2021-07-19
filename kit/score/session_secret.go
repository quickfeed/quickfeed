package score

import "os"

const (
	secretEnvName = "QUICKFEED_SESSION_SECRET"
)

var sessionSecret string

func init() {
	sessionSecret = os.Getenv(secretEnvName)
	// remove variable as soon as it has been read
	_ = os.Setenv(secretEnvName, "")
}
