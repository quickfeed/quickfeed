package env

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
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
func Root(paths ...string) string {
	if quickfeedRoot == "" {
		setRoot()
	}
	return filepath.Join(quickfeedRoot, filepath.Join(paths...))
}

func setRoot() {
	root, err := moduleRoot()
	if err != nil {
		// When the working directory is outside the repository, we must set the QUICKFEED env variable.
		wd, _ := os.Getwd()
		fmt.Printf("Working directory (%s) may be outside quickfeed's repository.\n", wd)
		fmt.Println("Please set the QUICKFEED environment variable to the root of the repository.")
		panic(fmt.Sprintf("Failed to determine root of the repository: %v", err))
	}
	os.Setenv("QUICKFEED", root)
	quickfeedRoot = root
}

// moduleRoot returns the root of the quickfeed module, that is, the closest
// directory at or above the working directory holding a go.mod file that
// declares the quickfeed module path.
//
// The search is lexical; it does not resolve symlinks. This keeps the returned
// path spelled the same way as the working directory, so that paths derived from
// it remain comparable to the paths reported by the go tool and by the shell.
func moduleRoot() (string, error) {
	// os.Getwd prefers $PWD, and thus keeps any symlinks it contains.
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := wd; ; {
		if err := checkModulePath(dir); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod file declaring module %s found in or above %s", quickfeedModulePath, wd)
		}
		dir = parent
	}
}

// checkModulePath checks that the given directory contains a go.mod file
// with the correct module path for QuickFeed.
func checkModulePath(dir string) error {
	modFile := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(modFile)
	if err != nil {
		return fmt.Errorf("reading %s: %w", modFile, err)
	}
	// The module path must match exactly; a prefix match would also accept
	// nested modules, such as the kit module.
	if path := modfile.ModulePath(data); path != quickfeedModulePath {
		return fmt.Errorf("invalid go.mod file: %s declares module %s", modFile, path)
	}
	return nil
}

// RootEnv returns the path $QUICKFEED/{envFile}.
func RootEnv(envFile string) string {
	return Root(envFile)
}

// PublicEnv returns the path $QUICKFEED/public/{envFile}.
func PublicEnv(envFile string) string {
	return Root("public", envFile)
}

// PublicDir returns the path to the public directory.
func PublicDir() string {
	return Root("public")
}

// DatabasePath returns the path to the database file.
func DatabasePath() string {
	return Root("qf.db")
}

// TestdataPath returns the path to the testdata/courses directory.
func TestdataPath() string {
	return Root("testdata", "courses")
}
