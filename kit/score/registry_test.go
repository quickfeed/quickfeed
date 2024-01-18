package score

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/kit/sh"
)

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
