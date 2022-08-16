package env

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
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

// Load loads environment variables from the given file, or from $QUICKFEED/.env.
// The variable's values are expanded with existing variables from the environment.
// It will not override a variable that already exists in the environment.
func Load(filename string) error {
	file, err := open(filename, false)
	if err != nil {
		return err
	}
	defer file.Close()

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

func open(filename string, write bool) (*os.File, error) {
	if filename == "" {
		filename = filepath.Join(quickfeedRoot, dotEnvPath)
		log.Printf("Loading environment variables from %s", filename)
	}
	if write {
		return os.OpenFile(filename, os.O_APPEND, 0644)
	}
	return os.Open(filename)
}

func Save(filename, key, id string) {
	file, err := open(filename, true)
	if err != nil {
		log.Fatalf("Failed to open %s: %v", filename, err)
	}
	defer file.Close()

	// Replace lines with new values.
	content, err := ioutil.ReadFile(file.Name())
	if err != nil {
		log.Fatalf("Failed to read %s: %v", file.Name(), err)
	}
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "QUICKFEED_APP_ID") || strings.HasPrefix(line, "QUICKFEED_APP_KEY") {
			lines[i] = ""
		}
	}
	lines = append(lines, fmt.Sprintf("QUICKFEED_APP_ID=%s", id))
	lines = append(lines, fmt.Sprintf("QUICKFEED_APP_KEY=%s", filepath.Join(quickfeedRoot, defaultKeyPath)))
	if _, err := os.Stat(defaultKeyPath); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(defaultKeyPath), 0755); err != nil {
			log.Fatalf("Failed to create directory for %s: %v", defaultKeyPath, err)
		}
	}
	if err := os.WriteFile(defaultKeyPath, []byte(key), 0644); err != nil {
		log.Fatalf("Failed to write %s: %v", filepath.Join(quickfeedRoot, defaultKeyPath), err)
	}
	if err := os.WriteFile(".env", []byte(strings.Join(lines, "\n")), 0644); err != nil {
		log.Fatalf("Failed to write %s: %v", file.Name(), err)
	}
}

func ignore(line string) bool {
	trimmedLine := strings.TrimSpace(line)
	return len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "#")
}
