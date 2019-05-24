package yamlparser_test

import (
	"testing"

	"github.com/autograde/aguis/yamlparser"
)

func TestParseWithInvalidDir(t *testing.T) {
	const dir = "invalid/dir"
	_, err := yamlparser.Parse(dir, 0)
	if err == nil {
		t.Errorf("want no such file or directory error, got nil")
	}
}

//TODO(Vera): fix time parsing
/*
func TestParse(t *testing.T) {
	const dir = "testrepos"
	deadline, err := time.Parse("02-01-2006 15:04", "27-08-2018 12:00")
	if err != nil {
		t.Fatal(err)
	}
	tstamp, err := tspb.TimestampProto(deadline)
	if err != nil {
		t.Fatal(err)
	}
	var (
		wantAssignment1 = &pb.Assignment{
			Id:          2,
			Name:        "Lab1",
			Language:    "java",
			Deadline:    tstamp,
			AutoApprove: false,
			Order:       2,
		}
	)

	assgns, err := yamlparser.Parse(dir, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(assgns) < 0 {
		t.Error("have 0 assignments, want 2")
	}

	if !reflect.DeepEqual(assgns[0], wantAssignment1) {
		t.Errorf("\nhave %+v \nwant %+v", assgns[0], wantAssignment1)
	}
}*/
