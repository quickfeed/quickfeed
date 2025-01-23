package ci

import (
	"os"
	"path/filepath"
)

// moduleCachePath returns quickfeed's go module cache.
// If the directory does not exist, it will be created.
func moduleCachePath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	goModCacheSrc := filepath.Join(homedir, GoModCache)
	if ok, _ := exists(goModCacheSrc); !ok {
		if err := os.MkdirAll(goModCacheSrc, 0o755); err != nil {
			return "", err
		}
	}
	return goModCacheSrc, nil
}
