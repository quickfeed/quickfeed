package env

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/quickfeed/quickfeed/kit/sh"
)

const dotEnvPath = ".env"

var quickfeedRoot string

func init() {
	quickfeedRoot = os.Getenv("QUICKFEED")
	if quickfeedRoot == "" {
		out, err := sh.Output("go list -m -f {{.Dir}}")
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

// Load loads environment variables from the given file, or from $QUICKFEED/.env.
// The variable's values are expanded with existing variables from the environment.
// It will not override a variable that already exists in the environment.
func Load(filename string) error {
	if filename == "" {
		filename = filepath.Join(quickfeedRoot, dotEnvPath)
	}
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close() // skipcq: GO-S2307

	log.Printf("Loading environment variables from %s", filename)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if ignore(line) {
			continue
		}
		key, val, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		k := strings.TrimSpace(key)
		if os.Getenv(k) != "" {
			// Ignore .env entries already set in the environment.
			continue
		}
		val = os.ExpandEnv(strings.Trim(strings.TrimSpace(val), `"`))
		os.Setenv(k, val)
	}

	return scanner.Err()
}

func ignore(line string) bool {
	trimmedLine := strings.TrimSpace(line)
	return len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "#")
}
