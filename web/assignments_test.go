package web_test

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
)

func TestUpdateAssignments(t *testing.T) {
	db, cleanup := qtest.TestDB(t)

	defer cleanup()
	client := web.MockClient(t, db, scm.WithMockOrgs(), nil)
	course := qtest.MockCourses[0]
	user := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, user, course)

	tests := []struct {
		name    string
		request *qf.CourseRequest
		wantErr error
	}{
		{
			name: "Invalid course ID",
			request: &qf.CourseRequest{
				CourseID: 111,
			},
			wantErr: connect.NewError(connect.CodeNotFound, errors.New("course not found")),
		},
		{
			name: "Invalid course request",
			request: &qf.CourseRequest{
				CourseID: course.GetID(),
			},
			wantErr: connect.NewError(connect.CodeNotFound, errors.New("failed to clone assignments repository")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := client.UpdateAssignments(context.Background(), &connect.Request[qf.CourseRequest]{Msg: test.request})
			qtest.CheckError(t, err, test.wantErr)
		})
	}
}
