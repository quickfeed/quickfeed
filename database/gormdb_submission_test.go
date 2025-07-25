package database_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
)

func TestGormDBGetSubmissionForUser(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	query := &qf.Submission{AssignmentID: 10, UserID: 10}
	if _, err := db.GetSubmission(query); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGetSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	submission := &qf.Submission{AssignmentID: 1, UserID: 1}
	submission1 := &qf.Submission{AssignmentID: 1, UserID: 2}

	var wantSubmissions []*qf.Submission
	tests := []struct {
		name          string
		query         *qf.Submission
		newSubmission *qf.Submission
		wantError     error
	}{
		{name: "No Assignment ID", query: &qf.Submission{}, wantError: gorm.ErrRecordNotFound},
		{name: "Invalid assignment ID", query: &qf.Submission{AssignmentID: 4, UserID: 2}, wantError: gorm.ErrRecordNotFound},
		{name: "First submission", query: &qf.Submission{AssignmentID: 1}, newSubmission: submission},
		{name: "Second submission", query: &qf.Submission{AssignmentID: 1}, newSubmission: submission1},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Only create the course and assignment once we have tested the GetAssignment error path
			if i == 1 {
				_, _, _ = qtest.SetupCourseAssignment(t, db)
				qtest.CreateFakeUser(t, db)
			}
			if test.newSubmission != nil {
				qtest.CreateSubmission(t, db, test.newSubmission)
				wantSubmissions = append(wantSubmissions, test.newSubmission)
			}

			submissions, err := db.GetSubmissions(test.query)
			qtest.CheckError(t, err, test.wantError)

			if test.wantError != nil {
				return
			}

			qtest.Diff(t, "GetSubmissions() = mismatch", submissions, wantSubmissions, protocmp.Transform())
		})
	}
}

func TestGormDBCreateSubmissionWithAutoApprove(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user, _, assignment := qtest.SetupCourseAssignment(t, db)

	assignment.AutoApprove = true
	assignment.ScoreLimit = 1

	if err := db.UpdateAssignments([]*qf.Assignment{assignment}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		in   *qf.Submission
		want *qf.Submission
	}{
		{name: "Approved", in: &qf.Submission{AssignmentID: assignment.ID, UserID: user.ID, Score: 1}, want: &qf.Submission{ID: 1, AssignmentID: assignment.ID, UserID: user.ID, Score: 1, Grades: []*qf.Grade{{UserID: user.ID, SubmissionID: 1, Status: qf.Submission_APPROVED}}}},
		{name: "NotApproved", in: &qf.Submission{AssignmentID: assignment.ID, UserID: user.ID, Score: 0}, want: &qf.Submission{ID: 2, AssignmentID: assignment.ID, UserID: user.ID, Score: 0, Grades: []*qf.Grade{{UserID: user.ID, SubmissionID: 2, Status: qf.Submission_NONE}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := db.CreateSubmission(tt.in); err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(tt.in, tt.want, protocmp.Transform()); diff != "" {
				t.Errorf("CreateSubmission(): (-got +want):\n%s", diff)
			}
		})
	}
}

func TestGormDBUpdateSubmissionReleaseToFalse(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user, _, assignment := qtest.SetupCourseAssignment(t, db)
	submission := &qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Released:     true,
	}
	if err := db.CreateSubmission(submission); err != nil {
		t.Fatal(err)
	}
	submission.Released = false
	if err := db.UpdateSubmission(submission); err != nil {
		t.Fatal(err)
	}
	gotSubmission, err := db.GetSubmission(submission)
	if err != nil {
		t.Fatal(err)
	}
	qtest.Diff(t, "Expected release to be false", gotSubmission, submission, protocmp.Transform())
}

func TestGormDBUpdateSubmissionZeroScore(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user, course, assignment := qtest.SetupCourseAssignment(t, db)

	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Score:        80,
	}); err != nil {
		t.Fatal(err)
	}

	submissions, err := db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Errorf("have %d submissions want %d", len(submissions), 1)
	}
	want := &qf.Submission{
		ID:           submissions[0].GetID(),
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Score:        80,
		Grades:       []*qf.Grade{{UserID: user.GetID(), SubmissionID: submissions[0].GetID(), Status: qf.Submission_NONE}},
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if diff := cmp.Diff(submissions[0], want, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same submission, but got (-sub +want):\n%s", diff)
	}

	// Set score to zero after having recorded a score of 80
	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Score:        0,
	}); err != nil {
		t.Fatal(err)
	}

	submissions, err = db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	want = &qf.Submission{
		ID:           submissions[0].GetID(),
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Score:        0,
		Grades:       []*qf.Grade{{UserID: user.GetID(), SubmissionID: submissions[0].GetID(), Status: qf.Submission_NONE}},
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if diff := cmp.Diff(submissions[0], want, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same submission, but got (-sub +want):\n%s", diff)
	}
}

func TestGormDBUpdateSubmission(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user, course, assignment := qtest.SetupCourseAssignment(t, db)

	// when we create a new submission for the same course lab and user, it will update the old one,
	// instead of creating an extra record
	// check that it is still approved after using create method

	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
	}); err != nil {
		t.Fatal(err)
	}

	// confirm that the submission is in the database
	submissions, err := db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}

	want := &qf.Submission{
		ID:           submissions[0].GetID(),
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Grades:       []*qf.Grade{{UserID: user.GetID(), SubmissionID: submissions[0].GetID(), Status: qf.Submission_NONE}},
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if diff := cmp.Diff(submissions[0], want, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same submission, but got (-sub +want):\n%s", diff)
	}

	if submissions[0].GetStatusByUser(want.GetUserID()) != qf.Submission_NONE {
		t.Errorf("expected submission to be 'not-approved' but got 'approved'")
	}

	// approved must stay false
	err = db.UpdateSubmission(submissions[0])
	if err != nil {
		t.Fatal(err)
	}
	submissions, err = db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}

	if submissions[0].GetStatusByUser(want.GetUserID()) != qf.Submission_NONE {
		t.Errorf("expected submission to be 'not-approved' but got 'approved'")
	}
	submissions[0].SetGrade(user.GetID(), qf.Submission_APPROVED)
	err = db.UpdateSubmission(submissions[0])
	if err != nil {
		t.Fatal(err)
	}
	submissions, err = db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if submissions[0].GetStatusByUser(want.GetUserID()) != qf.Submission_APPROVED {
		t.Errorf("expected submission to be 'approved' but got 'not-approved'")
	}
}

func TestGormDBGetNonExistingSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	if _, err := db.GetLastSubmissions(10, &qf.Submission{UserID: 10}); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBInsertSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	// expected to fail with record not found
	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: 1,
		UserID:       1,
	}); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal(err)
	}

	// create teacher, course, user (student) and assignment
	user, course, assignment := qtest.SetupCourseAssignment(t, db)

	// create a submission for the assignment for non-existing user; should fail
	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       3,
	}); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal(err)
	}

	// create another submission for the assignment; now it should succeed
	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
	}); err != nil {
		t.Fatal(err)
	}

	// confirm that the submission and its build info is in the database
	submissions, err := db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}
	gotSubmission := submissions[0]
	wantSubmission := &qf.Submission{
		ID:           gotSubmission.GetID(),
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Grades:       []*qf.Grade{{UserID: user.GetID(), SubmissionID: gotSubmission.GetID(), Status: qf.Submission_NONE}},
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}

	if diff := cmp.Diff(wantSubmission, gotSubmission, protocmp.Transform()); diff != "" {
		t.Errorf("GetLastSubmissions() mismatch (-wantSubmission, +gotSubmission):\n%s", diff)
	}
}

func TestGormDBInsertBadSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user, _, assignment := qtest.SetupCourseAssignment(t, db)

	tests := []struct {
		name    string
		in      *qf.Submission
		wantErr error
	}{
		{name: "Empty", in: &qf.Submission{}, wantErr: database.ErrInvalidAssignmentID},
		{name: "No UserID or GroupID", in: &qf.Submission{AssignmentID: 1}, wantErr: database.ErrInvalidSubmission},
		{name: "No AssignmentID", in: &qf.Submission{UserID: 1}, wantErr: database.ErrInvalidAssignmentID},
		{name: "Invalid AssignmentID", in: &qf.Submission{AssignmentID: 5, UserID: user.GetID()}, wantErr: gorm.ErrRecordNotFound},
		{name: "Non-existing user", in: &qf.Submission{AssignmentID: assignment.GetID(), UserID: 3}, wantErr: gorm.ErrRecordNotFound},
		{name: "Non-existing group", in: &qf.Submission{AssignmentID: assignment.GetID(), GroupID: 9}, wantErr: gorm.ErrRecordNotFound},
		{name: "Both UserID and GroupID", in: &qf.Submission{AssignmentID: 1, UserID: 1, GroupID: 2}, wantErr: database.ErrInvalidSubmission},
		{name: "Non-existing submission", in: &qf.Submission{ID: 1, AssignmentID: assignment.GetID(), UserID: user.GetID()}, wantErr: gorm.ErrRecordNotFound},
		{name: "valid submission", in: &qf.Submission{AssignmentID: assignment.GetID(), UserID: user.GetID()}, wantErr: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := db.CreateSubmission(test.in); !errors.Is(err, test.wantErr) {
				t.Fatalf("got error '%v' wanted '%v'", err, test.wantErr)
			}
		})
	}
}

func TestGormDBGetInsertSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	c1 := &qf.Course{ScmOrganizationID: 1, Year: 1}
	c2 := &qf.Course{ScmOrganizationID: 2, Year: 2}
	qtest.CreateCourse(t, db, admin, c1)
	qtest.CreateCourse(t, db, admin, c2)

	// create user and enroll as student
	user := qtest.CreateFakeUser(t, db)

	// enroll student in course c1
	qtest.EnrollStudent(t, db, user, c1)

	// Create some assignments
	assignment1 := qf.Assignment{
		Order:    1,
		CourseID: c1.GetID(),
	}
	if err := db.CreateAssignment(&assignment1); err != nil {
		t.Fatal(err)
	}
	assignment2 := qf.Assignment{
		Order:    2,
		CourseID: c1.GetID(),
	}
	if err := db.CreateAssignment(&assignment2); err != nil {
		t.Fatal(err)
	}
	assignment3 := qf.Assignment{
		Order:    1,
		CourseID: c2.GetID(),
	}
	if err := db.CreateAssignment(&assignment3); err != nil {
		t.Fatal(err)
	}

	// Create some submissions. We need IDs set here to be able
	// to compare local submission structs with database structs.
	submission1 := qf.Submission{
		UserID:       user.GetID(),
		AssignmentID: assignment1.GetID(),
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission1); err != nil {
		t.Fatal(err)
	}
	submission2 := qf.Submission{
		UserID:       user.GetID(),
		AssignmentID: assignment1.GetID(),
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission2); err != nil {
		t.Fatal(err)
	}
	submission3 := qf.Submission{
		UserID:       user.GetID(),
		AssignmentID: assignment2.GetID(),
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission3); err != nil {
		t.Fatal(err)
	}

	// Even if there is three submission, only the latest for each assignment should be returned

	submissions, err := db.GetLastSubmissions(c1.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	want := []*qf.Submission{&submission2, &submission3}
	if diff := cmp.Diff(submissions, want, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same submissions, but got (-sub +want):\n%s", diff)
	}
	data, err := db.GetLastSubmissions(c1.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	} else if len(data) != 2 {
		t.Errorf("Expected '%v' elements in the array, got '%v'", 2, len(data))
	}
	// Since there is no submissions, but the course and user exist, an empty array should be returned
	data, err = db.GetLastSubmissions(c2.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	} else if len(data) != 0 {
		t.Errorf("Expected '%v' elements in the array, got '%v'", 0, len(data))
	}
}

func TestGormDBCreateUpdateWithBuildInfoAndScores(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user, course, assignment := qtest.SetupCourseAssignment(t, db)

	// create a new submission, ensure that build info and scores are saved as well
	buildInfo := &score.BuildInfo{
		BuildDate: qtest.Timestamp(t, "2022-11-10T13:00:00"),
		BuildLog:  "Testing",
		ExecTime:  33333,
	}
	scores := []*score.Score{
		{
			Secret:   "secret",
			TestName: "Test1",
			Score:    10,
			MaxScore: 15,
			Weight:   1,
		},
		{
			Secret:   "secret",
			TestName: "Test2",
			Score:    0,
			MaxScore: 5,
			Weight:   1,
		},
	}
	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		BuildInfo:    buildInfo,
		Scores:       scores,
	}); err != nil {
		t.Fatal(err)
	}
	submissions, err := db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}

	buildInfo.SubmissionID = submissions[0].GetID()
	buildInfo.ID = 1
	if diff := cmp.Diff(buildInfo, submissions[0].GetBuildInfo(), protocmp.Transform()); diff != "" {
		t.Errorf("Expected same build info, but got (-got +want):\n%s", diff)
	}
	if diff := cmp.Diff(
		submissions[0].GetScores(),
		scores,
		protocmp.Transform(),
		protocmp.IgnoreFields(&score.Score{}, "ID", "SubmissionID", "Secret")); diff != "" {
		t.Errorf("Incorrect scores after first save (-want, +got):\n%s", diff)
	}

	// buildInfo record must be updated (have the same ID as before) instead
	// of saving a duplicate
	oldSubmissionID := submissions[0].GetID()
	updatedBuildInfo := &score.BuildInfo{
		BuildDate: qtest.Timestamp(t, "2022-11-10T15:00:00"),
		BuildLog:  "Updated",
		ExecTime:  12345,
	}
	scores[1].Score = 5
	for _, sc := range scores {
		sc.ID = 0
		sc.SubmissionID = 0
	}
	submissions[0].BuildInfo = updatedBuildInfo
	submissions[0].Scores = scores
	if err := db.CreateSubmission(submissions[0]); err != nil {
		t.Fatal(err)
	}
	submissions, err = db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}

	updatedBuildInfo.ID = submissions[0].GetBuildInfo().GetID()
	updatedBuildInfo.SubmissionID = oldSubmissionID
	if diff := cmp.Diff(submissions[0].GetBuildInfo(), updatedBuildInfo, protocmp.Transform()); diff != "" {
		t.Errorf("Expected updated build info, but got (-sub +want):\n%s", diff)
	}
	if diff := cmp.Diff(submissions[0].GetScores(), scores, protocmp.Transform(), protocmp.IgnoreFields(&score.Score{}, "Secret")); diff != "" {
		t.Errorf("Incorrect scores after update (-want, +got):\n%s", diff)
	}

	// attempting to update build info and scores with wrong submission ID must return an error
	submissions[0].ID = 123
	if err := db.CreateSubmission(submissions[0]); err == nil {
		t.Fatal("expected error: record not found")
	}
}

func TestGormDBSubmissionWithBuildDate(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user, course, assignment := qtest.SetupCourseAssignment(t, db)

	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		BuildInfo: &score.BuildInfo{
			BuildDate: qtest.Timestamp(t, "2022-11-12T13:00:00"),
		},
	}); err != nil {
		t.Fatal(err)
	}

	want := &qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       user.GetID(),
		Grades:       []*qf.Grade{{UserID: user.GetID(), Status: qf.Submission_NONE}},
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
		BuildInfo: &score.BuildInfo{
			BuildDate: qtest.Timestamp(t, "2022-11-12T13:00:00"),
		},
	}
	submission, err := db.GetSubmission(&qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	want.ID = submission.GetID()
	want.Grades[0].SubmissionID = submission.GetID()
	want.BuildInfo.ID = 1
	want.BuildInfo.SubmissionID = submission.GetID()
	if diff := cmp.Diff(submission, want, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same submission, but got (-sub +want):\n%s", diff)
	}

	submissions, err := db.GetLastSubmissions(course.GetID(), &qf.Submission{UserID: user.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	want.ID = submissions[0].GetID()
	if diff := cmp.Diff(submissions[0], want, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same submission, but got (-sub +want):\n%s", diff)
	}
}

// TestGormDBGetLastSubmission tests that the GetLastSubmission function returns the correct submission
// GetLastSubmission should return an error if the provided course ID is not related to the submission
// or if the submission does not exist.
func TestGormDBGetLastSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	admin := qtest.CreateFakeUser(t, db)
	c1 := &qf.Course{ScmOrganizationID: 1, Year: 1}
	c2 := &qf.Course{ScmOrganizationID: 2, Year: 2}
	qtest.CreateCourse(t, db, admin, c1)
	qtest.CreateCourse(t, db, admin, c2)
	// Create some assignments
	assignment1 := qf.Assignment{
		Order:    1,
		CourseID: c1.GetID(),
	}
	if err := db.CreateAssignment(&assignment1); err != nil {
		t.Fatal(err)
	}
	assignment2 := qf.Assignment{
		Order:    2,
		CourseID: c1.GetID(),
	}
	if err := db.CreateAssignment(&assignment2); err != nil {
		t.Fatal(err)
	}
	assignment3 := qf.Assignment{
		Order:    1,
		CourseID: c2.GetID(),
	}
	if err := db.CreateAssignment(&assignment3); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as student
	user := qtest.CreateFakeUser(t, db)
	// create a new submission
	submission := qf.Submission{
		AssignmentID: assignment1.GetID(),
		UserID:       user.GetID(),
	}
	if err := db.CreateSubmission(&submission); err != nil {
		t.Fatal(err)
	}

	// create a new submission
	submission2 := qf.Submission{
		AssignmentID: assignment2.GetID(),
		UserID:       user.GetID(),
	}
	if err := db.CreateSubmission(&submission2); err != nil {
		t.Fatal(err)
	}

	// create a new submission
	submission3 := qf.Submission{
		AssignmentID: assignment3.GetID(),
		UserID:       user.GetID(),
	}
	if err := db.CreateSubmission(&submission3); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		courseID     uint64
		failCourseID uint64
		want         *qf.Submission
	}{
		{
			name:         "assignment1, course1",
			courseID:     c1.GetID(),
			failCourseID: c2.GetID(),
			want:         &submission,
		},
		{
			name:         "assignment2, course1",
			courseID:     c1.GetID(),
			failCourseID: c2.GetID(),
			want:         &submission2,
		},
		{
			name:         "assignment3, course2",
			courseID:     c2.GetID(),
			failCourseID: c1.GetID(),
			want:         &submission3,
		},
	}

	for _, test := range tests {
		// Test that submission is returned only for the correct course
		_, err := db.GetLastSubmission(test.failCourseID, &qf.Submission{ID: test.want.GetID()})
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Errorf("Expected error: %v, got: %v", gorm.ErrRecordNotFound, err)
		}
		gotSubmission, err := db.GetLastSubmission(test.courseID, &qf.Submission{ID: test.want.GetID()})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(gotSubmission, test.want, protocmp.Transform()); diff != "" {
			t.Errorf("%s: Expected same submission, but got (-sub +want):\n%s", test.name, diff)
		}
	}

	// Test that non existing submission returns an error
	gotSubmission, err := db.GetLastSubmission(c1.GetID(), &qf.Submission{ID: 123})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Expected error: %v, got: %v", gorm.ErrRecordNotFound, err)
	}
	if gotSubmission != nil {
		t.Errorf("Expected nil submission, got: %v", gotSubmission)
	}
}

func TestCreateSubmissionEnsuresAllExpectedTests(t *testing.T) {
	tests := []struct {
		name              string
		expectedTests     []*qf.TestInfo
		initialScores     []*score.Score
		updateScores      []*score.Score
		wantInitialScores []*score.Score
		wantFinalScores   []*score.Score
	}{
		{
			name: "partial scores become complete on create and update",
			expectedTests: []*qf.TestInfo{
				{TestName: "test1", MaxScore: 10, Weight: 5},
				{TestName: "test2", MaxScore: 20, Weight: 10},
				{TestName: "test3", MaxScore: 15, Weight: 8},
			},
			initialScores: []*score.Score{
				{TestName: "test1", Score: 8, MaxScore: 10, Weight: 5},
				// test2 and test3 intentionally left out, will be added from template
			},
			wantInitialScores: []*score.Score{
				{TestName: "test1", Score: 8, MaxScore: 10, Weight: 5},  // Explicitly provided, with score 8
				{TestName: "test2", Score: 0, MaxScore: 20, Weight: 10}, // Added from template with zero score
				{TestName: "test3", Score: 0, MaxScore: 15, Weight: 8},  // Added from template with zero score
			},
			updateScores: []*score.Score{
				{TestName: "test1", Score: 10, MaxScore: 10, Weight: 5},  // Updated score
				{TestName: "test2", Score: 15, MaxScore: 20, Weight: 10}, // New test starts to work
				// test3 is still missing, will be added automatically
			},
			wantFinalScores: []*score.Score{
				{TestName: "test1", Score: 10, MaxScore: 10, Weight: 5},  // Updated score
				{TestName: "test2", Score: 15, MaxScore: 20, Weight: 10}, // New test starts to work
				{TestName: "test3", Score: 0, MaxScore: 15, Weight: 8},   // Added from template with zero score
			},
		},
		{
			name: "all expected tests present - no duplicates created",
			expectedTests: []*qf.TestInfo{
				{TestName: "test1", MaxScore: 10, Weight: 5},
				{TestName: "test2", MaxScore: 20, Weight: 10},
			},
			initialScores: []*score.Score{
				{TestName: "test1", Score: 8, MaxScore: 10, Weight: 5},
				{TestName: "test2", Score: 15, MaxScore: 20, Weight: 10},
			},
			updateScores: []*score.Score{
				{TestName: "test1", Score: 10, MaxScore: 10, Weight: 5},
				{TestName: "test2", Score: 20, MaxScore: 20, Weight: 10},
			},
			wantInitialScores: []*score.Score{
				{TestName: "test1", Score: 8, MaxScore: 10, Weight: 5},
				{TestName: "test2", Score: 15, MaxScore: 20, Weight: 10},
			},
			wantFinalScores: []*score.Score{
				{TestName: "test1", Score: 10, MaxScore: 10, Weight: 5},
				{TestName: "test2", Score: 20, MaxScore: 20, Weight: 10},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup := qtest.TestDB(t)
			defer cleanup()
			user, _, assignment := qtest.SetupCourseAssignment(t, db)

			// Update the assignment with expected tests
			assignment.ExpectedTests = tt.expectedTests
			if err := db.UpdateAssignments([]*qf.Assignment{assignment}); err != nil {
				t.Fatal(err)
			}

			initialSubmission := &qf.Submission{
				AssignmentID: assignment.GetID(),
				UserID:       user.GetID(),
				Score:        25,
				Scores:       tt.initialScores,
			}
			gotSubmission := verifySubmissionScores(t, db, initialSubmission, tt.wantInitialScores, "initial submission")

			updateSubmission := &qf.Submission{
				ID:           gotSubmission.GetID(),
				AssignmentID: assignment.GetID(),
				UserID:       user.GetID(),
				Score:        30,
				Scores:       tt.updateScores,
			}
			verifySubmissionScores(t, db, updateSubmission, tt.wantFinalScores, "final submission")
		})
	}
}

// Helper function to create and verify submission scores.
func verifySubmissionScores(t *testing.T, db database.Database, submission *qf.Submission, wantScores []*score.Score, description string) *qf.Submission {
	if err := db.CreateSubmission(submission); err != nil {
		t.Fatal(err)
	}

	gotSubmission, err := db.GetSubmission(&qf.Submission{
		AssignmentID: submission.AssignmentID,
		UserID:       submission.UserID,
	})
	if err != nil {
		t.Fatal(err)
	}

	gotScores := gotSubmission.GetScores()
	if len(gotScores) != len(wantScores) {
		t.Errorf("Expected %d scores in %s, got %d", len(wantScores), description, len(gotScores))
	}
	if diff := cmp.Diff(gotScores, wantScores, protocmp.Transform(), protocmp.IgnoreFields(&score.Score{}, "ID", "SubmissionID")); diff != "" {
		t.Errorf("%s scores mismatch (-got +want):\n%s", description, diff)
	}

	return gotSubmission
}
