package ag

import (
	"errors"
	"time"
)

var ErrMissingBuildInfo = errors.New("submission missing build information")

func (s *Submission) IsApproved() bool {
	return s.GetStatus() == Submission_APPROVED
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
