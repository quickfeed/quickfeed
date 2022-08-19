package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Save writes the given environment variables to the given file,
// replacing or leaving behind existing variables.
//
// If the file exists, it will be moved to a backup file.
// If the a backup file also exists, an error is returned.
func Save(filename string, env map[string]string) error {
	bakFilename := filename + ".bak"
	var (
		filenameExists    = exists(filename)
		bakFilenameExists = exists(bakFilename)
	)

	// We must load the file before renaming it below.
	content := load(filename)

	switch {
	case filenameExists && !bakFilenameExists:
		if err := os.Rename(filename, bakFilename); err != nil {
			return err
		}
	case filenameExists && bakFilenameExists:
		// Only report the base part of the backup file's name.
		// This may change in the future.
		return ExistsError(filepath.Base(bakFilename))
	}

	// Update the file with new environment variables.
	return update(filename, content, env)
}

// load reads the content of the given file assuming it exists.
// An empty string is returned if the file does not exist.
func load(filename string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}
	return string(content)
}

// update updates the file's content with the provided environment variables.
func update(filename, content string, env map[string]string) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()

	// Map of updated environment variables
	updated := make(map[string]bool)

	// Scan existing file's content and update existing environment variables.
	for _, line := range strings.Split(content, "\n") {
		key, val, found := strings.Cut(line, "=")
		if !found {
			// Leave non-environment and blank lines unchanged.
			fmt.Fprintln(file, line)
			continue
		}
		// Remove spaces around the key and value, if any.
		key, val = strings.TrimSpace(key), strings.TrimSpace(val)
		if v, ok := env[key]; ok {
			// Replace old value with new value.
			val = v
		}
		fmt.Fprintf(file, "%s=%s\n", key, val)
		updated[key] = true
	}

	// Write new lines for any new environment variables.
	for key, val := range env {
		if _, ok := updated[key]; ok {
			continue
		}
		fmt.Fprintf(file, "%s=%s\n", key, val)
	}
	return nil
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

type backupExistsError struct {
	filename string
}

func ExistsError(filename string) error {
	return backupExistsError{filename: filename}
}

func (e backupExistsError) Error() string {
	return fmt.Sprintf("%s already exists; check its content before removing and try again", e.filename)
}
