package env

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Prepared returns nil if the given env file exists and the corresponding backup file does not.
// Otherwise, it returns an error.
//
// If the env file does not exist, a MissingError returned.
// QuickFeed requires the env file to load (some) existing environment variables,
// even when creating a new GitHub app, and overwriting some environment variables.
// If the backup file exists, an ExistsError is returned.
// This is to avoid that QuickFeed overwrites an existing backup file.
func Prepared(filename string) error {
	bakFilename := filename + ".bak"
	if exists(bakFilename) {
		return ExistsError(bakFilename)
	}
	if !exists(filename) {
		return MissingError(filename)
	}
	return nil
}

// SetupEnvFiles creates the .env files if they do not exist.
func SetupEnvFiles(envFile string, dev bool) error {
	for i, envFile := range []string{RootEnv(envFile), PublicEnv(envFile)} {
		if !exists(envFile) {
			env, err := os.Create(envFile)
			if err != nil {
				return err
			}
			defer env.Close()

			template := ".env-template"
			if dev && i == 0 {
				template = fmt.Sprintf("%s-dev", template)
			}

			dir := filepath.Dir(envFile)
			envTemplate, err := os.Open(filepath.Join(dir, template))
			if err != nil {
				return err
			}
			_, err = io.Copy(env, envTemplate)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Save writes the given environment variables to the given file,
// replacing or leaving behind existing variables.
//
// If the file exists, it will be updated, but leaving a backup file.
// If a backup already exists it will be removed first.
func Save(filename string, env map[string]string) error {
	// Load the existing file's content before renaming it.
	content := load(filename)
	bakFilename := filename + ".bak"
	if exists(bakFilename) {
		if err := os.Remove(bakFilename); err != nil {
			return err
		}
	}
	if err := os.Rename(filename, bakFilename); err != nil {
		return err
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
		if _, err = fmt.Fprintf(file, "%s=%s\n", key, val); err != nil {
			return err
		}
	}
	return file.Close()
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

type missingEnvError struct {
	filename string
}

func MissingError(filename string) error {
	return missingEnvError{filename: filename}
}

func (e missingEnvError) Error() string {
	return fmt.Sprintf("missing required %q file", e.filename)
}
