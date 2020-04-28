package web

import (
	"context"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/assignments"
	"github.com/autograde/aguis/scm"
)

// getAssignments lists the assignments for the provided course.
func (s *AutograderService) getAssignments(courseID uint64) (*pb.Assignments, error) {
	allAssignments, err := s.db.GetAssignmentsByCourse(courseID, true)
	if err != nil {
		return nil, err
	}
	// Hack to ensure that assignments stored in database with wrong format
	// is displayed correctly in the frontend. This should ideally be removed
	// when the database no longer contains any incorrectly formatted dates.
	for _, assignment := range allAssignments {
		assignment.Deadline = assignments.FixDeadline(assignment.GetDeadline())
	}
	return &pb.Assignments{Assignments: allAssignments}, nil
}

// updateAssignments updates the assignments for the given course.
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

func (s *AutograderService) createBenchmark(query *pb.GradingBenchmark) (*pb.GradingBenchmark, error) {
	_, err := s.db.GetAssignment(&pb.Assignment{
		ID: query.AssignmentID,
	})
	if err != nil {
		return nil, err
	}
	err = s.db.CreateBenchmark(query)
	if err != nil {
		return nil, err
	}
	s.logger.Debugf("created a benchmark, make sure ID is set now: %+v", query)
	return query, nil
}

func (s *AutograderService) updateBenchmark(query *pb.GradingBenchmark) error {
	return s.db.UpdateBenchmark(query)
}

func (s *AutograderService) deleteBenchmark(query *pb.GradingBenchmark) error {
	return s.db.DeleteBenchmark(query)
}

func (s *AutograderService) createCriterion(query *pb.GradingCriterion) (*pb.GradingCriterion, error) {
	err := s.db.CreateCriterion(query)
	if err != nil {
		return nil, err
	}
	s.logger.Debugf("created a criterion, make sure ID is set now: %+v", query)
	return query, nil
}

func (s *AutograderService) updateCriterion(query *pb.GradingCriterion) error {
	return s.db.UpdateCriterion(query)
}

func (s *AutograderService) deleteCriterion(query *pb.GradingCriterion) error {
	return s.db.DeleteCriterion(query)
}
