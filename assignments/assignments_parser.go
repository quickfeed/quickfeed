package assignments

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	pb "github.com/autograde/quickfeed/ag"

	"gopkg.in/yaml.v2"
)

const (
	target                       = "assignment.yml"
	targetYaml                   = "assignment.yaml"
	criteriaFile                 = "criteria.json"
	scriptFile                   = "run.sh"
	scriptFolder                 = "scripts"
	dockerfile                   = "Dockerfile"
	defaultAutoApproveScoreLimit = 80
)

// assignmentData holds information about a single assignment.
// This is only used for parsing the 'assignment.yml' file.
// Note that the struct can be private, but the fields must be
// public to allow parsing.
type assignmentData struct {
	AssignmentID     uint   `yaml:"assignmentid"`
	Deadline         string `yaml:"deadline"`
	AutoApprove      bool   `yaml:"autoapprove"`
	ScoreLimit       uint   `yaml:"scorelimit"`
	IsGroupLab       bool   `yaml:"isgrouplab"`
	Reviewers        uint   `yaml:"reviewers"`
	ContainerTimeout uint   `yaml:"containertimeout"`
	SkipTests        bool   `yaml:"skiptests"`
}

// TODO(meling) this func should be renamed now that it does more than parseAssignments

// ParseAssignments recursively walks the given directory and parses
// any 'assignment.yml' files found and returns an array of assignments.
func parseAssignments(dir string, courseID uint64) ([]*pb.Assignment, string, error) {
	// check if directory exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, "", err
	}

	var assignments []*pb.Assignment
	var defaultScript string
	var courseDockerfile string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		assignmentName := filepath.Base(filepath.Dir(path))
		if !info.IsDir() {
			filename := filepath.Base(path)
			var contents []byte
			switch filename {
			case target, targetYaml, criteriaFile, scriptFile, dockerfile:
				contents, err = ioutil.ReadFile(path)
				if err != nil {
					return err
				}
			default:
				// no need to parse this file
				return nil
			}
			switch filename {
			case target, targetYaml:
				assignment, err := readAssignmentFile(contents, assignmentName, courseID)
				if err != nil {
					return err
				}
				assignments = append(assignments, assignment)

			case criteriaFile:
				if err := updateCriteriaFromFile(contents, assignmentName, assignments); err != nil {
					return err
				}

			case scriptFile:
				script, err := readScriptFile(contents, assignmentName, assignments)
				if err != nil {
					return err
				}
				defaultScript = script

			case dockerfile:
				courseDockerfile = string(contents)
			}
		}
		return nil
	})
	if err != nil {
		return nil, "", err
	}

	// if there is a script in `scripts` folder, save it for every assignment
	// that's missing the assignment specific script
	if defaultScript != "" {
		for _, assignment := range assignments {
			if assignment.ScriptFile == "" {
				assignment.ScriptFile = defaultScript
			}
		}
	}
	return assignments, courseDockerfile, nil
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

func updateCriteriaFromFile(criteria []byte, assignmentName string, assignments []*pb.Assignment) error {
	var benchmarks []*pb.GradingBenchmark
	if err := json.Unmarshal(criteria, &benchmarks); err != nil {
		return fmt.Errorf("could not unmarshal criteria.json: %s", err)
	}
	assignment := findAssignmentByName(assignments, assignmentName)
	if assignment == nil {
		return fmt.Errorf("could not find assignment %s for benchmark in %q", assignmentName, criteriaFile)
	}
	assignment.GradingBenchmarks = benchmarks
	return nil
}

func readScriptFile(contents []byte, assignmentName string, assignments []*pb.Assignment) (string, error) {
	if assignmentName != scriptFolder {
		assignment := findAssignmentByName(assignments, assignmentName)
		if assignment == nil {
			return "", fmt.Errorf("could not find assignment %s for script file", assignmentName)
		}
		assignment.ScriptFile = string(contents)
		return "", nil
	}
	return string(contents), nil
}

func readAssignmentFile(contents []byte, assignmentName string, courseID uint64) (*pb.Assignment, error) {
	var newAssignment assignmentData
	err := yaml.Unmarshal(contents, &newAssignment)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling assignment: %w", err)
	}
	// if no auto approve score limit is defined; use the default
	if newAssignment.ScoreLimit < 1 {
		newAssignment.ScoreLimit = defaultAutoApproveScoreLimit
	}

	// AssignmentID field from the parsed yaml is used to set Order, not assignment ID,
	// or it will cause a database constraint violation (IDs must be unique)
	// The Name field below is the folder name of the assignment.
	assignment := &pb.Assignment{
		CourseID:         courseID,
		Deadline:         FixDeadline(newAssignment.Deadline),
		Name:             assignmentName,
		Order:            uint32(newAssignment.AssignmentID),
		AutoApprove:      newAssignment.AutoApprove,
		ScoreLimit:       uint32(newAssignment.ScoreLimit),
		IsGroupLab:       newAssignment.IsGroupLab,
		Reviewers:        uint32(newAssignment.Reviewers),
		ContainerTimeout: uint32(newAssignment.ContainerTimeout),
	}
	return assignment, nil
}

func findAssignmentByName(assignments []*pb.Assignment, name string) *pb.Assignment {
	var found *pb.Assignment
	for _, assignment := range assignments {
		if assignment.Name == name {
			found = assignment
		}
	}
	return found
}
