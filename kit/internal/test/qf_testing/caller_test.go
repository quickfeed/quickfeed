// Package qf_testing is only used for testing the robustness of the internal
// [test.CallerName] function to ensure that module or package names with
// a testing suffix are not accidentally matched (in [test.unwindCallFrames])
// before reaching the actual std lib testing package in the call stack.
// This was introduced as a fix for issue #1387.
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
