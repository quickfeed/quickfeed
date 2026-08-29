package web_test

import (
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCreateNoteAccess(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"), web.WithInterceptors())
	_, course, assignment, student := qtest.SetupCourseAssignmentTeacherStudent(t, db)
	teacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, teacher, course)

	submission := &qf.Submission{AssignmentID: assignment.GetID(), UserID: student.GetID()}
	qtest.CreateSubmission(t, db, submission)

	req := &qf.Note{CourseID: course.GetID(), SubmissionID: submission.GetID(), Body: "fix issue B"}

	// A student must not be able to create a note.
	if _, err := client.CreateNote(client.Context(t, student), req); !qtest.CheckCode(t, err, connect.NewError(connect.CodePermissionDenied, errors.New("access denied for CreateNote: not teacher"))) {
		t.Errorf("student CreateNote() error = %v, want PermissionDenied", err)
	}

	// A teacher can create a note; the author is set server-side.
	note, err := client.CreateNote(client.Context(t, teacher), req)
	if err != nil {
		t.Fatalf("teacher CreateNote() unexpected error: %v", err)
	}
	if note.GetAuthorID() != teacher.GetID() {
		t.Errorf("note AuthorID = %d, want %d", note.GetAuthorID(), teacher.GetID())
	}
}

// TestNoteBodyRequired verifies that blank note bodies are rejected server-side
// on both create and update, independent of the UI's own trimming.
func TestNoteBodyRequired(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"), web.WithInterceptors())
	_, course, assignment, student := qtest.SetupCourseAssignmentTeacherStudent(t, db)
	teacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, teacher, course)

	submission := &qf.Submission{AssignmentID: assignment.GetID(), UserID: student.GetID()}
	qtest.CreateSubmission(t, db, submission)
	ctx := client.Context(t, teacher)

	emptyBodyErr := connect.NewError(connect.CodeInvalidArgument, errors.New("note body must not be empty"))

	// A whitespace-only body is rejected on create.
	if _, err := client.CreateNote(ctx, &qf.Note{
		CourseID: course.GetID(), SubmissionID: submission.GetID(), Body: "   ",
	}); !qtest.CheckCode(t, err, emptyBodyErr) {
		t.Errorf("CreateNote() with blank body error = %v, want %v", err, emptyBodyErr)
	}

	// Create a valid note, then attempt to blank it via update.
	note, err := client.CreateNote(ctx, &qf.Note{
		CourseID: course.GetID(), SubmissionID: submission.GetID(), Body: "original",
	})
	if err != nil {
		t.Fatalf("CreateNote() unexpected error: %v", err)
	}
	if _, err := client.UpdateNote(ctx, &qf.Note{
		CourseID: course.GetID(), ID: note.GetID(), Body: "",
	}); !qtest.CheckCode(t, err, emptyBodyErr) {
		t.Errorf("UpdateNote() with empty body error = %v, want %v", err, emptyBodyErr)
	}
}

// TestNoteSingleTargetRequired verifies that CreateNote rejects notes that
// reference no target or more than one target. A note naming targets in two
// courses must not slip past the cross-course check, which inspects only the
// first target it finds.
func TestNoteSingleTargetRequired(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"), web.WithInterceptors())
	_, course, assignment, student := qtest.SetupCourseAssignmentTeacherStudent(t, db)
	teacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, teacher, course)
	enrollment := qtest.GetEnrollment(t, db, student.GetID(), course.GetID())

	submission := &qf.Submission{AssignmentID: assignment.GetID(), UserID: student.GetID()}
	qtest.CreateSubmission(t, db, submission)
	ctx := client.Context(t, teacher)

	wantErr := connect.NewError(connect.CodeInvalidArgument, errors.New("note must reference exactly one submission, group, or enrollment"))
	for name, note := range map[string]*qf.Note{
		"no target": {CourseID: course.GetID(), Body: "note"},
		"two targets": {
			CourseID:     course.GetID(),
			SubmissionID: submission.GetID(),
			EnrollmentID: enrollment.GetID(),
			Body:         "note",
		},
	} {
		if _, err := client.CreateNote(ctx, note); !qtest.CheckCode(t, err, wantErr) {
			t.Errorf("CreateNote(%s) error = %v, want %v", name, err, wantErr)
		}
	}
}

// TestCreateNoteIgnoresServerOwnedFields verifies that client-supplied values
// for server-owned fields (ID, author, timestamps) are not trusted on create.
func TestCreateNoteIgnoresServerOwnedFields(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"), web.WithInterceptors())
	_, course, assignment, student := qtest.SetupCourseAssignmentTeacherStudent(t, db)
	teacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, teacher, course)

	submission := &qf.Submission{AssignmentID: assignment.GetID(), UserID: student.GetID()}
	qtest.CreateSubmission(t, db, submission)

	// The client attempts to dictate the primary key and author.
	note, err := client.CreateNote(client.Context(t, teacher), &qf.Note{
		ID:           999,
		CourseID:     course.GetID(),
		AuthorID:     student.GetID(),
		CreatedAt:    timestamppb.New(time.Unix(0, 0)),
		SubmissionID: submission.GetID(),
		Body:         "note body",
	})
	if err != nil {
		t.Fatalf("CreateNote() unexpected error: %v", err)
	}
	if note.GetID() == 999 {
		t.Error("CreateNote() used the client-supplied ID")
	}
	if note.GetAuthorID() != teacher.GetID() {
		t.Errorf("note AuthorID = %d, want the caller %d", note.GetAuthorID(), teacher.GetID())
	}
	if note.GetCreatedAt().AsTime().Unix() == 0 {
		t.Error("CreateNote() used the client-supplied CreatedAt")
	}
}

// TestUpdateDeleteNoteRequireID verifies that update/delete with a zero note ID
// are rejected instead of silently targeting an arbitrary note, and that an
// existing note is left untouched.
func TestUpdateDeleteNoteRequireID(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"), web.WithInterceptors())
	_, course, assignment, student := qtest.SetupCourseAssignmentTeacherStudent(t, db)
	teacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, teacher, course)

	submission := &qf.Submission{AssignmentID: assignment.GetID(), UserID: student.GetID()}
	qtest.CreateSubmission(t, db, submission)
	ctx := client.Context(t, teacher)

	if _, err := client.CreateNote(ctx, &qf.Note{
		CourseID: course.GetID(), SubmissionID: submission.GetID(), Body: "keep me",
	}); err != nil {
		t.Fatalf("CreateNote() unexpected error: %v", err)
	}

	missingIDErr := connect.NewError(connect.CodeInvalidArgument, errors.New("note ID is required"))

	// Update with a zero ID must not fall back to the first note in the table.
	if _, err := client.UpdateNote(ctx, &qf.Note{
		CourseID: course.GetID(), Body: "hijack",
	}); !qtest.CheckCode(t, err, missingIDErr) {
		t.Errorf("UpdateNote() with zero ID error = %v, want %v", err, missingIDErr)
	}

	// Delete with a zero ID must be rejected.
	if _, err := client.DeleteNote(ctx, &qf.Note{
		CourseID: course.GetID(),
	}); !qtest.CheckCode(t, err, missingIDErr) {
		t.Errorf("DeleteNote() with zero ID error = %v, want %v", err, missingIDErr)
	}

	// The existing note must be untouched.
	notes, err := client.GetNotes(ctx, &qf.NotesRequest{CourseID: course.GetID(), SubmissionID: submission.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if len(notes.GetNotes()) != 1 || notes.GetNotes()[0].GetBody() != "keep me" {
		t.Errorf("note was modified or removed: %+v", notes.GetNotes())
	}
}

// TestCreateNoteCrossCourse verifies that a teacher cannot attach a note to a
// target (submission, group, or enrollment) that belongs to a different course,
// even though the interceptor authorizes them for the request's course.
func TestCreateNoteCrossCourse(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"), web.WithInterceptors())

	// Course A: the teacher is authorized here.
	admin, courseA, _, _ := qtest.SetupCourseAssignmentTeacherStudent(t, db)
	teacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, teacher, courseA)

	// Course B: a separate course whose entities must not be reachable from A.
	courseB := &qf.Course{ScmOrganizationID: 2, Code: "B", Year: 2026}
	qtest.CreateCourse(t, db, admin, courseB)
	assignmentB := &qf.Assignment{CourseID: courseB.GetID(), Order: 1}
	qtest.CreateAssignment(t, db, assignmentB)
	studentB := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, studentB, courseB)
	enrollmentB := qtest.GetEnrollment(t, db, studentB.GetID(), courseB.GetID())
	submissionB := &qf.Submission{AssignmentID: assignmentB.GetID(), UserID: studentB.GetID()}
	qtest.CreateSubmission(t, db, submissionB)
	groupB := qtest.CreateGroup(t, db, &qf.Group{CourseID: courseB.GetID(), Name: "groupB", Users: []*qf.User{studentB}})

	ctx := client.Context(t, teacher)
	wantErr := connect.NewError(connect.CodePermissionDenied, errors.New("note target does not belong to the course"))
	for name, note := range map[string]*qf.Note{
		"submission": {SubmissionID: submissionB.GetID()},
		"group":      {GroupID: groupB.GetID()},
		"enrollment": {EnrollmentID: enrollmentB.GetID()},
	} {
		note.Body = "cross-course note"
		note.CourseID = courseA.GetID()
		if _, err := client.CreateNote(ctx, note); !qtest.CheckCode(t, err, wantErr) {
			t.Errorf("CreateNote(%s target from course B) error = %v, want PermissionDenied", name, err)
		}
	}
}

func TestUpdateAndDeleteNoteAuthorization(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"), web.WithInterceptors())
	admin, course, assignment, student := qtest.SetupCourseAssignmentTeacherStudent(t, db)
	author := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, author, course)
	otherTeacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, otherTeacher, course)

	submission := &qf.Submission{AssignmentID: assignment.GetID(), UserID: student.GetID()}
	qtest.CreateSubmission(t, db, submission)

	note, err := client.CreateNote(client.Context(t, author), &qf.Note{
		CourseID: course.GetID(), SubmissionID: submission.GetID(), Body: "original",
	})
	if err != nil {
		t.Fatalf("CreateNote() unexpected error: %v", err)
	}

	updateReq := func(body string) *qf.Note {
		return &qf.Note{CourseID: course.GetID(), ID: note.GetID(), Body: body}
	}

	// A different teacher may not update the note.
	if _, err := client.UpdateNote(client.Context(t, otherTeacher), updateReq("hijacked")); !qtest.CheckCode(t, err, connect.NewError(connect.CodePermissionDenied, errors.New("only the note's author may modify it"))) {
		t.Errorf("otherTeacher UpdateNote() error = %v, want PermissionDenied", err)
	}

	// The author may update the note.
	updated, err := client.UpdateNote(client.Context(t, author), updateReq("by author"))
	if err != nil {
		t.Fatalf("author UpdateNote() unexpected error: %v", err)
	}
	if updated.GetBody() != "by author" {
		t.Errorf("note body = %q, want %q", updated.GetBody(), "by author")
	}

	// An admin who is not the author may not update the note.
	if _, err := client.UpdateNote(client.Context(t, admin), updateReq("by admin")); !qtest.CheckCode(t, err, connect.NewError(connect.CodePermissionDenied, errors.New("only the note's author may modify it"))) {
		t.Errorf("admin UpdateNote() error = %v, want PermissionDenied", err)
	}

	// A non-author teacher may not delete the note.
	delReq := &qf.Note{CourseID: course.GetID(), ID: note.GetID()}
	if _, err := client.DeleteNote(client.Context(t, otherTeacher), delReq); !qtest.CheckCode(t, err, connect.NewError(connect.CodePermissionDenied, errors.New("only the note's author may modify it"))) {
		t.Errorf("otherTeacher DeleteNote() error = %v, want PermissionDenied", err)
	}

	// The author may delete the note.
	if _, err := client.DeleteNote(client.Context(t, author), delReq); err != nil {
		t.Fatalf("author DeleteNote() unexpected error: %v", err)
	}
}

func TestGetNotes(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"), web.WithInterceptors())
	_, course, assignment, student := qtest.SetupCourseAssignmentTeacherStudent(t, db)
	teacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, teacher, course)
	enrollment := qtest.GetEnrollment(t, db, student.GetID(), course.GetID())

	submission := &qf.Submission{AssignmentID: assignment.GetID(), UserID: student.GetID()}
	qtest.CreateSubmission(t, db, submission)

	ctx := client.Context(t, teacher)
	for _, note := range []*qf.Note{
		{CourseID: course.GetID(), SubmissionID: submission.GetID(), Body: "submission"},
		{CourseID: course.GetID(), EnrollmentID: enrollment.GetID(), Body: "enrollment"},
	} {
		if _, err := client.CreateNote(ctx, note); err != nil {
			t.Fatal(err)
		}
	}

	notes, err := client.GetNotes(ctx, &qf.NotesRequest{CourseID: course.GetID(), SubmissionID: submission.GetID()})
	if err != nil {
		t.Fatalf("GetNotes() unexpected error: %v", err)
	}
	if len(notes.GetNotes()) != 2 {
		t.Errorf("GetNotes() returned %d notes, want 2", len(notes.GetNotes()))
	}

	// A student must not be able to read notes.
	if _, err := client.GetNotes(client.Context(t, student), &qf.NotesRequest{CourseID: course.GetID(), SubmissionID: submission.GetID()}); !qtest.CheckCode(t, err, connect.NewError(connect.CodePermissionDenied, errors.New("access denied for GetNotes: not teacher"))) {
		t.Errorf("student GetNotes() error = %v, want PermissionDenied", err)
	}

	// GetCourseNotes returns every note in the course (both targets) for staff.
	courseNotes, err := client.GetCourseNotes(ctx, &qf.CourseRequest{CourseID: course.GetID()})
	if err != nil {
		t.Fatalf("GetCourseNotes() unexpected error: %v", err)
	}
	if len(courseNotes.GetNotes()) != 2 {
		t.Errorf("GetCourseNotes() returned %d notes, want 2", len(courseNotes.GetNotes()))
	}

	// A student must not be able to read course notes.
	if _, err := client.GetCourseNotes(client.Context(t, student), &qf.CourseRequest{CourseID: course.GetID()}); !qtest.CheckCode(t, err, connect.NewError(connect.CodePermissionDenied, errors.New("access denied for GetCourseNotes: not teacher"))) {
		t.Errorf("student GetCourseNotes() error = %v, want PermissionDenied", err)
	}
}
