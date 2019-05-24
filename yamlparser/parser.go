package yamlparser

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/models"
	tspb "github.com/golang/protobuf/ptypes"

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
				var tempAssignment *models.Assignment
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

				//deadline, err := time.Parse("02-01-2006 15:04", tempAssignment.Deadline)
				/*if err != nil {
					return err
				}*/
				tempAssignment.CourseID = courseID

				tstamp, err := tspb.TimestampProto(tempAssignment.Deadline)
				if err != nil {
					return status.Errorf(codes.Aborted, "cannot parse assignment deadline")
				}

				assignment := &pb.Assignment{
					Id:          tempAssignment.ID,
					CourseId:    courseID,
					Deadline:    tstamp,
					Language:    tempAssignment.Language,
					Name:        tempAssignment.Name,
					Order:       uint32(tempAssignment.ID),
					AutoApprove: tempAssignment.AutoApprove,
					IsGrouplab:  tempAssignment.IsGroupLab,
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
