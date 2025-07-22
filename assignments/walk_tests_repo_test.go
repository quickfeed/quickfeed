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
	// map of expected files in the testdata/tests folder
	// Note: run.sh is ignored by walkTestsRepository so they are not included here.
	wantFiles := map[string]struct{}{
		"testdata/tests/lab3/task-go-questions.md": {},
		"testdata/tests/lab3/task-learn-go.md":     {},
		"testdata/tests/lab3/task-tour-of-go.md":   {},
		"testdata/tests/scripts/Dockerfile":        {},
		"testdata/tests/lab1/assignment.yml":       {},
		"testdata/tests/lab1/tests.json":           {},
		"testdata/tests/lab2/assignment.yml":       {},
		"testdata/tests/lab2/tests.json":           {},
		"testdata/tests/lab3/assignment.yml":       {},
		"testdata/tests/lab4/assignment.yml":       {},
		"testdata/tests/lab4/criteria.json":        {},
		"testdata/tests/lab5/assignment.yml":       {},
		"testdata/tests/lab5/criteria.json":        {},
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
	for wantFilename := range wantFiles {
		if _, ok := files[wantFilename]; !ok {
			t.Errorf("missing file %q in %s", wantFilename, testsFolder)
		}
	}
	if len(files) != len(wantFiles) {
		t.Errorf("expected %d files, got %d", len(wantFiles), len(files))
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
			ExpectedTests: []*qf.TestInfo{
				{TestName: "TestGitQuestionsAG", MaxScore: 10, Weight: 1},
				{TestName: "TestMissingSemesterQuestionsAG", MaxScore: 9, Weight: 1},
				{TestName: "TestShellQuestionsAG", MaxScore: 20, Weight: 1},
			},
		},
		{
			Name:       "lab2",
			CourseID:   1,
			Order:      2,
			ScoreLimit: 80,
			Deadline:   qtest.Timestamp(t, "2019-01-31T16:00:00"),
			ExpectedTests: []*qf.TestInfo{
				{TestName: "Test0Formatting", MaxScore: 1, Weight: 5},
				{TestName: "Test0Lint", MaxScore: 1, Weight: 5},
				{TestName: "Test0TODOItems", MaxScore: 1, Weight: 5},
				{TestName: "Test0VetCheck", MaxScore: 1, Weight: 5},
				{TestName: "TestGrpc_ProtoGeneration", MaxScore: 2, Weight: 20},
				{TestName: "TestGrpc_RequestSequence", MaxScore: 14, Weight: 50},
				{TestName: "TestGrpc_ServerRaceCondition", MaxScore: 1, Weight: 50},
				{TestName: "TestNetworkQuestions", MaxScore: 5, Weight: 1},
				{TestName: "TestWeb_Counter", MaxScore: 5, Weight: 10},
				{TestName: "TestWeb_FizzBuzz", MaxScore: 18, Weight: 30},
				{TestName: "TestWeb_NonExisting", MaxScore: 6, Weight: 10},
				{TestName: "TestWeb_Redirect", MaxScore: 4, Weight: 20},
				{TestName: "TestWeb_Root", MaxScore: 1, Weight: 10},
				{TestName: "TestWeb_ServerFull", MaxScore: 39, Weight: 20},
			},
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
