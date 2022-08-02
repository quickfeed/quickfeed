package web

import (
	"fmt"
	"time"

	"github.com/quickfeed/quickfeed/assignments"
	"github.com/quickfeed/quickfeed/qf"
)

const reviewLayout = "02 Jan 15:04"

// getAssignments lists the assignments for the provided course.
func (s *QuickFeedService) getAssignments(courseID uint64) (*qf.Assignments, error) {
	allAssignments, err := s.db.GetAssignmentsByCourse(courseID, true)
	if err != nil {
		return nil, err
	}
	return &qf.Assignments{Assignments: allAssignments}, nil
}

// updateAssignments updates the assignments for the given course.
func (s *QuickFeedService) updateAssignments(courseID uint64) error {
	course, err := s.db.GetCourse(courseID, false)
	if err != nil {
		return fmt.Errorf("could not find course ID %d", courseID)
	}
	assignments.UpdateFromTestsRepo(s.logger, s.db, course)
	return nil
}

func (s *QuickFeedService) createBenchmark(query *qf.GradingBenchmark) (*qf.GradingBenchmark, error) {
	if _, err := s.db.GetAssignment(&qf.Assignment{
		ID: query.AssignmentID,
	}); err != nil {
		return nil, err
	}
	if err := s.db.CreateBenchmark(query); err != nil {
		return nil, err
	}
	return query, nil
}

func (s *QuickFeedService) updateBenchmark(query *qf.GradingBenchmark) error {
	return s.db.UpdateBenchmark(query)
}

func (s *QuickFeedService) deleteBenchmark(query *qf.GradingBenchmark) error {
	return s.db.DeleteBenchmark(query)
}

func (s *QuickFeedService) createCriterion(query *qf.GradingCriterion) (*qf.GradingCriterion, error) {
	if err := s.db.CreateCriterion(query); err != nil {
		return nil, err
	}
	return query, nil
}

func (s *QuickFeedService) updateCriterion(query *qf.GradingCriterion) error {
	return s.db.UpdateCriterion(query)
}

func (s *QuickFeedService) deleteCriterion(query *qf.GradingCriterion) error {
	return s.db.DeleteCriterion(query)
}

func (s *QuickFeedService) createReview(review *qf.Review) (*qf.Review, error) {
	submission, err := s.db.GetSubmission(&qf.Submission{ID: review.SubmissionID})
	if err != nil {
		return nil, err
	}
	assignment, err := s.db.GetAssignment(&qf.Assignment{ID: submission.AssignmentID})
	if err != nil {
		return nil, err
	}
	if len(submission.Reviews) >= int(assignment.Reviewers) {
		return nil, fmt.Errorf("failed to create a new review for submission %d to assignment %s: all %d reviews already created",
			submission.ID, assignment.Name, assignment.Reviewers)
	}
	review.Edited = time.Now().Format(reviewLayout)
	review.ComputeScore()

	benchmarks, err := s.db.GetBenchmarks(&qf.Assignment{ID: submission.AssignmentID})
	if err != nil {
		return nil, err
	}

	review.GradingBenchmarks = benchmarks
	for _, bm := range review.GradingBenchmarks {
		bm.ID = 0
		for _, c := range bm.Criteria {
			c.ID = 0
		}
	}
	if err := s.db.CreateReview(review); err != nil {
		return nil, err
	}

	return review, nil
}

func (s *QuickFeedService) updateReview(review *qf.Review) (*qf.Review, error) {
	if review.ID == 0 {
		return nil, fmt.Errorf("cannot update review with empty ID")
	}
	submission, err := s.db.GetSubmission(&qf.Submission{ID: review.SubmissionID})
	if err != nil {
		return nil, err
	}

	review.Edited = time.Now().Format(reviewLayout)
	review.ComputeScore()

	if err := s.db.UpdateReview(review); err != nil {
		return nil, err
	}

	for _, bm := range review.GradingBenchmarks {
		if err := s.db.UpdateBenchmark(bm); err != nil {
			return nil, err
		}
		for _, c := range bm.Criteria {
			if err := s.db.UpdateCriterion(c); err != nil {
				return nil, err
			}
		}
	}
	// Update the submission's score if the review score has changed.
	if submission.Score != review.Score {
		submission.Score = review.Score
		if err := s.db.UpdateSubmission(submission); err != nil {
			return nil, err
		}
	}
	return review, nil
}

func (s *QuickFeedService) getAssignmentWithCourse(query *qf.Assignment, withCourseInfo bool) (*qf.Assignment, *qf.Course, error) {
	assignment, err := s.db.GetAssignment(query)
	if err != nil {
		return nil, nil, err
	}
	course, err := s.db.GetCourse(assignment.CourseID, withCourseInfo)
	if err != nil {
		return nil, nil, err
	}
	return assignment, course, nil
}
