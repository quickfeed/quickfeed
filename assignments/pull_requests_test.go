package assignments

import (
	"context"
	"errors"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/scm"
)

func TestGetLinkedIssue(t *testing.T) {
	var wantIssueNumber uint64 = 30
	tests := map[string]struct {
		body string
		err  error
	}{
		"Simple":         {body: "Fixes #30", err: nil},
		"Not a number":   {body: "Fixes #30nan", err: ErrInvalidBody},
		"Invalid body":   {body: "Fixes #30nan #", err: ErrInvalidBody},
		"Invalid body 2": {body: "Fixes", err: ErrInvalidBody},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotIssueNumber, err := getLinkedIssue(tt.body)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Errorf("getLinkedIssue() error mismatch, got %v, expected %v", err, tt.err)
				}
				return
			}
			if gotIssueNumber != wantIssueNumber {
				t.Errorf("getLinkedIssue() = %d, expected %d", gotIssueNumber, wantIssueNumber)
			}
		})
	}
}

func TestAssignReviewers(t *testing.T) {
	type testUser struct {
		name string
		role pb.Enrollment_UserStatus
	}
	tests := map[string]struct {
		testUsers []testUser
	}{
		"Simple": {
			testUsers: []testUser{
				{name: "student1", role: pb.Enrollment_STUDENT},
				{name: "teacher1", role: pb.Enrollment_TEACHER},
			},
		},
		"No enrollments": {testUsers: []testUser{}},
		"fds": {
			testUsers: []testUser{
				{name: "student1", role: pb.Enrollment_STUDENT},
				{name: "teacher1", role: pb.Enrollment_TEACHER},
			},
		},
	}

	logger := qtest.Logger(t)
	ctx := context.Background()
	repo := &pb.Repository{HTMLURL: "irrelevant"}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			db, cleanup := qtest.TestDB(t)
			defer cleanup()
			admin := qtest.CreateNamedUser(t, db, 1, "admin")
			course := &pb.Course{Provider: "fake"}
			qtest.CreateCourse(t, db, admin, course)

			var nextRemoteID uint64 = 2
			for _, testUser := range tt.testUsers {
				user := qtest.CreateNamedUser(t, db, nextRemoteID, testUser.name)
				enrollment := &pb.Enrollment{UserID: user.GetID(), CourseID: course.GetID()}
				if err := db.CreateEnrollment(enrollment); err != nil {
					t.Fatal(err)
				}
				enrollment.Status = testUser.role
				if err := db.UpdateEnrollment(enrollment); err != nil {
					t.Fatal(err)
				}
				nextRemoteID++
			}
			sc, err := scm.NewSCMClient(logger, "fake", "irrelevant")
			if err != nil {
				t.Fatal(err)
			}
			if err = assignReviewers(ctx, sc, db, course, repo, 1); err != nil && errors.Is(err, scm.ErrNotSupported{}) {
				t.Fatal(err)
			}
		})
	}
}
