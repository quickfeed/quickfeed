package database_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/autograde/quickfeed/ag"
	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/kit/score"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
)

func TestGormDBGetSubmissionForUser(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	query := &pb.Submission{AssignmentID: 10, UserID: 10}
	if _, err := db.GetSubmission(query); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func setupCourseAssignment(t *testing.T, db database.Database) (*pb.User, *pb.Course, *pb.Assignment) {
	teacher := createFakeUser(t, db, 10)
	// create a course and an assignment
	course := &pb.Course{}
	if err := db.CreateCourse(teacher.ID, course); err != nil {
		t.Fatal(err)
	}
	assignment := &pb.Assignment{
		CourseID: course.ID,
		Order:    1,
	}
	if err := db.CreateAssignment(assignment); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as student
	user := createFakeUser(t, db, 11)
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
	}); err != nil {
		t.Fatal(err)
	}
	query := &pb.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
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

	if err := db.CreateSubmission(&pb.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Score:        80,
	}); err != nil {
		t.Fatal(err)
	}

	submissions, err := db.GetLastSubmissions(course.ID, &pb.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Errorf("have %d submissions want %d", len(submissions), 1)
	}
	want := &pb.Submission{
		ID:           submissions[0].ID,
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Score:        80,
		Status:       pb.Submission_NONE,
		Reviews:      []*ag.Review{},
		Scores:       []*score.Score{},
	}
	if diff := cmp.Diff(submissions[0], want, cmpopts.IgnoreUnexported(pb.Submission{})); diff != "" {
		t.Errorf("Expected same submission, but got (-sub +want):\n%s", diff)
	}

	// Set score to zero after having recorded a score of 80
	if err := db.CreateSubmission(&pb.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Score:        0,
	}); err != nil {
		t.Fatal(err)
	}

	submissions, err = db.GetLastSubmissions(course.ID, &pb.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	want = &pb.Submission{
		ID:           submissions[0].ID,
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Score:        0,
		Status:       pb.Submission_NONE,
		Reviews:      []*ag.Review{},
		Scores:       []*score.Score{},
	}
	if diff := cmp.Diff(submissions[0], want, cmpopts.IgnoreUnexported(pb.Submission{})); diff != "" {
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

	// create another submission for the assignment; now it should succeed
	if err := db.CreateSubmission(&pb.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
	}); err != nil {
		t.Fatal(err)
	}

	// confirm that the submission is in the database
	submissions, err := db.GetLastSubmissions(course.ID, &pb.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}
	want := &pb.Submission{
		ID:           submissions[0].ID,
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Status:       pb.Submission_NONE,
		Reviews:      []*pb.Review{},
		Scores:       []*score.Score{},
	}
	if diff := cmp.Diff(submissions[0], want, cmpopts.IgnoreUnexported(pb.Submission{})); diff != "" {
		t.Errorf("Expected same submission, but got (-sub +want):\n%s", diff)
	}

	if submissions[0].GetStatus() != pb.Submission_NONE {
		t.Errorf("expected submission to be 'not-approved' but got 'approved'")
	}

	// approved must stay false
	err = db.UpdateSubmission(submissions[0])
	if err != nil {
		t.Fatal(err)
	}
	submissions, err = db.GetLastSubmissions(course.ID, &pb.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if submissions[0].GetStatus() != pb.Submission_NONE {
		t.Errorf("expected submission to be 'not-approved' but got 'approved'")
	}
	submissions[0].Status = pb.Submission_APPROVED
	err = db.UpdateSubmission(submissions[0])
	if err != nil {
		t.Fatal(err)
	}
	submissions, err = db.GetLastSubmissions(course.ID, &pb.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if submissions[0].GetStatus() != pb.Submission_APPROVED {
		t.Errorf("expected submission to be 'approved' but got 'not-approved'")
	}
}

func TestGormDBGetNonExistingSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	if _, err := db.GetLastSubmissions(10, &pb.Submission{UserID: 10}); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBInsertSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	// expected to fail with record not found
	if err := db.CreateSubmission(&pb.Submission{
		AssignmentID: 1,
		UserID:       1,
	}); err != gorm.ErrRecordNotFound {
		t.Fatal(err)
	}

	// create teacher, course, user (student) and assignment
	user, course, assignment := setupCourseAssignment(t, db)

	// create a submission for the assignment for non-existing user; should fail
	if err := db.CreateSubmission(&pb.Submission{
		AssignmentID: assignment.ID,
		UserID:       3,
	}); err != gorm.ErrRecordNotFound {
		t.Fatal(err)
	}

	// create another submission for the assignment; now it should succeed
	if err := db.CreateSubmission(&pb.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
	}); err != nil {
		t.Fatal(err)
	}

	// confirm that the submission is in the database
	submissions, err := db.GetLastSubmissions(course.ID, &pb.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(submissions) != 1 {
		t.Fatalf("have %d submissions want %d", len(submissions), 1)
	}
	want := &pb.Submission{
		ID:           submissions[0].ID,
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Reviews:      []*pb.Review{},
		Scores:       []*score.Score{},
	}
	if !reflect.DeepEqual(submissions[0], want) {
		t.Errorf("have %#v want %#v", submissions[0], want)
	}
}

func TestGormDBGetInsertSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	teacher := createFakeUser(t, db, 10)
	// Create course c1 and c2
	c1 := pb.Course{OrganizationID: 1}
	if err := db.CreateCourse(teacher.ID, &c1); err != nil {
		t.Fatal(err)
	}
	c2 := pb.Course{OrganizationID: 2}
	if err := db.CreateCourse(teacher.ID, &c2); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as student
	user := createFakeUser(t, db, 11)

	// enroll student in course c1
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: c1.ID,
	}); err != nil {
		t.Fatal(err)
	}
	query := &pb.Enrollment{
		UserID:   user.ID,
		CourseID: c1.ID,
		Status:   pb.Enrollment_STUDENT,
	}
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}

	// Create some assignments
	assignment1 := pb.Assignment{
		Order:    1,
		CourseID: c1.ID,
	}
	if err := db.CreateAssignment(&assignment1); err != nil {
		t.Fatal(err)
	}
	assignment2 := pb.Assignment{
		Order:    2,
		CourseID: c1.ID,
	}
	if err := db.CreateAssignment(&assignment2); err != nil {
		t.Fatal(err)
	}
	assignment3 := pb.Assignment{
		Order:    1,
		CourseID: c2.ID,
	}
	if err := db.CreateAssignment(&assignment3); err != nil {
		t.Fatal(err)
	}

	// Create some submissions. We need IDs set here to be able
	// to compare local submission structs with database structs.
	submission1 := pb.Submission{
		UserID:       user.ID,
		AssignmentID: assignment1.ID,
		Reviews:      []*pb.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission1); err != nil {
		t.Fatal(err)
	}
	submission2 := pb.Submission{
		ID:           1,
		UserID:       user.ID,
		AssignmentID: assignment1.ID,
		Reviews:      []*pb.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission2); err != nil {
		t.Fatal(err)
	}
	submission3 := pb.Submission{
		ID:           2,
		UserID:       user.ID,
		AssignmentID: assignment2.ID,
		Reviews:      []*pb.Review{},
		Scores:       []*score.Score{},
	}
	if err := db.CreateSubmission(&submission3); err != nil {
		t.Fatal(err)
	}

	// Even if there is three submission, only the latest for each assignment should be returned

	submissions, err := db.GetLastSubmissions(c1.ID, &pb.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	want := []*pb.Submission{&submission2, &submission3}
	if !reflect.DeepEqual(submissions, want) {
		fmt.Println("Submissions in the database:")
		for _, s := range submissions {
			fmt.Printf("%+v\n", s)
		}
		fmt.Println("Expected submissions:")
		for _, s := range want {
			fmt.Printf("%+v\n", s)
		}
		t.Errorf("have %#v want %#v", submissions, want)
	}
	data, err := db.GetLastSubmissions(c1.ID, &pb.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	} else if len(data) != 2 {
		t.Errorf("Expected '%v' elements in the array, got '%v'", 2, len(data))
	}
	// Since there is no submissions, but the course and user exist, an empty array should be returned
	data, err = db.GetLastSubmissions(c2.ID, &pb.Submission{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	} else if len(data) != 0 {
		t.Errorf("Expected '%v' elements in the array, got '%v'", 0, len(data))
	}
}
