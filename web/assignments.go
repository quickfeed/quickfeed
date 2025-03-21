package web

import (
	"fmt"

	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *QuickFeedService) internalCreateReview(review *qf.Review) (*qf.Review, error) {
	submission, err := s.db.GetSubmission(&qf.Submission{ID: review.GetSubmissionID()})
	if err != nil {
		return nil, err
	}
	assignment, err := s.db.GetAssignment(&qf.Assignment{ID: submission.GetAssignmentID()})
	if err != nil {
		return nil, err
	}
	if len(submission.GetReviews()) >= int(assignment.GetReviewers()) {
		return nil, fmt.Errorf("failed to create a new review for submission %d to assignment %s: all %d reviews already created",
			submission.GetID(), assignment.GetName(), assignment.GetReviewers())
	}
	review.Edited = timestamppb.Now()
	review.ComputeScore()

	benchmarks, err := s.db.GetBenchmarks(&qf.Assignment{ID: submission.GetAssignmentID()})
	if err != nil {
		return nil, err
	}

	review.GradingBenchmarks = benchmarks
	for _, bm := range review.GetGradingBenchmarks() {
		bm.ID = 0
		for _, c := range bm.GetCriteria() {
			c.ID = 0
		}
	}
	if err := s.db.CreateReview(review); err != nil {
		return nil, err
	}

	return review, nil
}

func (s *QuickFeedService) internalUpdateReview(review *qf.Review) (*qf.Review, error) {
	if review.GetID() == 0 {
		return nil, fmt.Errorf("cannot update review with empty ID")
	}
	submission, err := s.db.GetSubmission(&qf.Submission{ID: review.GetSubmissionID()})
	if err != nil {
		return nil, err
	}

	review.Edited = timestamppb.Now()
	review.ComputeScore()

	if err := s.db.UpdateReview(review); err != nil {
		return nil, err
	}

	for _, bm := range review.GetGradingBenchmarks() {
		if err := s.db.UpdateBenchmark(bm); err != nil {
			return nil, err
		}
		for _, c := range bm.GetCriteria() {
			if err := s.db.UpdateCriterion(c); err != nil {
				return nil, err
			}
		}
	}
	// Update the submission's score if the review score has changed.
	if submission.GetScore() != review.GetScore() {
		submission.Score = review.GetScore()
		if err := s.db.UpdateSubmission(submission); err != nil {
			return nil, err
		}
	}
	return review, nil
}

func (s *QuickFeedService) getAssignmentWithCourse(query *qf.Assignment) (*qf.Assignment, *qf.Course, error) {
	assignment, err := s.db.GetAssignment(query)
	if err != nil {
		return nil, nil, err
	}
	course, err := s.db.GetCourse(assignment.GetCourseID())
	if err != nil {
		return nil, nil, err
	}
	return assignment, course, nil
}
