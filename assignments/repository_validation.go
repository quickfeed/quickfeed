package assignments

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
)

// ValidateCourseRepositories checks the tests repository's content and verifies
// that the tests and assignments repositories contain the same assignment folders.
// Problems are written to the course log and returned as a count; content problems
// do not make validation fail.
func ValidateCourseRepositories(ctx context.Context, testsDir, assignmentsDir string, courseID uint64) (int, error) {
	assignments, _, testsIssues, err := readTestsRepositoryContent(testsDir, courseID)
	if err != nil {
		return 0, fmt.Errorf("reading tests repository content: %w", err)
	}
	logRepositoryIssues(ctx, "tests repository issue", testsIssues)

	repositoryIssues, err := courseRepositoryIssues(testsDir, assignmentsDir, assignments)
	if err != nil {
		return len(testsIssues), err
	}
	logRepositoryIssues(ctx, "course repository issue", repositoryIssues)
	return len(testsIssues) + len(repositoryIssues), nil
}

func courseRepositoryIssues(testsDir, assignmentsDir string, parsedAssignments []*qf.Assignment) ([]RepoIssue, error) {
	testsFolders, err := topLevelFolders(testsDir)
	if err != nil {
		return nil, fmt.Errorf("reading tests repository folders: %w", err)
	}
	assignmentsFolders, err := topLevelFolders(assignmentsDir)
	if err != nil {
		return nil, fmt.Errorf("reading assignments repository folders: %w", err)
	}

	configuredTestsFolders, err := configuredAssignmentFolders(testsDir, testsFolders)
	if err != nil {
		return nil, err
	}
	// A folder without configuration is still an assignment folder when its
	// counterpart in the assignments repository establishes that name.
	for name := range assignmentsFolders {
		if testsFolders[name] {
			configuredTestsFolders[name] = true
		}
	}

	parsed := make(map[string]*qf.Assignment, len(parsedAssignments))
	for _, assignment := range parsedAssignments {
		parsed[assignment.GetName()] = assignment
	}

	names := unionNames(configuredTestsFolders, assignmentsFolders)
	var issues []RepoIssue
	for _, name := range names {
		inTests := configuredTestsFolders[name]
		inAssignments := assignmentsFolders[name]
		switch {
		case !inTests:
			issues = append(issues, RepoIssue{
				Assignment: name,
				File:       name,
				Problem:    fmt.Sprintf("assignment folder is missing from %q repository", qf.TestsRepo),
			})
			continue
		case !inAssignments:
			issues = append(issues, RepoIssue{
				Assignment: name,
				File:       name,
				Problem:    fmt.Sprintf("assignment folder is missing from %q repository", qf.AssignmentsRepo),
			})
		}

		structureIssues, err := assignmentStructureIssues(testsDir, name, parsed[name])
		if err != nil {
			return nil, err
		}
		issues = append(issues, structureIssues...)
	}
	return issues, nil
}

func assignmentStructureIssues(testsDir, name string, assignment *qf.Assignment) ([]RepoIssue, error) {
	assignmentPath := filepath.Join(name, assignmentFile)
	hasAssignment, err := regularFile(filepath.Join(testsDir, assignmentPath))
	if err != nil {
		return nil, err
	}
	hasTests, err := regularFile(filepath.Join(testsDir, name, testsFile))
	if err != nil {
		return nil, err
	}
	hasCriteria, err := regularFile(filepath.Join(testsDir, name, criteriaFile))
	if err != nil {
		return nil, err
	}

	var issues []RepoIssue
	// When tests.json or criteria.json exists, readTestsRepositoryContent already
	// reports the missing assignment file. Add it here only for otherwise empty
	// or unrecognized folders that its file-oriented walk cannot discover.
	if !hasAssignment && !hasTests && !hasCriteria {
		issues = append(issues, RepoIssue{
			Assignment: name,
			File:       assignmentPath,
			Problem:    fmt.Sprintf("missing %q", assignmentPath),
		})
	}
	// Auto-graded assignments get the more specific missing-tests issue from
	// missingTestsIssues. This check covers empty, broken, and manually graded
	// assignment folders for which neither scoring configuration exists.
	if !hasTests && !hasCriteria && (assignment == nil || assignment.GradedManually()) {
		issues = append(issues, RepoIssue{
			Assignment: name,
			File:       name,
			Problem:    fmt.Sprintf("missing %s or %s", testsFile, criteriaFile),
		})
	}
	return issues, nil
}

// topLevelFolders returns all possible assignment directories. Dot directories
// and the reserved scripts directory are repository metadata, not assignments.
func topLevelFolders(dir string) (map[string]bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	folders := make(map[string]bool)
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == scriptsDir || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		folders[entry.Name()] = true
	}
	return folders, nil
}

// configuredAssignmentFolders keeps only tests-side directories containing at
// least one assignment configuration file. Other top-level directories may hold
// shared packages and are not assignments unless the assignments repository has
// a folder with the same name.
func configuredAssignmentFolders(dir string, folders map[string]bool) (map[string]bool, error) {
	configured := make(map[string]bool)
	for name := range folders {
		for _, filename := range []string{assignmentFile, testsFile, criteriaFile} {
			exists, err := regularFile(filepath.Join(dir, name, filename))
			if err != nil {
				return nil, err
			}
			if exists {
				configured[name] = true
				break
			}
		}
	}
	return configured, nil
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

func logRepositoryIssues(ctx context.Context, message string, issues []RepoIssue) {
	logger := qlog.FromContext(ctx)
	for _, issue := range issues {
		logger.Warn(message,
			label.Assignment, issue.Assignment,
			label.Path, issue.File,
			"problem", issue.Problem,
		)
	}
}
