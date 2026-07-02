package database_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

func TestCreateNote(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user, course, assignment := qtest.SetupCourseAssignment(t, db)
	teacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, teacher, course)

	submission := &qf.Submission{AssignmentID: assignment.GetID(), UserID: user.GetID()}
	qtest.CreateSubmission(t, db, submission)

	note := &qf.Note{
		CourseID:     course.GetID(),
		AuthorID:     teacher.GetID(),
		SubmissionID: submission.GetID(),
		Body:         "fix issue B before approval",
	}
	if err := db.CreateNote(note); err != nil {
		t.Fatal(err)
	}
	if note.GetID() == 0 {
		t.Fatal("expected note ID to be set")
	}
	if note.GetCreatedAt() == nil || note.GetEditedAt() == nil {
		t.Error("expected created and edited timestamps to be set")
	}

	got, err := db.GetNote(&qf.Note{ID: note.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if got.GetBody() != note.GetBody() {
		t.Errorf("GetNote body = %q, want %q", got.GetBody(), note.GetBody())
	}
}

// TestGetNotesForSubmission verifies that fetching notes for a submission
// returns the submission's own notes as well as the associated group and
// enrollment notes.
func TestGetNotesForSubmission(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user, course, assignment := qtest.SetupCourseAssignment(t, db)
	teacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, teacher, course)
	enrollment := qtest.GetEnrollment(t, db, user.GetID(), course.GetID())

	group := qtest.CreateGroup(t, db, &qf.Group{
		CourseID: course.GetID(),
		Name:     "group1",
		Users:    []*qf.User{user},
	})

	submission := &qf.Submission{AssignmentID: assignment.GetID(), GroupID: group.GetID()}
	qtest.CreateSubmission(t, db, submission)

	notes := []*qf.Note{
		{CourseID: course.GetID(), AuthorID: teacher.GetID(), SubmissionID: submission.GetID(), Body: "submission note"},
		{CourseID: course.GetID(), AuthorID: teacher.GetID(), GroupID: group.GetID(), Body: "group note"},
		{CourseID: course.GetID(), AuthorID: teacher.GetID(), EnrollmentID: enrollment.GetID(), Body: "enrollment note"},
	}
	for _, n := range notes {
		if err := db.CreateNote(n); err != nil {
			t.Fatal(err)
		}
	}
	// A note on an unrelated group must not surface.
	otherGroup := qtest.CreateGroup(t, db, &qf.Group{
		CourseID: course.GetID(),
		Name:     "group2",
		Users:    []*qf.User{teacher},
	})
	if err := db.CreateNote(&qf.Note{CourseID: course.GetID(), AuthorID: teacher.GetID(), GroupID: otherGroup.GetID(), Body: "other group"}); err != nil {
		t.Fatal(err)
	}

	got, err := db.GetNotes(course.GetID(), submission.GetID(), 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("GetNotes returned %d notes, want 3", len(got))
	}
}

func TestUpdateAndDeleteNote(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user, course, assignment := qtest.SetupCourseAssignment(t, db)
	teacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, teacher, course)

	submission := &qf.Submission{AssignmentID: assignment.GetID(), UserID: user.GetID()}
	qtest.CreateSubmission(t, db, submission)

	note := &qf.Note{CourseID: course.GetID(), AuthorID: teacher.GetID(), SubmissionID: submission.GetID(), Body: "before"}
	if err := db.CreateNote(note); err != nil {
		t.Fatal(err)
	}

	note.Body = "after"
	if err := db.UpdateNote(note); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetNote(&qf.Note{ID: note.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if got.GetBody() != "after" {
		t.Errorf("UpdateNote body = %q, want %q", got.GetBody(), "after")
	}

	if err := db.DeleteNote(&qf.Note{ID: note.GetID()}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetNote(&qf.Note{ID: note.GetID()}); err == nil {
		t.Error("expected error fetching deleted note")
	}
}
