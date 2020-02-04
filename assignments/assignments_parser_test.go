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
	_, err = runner.Run(context.Background(), job, "")
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

	wantAssignment1 := &pb.Assignment{
		Name:        "For loops",
		Language:    "go",
		Deadline:    "27-08-2017 12:00",
		AutoApprove: false,
		Order:       1,
		ScoreLimit:  80,
	}

	wantAssignment2 := &pb.Assignment{
		Name:        "Nested loops",
		Language:    "java",
		Deadline:    "27-08-2018 12:00",
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
