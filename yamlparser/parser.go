package yamlparser

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	pb "github.com/autograde/aguis/ag"
	tspb "github.com/golang/protobuf/ptypes"

	"gopkg.in/yaml.v2"
)

const target = "assignment.yml"

// assignmentData holds information about a single assignment.
// This is only used for parsing the 'assignment.yml' file.
// Note that the struct can be private, but the fields must be
// public to allow parsing.
type assignmentData struct {
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
				var newAssignment assignmentData
				source, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				err = yaml.Unmarshal(source, &newAssignment)
				if err != nil {
					return err
				}

				// we need to parse the deadline in two stages;
				// first regular Go time.Time and then protobuf timestamp
				d, err := time.Parse("02-01-2006 15:04", newAssignment.Deadline)
				if err != nil {
					return err
				}
				deadline, err := tspb.TimestampProto(d)
				if err != nil {
					return err
				}

				assignment := &pb.Assignment{
					Id:          uint64(newAssignment.AssignmentID),
					CourseId:    courseID,
					Deadline:    deadline,
					Language:    strings.ToLower(newAssignment.Language),
					Name:        newAssignment.Name,
					Order:       uint32(newAssignment.AssignmentID),
					AutoApprove: newAssignment.AutoApprove,
					IsGrouplab:  newAssignment.IsGroupLab,
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
