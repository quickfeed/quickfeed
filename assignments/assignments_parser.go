package assignments

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/autograde/aguis/ag"

	"gopkg.in/yaml.v2"
)

const (
	target                       = "assignment.yml"
	targetYaml                   = "assignment.yaml"
	defaultAutoApproveScoreLimit = 80
)

// assignmentData holds information about a single assignment.
// This is only used for parsing the 'assignment.yml' file.
// Note that the struct can be private, but the fields must be
// public to allow parsing.
type assignmentData struct {
	AssignmentID uint   `yaml:"assignmentid"`
	Language     string `yaml:"language"`
	Deadline     string `yaml:"deadline"`
	AutoApprove  bool   `yaml:"autoapprove"`
	ScoreLimit   uint   `yaml:"scorelimit"`
	IsGroupLab   bool   `yaml:"isgrouplab"`
}

// ParseAssignments recursively walks the given directory and parses
// any 'assignment.yml' files found and returns an array of assignments.
// TODO(meling) Thinking: One complication with this approach is that we depend on the YAML's 'name' field
//   being the same as the assignment name in the folder structure in the assignments repository.
//   This is perhaps fine, but could be problematic if someone uses a name like "Lab assignment 1"
//   and the folder is named only "lab1". We should make this more robust; can we add a field to the
//   pb.Assignment type to hold the directory name, which should not be parsed from YAML, but computed
//   in assignment_parser.go, based on parent directory of the YAML. Issue is that we may need to add it to the DB.
func parseAssignments(dir string, courseID uint64) ([]*pb.Assignment, error) {
	// check if directory exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, err
	}

	var assignments []*pb.Assignment
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			filename := filepath.Base(path)
			if filename == target || filename == targetYaml {
				source, err := ioutil.ReadFile(path)
				if err != nil {
					return fmt.Errorf("could not to read %q file: %w", filename, err)
				}
				labDirectoryName := filepath.Base(filepath.Dir(path))
				var newAssignment assignmentData
				err = yaml.Unmarshal(source, &newAssignment)
				if err != nil {
					return fmt.Errorf("error unmarshalling assignment: %w", err)
				}
				// if no auto approve score limit is defined; use the default
				if newAssignment.ScoreLimit < 1 {
					newAssignment.ScoreLimit = defaultAutoApproveScoreLimit
				}

				// ID field from the parsed yaml is used to set Order, not assignment ID,
				// or it will cause a database constraint violation (IDs must be unique)
				assignment := &pb.Assignment{
					CourseID:    courseID,
					Deadline:    newAssignment.Deadline,
					Language:    strings.ToLower(newAssignment.Language),
					Name:        labDirectoryName,
					Order:       uint32(newAssignment.AssignmentID),
					AutoApprove: newAssignment.AutoApprove,
					ScoreLimit:  uint32(newAssignment.ScoreLimit),
					IsGroupLab:  newAssignment.IsGroupLab,
				}
				assignments = append(assignments, assignment)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return assignments, nil
}
