package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// removeSolutionTags removes `//go:build !solution` tags in Go files in the given repo.
//
// Run with the following command:
//
//	cm remove-solution-tags -repo <assignments | tests>
func removeSolutionTags(args []string) {
	if err := loadEnv(); err != nil {
		exitErr(err, "Error loading environment variables")
	}

	fs := flag.NewFlagSet(removeSolutionTagsCmd, flag.ExitOnError)
	var repo string
	fs.StringVar(&repo, "repo", "", "Repository name (one of: assignments | tests)")

	if err := fs.Parse(args); err != nil {
		exitErr(err, "Error parsing flags")
	}
	path := repoPathExists(repo)
	fmt.Printf("Removing solution build tags in %s\n", courseRepoPath(path))
	if err := walkRepo(path); err != nil {
		exitErr(err, "Error removing build tags")
	}
	fmt.Printf("Remove completed in %s\n", courseRepoPath(path))
}

// walkRepo walks the repository path and removes solution build tags from Go files.
func walkRepo(path string) error {
	err := filepath.WalkDir(path, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".go") {
			return removeSolutionTag(filePath)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking through directory: %w", err)
	}
	return nil
}

var solutionTag = []byte("//go:build !solution\n")

func removeSolutionTag(file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	for line := range bytes.Lines(content) {
		if !bytes.HasPrefix(line, solutionTag) {
			return nil
		}
		break
	}
	fmt.Printf("Removing solution build tag from %s\n", lastDirFile(file))
	// remove the solution tag and write the updated content back to the file
	content = content[len(solutionTag)+1:] // skip the solution tag and the extra newline
	return os.WriteFile(file, content, 0o644)
}
