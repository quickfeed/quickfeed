package qf

import (
	"time"
)

func (s *Submission) IsApproved(userID uint64) bool {
	for _, grade := range s.GetGrades() {
		if grade.GetUserID() == userID && grade.GetStatus() == Submission_APPROVED {
			return true
		}
	}
	return false
}

func (s *Submission) IsAllApproved() bool {
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

func (s *Submission) SetGrade(userID uint64, status Submission_Status) {
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
func (s *Submission) SetGradesIfApproved(a *Assignment, score uint32) {
	if a.GetAutoApprove() && score >= a.GetScoreLimit() {
		s.SetGradeAll(Submission_APPROVED)
	}
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

// Clean removes any score or reviews from the submission if it is not released.
// This is to prevent users from seeing the score or reviews of a submission that has not been released.
func (s *Submissions) Clean(userID uint64) {
	for _, submission := range s.GetSubmissions() {
		// Group submissions may have multiple grades, so we need to filter the grades by the user.
		submission.Grades = []*Grade{{
			UserID:       userID,
			SubmissionID: submission.GetID(),
			Status:       submission.GetStatusByUser(userID),
		}}
		// Released submissions, or submissions with no reviews need no cleaning.
		if submission.GetReleased() || len(submission.GetReviews()) == 0 {
			continue
		}
		// Remove any score, grades, or reviews if the submission is not released.
		submission.Score = 0
		submission.Grades = nil
		submission.Reviews = nil
	}
}
