package manifest_test

import (
	"os"
	"testing"

	"github.com/quickfeed/quickfeed/web/manifest"
)

func TestManifest(t *testing.T) {
	if os.Getenv("GITHUB_APP") == "" {
		t.Skip("Skipping test; GITHUB_APP is not set")
	}
	manifest.SetupGitHubApp(t)
}
