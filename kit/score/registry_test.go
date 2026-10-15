package score

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/kit/sh"
)

func TestPrintTestInfo(t *testing.T) {
	want := `{"TestName":"TestPanicTriangularBefore","MaxScore":8,"Weight":5}
{"TestName":"TestPanicTriangularPanic","MaxScore":8,"Weight":5}
{"TestName":"TestPanicTriangularAfter","MaxScore":8,"Weight":5}
{"TestName":"TestPanicHandler","MaxScore":8,"Weight":5}
{"TestName":"TestPanicTriangularPanicWithMsg","MaxScore":8,"Weight":5}
{"TestName":"TestPanicHandlerWithMsg","MaxScore":8,"Weight":5}
`
	got, err := sh.Output("env SCORE_INIT=1 go test -run TestTestNamePanic")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) < len(want) {
		t.Fatalf("PrintTestInfo() output too short: %q", got)
	}
	if diff := cmp.Diff(want, got[:len(want)]); diff != "" {
		t.Errorf("PrintTestInfo(): (-want +got):\n%s", diff)
	}
	got, err = sh.Output("go test -v -run TestTestNamePanic")
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(got, `{"TestName"`) {
		t.Errorf("PrintTestInfo() printed without SCORE_INIT: %q", got)
	}
}
