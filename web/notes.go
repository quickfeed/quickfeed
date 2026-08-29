package web

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
)

// CreateNote creates a new internal staff note attached to a submission, group, or enrollment.
// The validation interceptor rejects blank bodies and notes without exactly one target;
// the author and timestamps are set server-side; the access control interceptor restricts this to teachers.
func (s *QuickFeedService) CreateNote(ctx context.Context, in *qf.Note) (*qf.Note, error) {
	courseID := in.GetCourseID()
	// The interceptor only verifies the caller teaches courseID, not that the
	// note's target lives in that course; reject cross-course targets here.
	if !s.noteTargetInCourse(courseID, in) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("note target does not belong to the course"))
	}
	// Build the note from only the fields a client may set; the ID, author, and
	// timestamps are server-owned and must not be taken from the request.
	note := &qf.Note{
		CourseID:     courseID,
		AuthorID:     userID(ctx),
		Body:         in.GetBody(),
		SubmissionID: in.GetSubmissionID(),
		GroupID:      in.GetGroupID(),
		EnrollmentID: in.GetEnrollmentID(),
	}
	if err := s.db.CreateNote(note); err != nil {
		qlog.FromContext(ctx).Error("failed to create note",
			label.SubmissionID, note.GetSubmissionID(), label.GroupID, note.GetGroupID(), "enrollment_id", note.GetEnrollmentID(), label.Error, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create note"))
	}
	return note, nil
}

// UpdateNote updates the body of an existing note.
// Only the note's author may update it.
func (s *QuickFeedService) UpdateNote(ctx context.Context, in *qf.Note) (*qf.Note, error) {
	existing, err := s.authorizeNote(ctx, in)
	if err != nil {
		return nil, err
	}
	existing.Body = in.GetBody()
	if err := s.db.UpdateNote(existing); err != nil {
		qlog.FromContext(ctx).Error("failed to update note", "note_id", existing.GetID(), label.Error, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update note"))
	}
	updated, err := s.db.GetNote(&qf.Note{ID: existing.GetID()})
	if err != nil {
		qlog.FromContext(ctx).Error("failed to reload updated note", "note_id", existing.GetID(), label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to update note"))
	}
	return updated, nil
}

// DeleteNote removes an existing note.
// Only the note's author may delete it.
func (s *QuickFeedService) DeleteNote(ctx context.Context, in *qf.Note) (*qf.Void, error) {
	existing, err := s.authorizeNote(ctx, in)
	if err != nil {
		return nil, err
	}
	if err := s.db.DeleteNote(&qf.Note{ID: existing.GetID()}); err != nil {
		qlog.FromContext(ctx).Error("failed to delete note", "note_id", existing.GetID(), label.Error, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to delete note"))
	}
	return &qf.Void{}, nil
}

// GetNotes returns all internal notes relevant to the requested target.
func (s *QuickFeedService) GetNotes(ctx context.Context, in *qf.NotesRequest) (*qf.Notes, error) {
	notes, err := s.db.GetNotes(in.GetCourseID(), in.GetSubmissionID(), in.GetGroupID(), in.GetEnrollmentID())
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get notes",
			label.SubmissionID, in.GetSubmissionID(), label.GroupID, in.GetGroupID(), "enrollment_id", in.GetEnrollmentID(), label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get notes"))
	}
	return &qf.Notes{Notes: notes}, nil
}

// GetCourseNotes returns all internal notes for a course, used by staff
// overviews such as the members page to show per-student notes.
func (s *QuickFeedService) GetCourseNotes(ctx context.Context, in *qf.CourseRequest) (*qf.Notes, error) {
	notes, err := s.db.GetNotes(in.GetCourseID(), 0, 0, 0)
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get course notes", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get notes"))
	}
	return &qf.Notes{Notes: notes}, nil
}

// authorizeNote loads the note referenced by the request and verifies that it
// belongs to the request's course and that the caller is its author.
func (s *QuickFeedService) authorizeNote(ctx context.Context, in *qf.Note) (*qf.Note, error) {
	existing, err := s.db.GetNote(&qf.Note{ID: in.GetID()})
	if err != nil {
		if errors.Is(err, database.ErrEmptyNoteID) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("note ID is required"))
		}
		return nil, connect.NewError(connect.CodeNotFound, errors.New("note not found"))
	}
	if existing.GetCourseID() != in.GetCourseID() {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("note does not belong to the course"))
	}
	if existing.GetAuthorID() != userID(ctx) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("only the note's author may modify it"))
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
