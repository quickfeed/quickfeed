package assignments

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/ci"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestParseWithInvalidDir(t *testing.T) {
	const dir = "invalid/dir"
	_, _, err := parseAssignments(dir, 0)
	if err == nil {
		t.Errorf("want no such file or directory error, got nil")
	}
}

const (
	y1 = `assignmentid: 1
name: "For loops"
deadline: "27-08-2017 12:00"
autoapprove: false
`
	y2 = `assignmentid: 2
name: "Nested loops"
deadline: "27-08-2018 12:00"
autoapprove: false
`

	yUnknownFields = `assignmentid: 1
subject: "Go Programming for Fun and Profit"
name: "For loops"
deadline: "27-08-2017 12:00"
grading: "Pass/Fail"
expected_effort: "10 hours"
autoapprove: false
`

	script   = `Default script`
	script1  = `Script for Lab1`
	df       = `A dockerfile in training`
	criteria = `
	[
		{
			"heading": "First benchmark",
			"criteria": [
				{
					"description": "Test 1",
					"points": 5
				},
				{
					"description": "Test 2",
					"points": 10
				}
			]
		},
		{
			"heading": "Second benchmark",
			"criteria": [
				{
					"description": "Test 3",
					"points": 5
				}
			]
		}
	]`
)

func TestParse(t *testing.T) {
	testsDir, err := ioutil.TempDir("", pb.TestsRepo)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testsDir)

	job := &ci.Job{
		Commands: []string{
			"cd " + testsDir,
			"mkdir lab1",
			"mkdir lab2",
			"mkdir scripts",
		},
	}
	runner := ci.Local{}
	_, err = runner.Run(context.Background(), job)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(testsDir, "lab1", "assignment.yaml"), []byte(y1), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(testsDir, "lab2", "assignment.yaml"), []byte(y2), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(testsDir, "scripts", "run.sh"), []byte(script), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(testsDir, "lab1", "run.sh"), []byte(script1), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(testsDir, "scripts", "Dockerfile"), []byte(df), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(testsDir, "lab2", "criteria.json"), []byte(criteria), 0644)
	if err != nil {
		t.Fatal(err)
	}

	wantCriteria := []*pb.GradingBenchmark{{
		Heading: "First benchmark",
		Criteria: []*pb.GradingCriterion{
			{
				Description: "Test 1",
				Points:      5,
			},
			{
				Description: "Test 2",
				Points:      10,
			},
		},
	},
		{
			Heading: "Second benchmark",
			Criteria: []*pb.GradingCriterion{
				{
					Description: "Test 3",
					Points:      5,
				},
			},
		},
	}

	// We expect assignment names to be set based on
	// assignment folder names.
	wantAssignment1 := &pb.Assignment{
		Name:        "lab1",
		Deadline:    "2017-08-27T12:00:00",
		ScriptFile:  "Script for Lab1",
		AutoApprove: false,
		Order:       1,
		ScoreLimit:  80,
	}

	wantAssignment2 := &pb.Assignment{
		Name:              "lab2",
		Deadline:          "2018-08-27T12:00:00",
		ScriptFile:        "Default script",
		AutoApprove:       false,
		Order:             2,
		ScoreLimit:        80,
		GradingBenchmarks: wantCriteria,
	}

	assignments, dockerfile, err := parseAssignments(testsDir, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(assignments) != 2 {
		t.Errorf("len(assignments) = %d, want %d", len(assignments), 2)
	}
	if dockerfile != df {
		t.Errorf("Incorrect dockerfile\n Want: %s\n Got: %s\n", df, dockerfile)
	}
	if diff := cmp.Diff(assignments[0], wantAssignment1, cmpopts.IgnoreUnexported(pb.Assignment{})); diff != "" {
		t.Errorf("parseAssignments() mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(assignments[1], wantAssignment2, protocmp.Transform()); diff != "" {
		t.Errorf("parseAssignments() mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(assignments[1].GradingBenchmarks, wantCriteria, protocmp.Transform()); diff != "" {
		t.Errorf("parseAssignments() mismatch when parsing criteria (-want +got):\n%s", diff)
	}
}

func TestParseUnknownFields(t *testing.T) {
	testsDir, err := ioutil.TempDir("", pb.TestsRepo)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testsDir)

	job := &ci.Job{
		Commands: []string{
			"cd " + testsDir,
			"mkdir lab1",
		},
	}
	runner := ci.Local{}
	_, err = runner.Run(context.Background(), job)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(testsDir, "lab1", "assignment.yaml"), []byte(yUnknownFields), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// We expect assignment names to be set based on
	// assignment folder names.
	wantAssignment1 := &pb.Assignment{
		Name:        "lab1",
		Deadline:    "2017-08-27T12:00:00",
		AutoApprove: false,
		Order:       1,
		ScoreLimit:  80,
	}

	assignments, _, err := parseAssignments(testsDir, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(assignments) != 1 {
		t.Errorf("len(assignments) = %d, want %d", len(assignments), 1)
	}
	if diff := cmp.Diff(assignments[0], wantAssignment1, cmpopts.IgnoreUnexported(pb.Assignment{})); diff != "" {
		t.Errorf("parseAssignments() mismatch (-want +got):\n%s", diff)
	}
}

func TestFixDeadline(t *testing.T) {
	deadlineTests := []struct {
		in, want string
	}{
		{"2020-01-23T18:00:20", "2020-01-23T18:00:20"},
		{"2020-01-23 18:00:20", "2020-01-23T18:00:20"},
		{"2020-01-23T18:00", "2020-01-23T18:00:00"},
		{"2020-01-23 18:00", "2020-01-23T18:00:00"},
		{"2020-01-23T1800", "2020-01-23T18:00:00"},
		{"2020-01-23 1800", "2020-01-23T18:00:00"},
		{"2020-01-23T18", "2020-01-23T18:00:00"},
		{"2020-01-23 18", "2020-01-23T18:00:00"},
		{"2020-01-23 6pm", "2020-01-23T18:00:00"},
		{"2020-01-23 6am", "2020-01-23T06:00:00"},
		//
		{"2020-1-23T18:00:20", "2020-01-23T18:00:20"},
		{"2020-1-23 18:00:20", "2020-01-23T18:00:20"},
		{"2020-1-23T18:00", "2020-01-23T18:00:00"},
		{"2020-1-23 18:00", "2020-01-23T18:00:00"},
		{"2020-1-23T1800", "2020-01-23T18:00:00"},
		{"2020-1-23 1800", "2020-01-23T18:00:00"},
		{"2020-1-23T18", "2020-01-23T18:00:00"},
		{"2020-1-23 18", "2020-01-23T18:00:00"},
		{"2020-1-23 6pm", "2020-01-23T18:00:00"},
		{"2020-1-23 6am", "2020-01-23T06:00:00"},
		//
		{"2020-1-1T18:00:20", "2020-01-01T18:00:20"},
		{"2020-1-1 18:00:20", "2020-01-01T18:00:20"},
		{"2020-1-1T18:00", "2020-01-01T18:00:00"},
		{"2020-1-1 18:00", "2020-01-01T18:00:00"},
		{"2020-1-1T1800", "2020-01-01T18:00:00"},
		{"2020-1-1 1800", "2020-01-01T18:00:00"},
		{"2020-1-1T18", "2020-01-01T18:00:00"},
		{"2020-1-1 18", "2020-01-01T18:00:00"},
		{"2020-1-1 6pm", "2020-01-01T18:00:00"},
		{"2020-1-1 6am", "2020-01-01T06:00:00"},
		//
		{"23-01-2020T18:00:20", "2020-01-23T18:00:20"},
		{"23-01-2020 18:00:20", "2020-01-23T18:00:20"},
		{"23-01-2020T18:00", "2020-01-23T18:00:00"},
		{"23-01-2020 18:00", "2020-01-23T18:00:00"},
		{"23-01-2020T1800", "2020-01-23T18:00:00"},
		{"23-01-2020 1800", "2020-01-23T18:00:00"},
		{"23-01-2020T18", "2020-01-23T18:00:00"},
		{"23-01-2020 18", "2020-01-23T18:00:00"},
		{"23-01-2020 6pm", "2020-01-23T18:00:00"},
		{"23-01-2020 6am", "2020-01-23T06:00:00"},
		//
		{"23-1-2020T18:00:20", "2020-01-23T18:00:20"},
		{"23-1-2020 18:00:20", "2020-01-23T18:00:20"},
		{"23-1-2020T18:00", "2020-01-23T18:00:00"},
		{"23-1-2020 18:00", "2020-01-23T18:00:00"},
		{"23-1-2020T1800", "2020-01-23T18:00:00"},
		{"23-1-2020 1800", "2020-01-23T18:00:00"},
		{"23-1-2020T18", "2020-01-23T18:00:00"},
		{"23-1-2020 18", "2020-01-23T18:00:00"},
		{"23-1-2020 6pm", "2020-01-23T18:00:00"},
		{"23-1-2020 6am", "2020-01-23T06:00:00"},
		//
		{"1-1-2020T18:00:20", "2020-01-01T18:00:20"},
		{"1-1-2020 18:00:20", "2020-01-01T18:00:20"},
		{"1-1-2020T18:00", "2020-01-01T18:00:00"},
		{"1-1-2020 18:00", "2020-01-01T18:00:00"},
		{"1-1-2020T1800", "2020-01-01T18:00:00"},
		{"1-1-2020 1800", "2020-01-01T18:00:00"},
		{"1-1-2020T18", "2020-01-01T18:00:00"},
		{"1-1-2020 18", "2020-01-01T18:00:00"},
		{"1-1-2020 6pm", "2020-01-01T18:00:00"},
		{"1-1-2020 6am", "2020-01-01T06:00:00"},
		//
		{"1-12-2020T18:00:20", "2020-12-01T18:00:20"},
		{"1-12-2020 18:00:20", "2020-12-01T18:00:20"},
		{"1-12-2020T18:00", "2020-12-01T18:00:00"},
		{"1-12-2020 18:00", "2020-12-01T18:00:00"},
		{"1-12-2020T1800", "2020-12-01T18:00:00"},
		{"1-12-2020 1800", "2020-12-01T18:00:00"},
		{"1-12-2020T18", "2020-12-01T18:00:00"},
		{"1-12-2020 18", "2020-12-01T18:00:00"},
		{"1-12-2020 6pm", "2020-12-01T18:00:00"},
		{"1-12-2020 6am", "2020-12-01T06:00:00"},
		{"1-12-2020 6:59pm", "2020-12-01T18:59:00"},
		{"1-12-2020 6:59:30pm", "2020-12-01T18:59:30"},
	}
	for _, c := range deadlineTests {
		got := FixDeadline(c.in)
		if got != c.want {
			t.Errorf("FixDeadline(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
