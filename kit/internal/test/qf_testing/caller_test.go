package qf_testing

import (
	"testing"

	"github.com/quickfeed/quickfeed/kit/internal/test"
)

func callerName() string {
	return test.CallerName()
}

func TestCallerName(t *testing.T) {
	if callerName() != "callerName" {
		t.Errorf("callerName()=%s, want callerName", callerName())
	}
	if test.CallerName() != "TestCallerName" {
		t.Errorf("CallerName()=%s, want TestCallerName", test.CallerName())
	}
}
