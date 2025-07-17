package assignments

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

const testsFolder = "testdata/tests"

func TestWalkTestsRepository(t *testing.T) {
	wantFiles := map[string]struct{}{
		"testdata/tests/lab3/task-go-questions.md": {},
		"testdata/tests/lab3/task-learn-go.md":     {},
		"testdata/tests/lab3/task-tour-of-go.md":   {},
		"testdata/tests/scripts/Dockerfile":        {},
		"testdata/tests/scripts/run.sh":            {},
		"testdata/tests/lab1/assignment.yml":       {},
		"testdata/tests/lab1/run.sh":               {},
		"testdata/tests/lab2/assignment.yml":       {},
		"testdata/tests/lab3/assignment.yml":       {},
		"testdata/tests/lab4/assignment.yml":       {},
		"testdata/tests/lab4/criteria.json":        {},
		"testdata/tests/lab5/assignment.yml":       {},
		"testdata/tests/lab5/criteria.json":        {},
		"testdata/tests/lab6/assignment.yml":       {},
		"testdata/tests/lab6/tests.json":           {},
	}
	files, err := walkTestsRepository(testsFolder)
	if err != nil {
		t.Fatal(err)
	}
	for filename := range files {
		if _, ok := wantFiles[filename]; !ok {
			t.Errorf("unexpected file %q in %s", filename, testsFolder)
		}
	}
}

func TestReadTestsRepositoryContent(t *testing.T) {
	wantDockerfile := "FROM golang:1.24-alpine\nRUN apk update && apk add --no-cache git=~2.47 bash=~5.2.37 build-base=~0.5\nWORKDIR /quickfeed\n"
	wantAssignments := []*qf.Assignment{
		{
			Name:       "lab1",
			CourseID:   1,
			Order:      1,
			ScoreLimit: 80,
			Deadline:   qtest.Timestamp(t, "2019-01-24T14:00:00"),
		},
		{
			Name:       "lab2",
			CourseID:   1,
			Order:      2,
			ScoreLimit: 80,
			Deadline:   qtest.Timestamp(t, "2019-01-31T16:00:00"),
		},
		{
			Name:       "lab3",
			CourseID:   1,
			Order:      3,
			ScoreLimit: 80,
			Deadline:   qtest.Timestamp(t, "2019-02-14T23:00:00"),
			IsGroupLab: true,
			Tasks: []*qf.Task{
				{
					Title:           "Exercises from Tour of Go",
					AssignmentOrder: 3,
					Name:            "tour-of-go",
				},
				{
					Title:           "Go Exercises",
					AssignmentOrder: 3,
					Name:            "learn-go",
				},
				{
					Title:           "Multiple Choice Questions about Go Programming",
					AssignmentOrder: 3,
					Name:            "go-questions",
				},
			},
		},
		{
			Name:       "lab4",
			CourseID:   1,
			Order:      4,
			ScoreLimit: 80,
			Deadline:   qtest.Timestamp(t, "2019-03-15T16:00:00"),
			IsGroupLab: true,
			GradingBenchmarks: []*qf.GradingBenchmark{
				{
					Heading:  "Assignment 1",
					CourseID: 1,
					Criteria: []*qf.GradingCriterion{
						{
							CourseID:    1,
							Description: "Links work",
						},
						{
							CourseID:    1,
							Description: "Images are links, opening in a new tab",
						},
					},
				},
			},
		},
		{
			Name:       "lab5",
			CourseID:   1,
			Order:      5,
			ScoreLimit: 80,
			Deadline:   qtest.Timestamp(t, "2025-07-21T16:00:00"),
			IsGroupLab: true,
			GradingBenchmarks: []*qf.GradingBenchmark{
				{
					CourseID: 1,
					Criteria: []*qf.GradingCriterion{
						{
							CourseID: 1,
						},
					},
				},
			},
		},
	}

	gotAssignments, gotDockerfile, err := readTestsRepositoryContent(testsFolder, 1)
	if err != nil {
		t.Fatal(err)
	}
	if gotDockerfile != wantDockerfile {
		t.Errorf("got Dockerfile %q, want %q", gotDockerfile, wantDockerfile)
	}
	if diff := cmp.Diff(wantAssignments, gotAssignments, protocmp.Transform(), protocmp.IgnoreFields(&qf.Task{}, "body")); diff != "" {
		t.Errorf("readTestsRepositoryContent() mismatch (-wantAssignments +gotAssignments):\n%s", diff)
	}
}

func TestReadTestsRepositoryContentForInvalidCriteriaFiles(t *testing.T) {
	tests := []struct {
		name   string
		folder string
	}{
		{name: "invalidTypes", folder: "testdata/invalidJsonTests/invalidTypes"},
		{name: "negativeInteger", folder: "testdata/invalidJsonTests/negativeInteger"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			checkLabWithInvalidCriteriaFile(t, tc.folder)
		})
	}
}

func TestReadTestsRepositoryContentWithTestsJSON(t *testing.T) {
	// Test specifically for tests.json functionality
	gotAssignments, _, err := readTestsRepositoryContent(testsFolder, 1)
	if err != nil {
		t.Fatal(err)
	}
	
	// Find lab6 which has tests.json
	var lab6Assignment *qf.Assignment
	for _, assignment := range gotAssignments {
		if assignment.GetName() == "lab6" {
			lab6Assignment = assignment
			break
		}
	}
	
	if lab6Assignment == nil {
		t.Fatal("lab6 assignment not found")
	}
	
	submissions := lab6Assignment.GetSubmissions()
	if len(submissions) == 0 {
		t.Fatal("expected at least one submission (test info) for lab6")
	}
	
	// Check that the first submission has dummy ID 0 (test info)
	testInfoSubmission := submissions[0]
	if testInfoSubmission.GetID() != 0 {
		t.Errorf("expected test info submission to have ID 0, got %d", testInfoSubmission.GetID())
	}
	
	scores := testInfoSubmission.GetScores()
	if len(scores) != 2 {
		t.Errorf("expected 2 test scores, got %d", len(scores))
	}
	
	expectedTests := map[string]struct {
		maxScore int32
		weight   int32
	}{
		"TestExample1": {maxScore: 100, weight: 10},
		"TestExample2": {maxScore: 50, weight: 5},
	}
	
	for _, score := range scores {
		expected, ok := expectedTests[score.GetTestName()]
		if !ok {
			t.Errorf("unexpected test %s", score.GetTestName())
			continue
		}
		if score.GetMaxScore() != expected.maxScore {
			t.Errorf("test %s: expected max score %d, got %d", score.GetTestName(), expected.maxScore, score.GetMaxScore())
		}
		if score.GetWeight() != expected.weight {
			t.Errorf("test %s: expected weight %d, got %d", score.GetTestName(), expected.weight, score.GetWeight())
		}
	}
}

func checkLabWithInvalidCriteriaFile(t *testing.T, folder string) {
	_, _, err := readTestsRepositoryContent(folder, 1)
	if err == nil {
		t.Errorf("expected error")
	}
	if !isUnmarshalError(err) {
		t.Errorf("expected unmarshal error, got: %v", err)
	}
}

// Check if the error is related to invalid JSON unmarshalling.
func isUnmarshalError(e error) bool {
	return strings.Contains(e.Error(), "failed to unmarshal")
}
