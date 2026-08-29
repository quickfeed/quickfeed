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

	created := note.GetEditedAt().AsTime()
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
	// UpdateNote stamps the new edited time on the given note, so callers can
	// return it without reloading the row.
	if !note.GetEditedAt().AsTime().Equal(got.GetEditedAt().AsTime()) {
		t.Errorf("UpdateNote EditedAt = %v, want the stored %v", note.GetEditedAt().AsTime(), got.GetEditedAt().AsTime())
	}
	if !note.GetEditedAt().AsTime().After(created) {
		t.Errorf("UpdateNote EditedAt = %v, want after the creation time %v", note.GetEditedAt().AsTime(), created)
	}

	if err := db.DeleteNote(&qf.Note{ID: note.GetID()}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetNote(&qf.Note{ID: note.GetID()}); err == nil {
		t.Error("expected error fetching deleted note")
	}
}

// TestDeleteGroupDeletesGroupNotes verifies that notes attached to a group do
// not outlive the group; nothing enforces this with a foreign key constraint.
func TestDeleteGroupDeletesGroupNotes(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	student, course, assignment := qtest.SetupCourseAssignment(t, db)
	teacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, teacher, course)
	group := qtest.CreateGroup(t, db, &qf.Group{CourseID: course.GetID(), Name: "group", Users: []*qf.User{student}})

	submission := &qf.Submission{AssignmentID: assignment.GetID(), GroupID: group.GetID()}
	qtest.CreateSubmission(t, db, submission)

	groupNote := &qf.Note{CourseID: course.GetID(), AuthorID: teacher.GetID(), GroupID: group.GetID(), Body: "group note"}
	if err := db.CreateNote(groupNote); err != nil {
		t.Fatal(err)
	}
	// A note on the group's submission must survive; the submission does.
	submissionNote := &qf.Note{CourseID: course.GetID(), AuthorID: teacher.GetID(), SubmissionID: submission.GetID(), Body: "submission note"}
	if err := db.CreateNote(submissionNote); err != nil {
		t.Fatal(err)
	}

	if err := db.DeleteGroup(group.GetID()); err != nil {
		t.Fatal(err)
	}

	if _, err := db.GetNote(&qf.Note{ID: groupNote.GetID()}); err == nil {
		t.Error("expected the group's note to be deleted with the group")
	}
	if _, err := db.GetNote(&qf.Note{ID: submissionNote.GetID()}); err != nil {
		t.Errorf("expected the submission's note to survive: %v", err)
	}
}

// TestRejectEnrollmentDeletesEnrollmentNotes verifies that notes attached to an
// enrollment do not outlive the enrollment.
func TestRejectEnrollmentDeletesEnrollmentNotes(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	student, course, _ := qtest.SetupCourseAssignment(t, db)
	teacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, teacher, course)
	enrollment := qtest.GetEnrollment(t, db, student.GetID(), course.GetID())

	other := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, other, course)
	otherEnrollment := qtest.GetEnrollment(t, db, other.GetID(), course.GetID())

	note := &qf.Note{CourseID: course.GetID(), AuthorID: teacher.GetID(), EnrollmentID: enrollment.GetID(), Body: "student note"}
	if err := db.CreateNote(note); err != nil {
		t.Fatal(err)
	}
	otherNote := &qf.Note{CourseID: course.GetID(), AuthorID: teacher.GetID(), EnrollmentID: otherEnrollment.GetID(), Body: "other student note"}
	if err := db.CreateNote(otherNote); err != nil {
		t.Fatal(err)
	}

	if err := db.RejectEnrollment(student.GetID(), course.GetID()); err != nil {
		t.Fatal(err)
	}

	if _, err := db.GetNote(&qf.Note{ID: note.GetID()}); err == nil {
		t.Error("expected the enrollment's note to be deleted with the enrollment")
	}
	if _, err := db.GetNote(&qf.Note{ID: otherNote.GetID()}); err != nil {
		t.Errorf("expected another student's note to survive: %v", err)
	}
}
