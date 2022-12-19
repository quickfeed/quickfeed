package manifest_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/web/manifest"
)

func TestManifest(t *testing.T) {
	manifest.SetupGitHubApp(t)
}
