package assignments

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/quickfeed/quickfeed/qf"
)

const (
	assignmentFile     = "assignment.yml"
	assignmentFileYaml = "assignment.yaml"
	criteriaFile       = "criteria.json"
	testsFile          = "tests.json"
	dockerfile         = "Dockerfile"
	taskFilePattern    = "task-*.md"
)

var patterns = []string{
	assignmentFile,
	assignmentFileYaml,
	criteriaFile,
	testsFile,
	dockerfile,
	taskFilePattern,
}

// matchAny returns true if filename matches one of the target patterns.
func matchAny(filename string) bool {
	for _, pattern := range patterns {
		if ok, _ := filepath.Match(pattern, filename); ok {
			return true
		}
	}
	return false
}

// match returns true if filename matches the given pattern.
func match(filename, pattern string) bool {
	if ok, _ := filepath.Match(pattern, filename); ok {
		return true
	}
	return false
}

var processors = map[string]fileProcessor{
	criteriaFile: processCriteriaFile,
	testsFile:    processTestsFile,
}

// fileProcessor processes specific file types and updates the assignment
type fileProcessor func(contents []byte, assignment *qf.Assignment, courseID uint64) error

// processCriteriaFile handles criteria.json files
func processCriteriaFile(contents []byte, assignment *qf.Assignment, courseID uint64) error {
	var benchmarks []*qf.GradingBenchmark
	if err := json.Unmarshal(contents, &benchmarks); err != nil {
		return fmt.Errorf("failed to unmarshal %q: %s", criteriaFile, err)
	}
	// Benchmarks and criteria must have courseID for access control checks
	for _, bm := range benchmarks {
		bm.CourseID = courseID
		for _, c := range bm.GetCriteria() {
			c.CourseID = courseID
		}
	}
	assignment.GradingBenchmarks = benchmarks
	return nil
}

// processTestsFile handles tests.json files
func processTestsFile(contents []byte, assignment *qf.Assignment, _ uint64) error {
	var expectedTests []*qf.TestInfo
	if err := json.Unmarshal(contents, &expectedTests); err != nil {
		return fmt.Errorf("failed to unmarshal %q: %s", testsFile, err)
	}
	assignment.ExpectedTests = expectedTests
	return nil
}

// processTaskFile handles task-*.md files
func processTaskFile(contents []byte, assignment *qf.Assignment, filename string) error {
	taskName := taskName(filename)
	task, err := newTask(contents, assignment.GetOrder(), taskName)
	if err != nil {
		return err
	}
	assignment.Tasks = append(assignment.GetTasks(), task)
	return nil
}

// readTestsRepositoryContent reads dir and returns a list of assignments and
// the course's Dockerfile content if there exists a 'tests/scripts/Dockerfile'.
// Assignments are extracted from 'assignment.yml' files, one for each assignment.
func readTestsRepositoryContent(dir string, courseID uint64) ([]*qf.Assignment, string, error) {
	files, err := walkTestsRepository(dir)
	if err != nil {
		return nil, "", err
	}

	// Process assignment files first
	assignmentsMap, err := processAssignmentFiles(files, courseID)
	if err != nil {
		return nil, "", err
	}

	var courseDockerfile string

	// Process other files in tests repository
	for path, contents := range files {
		filename := filepath.Base(path)

		// Handle Dockerfile separately since it's not assignment-specific
		if filename == dockerfile {
			courseDockerfile = string(contents)
			continue
		}

		assignmentName := filepath.Base(filepath.Dir(path))
		assignment := assignmentsMap[assignmentName]

		// Process known file types registered in processors map
		if processor, exists := processors[filename]; exists {
			if err := processor(contents, assignment, courseID); err != nil {
				return nil, "", err
			}
		}

		// Process task files
		if match(filename, taskFilePattern) {
			if err := processTaskFile(contents, assignment, filename); err != nil {
				return nil, "", err
			}
		}
	}
	return sortAssignments(assignmentsMap), courseDockerfile, nil
}

// walkTestsRepository walks the tests repository and returns a map of file names and their contents.
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
		if !info.IsDir() && matchAny(info.Name()) {
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

// processAssignmentFiles processes assignment.yml/yaml files and returns assignments map.
func processAssignmentFiles(files map[string][]byte, courseID uint64) (map[string]*qf.Assignment, error) {
	assignmentsMap := make(map[string]*qf.Assignment)
	for path, contents := range files {
		assignmentName := filepath.Base(filepath.Dir(path))
		filename := filepath.Base(path)
		if filename == assignmentFile || filename == assignmentFileYaml {
			assignment, err := newAssignmentFromFile(contents, assignmentName, courseID)
			if err != nil {
				return nil, err
			}
			assignmentsMap[assignmentName] = assignment
		}
	}
	return assignmentsMap, nil
}

// sortAssignments converts map to sorted slice and sorts tasks within assignments.
func sortAssignments(assignmentsMap map[string]*qf.Assignment) []*qf.Assignment {
	assignments := make([]*qf.Assignment, 0, len(assignmentsMap))
	for _, assignment := range assignmentsMap {
		assignments = append(assignments, assignment)
		sort.Slice(assignment.GetTasks(), func(i, j int) bool {
			return assignment.GetTasks()[i].GetTitle() < assignment.GetTasks()[j].GetTitle()
		})
	}
	sort.Slice(assignments, func(i, j int) bool {
		return assignments[i].GetOrder() < assignments[j].GetOrder()
	})
	return assignments
}
