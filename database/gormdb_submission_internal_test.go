package database

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
)

func testGormDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	f, err := os.CreateTemp(t.TempDir(), "test.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	gormDB, err := NewGormDB(f.Name(), nil)
	if err != nil {
		t.Fatal(err)
	}

	return gormDB.conn, func() {
		if err := gormDB.Close(); err != nil {
			t.Error(err)
		}
		if err := os.Remove(f.Name()); err != nil {
			t.Error(err)
		}
	}
}

func TestBeforeCreate(t *testing.T) {
	submission2 := &qf.Submission{GroupID: 2}
	submission3 := &qf.Submission{GroupID: 1, Score: 1, AssignmentID: 1}
	submission4 := &qf.Submission{UserID: 2, Score: 1, AssignmentID: 2, Grades: []*qf.Grade{{UserID: 2, Status: qf.Submission_REJECTED}}}

	gormDB, cleanup := testGormDB(t)
	defer cleanup()

	gormDB.Create(&qf.Enrollment{GroupID: 1, UserID: 1})
	gormDB.Create(&qf.Enrollment{GroupID: 1, UserID: 2})

	gormDB.Create(&qf.Assignment{ID: 1, IsGroupLab: true, ScoreLimit: 1, AutoApprove: true})

	tests := []struct {
		name       string
		submission *qf.Submission
		db         *gorm.DB
		wantErr    error
		wantResult []*qf.Grade
	}{
		{name: "Test enrollment with no users", submission: submission2, wantErr: errors.New("group has no users")},
		{name: "Group should be assigned approved grades", submission: submission3, wantResult: []*qf.Grade{
			{UserID: 1, SubmissionID: submission3.GetID(), Status: qf.Submission_APPROVED},
			{UserID: 2, SubmissionID: submission3.GetID(), Status: qf.Submission_APPROVED},
		}},
		{name: "Test submission with no assignment", submission: submission4, wantErr: errors.New("submission must have an associated assignment"), wantResult: []*qf.Grade{{UserID: 2, SubmissionID: submission4.GetID(), Status: qf.Submission_REJECTED}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := beforeCreate(gormDB, test.submission)
			if err != nil && test.wantErr != nil && err.Error() != test.wantErr.Error() {
				t.Errorf("Expected err: %v, got: %v\n", err, test.wantErr)
			}
			if !cmp.Equal(test.wantResult, test.submission.GetGrades(), protocmp.Transform()) {
				t.Errorf("Expected grades: %v, got: %v\n", test.wantResult, test.submission.GetGrades())
			}
		})
	}
}
