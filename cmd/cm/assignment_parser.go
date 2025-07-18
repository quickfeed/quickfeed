package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	defaultGrading    = "Pass/fail"
	defaultAuto       = "Automatic"
	defaultManual     = "TA Approval"
	noApproval        = "No Approval"
	defaultIndividual = "Individually"
	defaultGroup      = "Group"
)

// AssignmentInfo contains fields present in the assignment.json files which accompany each lab.
type AssignmentInfo struct {
	Order         int    `json:"order"`
	Name          string `json:"name"`
	ScriptFile    string `json:"scriptfile"`
	Deadline      string `json:"deadline"` // time.Time
	ShortDeadline string `json:"-"`        // only date
	Year          string `json:"-"`        // derived from deadline
	AutoApprove   bool   `json:"autoapprove"`
	IsGroupLab    bool   `json:"isgrouplab"`
	ScoreLimit    int    `json:"scorelimit"`
	Reviewers     int    `json:"reviewers"`
	Title         string `json:"title"`
	HoursMin      int    `json:"hoursmin"`
	HoursMax      int    `json:"hoursmax"`
	// defaults to $COURSE $NAME
	Subject string `json:"subject"`
	// defaults to defaultGrading
	Grading string `json:"grading"`
	// SubmissionType is interpreted based on the value of IsGroupLab ("Individually" or "Group")
	SubmissionType string `json:"-"`
	ApproveType    string `json:"-"`
	CourseOrg      string `json:"-"` // this field is populated after parsing by the main script
}

func parseAssignment(filename string) (*AssignmentInfo, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	res := &AssignmentInfo{}
	
	// Parse JSON assignment file
	err = json.Unmarshal(contents, res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if res.Subject == "" {
		res.Subject = fmt.Sprintf("%s %s", course(), name())
	}

	if res.AutoApprove {
		res.ApproveType = defaultAuto
	} else {
		res.ApproveType = defaultManual
	}

	switch res.Grading {
	case "":
		res.Grading = defaultGrading
	case "No grading":
		res.ApproveType = noApproval
	}

	if res.IsGroupLab {
		res.SubmissionType = defaultGroup
	} else {
		res.SubmissionType = defaultIndividual
	}

	// copied from qf.TimeLayout
	const layout = "2006-01-02T15:04:05"
	deadline, err := time.Parse(layout, res.Deadline)
	if err == nil {
		res.Deadline = fmt.Sprintf("%s %d, %d %02d:%02d",
			deadline.Month(), deadline.Day(), deadline.Year(), deadline.Hour(), deadline.Minute())
		res.ShortDeadline = fmt.Sprintf("%s %d", deadline.Month(), deadline.Day())
		res.Year = fmt.Sprintf("%d", deadline.Year())
		if res.Year != year() {
			fmt.Printf("Warning(%s): deadline (%s) does not match environment variable $YEAR=%s\n", res.Name, res.Deadline, year())
		}
	} else {
		fmt.Printf("Warning(%s): failed to parse deadline: %s\n", res.Name, err)
	}
	return res, nil
}
