package ci

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// hasAssignment return nil if the assignment exists in the given clone dir.
func hasAssignment(cloneDir, currentAssignment string) error {
	// Check that there is code for the current assignment in clone dir
	if ok, err := exists(filepath.Join(cloneDir, currentAssignment)); !ok {
		return fmt.Errorf("directory %q does not contain %q: %w", cloneDir, currentAssignment, err)
	}
	return nil
}

// scanStudentRepo returns an error if the student's repository contains the session secret environment variable.
// Note: This scan may be costly for large repositories.
func scanStudentRepo(submittedDir, course, jobOwner string) error {
	defer timer(jobOwner, course, validationTimeGauge)()

	// Walk the student's submitted code directory
	files, err := walk(submittedDir)
	if err != nil {
		return err
	}
	// Ensure that the student code files does not contain the session secret environment variable.
	for file, content := range files {
		if strings.Contains(string(content), secretEnvName) {
			return fmt.Errorf("file %q in (%s/%s) contains the %q environment variable", filepath.Base(file), course, jobOwner, secretEnvName)
		}
		// We could add more checks here.
	}
	return nil
}

// exists returns true if path exists and is a directory.
// Otherwise, it returns false and an error.
func exists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil && fi.IsDir() {
		return true, nil
	}
	return false, err
}

// walk walks the student code repository and returns a map of file names and their contents.
func walk(dir string) (map[string][]byte, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, err
	}
	files := make(map[string][]byte)
	err := filepath.WalkDir(dir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			// Walk unable to read path; stop walking the tree
			return err
		}
		if !info.IsDir() {
			if files[path], err = os.ReadFile(path); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
