package yamlparser

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/autograde/aguis/ag"

	"gopkg.in/yaml.v2"
)

const target = "assignment.yml"

// NewAssignmentRequest represents a request for a new assignment.
type NewAssignmentRequest struct {
	AssignmentID uint   `yaml:"assignmentid"`
	Name         string `yaml:"name"`
	Language     string `yaml:"language"`
	Deadline     string `yaml:"deadline"`
	AutoApprove  bool   `yaml:"autoapprove"`
	IsGroupLab   bool   `yaml:"isgrouplab"`
}

// Parse recursively walks the given directory and parses any yaml files found
// and returns an array of assignments.
func Parse(dir string, courseID uint64) ([]*pb.Assignment, error) {
	// check if directory exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, err
	}

	var assignments []*pb.Assignment
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			filename := filepath.Base(path)
			if filename == target {
				var tempAssignment *pb.Assignment
				source, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				err = yaml.Unmarshal(source, &tempAssignment)
				if err != nil {
					return err
				}

				// convert to lowercase to normalize language name
				tempAssignment.Language = strings.ToLower(tempAssignment.Language)
				// parsing is not required as we use Timestamps
				/*deadline, err := time.Parse("02-01-2006 15:04", v.Deadline)
				if err != nil {
					return err
				}*/
				tempAssignment.CourseId = courseID

				/*
					assignment := &pb.Assignment{
						ID:          tempAssignment.Id,
						CourseId:    courseID,
						Deadline:    deadline,
						Language:    v.Language,
						Name:        v.Name,
						Order:       v.AssignmentID,
						AutoApprove: v.AutoApprove,
						IsGroupLab:  v.IsGroupLab,
					}*/

				assignments = append(assignments, tempAssignment)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return assignments, nil
}
