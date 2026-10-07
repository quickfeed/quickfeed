package score

import (
	"os"
	"strings"
)

const (
	secretEnvName = "QUICKFEED_SESSION_SECRET"
)

var sessionSecret string

func init() {
	sessionSecret = os.Getenv(secretEnvName)
	// remove variable as soon as it has been read
	_ = os.Setenv(secretEnvName, "")
}

// Redact replaces every occurrence of the given session secrets in output, so
// that output captured from a test run can be shown or logged safely. A run's
// secret is what proves a score line came from the course's tests, so it must
// never reach a student or a log.
func Redact(output string, secrets ...string) string {
	for _, secret := range secrets {
		if secret != "" {
			output = strings.ReplaceAll(output, secret, "[REDACTED]")
		}
	}
	return output
}
