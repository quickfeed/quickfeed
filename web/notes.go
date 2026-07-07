package web

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
)

// CreateNote creates a new internal staff note attached to a submission, group, or enrollment.
// The author and timestamps are set server-side; the access control interceptor restricts this to teachers.
func (s *QuickFeedService) CreateNote(ctx context.Context, in *qf.NoteRequest) (*qf.Note, error) {
	if err := checkNoteBody(in.GetNote().GetBody()); err != nil {
		return nil, err
	}
	if !hasSingleTarget(in.GetNote()) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("note must reference exactly one submission, group, or enrollment"))
	}
	courseID := in.GetCourseID()
	// The interceptor only verifies the caller teaches courseID, not that the
	// note's target lives in that course; reject cross-course targets here.
	if !s.noteTargetInCourse(courseID, in.GetNote()) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("note target does not belong to the course"))
	}
	// Build the note from only the fields a client may set; ID, author, course,
	// and timestamps are server-owned and must not be taken from the request.
	note := &qf.Note{
		CourseID:     courseID,
		AuthorID:     userID(ctx),
		Body:         in.GetNote().GetBody(),
		SubmissionID: in.GetNote().GetSubmissionID(),
		GroupID:      in.GetNote().GetGroupID(),
		EnrollmentID: in.GetNote().GetEnrollmentID(),
	}
	if err := s.db.CreateNote(note); err != nil {
		s.logger.Errorf("CreateNote failed for note %+v: %v", note, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create note"))
	}
	return note, nil
}

// UpdateNote updates the body of an existing note.
// Only the note's author or a site administrator may update it.
func (s *QuickFeedService) UpdateNote(ctx context.Context, in *qf.NoteRequest) (*qf.Note, error) {
	existing, err := s.authorizeNote(ctx, in)
	if err != nil {
		return nil, err
	}
	body := in.GetNote().GetBody()
	if err := checkNoteBody(body); err != nil {
		return nil, err
	}
	existing.Body = body
	if err := s.db.UpdateNote(existing); err != nil {
		s.logger.Errorf("UpdateNote failed for note %+v: %v", existing, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update note"))
	}
	updated, err := s.db.GetNote(&qf.Note{ID: existing.GetID()})
	if err != nil {
		s.logger.Errorf("UpdateNote failed to reload note %d: %v", existing.GetID(), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to update note"))
	}
	return updated, nil
}

// DeleteNote removes an existing note.
// Only the note's author or a site administrator may delete it.
func (s *QuickFeedService) DeleteNote(ctx context.Context, in *qf.NoteRequest) (*qf.Void, error) {
	existing, err := s.authorizeNote(ctx, in)
	if err != nil {
		return nil, err
	}
	if err := s.db.DeleteNote(&qf.Note{ID: existing.GetID()}); err != nil {
		s.logger.Errorf("DeleteNote failed for note %d: %v", existing.GetID(), err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to delete note"))
	}
	return &qf.Void{}, nil
}

// GetNotes returns all internal notes relevant to the requested target.
func (s *QuickFeedService) GetNotes(_ context.Context, in *qf.NotesRequest) (*qf.Notes, error) {
	notes, err := s.db.GetNotes(in.GetCourseID(), in.GetSubmissionID(), in.GetGroupID(), in.GetEnrollmentID())
	if err != nil {
		s.logger.Errorf("GetNotes failed for request %+v: %v", in, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get notes"))
	}
	return &qf.Notes{Notes: notes}, nil
}

// GetCourseNotes returns all internal notes for a course, used by staff
// overviews such as the members page to show per-student notes.
func (s *QuickFeedService) GetCourseNotes(_ context.Context, in *qf.CourseRequest) (*qf.Notes, error) {
	notes, err := s.db.GetNotes(in.GetCourseID(), 0, 0, 0)
	if err != nil {
		s.logger.Errorf("GetCourseNotes failed for course %d: %v", in.GetCourseID(), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get notes"))
	}
	return &qf.Notes{Notes: notes}, nil
}

// authorizeNote loads the note referenced by the request and verifies that it
// belongs to the request's course and that the caller is its author or an admin.
func (s *QuickFeedService) authorizeNote(ctx context.Context, in *qf.NoteRequest) (*qf.Note, error) {
	existing, err := s.db.GetNote(&qf.Note{ID: in.GetNote().GetID()})
	if err != nil {
		if errors.Is(err, database.ErrEmptyNoteID) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("note ID is required"))
		}
		return nil, connect.NewError(connect.CodeNotFound, errors.New("note not found"))
	}
	if existing.GetCourseID() != in.GetCourseID() {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("note does not belong to the course"))
	}

	// TODO(jostein): Currently the interceptor only checks that the caller is a teacher in the course, so any call that ends up here is from a teacher in the course.
	// This means that only admins that are *also teachers* in the course can update notes that they did not author.
	// If we want to allow *any admins* that are not also teachers in the course to update notes, we would need to change the interceptor to allow that.
	// This can be done by adding a "checkTeacherOrAdmin" to the access control interceptor.
	if existing.GetAuthorID() != userID(ctx) && !isAdmin(ctx) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("only the note's author or an administrator may modify it"))
	}
	return existing, nil
}

// noteTargetInCourse reports whether the note's single target (submission,
// group, or enrollment) belongs to the given course. A lookup failure is
// treated as "not in course" so a note is never attached to an entity the
// caller's course does not own.
func (s *QuickFeedService) noteTargetInCourse(courseID uint64, note *qf.Note) bool {
	switch {
	case note.GetSubmissionID() > 0:
		submission, err := s.db.GetSubmission(&qf.Submission{ID: note.GetSubmissionID()})
		if err != nil {
			return false
		}
		// A submission belongs to the course iff its assignment does.
		_, err = s.db.GetAssignment(&qf.Assignment{ID: submission.GetAssignmentID(), CourseID: courseID})
		return err == nil
	case note.GetGroupID() > 0:
		group, err := s.db.GetGroup(note.GetGroupID())
		return err == nil && group.GetCourseID() == courseID
	case note.GetEnrollmentID() > 0:
		enrollment, err := s.db.GetEnrollmentByID(note.GetEnrollmentID())
		return err == nil && enrollment.GetCourseID() == courseID
	}
	return false
}

// checkNoteBody rejects notes whose body is empty or only whitespace, so a blank
// note is never persisted even when the RPC is called directly, bypassing the UI.
func checkNoteBody(body string) error {
	if strings.TrimSpace(body) == "" {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("note body must not be empty"))
	}
	return nil
}

// hasSingleTarget returns true if the note references exactly one of a
// submission, group, or enrollment.
func hasSingleTarget(note *qf.Note) bool {
	targets := 0
	if note.GetSubmissionID() > 0 {
		targets++
	}
	if note.GetGroupID() > 0 {
		targets++
	}
	if note.GetEnrollmentID() > 0 {
		targets++
	}
	return targets == 1
}
