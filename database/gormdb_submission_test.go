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

func setupCourseAssignment(t *testing.T, db database.Database) (*qf.User, *qf.Course, *qf.Assignment) {
	// create a course and an assignment
	admin := qtest.CreateFakeUser(t, db, 10)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)
	assignment := &qf.Assignment{
		CourseID: course.ID,
		Order:    1,
	}
	if err := db.CreateAssignment(assignment); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as student
	user := qtest.CreateFakeUser(t, db, 11)
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Fatal(err)
	}
	query := &qf.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}
	return user, course, assignment
}

func TestGormDBUpdateSubmissionZeroScore(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user, course, assignment := setupCourseAssignment(t, db)

	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Score:        80,
	}); err != nil {
		t.Fatal(err)
	}

	submissions, err := db.GetLastSubmissions(course.ID, &qf.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Errorf("have %d submissions want %d", len(submissions), 1)
	}
	want := &qf.Submission{
		ID:           submissions[0].ID,
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Score:        80,
		Status:       qf.Submission_NONE,
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if diff := cmp.Diff(submissions[0], want, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same submission, but got (-sub +want):\n%s", diff)
	}

	// Set score to zero after having recorded a score of 80
	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Score:        0,
	}); err != nil {
		t.Fatal(err)
	}

	submissions, err = db.GetLastSubmissions(course.ID, &qf.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	want = &qf.Submission{
		ID:           submissions[0].ID,
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Score:        0,
		Status:       qf.Submission_NONE,
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
	user, course, assignment := setupCourseAssignment(t, db)

	// when we create a new submission for the same course lab and user, it will update the old one,
	// instead of creating an extra record
	// check that it is still approved after using create method

	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
	}); err != nil {
		t.Fatal(err)
	}

	// confirm that the submission is in the database
	submissions, err := db.GetLastSubmissions(course.ID, &qf.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}

	want := &qf.Submission{
		ID:           submissions[0].ID,
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Status:       qf.Submission_NONE,
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if diff := cmp.Diff(submissions[0], want, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same submission, but got (-sub +want):\n%s", diff)
	}

	if submissions[0].GetStatus() != qf.Submission_NONE {
		t.Errorf("expected submission to be 'not-approved' but got 'approved'")
	}

	// approved must stay false
	err = db.UpdateSubmission(submissions[0])
	if err != nil {
		t.Fatal(err)
	}
	submissions, err = db.GetLastSubmissions(course.ID, &qf.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}

	if submissions[0].GetStatus() != qf.Submission_NONE {
		t.Errorf("expected submission to be 'not-approved' but got 'approved'")
	}
	submissions[0].Status = qf.Submission_APPROVED
	err = db.UpdateSubmission(submissions[0])
	if err != nil {
		t.Fatal(err)
	}
	submissions, err = db.GetLastSubmissions(course.ID, &qf.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if submissions[0].GetStatus() != qf.Submission_APPROVED {
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
	user, course, assignment := setupCourseAssignment(t, db)

	// create a submission for the assignment for non-existing user; should fail
	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.ID,
		UserID:       3,
	}); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal(err)
	}

	// create another submission for the assignment; now it should succeed
	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
	}); err != nil {
		t.Fatal(err)
	}

	// confirm that the submission and its build info is in the database
	submissions, err := db.GetLastSubmissions(course.ID, &qf.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}
	gotSubmission := submissions[0]
	wantSubmission := &qf.Submission{
		ID:           gotSubmission.ID,
		AssignmentID: assignment.ID,
		UserID:       user.ID,
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

	// expected to fail
	if err := db.CreateSubmission(&qf.Submission{}); !errors.Is(err, database.ErrInvalidAssignmentID) {
		t.Fatal(err)
	}
	// expected to fail
	if err := db.CreateSubmission(&qf.Submission{AssignmentID: 1}); !errors.Is(err, database.ErrInvalidSubmission) {
		t.Fatal(err)
	}
	// expected to fail
	if err := db.CreateSubmission(&qf.Submission{UserID: 1}); !errors.Is(err, database.ErrInvalidAssignmentID) {
		t.Fatal(err)
	}
	// expected to fail with record not found
	if err := db.CreateSubmission(&qf.Submission{AssignmentID: 1, UserID: 1}); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal(err)
	}
	// expected to fail with record not found
	if err := db.CreateSubmission(&qf.Submission{AssignmentID: 1, GroupID: 6}); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal(err)
	}

	// create teacher, course, user (student) and assignment
	user, _, assignment := setupCourseAssignment(t, db)

	// create a submission for the assignment for non-existing user; should fail
	if err := db.CreateSubmission(&qf.Submission{AssignmentID: assignment.ID, UserID: 3}); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal(err)
	}
	// create a submission for the assignment for non-existing user; should fail
	if err := db.CreateSubmission(&qf.Submission{AssignmentID: assignment.ID, GroupID: 9}); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal(err)
	}

	// create another submission for the assignment; now it should succeed
	if err := db.CreateSubmission(&qf.Submission{AssignmentID: assignment.ID, UserID: user.ID}); err != nil {
		t.Fatal(err)
	}
}

func TestGormDBGetInsertSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)
	c1 := &qf.Course{OrganizationID: 1, Year: 1}
	c2 := &qf.Course{OrganizationID: 2, Year: 2}
	qtest.CreateCourse(t, db, admin, c1)
	qtest.CreateCourse(t, db, admin, c2)

	// create user and enroll as student
	user := qtest.CreateFakeUser(t, db, 11)

	// enroll student in course c1
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: c1.ID,
	}); err != nil {
		t.Fatal(err)
	}
	query := &qf.Enrollment{
		UserID:   user.ID,
		CourseID: c1.ID,
		Status:   qf.Enrollment_STUDENT,
	}
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	// Create some assignments
	assignment1 := qf.Assignment{
		Order:    1,
		CourseID: c1.ID,
	}
	if err := db.CreateAssignment(&assignment1); err != nil {
		t.Fatal(err)
	}
	assignment2 := qf.Assignment{
		Order:    2,
		CourseID: c1.ID,
	}
	if err := db.CreateAssignment(&assignment2); err != nil {
		t.Fatal(err)
	}
	assignment3 := qf.Assignment{
		Order:    1,
		CourseID: c2.ID,
	}
	if err := db.CreateAssignment(&assignment3); err != nil {
		t.Fatal(err)
	}

	// Create some submissions. We need IDs set here to be able
	// to compare local submission structs with database structs.
	submission1 := qf.Submission{
		UserID:       user.ID,
		AssignmentID: assignment1.ID,
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission1); err != nil {
		t.Fatal(err)
	}
	submission2 := qf.Submission{
		UserID:       user.ID,
		AssignmentID: assignment1.ID,
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission2); err != nil {
		t.Fatal(err)
	}
	submission3 := qf.Submission{
		UserID:       user.ID,
		AssignmentID: assignment2.ID,
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission3); err != nil {
		t.Fatal(err)
	}

	// Even if there is three submission, only the latest for each assignment should be returned

	submissions, err := db.GetLastSubmissions(c1.ID, &qf.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	want := []*qf.Submission{&submission2, &submission3}
	if diff := cmp.Diff(submissions, want, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same submissions, but got (-sub +want):\n%s", diff)
	}
	data, err := db.GetLastSubmissions(c1.ID, &qf.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	} else if len(data) != 2 {
		t.Errorf("Expected '%v' elements in the array, got '%v'", 2, len(data))
	}
	// Since there is no submissions, but the course and user exist, an empty array should be returned
	data, err = db.GetLastSubmissions(c2.ID, &qf.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	} else if len(data) != 0 {
		t.Errorf("Expected '%v' elements in the array, got '%v'", 0, len(data))
	}
}

func TestGormDBCreateUpdateWithBuildInfoAndScores(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	user, course, assignment := setupCourseAssignment(t, db)

	// create a new submission, ensure that build info and scores are saved as well
	buildInfo := &score.BuildInfo{
		BuildDate: "2022-11-10T13:00:00",
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
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		BuildInfo:    buildInfo,
		Scores:       scores,
	}); err != nil {
		t.Fatal(err)
	}
	submissions, err := db.GetLastSubmissions(course.ID, &qf.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}

	buildInfo.SubmissionID = submissions[0].ID
	buildInfo.ID = 1
	if diff := cmp.Diff(buildInfo, submissions[0].BuildInfo, protocmp.Transform()); diff != "" {
		t.Errorf("Expected same build info, but got (-got +want):\n%s", diff)
	}
	if diff := cmp.Diff(
		submissions[0].Scores,
		scores,
		protocmp.Transform(),
		protocmp.IgnoreFields(&score.Score{}, "ID", "SubmissionID", "Secret")); diff != "" {
		t.Errorf("Incorrect scores after first save (-want, +got):\n%s", diff)
	}

	// buildInfo record must be updated (have the same ID as before) instead
	// of saving a duplicate
	oldSubmissionID := submissions[0].ID
	updatedBuildInfo := &score.BuildInfo{
		BuildDate: "2022-11-10T15:00:00",
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
	submissions, err = db.GetLastSubmissions(course.ID, &qf.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}

	updatedBuildInfo.ID = submissions[0].BuildInfo.ID
	updatedBuildInfo.SubmissionID = oldSubmissionID
	if diff := cmp.Diff(submissions[0].BuildInfo, updatedBuildInfo, protocmp.Transform()); diff != "" {
		t.Errorf("Expected updated build info, but got (-sub +want):\n%s", diff)
	}
	if diff := cmp.Diff(submissions[0].Scores, scores, protocmp.Transform(), protocmp.IgnoreFields(&score.Score{}, "Secret")); diff != "" {
		t.Errorf("Incorrect scores after update (-want, +got):\n%s", diff)
	}

	// attempting to update build info and scores with wrong submission ID must return an error
	submissions[0].ID = 123
	if err := db.CreateSubmission(submissions[0]); err == nil {
		t.Fatal("expected error: record not found")
	}
}

func TestGormDBGetLastSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	admin := qtest.CreateFakeUser(t, db, 10)
	c1 := &qf.Course{OrganizationID: 1, Year: 1}
	c2 := &qf.Course{OrganizationID: 2, Year: 2}
	qtest.CreateCourse(t, db, admin, c1)
	qtest.CreateCourse(t, db, admin, c2)
	// Create some assignments
	assignment1 := qf.Assignment{
		Order:    1,
		CourseID: c1.ID,
	}
	if err := db.CreateAssignment(&assignment1); err != nil {
		t.Fatal(err)
	}
	assignment2 := qf.Assignment{
		Order:    2,
		CourseID: c1.ID,
	}
	if err := db.CreateAssignment(&assignment2); err != nil {
		t.Fatal(err)
	}
	assignment3 := qf.Assignment{
		Order:    1,
		CourseID: c2.ID,
	}
	if err := db.CreateAssignment(&assignment3); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as student
	user := qtest.CreateFakeUser(t, db, 11)
	// create a new submission
	submission := qf.Submission{
		AssignmentID: assignment1.ID,
		UserID:       user.ID,
	}
	if err := db.CreateSubmission(&submission); err != nil {
		t.Fatal(err)
	}

	// create a new submission
	submission2 := qf.Submission{
		AssignmentID: assignment2.ID,
		UserID:       user.ID,
	}
	if err := db.CreateSubmission(&submission2); err != nil {
		t.Fatal(err)
	}

	// create a new submission
	submission3 := qf.Submission{
		AssignmentID: assignment3.ID,
		UserID:       user.ID,
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
			courseID:     c1.ID,
			failCourseID: c2.ID,
			want:         &submission,
		},
		{
			name:         "assignment2, course1",
			courseID:     c1.ID,
			failCourseID: c2.ID,
			want:         &submission2,
		},
		{
			name:         "assignment3, course2",
			courseID:     c2.ID,
			failCourseID: c1.ID,
			want:         &submission3,
		},
	}

	for _, test := range tests {
		// Test that submission is returned only for the correct course
		_, err := db.GetLastSubmission(test.failCourseID, &qf.Submission{ID: test.want.ID})
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Errorf("Expected error: %v, got: %v", gorm.ErrRecordNotFound, err)
		}
		gotSubmission, err := db.GetLastSubmission(test.courseID, &qf.Submission{ID: test.want.ID})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(gotSubmission, test.want, protocmp.Transform()); diff != "" {
			t.Errorf("%s: Expected same submission, but got (-sub +want):\n%s", test.name, diff)
		}
	}

	// Test that non existing submission returns an error
	gotSubmission, err := db.GetLastSubmission(c1.ID, &qf.Submission{ID: 123})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Expected error: %v, got: %v", gorm.ErrRecordNotFound, err)
	}
	if gotSubmission != nil {
		t.Errorf("Expected nil submission, got: %v", gotSubmission)
	}
}
