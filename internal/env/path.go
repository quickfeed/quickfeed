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
	if quickfeedRoot == "" {
		out, err := sh.Output("go list -m -f {{.Dir}} " + quickfeedModulePath)
		if err != nil {
			log.Fatalf("Failed to set QUICKFEED variable: %v", err)
		}
		quickfeedRoot = strings.TrimSpace(out)
		os.Setenv("QUICKFEED", quickfeedRoot)
	}
}

// Root returns the root directory as defined by $QUICKFEED.
func Root() string {
	return quickfeedRoot
}

// RootEnv returns the path $QUICKFEED/{envFile}.
func RootEnv(envFile string) string {
	return filepath.Join(quickfeedRoot, envFile)
}

// PublicEnv returns the path $QUICKFEED/public/{envFile}.
func PublicEnv(envFile string) string {
	return filepath.Join(quickfeedRoot, "public", envFile)
}
