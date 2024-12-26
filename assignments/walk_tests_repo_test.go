package assignments

import (
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
	wantDockerfile := `FROM golang:1.19-alpine
RUN apk update && apk add --no-cache git bash build-base
WORKDIR /quickfeed
`
	courseID := uint64(1)

	wantAssignments := []*qf.Assignment{
		{
			Name:       "lab1",
			CourseID:   courseID,
			Order:      1,
			ScoreLimit: 80,
			Deadline:   qtest.Timestamp(t, "2019-01-24T14:00:00"),
		},
		{
			Name:       "lab2",
			CourseID:   courseID,
			Order:      2,
			ScoreLimit: 80,
			Deadline:   qtest.Timestamp(t, "2019-01-31T16:00:00"),
		},
		{
			Name:       "lab3",
			CourseID:   courseID,
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
			CourseID:   courseID,
			Order:      4,
			ScoreLimit: 80,
			Deadline:   qtest.Timestamp(t, "2019-03-15T16:00:00"),
			IsGroupLab: true,
			GradingBenchmarks: []*qf.GradingBenchmark{
				{
					CourseID: courseID, // Confirm that courseID is set
					Criteria: []*qf.GradingCriterion{
						{
							CourseID: courseID, // Confirm that courseID is set
						},
					},
				},
			},
		},
	}

	gotAssignments, gotDockerfile, err := readTestsRepositoryContent(testsFolder, courseID)
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
