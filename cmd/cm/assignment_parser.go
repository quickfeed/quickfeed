package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	defaultGrading    = "Pass/fail"
	defaultAuto       = "Automatic"
	defaultManual     = "TA Approval"
	noApproval        = "No Approval"
	defaultIndividual = "Individually"
	defaultGroup      = "Group"
)

// AssignmentInfo contains fields present in the assignment.yml or assignment.json files which accompany each lab.
type AssignmentInfo struct {
	Order         int    `yaml:"order" json:"order"`
	Name          string `yaml:"name" json:"name"`
	ScriptFile    string `yaml:"scriptfile" json:"scriptfile"`
	Deadline      string `yaml:"deadline" json:"deadline"` // time.Time
	ShortDeadline string `yaml:"-" json:"-"`               // only date
	Year          string `yaml:"-" json:"-"`               // derived from deadline
	AutoApprove   bool   `yaml:"autoapprove" json:"autoapprove"`
	IsGroupLab    bool   `yaml:"isgrouplab" json:"isgrouplab"`
	ScoreLimit    int    `yaml:"scorelimit" json:"scorelimit"`
	Reviewers     int    `yaml:"reviewers" json:"reviewers"`
	Title         string `yaml:"title" json:"title"`
	HoursMin      int    `yaml:"hoursmin" json:"hoursmin"`
	HoursMax      int    `yaml:"hoursmax" json:"hoursmax"`
	// defaults to $COURSE $NAME
	Subject string `yaml:"subject" json:"subject"`
	// defaults to defaultGrading
	Grading string `yaml:"grading" json:"grading"`
	// SubmissionType is interpreted based on the value of IsGroupLab ("Individually" or "Group")
	SubmissionType string `yaml:"-" json:"-"`
	ApproveType    string `yaml:"-" json:"-"`
	CourseOrg      string `yaml:"-" json:"-"` // this field is populated after parsing by the main script
}

func parseAssignment(filename string) (*AssignmentInfo, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	res := &AssignmentInfo{}
	
	// Determine file format based on extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		err = json.Unmarshal(contents, res)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
	case ".yml", ".yaml":
		err = yaml.UnmarshalStrict(contents, res)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	if res.Subject == "" {
		res.Subject = fmt.Sprintf("%s %s", strings.ToUpper(course()), name())
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
