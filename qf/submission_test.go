package qf_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

func TestNewestSubmissionDate(t *testing.T) {
	submission := &qf.Submission{}
	tim := time.Now()
	newSubmissionDate := submission.NewestSubmissionDate(tim)
	if !newSubmissionDate.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, tim)
	}

	submission = &qf.Submission{}
	newSubmissionDate = submission.NewestSubmissionDate(tim)
	if !newSubmissionDate.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, tim)
	}

	submission = &qf.Submission{
		BuildInfo: &score.BuildInfo{},
	}
	newSubmissionDate = submission.NewestSubmissionDate(tim)
	if !newSubmissionDate.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, tim)
	}

	submission = &qf.Submission{
		BuildInfo: &score.BuildInfo{
			SubmissionDate: &timestamppb.Timestamp{},
		},
	}
	newSubmissionDate = submission.NewestSubmissionDate(tim)
	if !newSubmissionDate.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, tim)
	}

	// Seems like the conversion from time.Time to timestamppb.Timestamp is not
	// exact, so we need to add a second to make sure the build date is newer.
	buildDate := time.Now().Add(1 * time.Second)
	submission = &qf.Submission{
		BuildInfo: &score.BuildInfo{
			SubmissionDate: timestamppb.New(buildDate),
		},
	}
	newSubmissionDate = submission.NewestSubmissionDate(tim)
	if newSubmissionDate.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, buildDate)
	}
	if newSubmissionDate.Before(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, buildDate)
	}
	if newSubmissionDate.After(buildDate) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, buildDate)
	}
	if !newSubmissionDate.After(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, buildDate)
	}

	zero := time.Time{}
	newSubmissionDate = submission.NewestSubmissionDate(zero)
	if newSubmissionDate.Equal(zero) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", zero, newSubmissionDate, buildDate)
	}
	if newSubmissionDate.Before(zero) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", zero, newSubmissionDate, buildDate)
	}
	if newSubmissionDate.After(buildDate) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", zero, newSubmissionDate, buildDate)
	}
	if !newSubmissionDate.After(zero) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", zero, newSubmissionDate, buildDate)
	}
}

func TestByUser(t *testing.T) {
	submission := &qf.Submission{}
	if submission.ByUser(0) {
		t.Errorf("ByUser(0) = true, expected false\n")
	}

	submission = &qf.Submission{
		UserID: 1,
	}
	if !submission.ByUser(1) {
		t.Errorf("ByUser(1) = false, expected true\n")
	}

	submission = &qf.Submission{
		GroupID: 1,
	}
	if submission.ByUser(1) {
		t.Errorf("ByUser(1) = true, expected false\n")
	}

	// submissions with both user and group ID are invalid
	submission = &qf.Submission{
		UserID:  1,
		GroupID: 2,
	}
	if submission.ByUser(1) {
		t.Errorf("ByUser(1) = true, expected false\n")
	}
}

func TestByGroup(t *testing.T) {
	submission := &qf.Submission{}
	if submission.ByGroup(0) {
		t.Errorf("ByGroup(0) = true, expected false\n")
	}

	submission = &qf.Submission{
		GroupID: 1,
	}
	if !submission.ByGroup(1) {
		t.Errorf("ByGroup(1) = false, expected true\n")
	}

	submission = &qf.Submission{
		UserID: 1,
	}
	if submission.ByGroup(1) {
		t.Errorf("ByGroup(1) = true, expected false\n")
	}

	// submissions with both user and group ID are invalid
	submission = &qf.Submission{
		UserID:  1,
		GroupID: 2,
	}
	if submission.ByGroup(1) {
		t.Errorf("ByGroup(1) = true, expected false\n")
	}
}

func TestCountApprovedSubmissions(t *testing.T) {
	enroll1 := &qf.Enrollment{ID: 10, UserID: 1}
	enroll2 := &qf.Enrollment{ID: 20, UserID: 2}
	enroll3 := &qf.Enrollment{ID: 30, UserID: 3}
	enroll4 := &qf.Enrollment{ID: 40, UserID: 4}

	submissions := qf.CourseSubmissions{
		Submissions: map[uint64]*qf.Submissions{
			enroll1.ID: {
				Submissions: []*qf.Submission{
					// total approved = 3
					{ID: 1, AssignmentID: 1, UserID: enroll1.UserID, Grades: []*qf.Grade{{UserID: enroll1.UserID, Status: qf.Submission_APPROVED}}},
					{ID: 2, AssignmentID: 2, GroupID: 20, Grades: []*qf.Grade{{UserID: enroll1.UserID, Status: qf.Submission_APPROVED}}},
					{ID: 3, AssignmentID: 3, UserID: enroll1.UserID, Grades: []*qf.Grade{{UserID: 3, Status: qf.Submission_APPROVED}}},
					// duplicate approved assignment should be ignored
					{ID: 2, AssignmentID: 3, UserID: enroll1.UserID, Grades: []*qf.Grade{{UserID: enroll1.UserID, Status: qf.Submission_APPROVED}}},
				},
			},
			enroll2.ID: {
				Submissions: []*qf.Submission{
					// total approved = 4
					{ID: 1, AssignmentID: 1, GroupID: 30, Grades: []*qf.Grade{{UserID: enroll2.UserID, Status: qf.Submission_APPROVED}}},
					{ID: 2, AssignmentID: 2, UserID: enroll2.UserID, Grades: []*qf.Grade{{UserID: enroll2.UserID, Status: qf.Submission_APPROVED}}},
					{ID: 3, AssignmentID: 3, UserID: enroll2.UserID, Grades: []*qf.Grade{{UserID: enroll2.UserID, Status: qf.Submission_APPROVED}}},
					{ID: 4, AssignmentID: 4, UserID: enroll2.UserID, Grades: []*qf.Grade{{UserID: enroll2.UserID, Status: qf.Submission_APPROVED}}},
				},
			},
			enroll3.ID: {
				Submissions: []*qf.Submission{
					// total approved = 1
					{ID: 1, AssignmentID: 1, UserID: enroll3.UserID, Grades: []*qf.Grade{
						{UserID: enroll3.UserID, Status: qf.Submission_APPROVED},
						// duplicate grade should be ignored
						{UserID: enroll3.UserID, Status: qf.Submission_APPROVED},
					}},
				},
			},
			enroll4.ID: {
				Submissions: []*qf.Submission{
					// total approved = 1
					{ID: 1, AssignmentID: 1, UserID: enroll4.UserID, Grades: []*qf.Grade{{UserID: enroll4.UserID, Status: qf.Submission_APPROVED}}},
					// duplicate assignment should be ignored
					{ID: 1, AssignmentID: 1, GroupID: 40, Grades: []*qf.Grade{{UserID: enroll4.UserID, Status: qf.Submission_APPROVED}}},
					{ID: 2, AssignmentID: 2, UserID: enroll4.UserID, Grades: []*qf.Grade{{UserID: enroll4.UserID, Status: qf.Submission_NONE}}},
					// user has no grade for this assignment
					{ID: 3, AssignmentID: 3, GroupID: 40, Grades: []*qf.Grade{{UserID: 10, Status: qf.Submission_APPROVED}}},
				},
			},
		},
	}

	tests := []*struct {
		enrollment *qf.Enrollment
		want       uint64
	}{
		{enroll1, 3},
		{enroll2, 4},
		{enroll3, 1},
		{enroll4, 1},
	}

	for _, test := range tests {
		enrollment := test.enrollment
		enrollment.CountApprovedSubmissions(submissions.For(enrollment.GetID()))
		if enrollment.GetTotalApproved() != test.want {
			t.Errorf("expected enrollment(id=%d) total approved %d, got %d", enrollment.GetID(), test.want, enrollment.GetTotalApproved())
		}
	}
}

func TestBeforeCreate(t *testing.T) {
	submission1 := &qf.Submission{ID: 1, UserID: 1, Grades: []*qf.Grade{{UserID: 1, Status: qf.Submission_APPROVED}}}
	submission2 := &qf.Submission{ID: 2, GroupID: 2}
	submission3 := &qf.Submission{ID: 3, GroupID: 1, Grades: []*qf.Grade{{UserID: 2, Status: qf.Submission_REJECTED}, {UserID: 3, Status: qf.Submission_REVISION}}}
	// Not sure if its necessary to check this case; submission3 := &qf.Submission{UserID: 1, GroupID: 1}

	gormDB, cleanup := qtest.TestGormDB(t, &qf.Enrollment{})
	defer cleanup()

	gormDB.Create(&qf.Enrollment{GroupID: 1, UserID: 1, CourseID: 1})
	gormDB.Create(&qf.Enrollment{GroupID: 1, UserID: 2, CourseID: 1})
	gormDB.Create(&qf.Enrollment{GroupID: 1, UserID: 3, CourseID: 1})

	tests := map[string]*struct {
		submission *qf.Submission
		db         *gorm.DB
		wantErr    error
		wantResult []*qf.Grade
	}{
		"Group and user Id is nil":      {&qf.Submission{}, nil, errors.New("submission must have either user or group ID"), nil},
		"User should be assigned grade": {submission1, nil, nil, []*qf.Grade{{UserID: 1, SubmissionID: submission1.GetID(), Status: qf.Submission_APPROVED}}},
		"Test enrollment with no users": {submission2, gormDB, errors.New("group has no users"), nil},
		"Group should be assigned grades": {submission3, gormDB, nil, []*qf.Grade{
			{UserID: 1, SubmissionID: submission3.GetID(), Status: qf.Submission_NONE},
			{UserID: 2, SubmissionID: submission3.GetID(), Status: qf.Submission_REJECTED},
			{UserID: 3, SubmissionID: submission3.GetID(), Status: qf.Submission_REVISION},
		}},
	}

	for key, test := range tests {
		t.Run(key, func(t *testing.T) {
			err := test.submission.BeforeCreate(test.db)
			if (err == nil) != (test.wantErr == nil) || err != nil && test.wantErr != nil && err.Error() != test.wantErr.Error() {
				t.Errorf("Expected err: %v, got: %v\n", err, test.wantErr)
			}
			if !reflect.DeepEqual(test.wantResult, test.submission.Grades) {
				t.Errorf("Expected grades: %v, got: %v\n", test.wantResult, test.submission.Grades)
			}
		})
	}
}
