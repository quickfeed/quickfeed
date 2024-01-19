package test_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quickfeed/quickfeed/kit/internal/test"
	"github.com/quickfeed/quickfeed/kit/internal/test/testdata/a"
)

func fib() {}

func TestTestNamePanic(t *testing.T) {
	const (
		notFunc     = "not a function: "
		notTestFunc = "not a test function: "
		missingT    = "test function missing *testing.T argument: "
	)
	tests := []struct {
		typ    string
		inName string
		in     any
		want   string
	}{
		{"____NotFunc", "TestHurricane", "TestHurricane", notFunc + "TestHurricane"},
		{"NotTestFunc", "fib", fib, notTestFunc + "fib"},
		{"NotTestFunc", "TestNoArgs", a.TestNoArgs, notTestFunc + "TestNoArgs"},
		{"NotTestFunc", "NotATest", a.NotATest, notTestFunc + "NotATest"},
		{"NotTestFunc", "TesttooManyParams", a.TesttooManyParams, notTestFunc + "TesttooManyParams"},
		{"NotTestFunc", "TesttooManyNames", a.TesttooManyNames, notTestFunc + "TesttooManyNames"},
		{"NotTestFunc", "TestNoTParam", a.TestNoTParam, missingT + "TestNoTParam"},
		{"NotTestFunc", "TestnoTParam", a.TestnoTParam, missingT + "TestnoTParam"},
		{"___TestFunc", "TestFire", a.TestFire, ""},
	}
	for _, tt := range tests {
		t.Run(tt.typ+"/"+tt.inName, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					out := strings.TrimSpace(fmt.Sprintln(r))
					// ignore the file name and line number in the prefix of out
					if !strings.HasSuffix(out, tt.want) {
						t.Errorf("test.Name('%s')='%s', expected '%s'", tt.inName, out, tt.want)
					}
					if tt.want == "" {
						t.Errorf("test.Name('%s')='%s', not expected to fail", tt.inName, out)
					}
				}
			}()
			if len(tt.want) > 0 {
				t.Errorf("test.Name('%s')='%s', expected to fail", tt.inName, test.Name(tt.in))
			}
		})
	}
}

// isCallerShouldBeFalse returns false since IsCaller is not called from within TestIsCaller.
func isCallerShouldBeFalse() bool {
	return test.IsCaller("TestIsCaller")
}

func TestIsCaller(t *testing.T) {
	if isCallerShouldBeFalse() {
		t.Errorf("IsCaller incorrectly identified a non-test function as the calling function")
	}
	tt := []struct {
		funcName string
		want     bool
	}{
		{"TestIsCaller", true},
		{"TestTestNamePanic", false}, // TestTestNamePanic exists but is not the calling function
		{"TestNonExistentTest", false},
		{"NotATestFunc", false},
	}
	for _, tc := range tt {
		t.Run(tc.funcName, func(t *testing.T) {
			if test.IsCaller(tc.funcName) != tc.want {
				t.Errorf("IsCaller('%s')=%v, want %v", tc.funcName, !tc.want, tc.want)
			}
		})
	}
}

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
	tt := []struct {
		funcName string
		want     string
	}{
		{"SubtestCaller", "func1"},
	}
	for _, tc := range tt {
		t.Run(tc.funcName, func(t *testing.T) {
			if got := test.CallerName(); got != tc.want {
				t.Errorf("CallerName()=%s, want %s", got, tc.want)
			}
		})
	}
}
