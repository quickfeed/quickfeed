package score_test

import (
	"os"
	"testing"
)

const (
	secretEnvName = "QUICKFEED_SESSION_SECRET"
)

// To run this test, use this command:
//   QUICKFEED_SESSION_SECRET=hei go test -v -run TestSessionSecret
//
func TestSessionSecret(t *testing.T) {
	sessionSecret := os.Getenv(secretEnvName)
	if sessionSecret != "" {
		t.Fatalf("Unexpected access to %s=%s", secretEnvName, sessionSecret)
	}
}
