package assignments

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/qf"
)

const (
	assignmentFile = "assignment.json"
	criteriaFile   = "criteria.json"
	testsFile      = "tests.json"
	scriptsDir     = "scripts"
)

// RepoIssue describes a content problem detected in the tests repository,
// such as a malformed json file or a missing assignment.json. Issues are
// reported to the teaching staff via the course log and the UpdateAssignments
// RPC; they do not abort the assignment update.
type RepoIssue struct {
	Assignment string // assignment folder name; empty for repository-level issues
	File       string // repository-relative path, e.g., "lab1/tests.json"
	Problem    string
}

func (i RepoIssue) String() string {
	if i.File == "" {
		return i.Problem
	}
	return fmt.Sprintf("%s: %s", i.File, i.Problem)
}

// filesForBuildContext specifies files for the Docker build context.
// Add more files to support more dependencies for different courses.
var filesForBuildContext = map[string]bool{
	ci.Dockerfile: true,
	"go.mod":      true,
	"go.sum":      true,
}

var patterns = []string{
	assignmentFile,
	criteriaFile,
	testsFile,
	ci.Dockerfile,
}

// matchAny returns true if filename matches one of the target patterns
// or one of the files for build context.
func matchAny(filename string) bool {
	for _, pattern := range patterns {
		if ok, _ := filepath.Match(pattern, filename); ok {
			return true
		}
	}
	return filesForBuildContext[filename]
}

// match returns true if filename matches the given pattern.
func match(filename, pattern string) bool {
	if ok, _ := filepath.Match(pattern, filename); ok {
		return true
	}
	return false
}

// lookupFileProcessor returns the file processor for the given filename, if exists.
func lookupFileProcessor(filename string) (fileProcessor, bool) {
	for pattern, processor := range processors {
		if match(filename, pattern) {
			return processor, true
		}
	}
	return nil, false
}

var processors = map[string]fileProcessor{
	criteriaFile: processCriteriaFile,
	testsFile:    processTestsFile,
}

// fileProcessor processes specific file types and updates the assignment.
// It returns a list of entry-level problems for content that is usable but
// flawed, or an error if the file cannot be used at all.
type fileProcessor func(contents []byte, assignment *qf.Assignment, courseID uint64) ([]string, error)

// processCriteriaFile handles criteria.json files
func processCriteriaFile(contents []byte, assignment *qf.Assignment, courseID uint64) ([]string, error) {
	var benchmarks []*qf.GradingBenchmark
	if err := json.Unmarshal(contents, &benchmarks); err != nil {
		return nil, fmt.Errorf("unmarshaling %q: %w", criteriaFile, err)
	}
	var problems []string
	// Benchmarks and criteria must have courseID for access control checks
	for _, bm := range benchmarks {
		bm.CourseID = courseID
		if bm.GetHeading() == "" {
			problems = append(problems, "benchmark with empty heading")
		}
		for _, c := range bm.GetCriteria() {
			c.CourseID = courseID
			if c.GetDescription() == "" {
				problems = append(problems, fmt.Sprintf("criterion with empty description in benchmark %q", bm.GetHeading()))
			}
		}
	}
	assignment.GradingBenchmarks = benchmarks
	return problems, nil
}

// processTestsFile handles tests.json files
func processTestsFile(contents []byte, assignment *qf.Assignment, _ uint64) ([]string, error) {
	var expectedTests []*qf.TestInfo
	if err := json.Unmarshal(contents, &expectedTests); err != nil {
		return nil, fmt.Errorf("unmarshaling %q: %w", testsFile, err)
	}
	// Mirror the runtime rules of kit/score.Score.isValid: entries that would
	// be rejected when parsing test output are dropped here, with an issue
	// telling the teaching staff why.
	var problems []string
	validTests := make([]*qf.TestInfo, 0, len(expectedTests))
	seen := make(map[string]bool)
	for _, test := range expectedTests {
		switch {
		case test.GetTestName() == "":
			problems = append(problems, "test entry with empty test name")
		case seen[test.GetTestName()]:
			problems = append(problems, fmt.Sprintf("duplicate test name %q", test.GetTestName()))
		case test.GetMaxScore() < 1:
			problems = append(problems, fmt.Sprintf("test %q must have max score greater than 0", test.GetTestName()))
		case test.GetWeight() < 1:
			problems = append(problems, fmt.Sprintf("test %q must have weight greater than 0", test.GetTestName()))
		default:
			seen[test.GetTestName()] = true
			validTests = append(validTests, test)
		}
	}
	assignment.ExpectedTests = validTests
	return problems, nil
}

// readTestsRepositoryContent reads dir and returns a sorted list of assignments,
// a map with the docker build context as defined by the filesForBuildContext
// variable, and a list of content issues found in the repository.
// Assignments are extracted from 'assignment.json' files, one for each assignment.
// An assignment whose json files cannot be parsed is excluded from the returned
// list (and reported as an issue), so that a typo in a file is not interpreted
// as removal of the assignment's tests or criteria.
func readTestsRepositoryContent(dir string, courseID uint64) ([]*qf.Assignment, map[string]string, []RepoIssue, error) {
	files, err := walkTestsRepository(dir)
	if err != nil {
		return nil, nil, nil, err
	}

	// Process assignment files first
	assignmentsMap, broken, issues := processAssignmentFiles(files, courseID)

	buildContext := make(map[string]string)

	// Process other files in tests repository
	for _, path := range slices.Sorted(maps.Keys(files)) {
		contents := files[path]
		filename := filepath.Base(path)
		parts := strings.Split(filepath.ToSlash(path), "/")
		switch {
		case len(parts) == 1:
			// Repository root: only build context files are expected here.
			if filesForBuildContext[filename] {
				buildContext[filename] = string(contents)
			} else {
				issues = append(issues, RepoIssue{File: path, Problem: "file must be inside an assignment folder"})
			}

		case len(parts) == 2 && parts[0] == scriptsDir:
			// The scripts folder holds the Dockerfile and its build context.
			if filesForBuildContext[filename] {
				buildContext[filename] = string(contents)
			}

		case len(parts) == 2:
			issues = append(issues, processAssignmentFolderFile(path, contents, assignmentsMap, broken, courseID)...)

		default:
			// Files nested deeper than the assignment folders (internal packages,
			// testdata, and similar) are not assignment configuration; ignore.
		}
	}
	issues = append(issues, missingTestsIssues(assignmentsMap)...)
	return sortAssignments(assignmentsMap), buildContext, issues, nil
}

// processAssignmentFolderFile handles a file in an assignment folder, updating
// the folder's assignment via the file's processor. An assignment whose file
// cannot be used is removed from assignmentsMap and marked broken, so that a
// typo in the file is not interpreted as removal of its tests or criteria.
func processAssignmentFolderFile(path string, contents []byte, assignmentsMap map[string]*qf.Assignment, broken map[string]bool, courseID uint64) []RepoIssue {
	assignmentName := filepath.Dir(path)
	filename := filepath.Base(path)
	if filename == assignmentFile {
		return nil // already processed by processAssignmentFiles
	}
	if filename == ci.Dockerfile {
		return []RepoIssue{{Assignment: assignmentName, File: path,
			Problem: fmt.Sprintf("%s must be in the %s folder", ci.Dockerfile, scriptsDir)}}
	}
	processor, exists := lookupFileProcessor(filename)
	if !exists {
		return nil // other files in assignment folders (go.mod, go.sum) are ignored
	}
	if broken[assignmentName] {
		return nil // the folder's assignment.json failed to parse; already reported
	}
	assignment, exists := assignmentsMap[assignmentName]
	if !exists {
		return []RepoIssue{{Assignment: assignmentName, File: path,
			Problem: fmt.Sprintf("missing %q", filepath.Join(assignmentName, assignmentFile))}}
	}
	problems, err := processor(contents, assignment, courseID)
	if err != nil {
		broken[assignmentName] = true
		delete(assignmentsMap, assignmentName)
		return []RepoIssue{{Assignment: assignmentName, File: path, Problem: err.Error()}}
	}
	issues := make([]RepoIssue, 0, len(problems))
	for _, problem := range problems {
		issues = append(issues, RepoIssue{Assignment: assignmentName, File: path, Problem: problem})
	}
	return issues
}

// missingTestsIssues warns about auto-graded assignments without expected
// tests: such assignments record a zero score for every submission, since
// scores not listed in tests.json are discarded.
func missingTestsIssues(assignmentsMap map[string]*qf.Assignment) []RepoIssue {
	var issues []RepoIssue
	for _, name := range slices.Sorted(maps.Keys(assignmentsMap)) {
		assignment := assignmentsMap[name]
		if assignment.GradedManually() || len(assignment.GetGradingBenchmarks()) > 0 {
			continue
		}
		if len(assignment.GetExpectedTests()) == 0 {
			issues = append(issues, RepoIssue{Assignment: name, File: filepath.Join(name, testsFile),
				Problem: "missing or empty tests.json: all submissions will score zero"})
		}
	}
	return issues
}

// walkTestsRepository walks the tests repository and returns a map of
// repository-relative file names and their contents.
func walkTestsRepository(dir string) (map[string][]byte, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, err
	}
	files := make(map[string][]byte)
	err := filepath.WalkDir(dir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			// Walk unable to read path; stop walking the tree
			return err
		}
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		if !info.IsDir() && matchAny(info.Name()) {
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			if files[relPath], err = os.ReadFile(path); err != nil {
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

// processAssignmentFiles processes assignment.json files and returns the
// assignments map, the set of assignment folders whose assignment.json could
// not be parsed, and the issues found.
func processAssignmentFiles(files map[string][]byte, courseID uint64) (map[string]*qf.Assignment, map[string]bool, []RepoIssue) {
	assignmentsMap := make(map[string]*qf.Assignment)
	broken := make(map[string]bool)
	var issues []RepoIssue
	for _, path := range slices.Sorted(maps.Keys(files)) {
		if filepath.Base(path) != assignmentFile {
			continue
		}
		parts := strings.Split(filepath.ToSlash(path), "/")
		if len(parts) != 2 || parts[0] == scriptsDir {
			continue // handled by the main loop in readTestsRepositoryContent
		}
		assignmentName := parts[0]
		assignment, err := newAssignmentFromFile(files[path], assignmentName, courseID)
		if err != nil {
			issues = append(issues, RepoIssue{Assignment: assignmentName, File: path, Problem: err.Error()})
			broken[assignmentName] = true
			continue
		}
		assignmentsMap[assignmentName] = assignment
	}
	return assignmentsMap, broken, issues
}

// sortAssignments converts map to sorted slice.
func sortAssignments(assignmentsMap map[string]*qf.Assignment) []*qf.Assignment {
	assignments := make([]*qf.Assignment, 0, len(assignmentsMap))
	for _, assignment := range assignmentsMap {
		assignments = append(assignments, assignment)
	}
	sort.Slice(assignments, func(i, j int) bool {
		return assignments[i].GetOrder() < assignments[j].GetOrder()
	})
	return assignments
}
