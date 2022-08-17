package qf

import (
	"errors"
	"time"
)

var ErrMissingBuildInfo = errors.New("submission missing build information")

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

func (s *Submission) GetStatusByUser(userID uint64) Submission_Status {
	for idx, grade := range s.GetGrades() {
		if grade.GetUserID() == userID {
			return s.Grades[idx].GetStatus()
		}
	}
	return Submission_NONE
}

func (s *Submission) SetGrade(userID uint64, status Submission_Status) {
	for idx, grade := range s.GetGrades() {
		if grade.GetUserID() == userID {
			s.Grades[idx].Status = status
			return
		}
	}
}

func (s *Submission) SetGradeAll(status Submission_Status) {
	for idx := range s.Grades {
		s.Grades[idx].Status = status
	}
}

// NewestBuildDate returns the submission's build date if newer than the provided submission date.
// Otherwise, the provided submission date is returned, i.e., if it is newer.
func (s *Submission) NewestBuildDate(submissionDate time.Time) (t time.Time, err error) {
	if s == nil || s.BuildInfo == nil {
		return t, ErrMissingBuildInfo
	}
	currentSubmissionDate, err := time.Parse(TimeLayout, s.BuildInfo.BuildDate)
	if err != nil {
		return t, err
	}
	if currentSubmissionDate.After(submissionDate) {
		submissionDate = currentSubmissionDate
	}
	return submissionDate, nil
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
	for _, submission := range s.Submissions {
		// Only send the grade belonging to the user requesting the submission
		submission.Grades = []*Grade{{
			UserID:       userID,
			SubmissionID: submission.GetID(),
			Status:       submission.GetStatusByUser(userID),
		}}

		// Released submissions, or submissions with no reviews need no cleaning.
		if submission.GetReleased() || len(submission.GetReviews()) == 0 {
			continue
		}
		// Remove any score or reviews if the submission is not released.
		submission.Score = 0
		submission.Reviews = nil
	}
}
