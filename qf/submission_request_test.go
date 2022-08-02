package qf_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/qf"
)

func TestIncludeGroup(t *testing.T) {
	tests := []struct {
		name       string
		req        *qf.SubmissionsForCourseRequest
		assignment *qf.Assignment
		want       bool
	}{
		{
			name:       "all I",
			req:        &qf.SubmissionsForCourseRequest{Type: qf.SubmissionsForCourseRequest_ALL},
			assignment: &qf.Assignment{IsGroupLab: true},
			want:       true,
		},
		{
			name:       "all II",
			req:        &qf.SubmissionsForCourseRequest{Type: qf.SubmissionsForCourseRequest_ALL},
			assignment: &qf.Assignment{IsGroupLab: false},
			want:       true,
		},
		{
			name:       "group I",
			req:        &qf.SubmissionsForCourseRequest{Type: qf.SubmissionsForCourseRequest_GROUP},
			assignment: &qf.Assignment{IsGroupLab: true},
			want:       true,
		},
		{
			name:       "group II",
			req:        &qf.SubmissionsForCourseRequest{Type: qf.SubmissionsForCourseRequest_GROUP},
			assignment: &qf.Assignment{IsGroupLab: false},
			want:       false,
		},
		{
			name:       "individual I",
			req:        &qf.SubmissionsForCourseRequest{Type: qf.SubmissionsForCourseRequest_INDIVIDUAL},
			assignment: &qf.Assignment{IsGroupLab: true},
			want:       false,
		},
		{
			name:       "individual II",
			req:        &qf.SubmissionsForCourseRequest{Type: qf.SubmissionsForCourseRequest_INDIVIDUAL},
			assignment: &qf.Assignment{IsGroupLab: false},
			want:       true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.req.Include(test.assignment); got != test.want {
				t.Errorf("IncludeGroup() = %v, want %v", got, test.want)
			}
		})
	}
}
