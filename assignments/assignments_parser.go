package assignments

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	pb "github.com/autograde/quickfeed/ag"

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
	AssignmentID     uint   `yaml:"assignmentid"`
	ScriptFile       string `yaml:"scriptfile"`
	Deadline         string `yaml:"deadline"`
	AutoApprove      bool   `yaml:"autoapprove"`
	ScoreLimit       uint   `yaml:"scorelimit"`
	IsGroupLab       bool   `yaml:"isgrouplab"`
	Reviewers        uint   `yaml:"reviewers"`
	ContainerTimeout uint   `yaml:"containertimeout"`
	SkipTests        bool   `yaml:"skiptests"`
}

// ParseAssignments recursively walks the given directory and parses
// any 'assignment.yml' files found and returns an array of assignments.
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
				var newAssignment assignmentData
				err = yaml.Unmarshal(source, &newAssignment)
				if err != nil {
					return fmt.Errorf("error unmarshalling assignment: %w", err)
				}
				// if no auto approve score limit is defined; use the default
				if newAssignment.ScoreLimit < 1 {
					newAssignment.ScoreLimit = defaultAutoApproveScoreLimit
				}
				if newAssignment.ScriptFile == "" {
					return fmt.Errorf("error unmarshalling assignment: missing field 'scriptfile'")
				}

				// AssignmentID field from the parsed yaml is used to set Order, not assignment ID,
				// or it will cause a database constraint violation (IDs must be unique)
				// The Name field below is the folder name of the assignment.
				assignment := &pb.Assignment{
					CourseID:         courseID,
					Deadline:         FixDeadline(newAssignment.Deadline),
					ScriptFile:       strings.ToLower(newAssignment.ScriptFile),
					Name:             filepath.Base(filepath.Dir(path)),
					Order:            uint32(newAssignment.AssignmentID),
					AutoApprove:      newAssignment.AutoApprove,
					ScoreLimit:       uint32(newAssignment.ScoreLimit),
					IsGroupLab:       newAssignment.IsGroupLab,
					Reviewers:        uint32(newAssignment.Reviewers),
					ContainerTimeout: uint32(newAssignment.ContainerTimeout),
					SkipTests:        newAssignment.SkipTests,
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

func FixDeadline(in string) string {
	wantLayout := "2006-01-02T15:04:05"
	acceptedLayouts := []string{
		"2006-1-2T15:04:05",
		"2006-1-2 15:04:05",
		"2006-1-2T15:04",
		"2006-1-2 15:04",
		"2006-1-2T1504",
		"2006-1-2 1504",
		"2006-1-2T15",
		"2006-1-2 15",
		"2006-1-2 3pm",
		"2006-1-2 3:04pm",
		"2006-1-2 3:04:05pm",
		"2-1-2006T15:04:05",
		"2-1-2006 15:04:05",
		"2-1-2006T15:04",
		"2-1-2006 15:04",
		"2-1-2006T1504",
		"2-1-2006 1504",
		"2-1-2006T15",
		"2-1-2006 15",
		"2-1-2006 3pm",
		"2-1-2006 3:04pm",
		"2-1-2006 3:04:05pm",
	}
	for _, layout := range acceptedLayouts {
		t, err := time.Parse(layout, in)
		if err != nil {
			continue
		}
		return t.Format(wantLayout)
	}
	return "Invalid date format: " + in
}
