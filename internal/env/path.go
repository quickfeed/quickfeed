package env

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/quickfeed/quickfeed/kit/sh"
)

const (
	dotEnvPath          = ".env"
	quickfeedModulePath = "github.com/quickfeed/quickfeed"
)

var quickfeedRoot string

func init() {
	quickfeedRoot = os.Getenv("QUICKFEED")
}

// root returns the root directory as defined by $QUICKFEED or
// sets it relative to the quickfeed module's root.
func root() string {
	if quickfeedRoot != "" {
		return quickfeedRoot
	}

	out, err := sh.Output("go list -m -f {{.Dir}} " + quickfeedModulePath)
	if err != nil {
		log.Fatalf("Failed to set QUICKFEED variable: %v", err)
	}
	quickfeedRoot = strings.TrimSpace(out)
	os.Setenv("QUICKFEED", quickfeedRoot)

	return quickfeedRoot
}

// RootEnv returns the path $QUICKFEED/{envFile}.
func RootEnv(envFile string) string {
	return filepath.Join(root(), envFile)
}

// PublicEnv returns the path $QUICKFEED/public/{envFile}.
func PublicEnv(envFile string) string {
	return filepath.Join(root(), "public", envFile)
}

// TestdataPath returns the path to the testdata/courses directory.
func TestdataPath() string {
	return filepath.Join(root(), "testdata", "courses")
}
