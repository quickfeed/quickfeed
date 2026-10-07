package score

import (
	"os"
	"testing"
)

// To run this test, use this command:
//
//	QUICKFEED_SESSION_SECRET=hei go test -v -run TestSessionSecret
func TestSessionSecret(t *testing.T) {
	sessionSecret := os.Getenv(secretEnvName)
	if sessionSecret != "" {
		t.Fatalf("Unexpected access to %s=%s", secretEnvName, sessionSecret)
	}
}

func TestRedact(t *testing.T) {
	const secret = "quickfeed-session-secret"
	got := Redact("failure: "+secret+" repeated "+secret, secret)
	if want := "failure: [REDACTED] repeated [REDACTED]"; got != want {
		t.Errorf("Redact() = %q, want %q", got, want)
	}
	if got := Redact("nothing to redact", ""); got != "nothing to redact" {
		t.Errorf("Redact() with an empty secret = %q, want the output unchanged", got)
	}
}
