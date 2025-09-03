package ci

import (
	"os"
	"path/filepath"
	"testing"
)

func TestModuleCachePath(t *testing.T) {
	// Uncomment the following line to run the test locally; do not commit the change
	t.Skip("Only for local testing; should not be run on quickfeed server")
	homedir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(filepath.Join(homedir, GoModCache))

	path, err := moduleCachePath()
	if err != nil {
		t.Error(err)
	}
	if path != filepath.Join(homedir, GoModCache) {
		t.Errorf("moduleCachePath() = %s, want %s", path, filepath.Join(homedir, GoModCache))
	}
	if ok, err := exists(path); !ok {
		t.Errorf("%s does not exist: %v", path, err)
	}
	t.Logf("path: %s", path)
}
