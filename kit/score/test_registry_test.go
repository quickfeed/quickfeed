package score

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/autograde/quickfeed/kit/score/testdata/a"
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
		in     interface{}
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
	for _, test := range tests {
		t.Run(test.typ+"/"+test.inName, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					out := strings.TrimSpace(fmt.Sprintln(r))
					// ignore the file name and line number in the prefix of out
					if !strings.HasSuffix(out, test.want) {
						t.Errorf("testName('%s')='%s', expected '%s'", test.inName, out, test.want)
					}
					if len(test.want) == 0 {
						t.Errorf("testName('%s')='%s', not expected to fail", test.inName, out)
					}
					t.Logf("recovered %v", r)
				}
			}()
			tName := testName(test.in)
			if len(test.want) > 0 {
				t.Errorf("testName('%s')='%s', expected to fail", test.inName, tName)
			}
		})
	}
}
