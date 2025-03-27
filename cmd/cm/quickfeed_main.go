package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func addMainTests() {
	if err := walkDir(gitRoot); err != nil {
		exitErr(err, "Error walking directory")
	}
}

//go:embed main_qf_test.tmpl
var mainTmplFS embed.FS

const (
	mainTmplFile = "main_qf_test.tmpl"
	mainFile     = "main_qf_test.go"
	qfTestSuffix = "_qf_test.go"
)

// walkDir walks the given path and checks for _qf_test.go files and
// adds or removes the main_qf_test.go file accordingly.
func walkDir(path string) error {
	dirsWithQFTests := make(map[string][]string)
	err := filepath.WalkDir(path, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), qfTestSuffix) {
			dir := filepath.Dir(filePath)
			dirsWithQFTests[dir] = append(dirsWithQFTests[dir], filepath.Base(filePath))
		}
		return nil
	})
	if err != nil {
		return err
	}

	for dir, files := range dirsWithQFTests {
		if len(files) > 0 {
			if len(files) == 1 && slices.Contains(files, mainFile) {
				// remove main_qf_test.go if there is no actual _qf_test.go files
				if err := removeMainTest(dir); err != nil {
					exitErr(err, "Error removing QuickFeed test")
				}
				continue
			}
			// update main_qf_test.go if there are at least one actual _qf_test.go files
			if err := updateMainTest(dir); err != nil {
				exitErr(err, "Error updating QuickFeed test")
			}
		}
	}
	return nil
}

func removeMainTest(dir string) error {
	fmt.Printf("Removing %q from %s\n", mainFile, courseRepoPath(dir))
	return os.Remove(filepath.Join(dir, mainFile))
}

func updateMainTest(dir string) error {
	fmt.Printf("Updating %q in %s\n", mainFile, courseRepoPath(dir))
	return generateGoFromTemplate(dir, mainFile, mainTmplFile, mainTmplFS)
}
