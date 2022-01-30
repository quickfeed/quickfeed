package assignments

import (
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/google/go-cmp/cmp"
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
	wantDockerfile := `FROM golang:1.17-alpine
RUN apk update && apk add --no-cache git bash build-base
WORKDIR /quickfeed
`

	wantAssignments := []*pb.Assignment{
		{
			Name:       "lab1",
			CourseID:   1,
			Order:      1,
			ScoreLimit: 80,
			Deadline:   "2019-01-24T14:00:00",
			ScriptFile: `#image/quickfeed:go

printf "Custom lab1 script\n"
`,
		},
		{
			Name:       "lab2",
			CourseID:   1,
			Order:      2,
			ScoreLimit: 80,
			Deadline:   "2019-01-31T16:00:00",
			ScriptFile: `#image/quickfeed:go

printf "Default script\n"
`,
		},
		{
			Name:       "lab3",
			CourseID:   1,
			Order:      3,
			ScoreLimit: 80,
			Deadline:   "2019-02-14T23:00:00",
			IsGroupLab: true,
			ScriptFile: `#image/quickfeed:go

printf "Default script\n"
`,
			Tasks: []*pb.Task{
				{Title: "Exercises from Tour of Go"},
				{Title: "Go Exercises"},
				{Title: "Multiple Choice Questions about Go Programming"},
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
	if diff := cmp.Diff(wantAssignments, gotAssignments, protocmp.Transform(), protocmp.IgnoreFields(&pb.Task{}, "Body")); diff != "" {
		t.Errorf("readTestsRepositoryContent() mismatch (-wantAssignments +gotAssignments):\n%s", diff)
	}
}
