package qf_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/qf"
)

func TestIncludeAssignment(t *testing.T) {
	tests := []struct {
		name       string
		req        *qf.SubmissionRequest
		assignment *qf.Assignment
		want       bool
	}{
		{
			name: "all I",
			req: &qf.SubmissionRequest{
				FetchMode: &qf.SubmissionRequest_Type{Type: qf.SubmissionRequest_ALL},
			},
			assignment: &qf.Assignment{IsGroupLab: true},
			want:       true,
		},
		{
			name: "all II",
			req: &qf.SubmissionRequest{
				FetchMode: &qf.SubmissionRequest_Type{Type: qf.SubmissionRequest_ALL},
			},
			assignment: &qf.Assignment{IsGroupLab: false},
			want:       true,
		},
		{
			name: "group I",
			req: &qf.SubmissionRequest{
				FetchMode: &qf.SubmissionRequest_Type{Type: qf.SubmissionRequest_GROUP},
			},
			assignment: &qf.Assignment{IsGroupLab: true},
			want:       true,
		},
		{
			name: "group II",
			req: &qf.SubmissionRequest{
				FetchMode: &qf.SubmissionRequest_Type{Type: qf.SubmissionRequest_GROUP},
			},
			assignment: &qf.Assignment{IsGroupLab: false},
			want:       false,
		},
		{
			name: "individual I",
			req: &qf.SubmissionRequest{
				FetchMode: &qf.SubmissionRequest_Type{Type: qf.SubmissionRequest_USER},
			},
			assignment: &qf.Assignment{IsGroupLab: true},
			want:       false,
		},
		{
			name: "individual II",
			req: &qf.SubmissionRequest{
				FetchMode: &qf.SubmissionRequest_Type{Type: qf.SubmissionRequest_USER},
			},
			assignment: &qf.Assignment{IsGroupLab: false},
			want:       true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.req.Include(test.assignment); got != test.want {
				t.Errorf("Include(%v) = %v, want %v", test.assignment, got, test.want)
			}
		})
	}
}
