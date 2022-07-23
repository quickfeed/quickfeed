package env

import (
	"bufio"
	"os"
	"strings"
)

var dotEnvPath = ".env"

// Load loads environment variables from .env in the current folder.
// Note that the variable's values are not expanded.
func Load(filename string) error {
	if filename == "" {
		filename = dotEnvPath
	}
	file, err := os.Open(filename)
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
		os.Setenv(k, strings.TrimSpace(val))
	}

	return scanner.Err()
}

func ignore(line string) bool {
	trimmedLine := strings.TrimSpace(line)
	return len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "#")
}
