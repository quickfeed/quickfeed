package yamlparser_test

import (
	"reflect"
	"testing"

	"github.com/autograde/aguis/yamlparser"
)

func TestParseWithInvalidDir(t *testing.T) {
	const dir = "invalid/dir"
	_, err := yamlparser.Parse(dir)
	if err == nil {
		t.Errorf("want no such file or directory error, got nil")
	}
}

func TestParse(t *testing.T) {
	const dir = "testrepos"
	var (
		wantAssignment1 = yamlparser.NewAssignmentRequest{
			AssignmentID: 2,
			Name:         "Lab1",
			Language:     "java",
			Deadline:     "27-08-2018 12:00",
			AutoApprove:  false,
		}
	)

	assgns, err := yamlparser.Parse(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(assgns) < 0 {
		t.Error("have 0 assignments, want 2")
	}

	if !reflect.DeepEqual(assgns[0], wantAssignment1) {
		t.Errorf("have assignment %+v want %+v", assgns[0], wantAssignment1)
	}
}
