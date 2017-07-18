package yamlparser

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type NewAssignmentRequest struct {
	AssignmentID uint   `yaml: "assignmentid"`
	Name         string `yaml: "name"`
	Language     string `yaml: "language"`
	CourseCode   string `yaml:"coursecode"`
	Deadline     string `yaml:"deadline"`
	Autoapprove  bool   `yaml: "autoapprove"`
}

func Parse(dir string) ([]NewAssignmentRequest, error) {
	// check if directory exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, err
	}

	var assignments []NewAssignmentRequest
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			filename := filepath.Base(path)
			if filename == "assignment.yml" {
				var assignment NewAssignmentRequest
				source, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}

				err = yaml.Unmarshal(source, &assignment)
				if err != nil {
					return err
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
