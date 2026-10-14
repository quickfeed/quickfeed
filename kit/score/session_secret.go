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

// Redact replaces every occurrence of secret in output with [REDACTED].
// An empty secret leaves output unchanged.
func Redact(output, secret string) string {
	if secret == "" {
		return output
	}
	return strings.ReplaceAll(output, secret, "[REDACTED]")
}
