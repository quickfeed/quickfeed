package web_test

import (
	"testing"
)

func assertCode(t *testing.T, haveCode, wantCode int) {
	t.Helper()
	if haveCode != wantCode {
		t.Errorf("have status code %d want %d", haveCode, wantCode)
	}
}
