package database_test

import (
	"testing"

	pb "github.com/autograde/aguis/ag"
)

func TestGetNextAssignment(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	_, err := db.GetNextAssignment(0, 0, 0)
	if err == nil {
		t.Fatal("expected error 'record not found'")
	}

	course := pb.Course{
		Name:           "Distributed Systems",
		Code:           "DAT520",
		Year:           2018,
		Tag:            "Spring",
		Provider:       "fake",
		OrganizationID: 1,
	}

	// create course as teacher
	teacher := createFakeUser(t, db, 10)
	if err := db.CreateCourse(teacher.ID, &course); err != nil {
		t.Fatal(err)
	}

	// create and enroll user as student
	user := createFakeUser(t, db, 11)
	if err := db.CreateEnrollment(&pb.Enrollment{CourseID: course.ID, UserID: user.ID}); err != nil {
		t.Fatal(err)
	}
	if err = db.EnrollStudent(user.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	// create group with single student
	group := pb.Group{
		CourseID: course.ID,
		Users: []*pb.User{
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

	// create assignments for course
	assignment1 := pb.Assignment{CourseID: course.ID, Order: 1}
	if err := db.CreateAssignment(&assignment1); err != nil {
		t.Fatal(err)
	}
	assignment2 := pb.Assignment{CourseID: course.ID, Order: 2}
	if err := db.CreateAssignment(&assignment2); err != nil {
		t.Fatal(err)
	}
	assignment3 := pb.Assignment{CourseID: course.ID, Order: 3, IsGroupLab: true}
	if err := db.CreateAssignment(&assignment3); err != nil {
		t.Fatal(err)
	}
	assignment4 := pb.Assignment{CourseID: course.ID, Order: 4}
	if err := db.CreateAssignment(&assignment4); err != nil {
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
	if nxtUnapproved.ID != assignment1.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assignment1.ID, nxtUnapproved.ID)
	}

	// send new submission for assignment1
	submission1 := pb.Submission{AssignmentID: assignment1.ID, UserID: user.ID}
	if err := db.CreateSubmission(&submission1); err != nil {
		t.Fatal(err)
	}

	// send another submission for assignment1
	// will update the previous one, ID will stay the same
	submission2 := pb.Submission{AssignmentID: assignment1.ID, UserID: user.ID}
	if err := db.CreateSubmission(&submission2); err != nil {
		t.Fatal(err)
	}
	// send new submission for assignment2
	submission3 := pb.Submission{AssignmentID: assignment2.ID, UserID: user.ID}
	if err := db.CreateSubmission(&submission3); err != nil {
		t.Fatal(err)
	}
	// send new submission for assignment3
	submission4 := pb.Submission{AssignmentID: assignment3.ID, GroupID: group.ID}
	if err := db.CreateSubmission(&submission4); err != nil {
		t.Fatal(err)
	}

	// we haven't approved any of the submissions yet; expect same result as above
	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if nxtUnapproved.ID != assignment1.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assignment1.ID, nxtUnapproved.ID)
	}

	// approve submission for assignment1
	if err := db.UpdateSubmission(submission1.ID, true); err != nil {
		t.Fatal(err)
	}

	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, 0)
	if err != nil {
		t.Fatal(err)
	}

	if nxtUnapproved.ID != assignment2.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assignment2.ID, nxtUnapproved.ID)
	}

	// approve submission for assignment2
	if err := db.UpdateSubmission(submission3.ID, true); err != nil {
		t.Fatal(err)
	}

	// now the first two labs are approved,
	// the third one is unapproved and a group lab
	// if we only provide user ID, next unapproved must be assignment4 (group lab will be ignored)
	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if nxtUnapproved.ID != assignment4.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assignment4.ID, nxtUnapproved.ID)
	}

	// for the group next unapproved should be assignment3 - the group lab
	nxtUnapproved, err = db.GetNextAssignment(course.ID, 0, group.ID)
	if err != nil {
		t.Fatal(err)
	}
	if nxtUnapproved.ID != assignment3.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assignment3.ID, nxtUnapproved.ID)
	}

	// must also return assignment3 if both user and group IDs are provided
	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, group.ID)
	if err != nil {
		t.Fatal(err)
	}
	if nxtUnapproved.ID != assignment3.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assignment3.ID, nxtUnapproved.ID)
	}

	// approve submission for assignment3 (the group lab)
	if err := db.UpdateSubmission(submission4.ID, true); err != nil {
		t.Fatal(err)
	}

	// now next unapproved must be assignment4 for user, and all approved for group
	nxtUnapprovedForUser, err := db.GetNextAssignment(course.ID, user.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if nxtUnapprovedForUser.ID != assignment4.ID {
		t.Errorf("expected unapproved assignment to be %v, got %v", assignment4.ID, nxtUnapproved.ID)
	}
	nxtUnapprovedForGroup, err := db.GetNextAssignment(course.ID, 0, group.ID)
	if nxtUnapprovedForGroup != nil || err == nil {
		t.Fatal("expected error 'all assignments approved'")
	}
	// then create and approve submission for assignment4
	submission5 := &pb.Submission{AssignmentID: assignment4.ID, UserID: user.ID}
	if err := db.CreateSubmission(submission5); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateSubmission(submission5.ID, true); err != nil {
		t.Fatal(err)
	}

	// all assignments have been approved
	nxtUnapproved, err = db.GetNextAssignment(course.ID, user.ID, group.ID)
	if nxtUnapproved != nil || err == nil {
		t.Fatal("expected error 'all assignments approved'")
	}
}
