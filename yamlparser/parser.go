package yamlparser

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/autograde/aguis/models"
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
func Parse(dir string, courseID uint64) ([]*models.Assignment, error) {
	// check if directory exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, err
	}

	var assignments []*models.Assignment
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			filename := filepath.Base(path)
			if filename == target {
				var v NewAssignmentRequest
				source, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				err = yaml.Unmarshal(source, &v)
				if err != nil {
					return err
				}

				// convert to lowercase to normalize language name
				v.Language = strings.ToLower(v.Language)
				deadline, err := time.Parse("02-01-2006 15:04", v.Deadline)
				if err != nil {
					return err
				}

				assignment := &models.Assignment{
					ID:          uint64(v.AssignmentID),
					CourseID:    courseID,
					Deadline:    deadline,
					Language:    v.Language,
					Name:        v.Name,
					Order:       v.AssignmentID,
					AutoApprove: v.AutoApprove,
					IsGroupLab:  v.IsGroupLab,
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
