package web

import (
	"context"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/assignments"
	"github.com/autograde/aguis/scm"
)

// getAssignments lists the assignments for the provided course.
func (s *AutograderService) getAssignments(courseID uint64) (*pb.Assignments, error) {
	assignments, err := s.db.GetAssignmentsByCourse(courseID)
	if err != nil {
		return nil, err
	}
	return &pb.Assignments{Assignments: assignments}, nil
}

// updateAssignments updates the assignments for the given course
func (s *AutograderService) updateAssignments(ctx context.Context, sc scm.SCM, courseID uint64) error {
	course, err := s.db.GetCourse(courseID, false)
	if err != nil {
		return err
	}
	assignments, err := assignments.FetchAssignments(ctx, sc, course)
	if err != nil {
		return err
	}
	if err = s.db.UpdateAssignments(assignments); err != nil {
		return err
	}
	return nil
}
