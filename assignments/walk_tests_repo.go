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
	scriptFile         = "run.sh"
	scriptFolder       = "scripts"
	dockerfile         = "Dockerfile"
	taskFilePattern    = "task-*.md"
)

var patterns = []string{
	assignmentFile,
	assignmentFileYaml,
	criteriaFile,
	scriptFile,
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

// readTestsRepositoryContent reads dir and returns a list of assignments and
// the course's Dockerfile content if there exists a 'tests/scripts/Dockerfile'.
// Assignments are extracted from 'assignment.yml' files, one for each assignment.
func readTestsRepositoryContent(dir string, courseID uint64) ([]*qf.Assignment, string, error) {
	files, err := walkTestsRepository(dir)
	if err != nil {
		return nil, "", err
	}

	// Process all assignment.yml files first
	assignmentsMap := make(map[string]*qf.Assignment)
	for path, contents := range files {
		assignmentName := filepath.Base(filepath.Dir(path))
		switch filepath.Base(path) {
		case assignmentFile, assignmentFileYaml:
			assignment, err := newAssignmentFromFile(contents, assignmentName, courseID)
			if err != nil {
				return nil, "", err
			}
			assignmentsMap[assignmentName] = assignment
		}
	}

	var defaultScriptContent string
	var courseDockerfile string

	// Process other files in tests repository
	for path, contents := range files {
		assignmentName := filepath.Base(filepath.Dir(path))
		filename := filepath.Base(path)

		switch filename {
		case criteriaFile:
			var benchmarks []*qf.GradingBenchmark
			if err := json.Unmarshal(contents, &benchmarks); err != nil {
				return nil, "", fmt.Errorf("failed to unmarshal %q: %s", criteriaFile, err)
			}
			// Benchmarks and criteria must have courseID
			// for access control checks.
			for _, bm := range benchmarks {
				bm.CourseID = courseID
				for _, c := range bm.Criteria {
					c.CourseID = courseID
				}
			}
			assignmentsMap[assignmentName].GradingBenchmarks = benchmarks

		case scriptFile:
			if assignmentName != scriptFolder {
				// Found assignment-specific run script
				assignmentsMap[assignmentName].RunScriptContent = string(contents)
			} else {
				defaultScriptContent = string(contents)
			}

		case dockerfile:
			courseDockerfile = string(contents)
		}

		if match(filename, taskFilePattern) {
			assignment := assignmentsMap[assignmentName]
			taskName := taskName(filename)
			task, err := newTask(contents, assignment.GetOrder(), taskName)
			if err != nil {
				return nil, "", err
			}
			assignmentsMap[assignmentName].Tasks = append(assignmentsMap[assignmentName].Tasks, task)
		}
	}

	// If there is a run.sh script in the scripts folder, save it for each
	// assignment that is missing an assignment-specific script.
	if defaultScriptContent != "" {
		for _, assignment := range assignmentsMap {
			if assignment.RunScriptContent == "" {
				assignment.RunScriptContent = defaultScriptContent
			}
		}
	}

	assignments := make([]*qf.Assignment, 0)
	for _, assignment := range assignmentsMap {
		assignments = append(assignments, assignment)
		sort.Slice(assignment.Tasks, func(i, j int) bool {
			return assignment.Tasks[i].Title < assignment.Tasks[j].Title
		})
	}
	sort.Slice(assignments, func(i, j int) bool {
		return assignments[i].Order < assignments[j].Order
	})

	return assignments, courseDockerfile, nil
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
