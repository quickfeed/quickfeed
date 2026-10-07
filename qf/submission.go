package qf

import (
	"time"

	"google.golang.org/protobuf/proto"
)

// Failed reports whether the submission's latest test run failed.
func (s *Submission) Failed() bool {
	return s.GetBuildInfo().Failed()
}

func (s *Submission) IsApproved(userID uint64) bool {
	for _, grade := range s.GetGrades() {
		if grade.GetUserID() == userID && grade.GetStatus() == Submission_APPROVED {
			return true
		}
	}
	return false
}

func (s *Submission) IsAllApproved() bool {
	if len(s.GetGrades()) == 0 {
		return false
	}
	for _, grade := range s.GetGrades() {
		if grade.GetStatus() != Submission_APPROVED {
			return false
		}
	}
	return true
}

func (s *Submission) GetStatuses() []Submission_Status {
	statuses := make([]Submission_Status, len(s.GetGrades()))
	for idx, grade := range s.GetGrades() {
		statuses[idx] = grade.GetStatus()
	}
	return statuses
}

func (s *Submission) GetStatusByUser(userID uint64) Submission_Status {
	for idx, grade := range s.GetGrades() {
		if grade.GetUserID() == userID {
			return s.GetGrades()[idx].GetStatus()
		}
	}
	return Submission_NONE
}

// SetGrade sets the submission's grade and score.
// If UserID is 0, the grade is set for all users of a group submission.
func (s *Submission) SetGrade(grade *Grade) {
	if grade.GetUserID() == 0 {
		// Set grade for all users if no specific user is provided.
		s.SetGradeAll(grade.GetStatus())
		return
	}
	s.SetGradeByUser(grade.GetUserID(), grade.GetStatus())
}

func (s *Submission) SetGradeByUser(userID uint64, status Submission_Status) {
	for idx, grade := range s.GetGrades() {
		if grade.GetUserID() == userID {
			s.GetGrades()[idx].Status = status
			return
		}
	}
}

func (s *Submission) SetGradeAll(status Submission_Status) {
	for idx := range s.GetGrades() {
		s.GetGrades()[idx].Status = status
	}
}

// SetGradesIfApproved marks the submission approved for all group members
// or a single user if the assignment is autoapprove and
// the score is greater or equal to the assignment's score limit.
// A submission to a reviewed assignment is approved only when its reviews are complete.
func (s *Submission) SetGradesIfApproved(a *Assignment, score uint32) {
	if a.GetAutoApprove() && score >= a.GetScoreLimit() && (!a.GradedManually() || s.ReviewsComplete(a)) {
		s.SetGradeAll(Submission_APPROVED)
	}
}

// ReviewsComplete returns true if the submission has as many reviews as the
// assignment has reviewers, and every review is complete.
func (s *Submission) ReviewsComplete(a *Assignment) bool {
	if len(s.GetReviews()) < int(a.GetReviewers()) {
		return false
	}
	for _, review := range s.GetReviews() {
		if !review.Complete() {
			return false
		}
	}
	return true
}

// ComputeScore sets the submission's review score to the average percentage
// score of its reviews, and its score to the test and review scores weighted
// by the assignment's review weight. Reviews count only if the assignment has reviewers.
func (s *Submission) ComputeScore(a *Assignment) {
	var sum uint64
	for _, review := range s.GetReviews() {
		sum += uint64(review.PercentScore())
	}
	s.ReviewScore = 0
	if n := uint64(len(s.GetReviews())); n > 0 {
		s.ReviewScore = uint32(sum / n)
	}
	s.Score = a.WeightedScore(s.GetTestScore(), s.GetReviewScore())
}

// NewestSubmissionDate returns the submission's submission date if newer than the provided date.
// Otherwise, the provided date is returned, i.e., if it is newer.
func (s *Submission) NewestSubmissionDate(submissionDate time.Time) time.Time {
	currentSubmissionDate := s.GetBuildInfo().GetSubmissionDate().AsTime()
	if currentSubmissionDate.After(submissionDate) {
		return currentSubmissionDate
	}
	return submissionDate
}

func (s *Submission) ByUser(userID uint64) bool {
	return s.GetGroupID() == 0 && s.GetUserID() > 0 && s.GetUserID() == userID
}

func (s *Submission) ByGroup(groupID uint64) bool {
	return s.GetUserID() == 0 && s.GetGroupID() > 0 && s.GetGroupID() == groupID
}

// Clean removes any review score or reviews from the submission to prevent
// a student from seeing them before the submission has been graded.
// The student still sees the test score.
func (s *Submissions) Clean(userID uint64) {
	for _, submission := range s.GetSubmissions() {
		submission.clean(userID)
	}
}

// CleanCopy returns a copy of the submission as the given user may see it; see Submissions.Clean.
func (s *Submission) CleanCopy(userID uint64) *Submission {
	submission := proto.CloneOf(s)
	submission.clean(userID)
	return submission
}

func (s *Submission) clean(userID uint64) {
	// Group submissions may have multiple grades, so we need to filter the grades by the user.
	grade := &Grade{
		UserID:       userID,
		SubmissionID: s.GetID(),
		Status:       s.GetStatusByUser(userID),
	}
	s.Grades = []*Grade{grade}
	if grade.GetStatus() != Submission_NONE {
		return
	}
	// Remove the review score if the submission has not been graded,
	// and also the grades and reviews if it has been reviewed.
	s.Score = s.GetTestScore()
	s.ReviewScore = 0
	if len(s.GetReviews()) > 0 {
		s.Grades = nil
		s.Reviews = nil
	}
}
