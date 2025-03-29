package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// updateDocTags replaces various tags in Markdown files in the given repo
// with values from the environment. The tags are defined in the replacements
// function.
//
// Run with the following command:
//
//	cm update-doc-tags -repo <assignments | tests | info>
func updateDocTags(args []string) {
	if err := loadEnv(); err != nil {
		exitErr(err, "Error loading environment variables")
	}

	fs := flag.NewFlagSet(updateDocTagsCmd, flag.ExitOnError)
	var repo string
	fs.StringVar(&repo, "repo", "", "Repository name (one of: assignments | tests | info)")

	if err := fs.Parse(args); err != nil {
		exitErr(err, "Error parsing flags")
	}
	path := repoPathExists(repo)
	fmt.Printf("Updating course tags in %s\n", courseRepoPath(path))
	if err := replaceTags(path); err != nil {
		exitErr(err, "Error replacing tags")
	}
	fmt.Printf("Replacements completed in %s\n", courseRepoPath(path))
}

func repoPath(repo string) string {
	return filepath.Join(filepath.Dir(gitRoot), year(), repo)
}

func repoPathExists(repo string) string {
	path := repoPath(repo)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		exitErr(err, "Repository does not exist")
	}
	return path
}

// replaceTags walks the repository path and updates Markdown files
// with the values from the environment.
func replaceTags(path string) error {
	// initialize the replacements to be made
	replacements := replacements()

	err := filepath.WalkDir(path, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(filePath, ".md") {
			return updateFile(filePath, replacements)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking directory: %w", err)
	}
	return nil
}

func updateFile(file string, replacements map[string]string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	updatedContent := string(content)
	fileChanged := false
	for key, val := range replacements {
		re := regexp.MustCompile(regexp.QuoteMeta(key))
		// check if 'key' exists in the file
		if re.MatchString(updatedContent) {
			fileChanged = true
			fmt.Printf("Replacing '%s' with '%s' in %s\n", key, val, lastDirFile(file))
		}
		updatedContent = re.ReplaceAllString(updatedContent, val)
	}

	if fileChanged {
		if err := os.WriteFile(file, []byte(updatedContent), 0o644); err != nil {
			return fmt.Errorf("failed to write file: %s, %w", file, err)
		}
	}
	return nil
}

// courseRepoPath returns the path relative to the course repository.
// If the course name is not found in the path, the full path is returned.
func courseRepoPath(path string) string {
	courseName := filepath.Base(gitRoot)
	cleanPath := filepath.Clean(path)
	pathElements := strings.Split(cleanPath, string(filepath.Separator))
	for i, element := range slices.Backward(pathElements) {
		if element == courseName {
			return filepath.Join(pathElements[i:]...)
		}
	}
	// return the full path if course name is not found
	return path
}
