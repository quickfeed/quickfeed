package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func renameLegacyTests() {
	if err := renameTestFiles(gitRoot); err != nil {
		exitErr(err, "Error renaming legacy test files")
	}
}

// renameTestFiles renames all legacy *_ag_test.go files to *_qf_test.go in path.
func renameTestFiles(path string) error {
	return filepath.WalkDir(path, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", filePath, err)
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), "_ag_test.go") {
			newName := strings.Replace(d.Name(), "_ag_test.go", "_qf_test.go", 1)
			newPath := filepath.Join(filepath.Dir(filePath), newName)
			if err := os.Rename(filePath, newPath); err != nil {
				return fmt.Errorf("failed to rename file %s to %s: %w", filePath, newPath, err)
			}
			fmt.Printf("Renamed: %s -> %s\n", filePath, newPath)
		}
		return nil
	})
}
