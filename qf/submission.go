package qf

import (
	"errors"
	"time"
)

var ErrMissingBuildInfo = errors.New("submission missing build information")

func (s *Submission) IsApproved() bool {
	return s.GetStatus() == Submission_APPROVED
}

// NewestSubmissionDate returns the submission's build date if newer than the provided submission date.
// Otherwise, the provided submission date is returned, i.e., if it is newer.
func (s *Submission) NewestSubmissionDate(submissionDate time.Time) (t time.Time, err error) {
	if s == nil || s.BuildInfo == nil || s.BuildInfo.BuildDate == nil || s.BuildInfo.SubmissionDate == nil {
		return t, ErrMissingBuildInfo
	}
	currentSubmissionDate := s.BuildInfo.SubmissionDate.AsTime()
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
func (s *Submissions) Clean() {
	for _, submission := range s.Submissions {
		// Released submissions, or submissions with no reviews need no cleaning.
		if submission.GetReleased() || len(submission.GetReviews()) == 0 {
			continue
		}
		// Remove any score, status, or reviews if the submission is not released.
		submission.Score = 0
		submission.Status = Submission_NONE
		submission.Reviews = nil
	}
}
