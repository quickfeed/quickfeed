package assignments

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/ci"
)

func TestParseWithInvalidDir(t *testing.T) {
	const dir = "invalid/dir"
	_, err := parseAssignments(dir, 0)
	if err == nil {
		t.Errorf("want no such file or directory error, got nil")
	}
}

const (
	y1 = `assignmentid: 1
name: "For loops"
language: "Go"
deadline: "27-08-2017 12:00"
autoapprove: false
`
	y2 = `assignmentid: 2
name: "Nested loops"
language: "Java"
deadline: "27-08-2018 12:00"
autoapprove: false
`
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
		},
	}
	runner := ci.Local{}
	_, err = runner.Run(context.Background(), job, "", 0)
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

	// We expect assignment names to be set based on
	// assignment folder names.
	wantAssignment1 := &pb.Assignment{
		Name:        "lab1",
		Language:    "go",
		Deadline:    "2017-08-27T12:00:00",
		AutoApprove: false,
		Order:       1,
		ScoreLimit:  80,
	}

	wantAssignment2 := &pb.Assignment{
		Name:        "lab2",
		Language:    "java",
		Deadline:    "2018-08-27T12:00:00",
		AutoApprove: false,
		Order:       2,
		ScoreLimit:  80,
	}

	assignments, err := parseAssignments(testsDir, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(assignments) < 1 {
		t.Error("have 0 assignments, want 2")
	}

	if !reflect.DeepEqual(assignments[0], wantAssignment1) {
		t.Errorf("\nhave %+v \nwant %+v", assignments[0], wantAssignment1)
	}
	if !reflect.DeepEqual(assignments[1], wantAssignment2) {
		t.Errorf("\nhave %+v \nwant %+v", assignments[1], wantAssignment2)
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
			t.Errorf("fixDeadline(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
