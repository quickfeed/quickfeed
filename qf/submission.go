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

// Clean removes what the given user may not see of the submissions before
// they are graded: the score and review score of submissions to the given
// assignments that have reviewers, and the reviews. The test score remains.
// A submission to an assignment not among the given ones is treated as reviewed.
func (s *Submissions) Clean(userID uint64, assignments []*Assignment) {
	byID := make(map[uint64]*Assignment, len(assignments))
	for _, assignment := range assignments {
		byID[assignment.GetID()] = assignment
	}
	for _, submission := range s.GetSubmissions() {
		assignment, ok := byID[submission.GetAssignmentID()]
		submission.clean(userID, !ok || assignment.GradedManually())
	}
}

// CleanCopy returns a copy of the submission to the given assignment as the
// given user may see it; see Submissions.Clean.
func (s *Submission) CleanCopy(userID uint64, assignment *Assignment) *Submission {
	submission := proto.CloneOf(s)
	submission.clean(userID, assignment.GradedManually())
	return submission
}

func (s *Submission) clean(userID uint64, reviewed bool) {
	// Group submissions may have multiple grades, so we need to filter the grades by the user.
	grade := &Grade{
		UserID:       userID,
		SubmissionID: s.GetID(),
		Status:       s.GetStatusByUser(userID),
	}
	s.Grades = []*Grade{grade}
	if grade.GetStatus() != Submission_NONE || !reviewed {
		return
	}
	// The score includes the review score, so both stay hidden until the
	// submission is graded, as do the grades and reviews once it has reviews.
	s.Score = 0
	s.ReviewScore = 0
	if len(s.GetReviews()) > 0 {
		s.Grades = nil
		s.Reviews = nil
	}
}
