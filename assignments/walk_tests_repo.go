package assignments

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	pb "github.com/autograde/quickfeed/ag"
)

const (
	assignmentFile     = "assignment.yml"
	assignmentFileYaml = "assignment.yaml"
	criteriaFile       = "criteria.json"
	scriptTemplateFile = "run.sh"
	scriptFolder       = "scripts"
	dockerfile         = "Dockerfile"
	taskFilePattern    = "task-*.md"
)

var patterns = []string{
	assignmentFile,
	assignmentFileYaml,
	criteriaFile,
	scriptTemplateFile,
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

// readTestsRepositoryContent reads dir and returns a list of assignments and the course's Dockerfile.
func readTestsRepositoryContent(dir string, courseID uint64) ([]*pb.Assignment, string, error) {
	files, err := walkTestsRepository(dir)
	if err != nil {
		return nil, "", err
	}

	// Process all assignment.yml files first
	assignmentsMap := make(map[string]*pb.Assignment)
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

	var defaultScriptTemplate string
	var courseDockerfile string

	// Process other files in tests repository
	for path, contents := range files {
		assignmentName := filepath.Base(filepath.Dir(path))

		switch filepath.Base(path) {
		case criteriaFile:
			var benchmarks []*pb.GradingBenchmark
			if err := json.Unmarshal(contents, &benchmarks); err != nil {
				return nil, "", fmt.Errorf("failed to unmarshal %q: %s", criteriaFile, err)
			}
			assignmentsMap[assignmentName].GradingBenchmarks = benchmarks

		case scriptTemplateFile:
			if assignmentName != scriptFolder {
				// Found assignment-specific script template
				assignmentsMap[assignmentName].ScriptFile = string(contents)
			} else {
				defaultScriptTemplate = string(contents)
			}

		case dockerfile:
			courseDockerfile = string(contents)
		}

		if match(filepath.Base(path), taskFilePattern) {
			assignment := assignmentsMap[assignmentName]
			task, err := newTask(contents, assignment)
			if err != nil {
				return nil, "", err
			}
			assignmentsMap[assignmentName].Tasks = append(assignmentsMap[assignmentName].Tasks, task)
		}
	}

	// If there is a run.sh script template in the scripts folder, save it
	// for each assignment that is missing an assignment-specific script template.
	if defaultScriptTemplate != "" {
		for _, assignment := range assignmentsMap {
			if assignment.ScriptFile == "" {
				assignment.ScriptFile = defaultScriptTemplate
			}
		}
	}

	assignments := make([]*pb.Assignment, 0)
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
