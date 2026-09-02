package assignments

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
)

// ignoreFile is an optional file in the tests repository root listing top-level
// folders, in either repository, that are not assignments.
const ignoreFile = ".quickfeedignore"

// courseRepositoryIssues reports assignment folders that are not aligned across
// the two course repositories. Every top-level folder in either repository is an
// assignment candidate, except repository metadata, the reserved scripts folder,
// and the folders listed in the tests repository's ignore file.
func courseRepositoryIssues(testsDir, assignmentsDir string, parsedAssignments []*qf.Assignment) ([]RepoIssue, error) {
	ignored, issues, err := readIgnoreFile(testsDir)
	if err != nil {
		return nil, err
	}
	testsFolders, err := topLevelFolders(testsDir, ignored)
	if err != nil {
		return nil, fmt.Errorf("reading tests repository folders: %w", err)
	}
	assignmentsFolders, err := topLevelFolders(assignmentsDir, ignored)
	if err != nil {
		return nil, fmt.Errorf("reading assignments repository folders: %w", err)
	}

	parsed := make(map[string]*qf.Assignment, len(parsedAssignments))
	for _, assignment := range parsedAssignments {
		parsed[assignment.GetName()] = assignment
	}

	for _, name := range unionNames(testsFolders, assignmentsFolders) {
		folderIssues, err := assignmentFolderIssues(testsDir, name, testsFolders[name], assignmentsFolders[name], parsed[name])
		if err != nil {
			return nil, err
		}
		issues = append(issues, folderIssues...)
	}
	return issues, nil
}

// assignmentFolderIssues reports the problems found for a single candidate
// assignment folder.
//
// A folder holding no assignment configuration at all is reported once, and
// nothing further is checked for it: QuickFeed cannot tell an assignment whose
// assignment.json was forgotten from shared course code that belongs in the
// tests repository. That single issue names both remedies, so an unconfigured
// folder costs one line whichever it turns out to be. The remaining checks
// apply to folders that are known to be assignments.
func assignmentFolderIssues(testsDir, name string, inTests, inAssignments bool, assignment *qf.Assignment) ([]RepoIssue, error) {
	if !inTests {
		return []RepoIssue{{
			Assignment: name,
			File:       name,
			Problem: fmt.Sprintf("assignment folder is missing from %q repository; add it or list the folder in %s",
				qf.TestsRepo, ignoreFile),
			Transient: true,
		}}, nil
	}
	files, err := readAssignmentFiles(testsDir, name)
	if err != nil {
		return nil, err
	}
	if !files.configured() {
		// readTestsRepositoryContent cannot discover this folder, since its
		// file-oriented walk only sees the configuration files that are absent here.
		return []RepoIssue{{
			Assignment: name,
			File:       name,
			Problem: fmt.Sprintf("no assignment configuration found; add %q if this is an assignment, or list the folder in %s",
				filepath.Join(name, assignmentFile), ignoreFile),
		}}, nil
	}

	var issues []RepoIssue
	if !inAssignments {
		issues = append(issues, RepoIssue{
			Assignment: name,
			File:       name,
			Problem:    fmt.Sprintf("assignment folder is missing from %q repository", qf.AssignmentsRepo),
			Transient:  true,
		})
	}
	// Auto-graded assignments get the more specific missing-tests issue from
	// missingTestsIssues. This check covers broken and manually graded
	// assignment folders for which neither scoring configuration exists.
	if !files.tests && !files.criteria && (assignment == nil || assignment.GradedManually()) {
		issues = append(issues, RepoIssue{
			Assignment: name,
			File:       name,
			Problem:    fmt.Sprintf("missing %s or %s", testsFile, criteriaFile),
		})
	}
	return issues, nil
}

// assignmentFiles records which configuration files an assignment folder in the
// tests repository contains.
type assignmentFiles struct {
	assignment bool
	tests      bool
	criteria   bool
}

// configured reports whether the folder holds any assignment configuration, and
// is therefore intended as an assignment rather than shared course code.
func (f assignmentFiles) configured() bool {
	return f.assignment || f.tests || f.criteria
}

func readAssignmentFiles(testsDir, name string) (assignmentFiles, error) {
	var files assignmentFiles
	for _, entry := range []struct {
		filename string
		found    *bool
	}{
		{assignmentFile, &files.assignment},
		{testsFile, &files.tests},
		{criteriaFile, &files.criteria},
	} {
		exists, err := regularFile(filepath.Join(testsDir, name, entry.filename))
		if err != nil {
			return files, err
		}
		*entry.found = exists
	}
	return files, nil
}

// topLevelFolders returns the assignment candidates in dir. Dot directories and
// the reserved scripts directory are repository metadata, not assignments, and
// so are the folders named by the course's ignore file.
func topLevelFolders(dir string, ignored map[string]bool) (map[string]bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	folders := make(map[string]bool)
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || name == scriptsDir || ignored[name] || strings.HasPrefix(name, ".") {
			continue
		}
		folders[name] = true
	}
	return folders, nil
}

// readIgnoreFile reads the optional ignore file from the tests repository root.
// Each line names one top-level folder, in either repository, that is not an
// assignment; blank lines and lines starting with '#' are skipped. Only exact
// folder names are supported; path and glob entries are reported and ignored,
// so that a misunderstood entry does not silently hide an assignment folder.
func readIgnoreFile(testsDir string) (map[string]bool, []RepoIssue, error) {
	contents, err := os.ReadFile(filepath.Join(testsDir, ignoreFile))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("reading %q: %w", ignoreFile, err)
	}
	ignored := make(map[string]bool)
	var issues []RepoIssue
	for line := range strings.Lines(string(contents)) {
		entry := strings.TrimSpace(line)
		if entry == "" || strings.HasPrefix(entry, "#") {
			continue
		}
		if strings.ContainsAny(entry, `/\*?[`) {
			issues = append(issues, RepoIssue{
				File:    ignoreFile,
				Problem: fmt.Sprintf("invalid entry %q: only top-level folder names are supported", entry),
			})
			continue
		}
		ignored[entry] = true
	}
	return ignored, issues, nil
}

func unionNames(left, right map[string]bool) []string {
	names := make([]string, 0, len(left)+len(right))
	for name := range left {
		names = append(names, name)
	}
	for name := range right {
		if !left[name] {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

func regularFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return info.Mode().IsRegular(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("checking %q: %w", path, err)
}

// logRepositoryIssues writes issues to the course log. Most issues are warnings,
// but transient issues are logged at info level because they only reflect the
// temporary mismatch between the two pushes needed to add an assignment.
// The frontend issue count reports the state after the latest push.
func logRepositoryIssues(ctx context.Context, message string, issues []RepoIssue) {
	logger := qlog.FromContext(ctx)
	for _, issue := range issues {
		level := slog.LevelWarn
		if issue.Transient {
			level = slog.LevelInfo
		}
		logger.Log(
			ctx, level,
			message,
			label.Assignment, issue.Assignment,
			label.Path, issue.File,
			"problem", issue.Problem,
		)
	}
}
