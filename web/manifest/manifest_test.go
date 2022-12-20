package manifest_test

import (
	"os"
	"testing"

	"github.com/quickfeed/quickfeed/web/manifest"
)

func TestCreateQuickFeedApp(t *testing.T) {
	if os.Getenv("GITHUB_APP") == "" {
		t.Skipf("Skipping test. To run: GITHUB_APP=1 go test -v -run %s", t.Name())
	}
	manifest.CreateQuickFeedApp(t)
}
