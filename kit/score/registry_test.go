package score

import (
	"fmt"
	"strings"
	"testing"

	"github.com/autograde/quickfeed/kit/score/testdata/a"
	"github.com/autograde/quickfeed/kit/sh"
	"github.com/google/go-cmp/cmp"
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
				}
			}()
			tName := testName(test.in)
			if len(test.want) > 0 {
				t.Errorf("testName('%s')='%s', expected to fail", test.inName, tName)
			}
		})
	}
}

func TestPrintTestInfoOrder(t *testing.T) {
	got, err := sh.Output("go test -run TestTestNamePanic")
	if err != nil {
		t.Fatal(err)
	}
	expectedPrefixOrder := `Registration order:
{"TestName":"TestPanicTriangularBefore","MaxScore":8,"Weight":5}
{"TestName":"TestPanicTriangularPanic","MaxScore":8,"Weight":5}
{"TestName":"TestPanicTriangularAfter","MaxScore":8,"Weight":5}
{"TestName":"TestPanicHandler","MaxScore":8,"Weight":5}
{"TestName":"TestPanicTriangularPanicWithMsg","MaxScore":8,"Weight":5}
{"TestName":"TestPanicHandlerWithMsg","MaxScore":8,"Weight":5}
Sorted order:
{"TestName":"TestPanicHandler","MaxScore":8,"Weight":5}
{"TestName":"TestPanicHandlerWithMsg","MaxScore":8,"Weight":5}
{"TestName":"TestPanicTriangularAfter","MaxScore":8,"Weight":5}
{"TestName":"TestPanicTriangularBefore","MaxScore":8,"Weight":5}
{"TestName":"TestPanicTriangularPanic","MaxScore":8,"Weight":5}
{"TestName":"TestPanicTriangularPanicWithMsg","MaxScore":8,"Weight":5}
`
	if diff := cmp.Diff(expectedPrefixOrder, got[:len(expectedPrefixOrder)]); diff != "" {
		t.Errorf("PrintTestInfo(): (-want +got):\n%s", diff)
	}
}
