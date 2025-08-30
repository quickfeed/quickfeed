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

	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors())
	course := qtest.MockCourses[0]
	user := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, user, course)
	cookie := client.Cookie(t, user)

	tests := []struct {
		name    string
		request *qf.CourseRequest
		wantErr error
	}{
		{
			name: "Invalid course ID (permission denied)",
			request: &qf.CourseRequest{
				CourseID: 111,
			},
			wantErr: connect.NewError(connect.CodePermissionDenied, errors.New("access denied for UpdateAssignments: required roles [teacher] not satisfied by claims: UserID: 1 (admin): Courses: map[1:TEACHER], Groups: []")),
		},
		{
			name: "Valid course ID but failed to clone repository",
			request: &qf.CourseRequest{
				CourseID: course.GetID(),
			},
			wantErr: connect.NewError(connect.CodeNotFound, errors.New("failed to clone assignments repository")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := client.UpdateAssignments(context.Background(), qtest.RequestWithCookie(test.request, cookie))
			qtest.CheckError(t, err, test.wantErr)
			// Check error code and that message contains expected key phrase
			gotCode := connect.CodeOf(err)
			wantCode := connect.CodeOf(test.wantErr)
			if gotCode != wantCode {
				t.Errorf("expected error code %v, got %v", wantCode, gotCode)
			}
		})
	}
}
