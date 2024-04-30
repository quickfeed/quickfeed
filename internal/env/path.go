package env

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

var quickfeedRoot string

const quickfeedModulePath = "github.com/quickfeed/quickfeed"

func init() {
	quickfeedRoot = os.Getenv("QUICKFEED")
}

// Root returns the root directory as defined by $QUICKFEED or
// sets it relative to the quickfeed module's root.
// This function will panic if called when the working directory
// is not within the quickfeed repository. In this case, the
// environment variable $QUICKFEED must be set manually.
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
		// When the working directory is outside the git repository, we must set the QUICKFEED env variable.
		wd, _ := os.Getwd()
		fmt.Printf("Working directory (%s) may be outside quickfeed's git repository.\n", wd)
		fmt.Println("Please set the QUICKFEED environment variable to the root of the repository.")
		panic(fmt.Sprintf("Failed to determine root of the git repository: %v", err))
	}
	if err := checkModulePath(root); err != nil {
		panic(fmt.Sprintf("Invalid module path: %v", err))
	}
	os.Setenv("QUICKFEED", root)
	quickfeedRoot = root
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

// checkModulePath checks that the root directory contains a go.mod file
// with the correct module path for QuickFeed.
func checkModulePath(root string) error {
	modFile := filepath.Join(root, "go.mod")
	data, err := os.ReadFile(modFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %v", modFile, err)
	}
	if !bytes.Contains(data, []byte("module "+quickfeedModulePath)) {
		return fmt.Errorf("invalid go.mod file: %s", modFile)
	}
	return nil
}

// RootEnv returns the path $QUICKFEED/{envFile}.
func RootEnv(envFile string) string {
	return filepath.Join(Root(), envFile)
}

// PublicEnv returns the path $QUICKFEED/public/{envFile}.
func PublicEnv(envFile string) string {
	return filepath.Join(Root(), "public", envFile)
}

// PublicDir returns the path to the public directory.
func PublicDir() string {
	return filepath.Join(Root(), "public")
}

// DatabasePath returns the path to the database file.
func DatabasePath() string {
	return filepath.Join(Root(), "qf.db")
}

// TestdataPath returns the path to the testdata/courses directory.
func TestdataPath() string {
	return filepath.Join(Root(), "testdata", "courses")
}
