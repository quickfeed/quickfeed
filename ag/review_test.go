package ag_test

import (
	"testing"

	"github.com/autograde/quickfeed/ag"
	"github.com/google/go-cmp/cmp"
)

var tst = []*ag.GradingBenchmark{
	{
		ID:           0,
		AssignmentID: 10,
		Heading:      "Steve",
		Comment:      "Jobs",
		Criteria: []*ag.GradingCriterion{
			{ID: 1, Points: 50, BenchmarkID: 0, Description: "Ping"},
			{ID: 2, Points: 50, BenchmarkID: 0, Description: "Pong"},
		},
	},
	{
		ID:           1,
		AssignmentID: 10,
		Heading:      "Johnny",
		Comment:      "Ive",
		Criteria: []*ag.GradingCriterion{
			{ID: 1, Points: 50, BenchmarkID: 0, Description: "Ding"},
			{ID: 2, Points: 50, BenchmarkID: 0, Description: "Dong"},
		},
	},
}

func TestReviewMarshalString(t *testing.T) {
	r := &ag.Review{}
	err := r.MarshalReviewString()
	if err != nil {
		t.Fatal(err)
	}
	if r.Review != "" {
		t.Errorf("MarshalReviewString() = %s, expected empty review", r.Review)
	}
	r.Benchmarks = tst

	// Based on protobuf/jsonpb package (being deprecated)
	// want := `{"assignmentID":"10","heading":"Steve","comment":"Jobs","criteria":[{"ID":"1","points":"50","description":"Ping"},{"ID":"2","points":"50","description":"Pong"}]}; {"ID":"1","assignmentID":"10","heading":"Johnny","comment":"Ive","criteria":[{"ID":"1","points":"50","description":"Ding"},{"ID":"2","points":"50","description":"Dong"}]}`
	// Based on stdlib json package
	want := `{"assignmentID":10,"heading":"Steve","comment":"Jobs","criteria":[{"ID":1,"points":50,"description":"Ping"},{"ID":2,"points":50,"description":"Pong"}]}; {"ID":1,"assignmentID":10,"heading":"Johnny","comment":"Ive","criteria":[{"ID":1,"points":50,"description":"Ding"},{"ID":2,"points":50,"description":"Dong"}]}`
	err = r.MarshalReviewString()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(r.Review, want); diff != "" {
		t.Errorf("r.MarshalReviewString() mismatch (-want +got):\n%s", diff)
	}
}

func TestReviewUnmarshalString(t *testing.T) {
	// Based on protobuf/jsonpb package (being deprecated)
	// in := `{"assignmentID":"10","heading":"Steve","comment":"Jobs","criteria":[{"ID":"1","points":"50","description":"Ping"},{"ID":"2","points":"50","description":"Pong"}]}; {"ID":"1","assignmentID":"10","heading":"Johnny","comment":"Ive","criteria":[{"ID":"1","points":"50","description":"Ding"},{"ID":"2","points":"50","description":"Dong"}]}`
	// Based on stdlib json package
	in := `{"assignmentID":10,"heading":"Steve","comment":"Jobs","criteria":[{"ID":1,"points":50,"description":"Ping"},{"ID":2,"points":50,"description":"Pong"}]}; {"ID":1,"assignmentID":10,"heading":"Johnny","comment":"Ive","criteria":[{"ID":1,"points":50,"description":"Ding"},{"ID":2,"points":50,"description":"Dong"}]}`
	r := &ag.Review{Review: in}
	err := r.UnmarshalReviewString()
	if err != nil {
		t.Fatal(err)
	}
	want := tst
	if diff := cmp.Diff(r.Benchmarks, want); diff != "" {
		t.Errorf("r.UnmarshalReviewString() mismatch (-want +got):\n%s", diff)
	}
}
