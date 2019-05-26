package web_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/web"
	tspb "github.com/golang/protobuf/ptypes"
)

func TestParseWithInvalidDir(t *testing.T) {
	const dir = "invalid/dir"
	_, err := web.ParseAssignments(dir, 0)
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
	testsDir, err := ioutil.TempDir("", web.TestsRepo)
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
	_, err = runner.Run(context.Background(), job)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(testsDir, "lab1", "assignment.yml"), []byte(y1), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(testsDir, "lab2", "assignment.yml"), []byte(y2), 0644)
	if err != nil {
		t.Fatal(err)
	}

	d, err := time.Parse("02-01-2006 15:04", "27-08-2017 12:00")
	if err != nil {
		t.Fatal(err)
	}
	deadline, err := tspb.TimestampProto(d)
	if err != nil {
		t.Fatal(err)
	}
	wantAssignment1 := &pb.Assignment{
		Id:          1,
		Name:        "For loops",
		Language:    "go",
		Deadline:    deadline,
		AutoApprove: false,
		Order:       1,
	}

	d, err = time.Parse("02-01-2006 15:04", "27-08-2018 12:00")
	if err != nil {
		t.Fatal(err)
	}
	deadline, err = tspb.TimestampProto(d)
	if err != nil {
		t.Fatal(err)
	}
	wantAssignment2 := &pb.Assignment{
		Id:          2,
		Name:        "Nested loops",
		Language:    "java",
		Deadline:    deadline,
		AutoApprove: false,
		Order:       2,
	}

	assignments, err := web.ParseAssignments(testsDir, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(assignments) < 0 {
		t.Error("have 0 assignments, want 2")
	}

	if !reflect.DeepEqual(assignments[0], wantAssignment1) {
		t.Errorf("\nhave %+v \nwant %+v", assignments[0], wantAssignment1)
	}
	if !reflect.DeepEqual(assignments[1], wantAssignment2) {
		t.Errorf("\nhave %+v \nwant %+v", assignments[1], wantAssignment2)
	}
}
