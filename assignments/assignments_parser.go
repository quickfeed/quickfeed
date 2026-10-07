package assignments

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const defaultAutoApproveScoreLimit = 80

// assignmentData holds information about a single assignment.
// This is only used for parsing the 'assignment.json' file.
// Note that the struct can be private, but the fields must be
// public to allow parsing.
type assignmentData struct {
	Order            uint32  `json:"order"`
	Deadline         string  `json:"deadline"`
	IsGroupLab       bool    `json:"isgrouplab"`
	AutoApprove      bool    `json:"autoapprove"`
	ScoreLimit       uint32  `json:"scorelimit"`
	Reviewers        uint32  `json:"reviewers"`
	ReviewWeight     *uint32 `json:"reviewweight"`
	ContainerTimeout uint32  `json:"containertimeout"`
}

// newAssignmentFromFile returns the assignment described by contents, and
// whether contents leave out reviewweight; the default review weight depends
// on the assignment's tests, which are not known until tests.json is read.
func newAssignmentFromFile(contents []byte, assignmentName string, courseID uint64) (*qf.Assignment, bool, error) {
	var newAssignment assignmentData
	err := json.Unmarshal(contents, &newAssignment)
	if err != nil {
		return nil, false, fmt.Errorf("unmarshaling assignment: %w", err)
	}
	if newAssignment.Order < 1 {
		return nil, false, fmt.Errorf("assignment order must be greater than 0")
	}
	if newAssignment.ReviewWeight != nil && *newAssignment.ReviewWeight > 100 {
		return nil, false, fmt.Errorf("reviewweight must be at most 100, got %d", *newAssignment.ReviewWeight)
	}
	// if no auto approve score limit is defined; use the default
	if newAssignment.ScoreLimit < 1 {
		newAssignment.ScoreLimit = defaultAutoApproveScoreLimit
	}
	deadline, err := FixDeadline(newAssignment.Deadline)
	if err != nil {
		return nil, false, fmt.Errorf("parsing deadline: %w", err)
	}
	// AssignmentID field from the parsed json is used to set Order, not assignment ID,
	// or it will cause a database constraint violation (IDs must be unique)
	// The Name field below is the folder name of the assignment.
	assignment := &qf.Assignment{
		CourseID:         courseID,
		Deadline:         deadline,
		Name:             assignmentName,
		Order:            newAssignment.Order,
		IsGroupLab:       newAssignment.IsGroupLab,
		AutoApprove:      newAssignment.AutoApprove,
		ScoreLimit:       newAssignment.ScoreLimit,
		Reviewers:        newAssignment.Reviewers,
		ContainerTimeout: newAssignment.ContainerTimeout,
	}
	if newAssignment.ReviewWeight != nil {
		assignment.ReviewWeight = *newAssignment.ReviewWeight
	}
	return assignment, newAssignment.ReviewWeight == nil, nil
}

func FixDeadline(in string) (*timestamppb.Timestamp, error) {
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
		return timestamppb.New(t), nil
	}
	return nil, fmt.Errorf("invalid date format: %s", in)
}
