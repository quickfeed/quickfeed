package yamlparser_test

import (
	"reflect"
	"testing"
	"time"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/yamlparser"
	tspb "github.com/golang/protobuf/ptypes"
)

func TestParseWithInvalidDir(t *testing.T) {
	const dir = "invalid/dir"
	_, err := yamlparser.Parse(dir, 0)
	if err == nil {
		t.Errorf("want no such file or directory error, got nil")
	}
}

func TestParse(t *testing.T) {
	const dir = "testrepos"
	d, err := time.Parse("02-01-2006 15:04", "27-08-2018 12:00")
	if err != nil {
		t.Fatal(err)
	}
	deadline, err := tspb.TimestampProto(d)
	if err != nil {
		t.Fatal(err)
	}
	var (
		wantAssignment1 = &pb.Assignment{
			Id:          2,
			Name:        "Lab1",
			Language:    "java",
			Deadline:    deadline,
			AutoApprove: false,
			Order:       2,
		}
	)

	assignments, err := yamlparser.Parse(dir, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(assignments) < 0 {
		t.Error("have 0 assignments, want 2")
	}

	if !reflect.DeepEqual(assignments[0], wantAssignment1) {
		t.Errorf("\nhave %+v \nwant %+v", assignments[0], wantAssignment1)
	}
}
