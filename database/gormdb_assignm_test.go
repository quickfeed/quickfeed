package database_test

import (
	"testing"

	"github.com/autograde/aguis/models"
)

func TestGetNextAssignment(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	_, err := db.GetNextAssignment(0, 0, 0)
	if err == nil {
		t.Fatal("expected error 'record not found'")
	}

	// Create course
	course := models.Course{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryID: 1,
	}
	if err := db.CreateCourse(&course); err != nil {
		t.Fatal(err)
	}

	// Create and enroll user
	var user models.User
	if err := db.CreateUserFromRemoteIdentity(&user, &models.RemoteIdentity{}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&models.Enrollment{CourseID: course.ID, UserID: user.ID}); err != nil {
		t.Fatal(err)
	}
	if err = db.EnrollStudent(user.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	// Create group
	group := models.Group{
		CourseID: course.ID,
		Users: []*models.User{
			{ID: user.ID},
		},
	}
	if err := db.CreateGroup(&group); err != nil {
		t.Fatal(err)
	}

	_, err = db.GetNextAssignment(course.ID, user.ID, group.ID)
	if err == nil {
		t.Fatal("expected error 'no assignments found for course 1'")
	}

	// Create assignments
	assigment1 := models.Assignment{CourseID: course.ID, Order: 1}
	if err := db.CreateAssignment(&assigment1); err != nil {
		t.Fatal(err)
	}
	assigment2 := models.Assignment{CourseID: course.ID, Order: 2}
	if err := db.CreateAssignment(&assigment2); err != nil {
		t.Fatal(err)
	}
	assigment3 := models.Assignment{CourseID: course.ID, Order: 3, IsGroupLab: true}
	if err := db.CreateAssignment(&assigment3); err != nil {
		t.Fatal(err)
	}
	assigment4 := models.Assignment{CourseID: course.ID, Order: 4}
	if err := db.CreateAssignment(&assigment4); err != nil {
		t.Fatal(err)
	}

	_, err = db.GetNextAssignment(course.ID, 0, 0)
	if err == nil {
		t.Fatal("expected error 'record not found'")
	}

	nxtUnapproved, err := db.GetNextAssignment(course.ID, user.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if nxtUnapproved.ID != assigment1.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assigment1.ID, nxtUnapproved.ID)
	}

	// send new submission for assignment1
	submission1 := models.Submission{AssignmentID: assigment1.ID, UserID: user.ID}
	if err := db.CreateSubmission(&submission1); err != nil {
		t.Fatal(err)
	}
	// send another submission for assignment1
	submission2 := models.Submission{AssignmentID: assigment1.ID, UserID: user.ID}
	if err := db.CreateSubmission(&submission2); err != nil {
		t.Fatal(err)
	}
	// send new submission for assignment2
	submission3 := models.Submission{AssignmentID: assigment2.ID, UserID: user.ID}
	if err := db.CreateSubmission(&submission3); err != nil {
		t.Fatal(err)
	}
	// send new submission for assignment3
	// submission4 := models.Submission{AssignmentID: assigment3.ID, UserID: user.ID, GroupID: group.ID}
	submission4 := models.Submission{AssignmentID: assigment3.ID, GroupID: group.ID}
	if err := db.CreateSubmission(&submission4); err != nil {
		t.Fatal(err)
	}

	// we haven't approved any of the submissions yet; expect same result as above

	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if nxtUnapproved.ID != assigment1.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assigment1.ID, nxtUnapproved.ID)
	}

	// approve submission1
	if err := db.UpdateSubmissionByID(submission1.ID, true); err != nil {
		t.Fatal(err)
	}

	// we have approved the first submission of the first assignment, but since
	// we two submissions for assignment1, this won't change the next to approve.
	// TODO Is this the desired semantics for this??
	// That is, it seems more reasonable to have a function ApproveAssignment(assignment, user)
	// that finds the latest submission for the user and marks it approved.
	// That is, maybe the UpdateSubmissionByID shouldn't be exported.

	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if nxtUnapproved.ID != assigment1.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assigment1.ID, nxtUnapproved.ID)
	}

	// approve submission2
	if err := db.UpdateSubmissionByID(submission2.ID, true); err != nil {
		t.Fatal(err)
	}

	// now the first assignment is approved, moving on to the second

	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if nxtUnapproved.ID != assigment2.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assigment2.ID, nxtUnapproved.ID)
	}

	// approve submission3
	if err := db.UpdateSubmissionByID(submission3.ID, true); err != nil {
		t.Fatal(err)
	}

	// now the second assignment is approved, moving on to the third
	// this fails because the next assignment to approve is a group lab,
	// and we don't provide a group id.

	_, err = db.GetNextAssignment(course.ID, user.ID, 0)
	if err == nil {
		t.Fatal("expected error 'record not found'")
	}

	// moving on to the third assignment, using the group id this time.
	// fails because user id must be provided.

	_, err = db.GetNextAssignment(course.ID, 0, group.ID)
	if err == nil {
		t.Fatal("expected error 'user id must be provided'")
	}

	// moving on to the third assignment, using both user id and group id this time.

	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, group.ID)
	if err != nil {
		t.Fatal(err)
	}
	if nxtUnapproved.ID != assigment3.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assigment3.ID, nxtUnapproved.ID)
	}

	// approve submission4 for assignment3 (the group lab)
	if err := db.UpdateSubmissionByID(submission4.ID, true); err != nil {
		t.Fatal(err)
	}

	// approving the 4th submission (for assignment3, which is a group lab),
	// should fail because we only provide user id, and no group.ID.

	_, err = db.GetNextAssignment(course.ID, user.ID, 0)
	if err == nil {
		t.Fatal("expected error 'user id must be provided'")
	}

	// here it should pass since we also provide the group id.

	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, group.ID)
	if err != nil {
		t.Fatal(err)
	}
	if nxtUnapproved.ID != assigment4.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assigment4.ID, nxtUnapproved.ID)
	}

	// send new submission for assignment4
	submission5 := models.Submission{AssignmentID: assigment4.ID, UserID: user.ID}
	if err := db.CreateSubmission(&submission5); err != nil {
		t.Fatal(err)
	}
	// approve submission5
	if err := db.UpdateSubmissionByID(submission5.ID, true); err != nil {
		t.Fatal(err)
	}

	// all assignments have been approved

	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, group.ID)
	if nxtUnapproved != nil || err == nil {
		t.Fatal("expected error 'all assignments approved'")
	}
}
