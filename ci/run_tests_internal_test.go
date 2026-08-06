package ci

import "testing"

func TestRedactOutput(t *testing.T) {
	const secret = "quickfeed-session-secret"
	got := redactOutput("failure: "+secret+" repeated "+secret, secret)
	if want := "failure: [REDACTED] repeated [REDACTED]"; got != want {
		t.Errorf("redactOutput() = %q, want %q", got, want)
	}
}
