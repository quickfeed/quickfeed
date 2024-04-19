package env

import (
	"log"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

const (
	dotEnvPath = ".env"
)

var quickfeedRoot string

func init() {
	quickfeedRoot = os.Getenv("QUICKFEED")
}

// Root returns the root directory as defined by $QUICKFEED or
// sets it relative to the quickfeed module's root.
func Root() string {
	if quickfeedRoot != "" {
		return quickfeedRoot
	}
	setRoot()
	return quickfeedRoot
}

func setRoot() {
	root, err := gitRoot()
	if err != nil {
		log.Println("When running outside QuickFeed's git repository you must set the QUICKFEED environment variable manually.")
		log.Fatalf("Failed to set QUICKFEED variable: %v", err)
	}
	os.Setenv("QUICKFEED", root)
	quickfeedRoot = root
}

// RootEnv returns the path $QUICKFEED/{envFile}.
func RootEnv(envFile string) string {
	return filepath.Join(Root(), envFile)
}

// PublicEnv returns the path $QUICKFEED/public/{envFile}.
func PublicEnv(envFile string) string {
	return filepath.Join(Root(), "public", envFile)
}

// TestdataPath returns the path to the testdata/courses directory.
func TestdataPath() string {
	return filepath.Join(Root(), "testdata", "courses")
}

// gitRoot return the root of the Git repository.
func gitRoot() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// PlainOpen opens a git repository from the given path and searches upwards.
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return "", err
	}
	w, err := repo.Worktree()
	if err != nil {
		return "", err
	}
	return w.Filesystem.Root(), nil
}
