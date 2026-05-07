package ci

import (
	"os"
	"path/filepath"
)

// moduleCachePath returns quickfeed's go module cache path on the host.
// If the directory does not exist, it will be created.
func moduleCachePath() (string, error) {
	return hostCacheDir(GoModCache)
}

// goCachePath returns quickfeed's Go build cache path on the host.
// If the directory does not exist, it will be created.
func goCachePath() (string, error) {
	return hostCacheDir(GoCache)
}

// golangciLintCachePath returns quickfeed's golangci-lint cache path on the host.
// If the directory does not exist, it will be created.
func golangciLintCachePath() (string, error) {
	return hostCacheDir(GolangciLintCache)
}

// hostCacheDir returns a cache directory under $HOME with the given name,
// creating it if necessary. Directories are created owned by the current user,
// which ensures containers running as that user can read and write them.
func hostCacheDir(name string) (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(homedir, name)
	if ok, _ := exists(dir); !ok {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
	}
	return dir, nil
}
