package assignments

import (
	"fmt"
	"time"

	pb "github.com/autograde/quickfeed/ag"

	"gopkg.in/yaml.v2"
)

const defaultAutoApproveScoreLimit = 80

// assignmentData holds information about a single assignment.
// This is only used for parsing the 'assignment.yml' file.
// Note that the struct can be private, but the fields must be
// public to allow parsing.
type assignmentData struct {
	AssignmentID     uint   `yaml:"assignmentid"` // deprecated: use Order instead
	Order            uint32 `yaml:"order"`
	Deadline         string `yaml:"deadline"`
	IsGroupLab       bool   `yaml:"isgrouplab"`
	AutoApprove      bool   `yaml:"autoapprove"`
	ScoreLimit       uint32 `yaml:"scorelimit"`
	Reviewers        uint32 `yaml:"reviewers"`
	ContainerTimeout uint32 `yaml:"containertimeout"`
}

func newAssignmentFromFile(contents []byte, assignmentName string, courseID uint64) (*pb.Assignment, error) {
	var newAssignment assignmentData
	err := yaml.Unmarshal(contents, &newAssignment)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling assignment: %w", err)
	}
	// if no auto approve score limit is defined; use the default
	if newAssignment.ScoreLimit < 1 {
		newAssignment.ScoreLimit = defaultAutoApproveScoreLimit
	}
	if newAssignment.AssignmentID > 0 && newAssignment.Order == 0 {
		newAssignment.Order = uint32(newAssignment.AssignmentID)
	}

	// AssignmentID field from the parsed yaml is used to set Order, not assignment ID,
	// or it will cause a database constraint violation (IDs must be unique)
	// The Name field below is the folder name of the assignment.
	assignment := &pb.Assignment{
		CourseID:         courseID,
		Deadline:         FixDeadline(newAssignment.Deadline),
		Name:             assignmentName,
		Order:            newAssignment.Order,
		IsGroupLab:       newAssignment.IsGroupLab,
		AutoApprove:      newAssignment.AutoApprove,
		ScoreLimit:       newAssignment.ScoreLimit,
		Reviewers:        newAssignment.Reviewers,
		ContainerTimeout: newAssignment.ContainerTimeout,
	}
	return assignment, nil
}

func FixDeadline(in string) string {
	wantLayout := pb.TimeLayout
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
