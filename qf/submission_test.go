package qf_test

import (
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
)

func TestNewestSubmissionDate(t *testing.T) {
	submission := &qf.Submission{}
	tim := time.Now()
	newBuildDate, err := submission.NewestSubmissionDate(tim)
	if err == nil {
		t.Errorf("NewestBuildDate(%v) = %v, expected error '%v'\n", tim, newBuildDate, qf.ErrMissingBuildInfo)
	}

	submission = &qf.Submission{}
	newBuildDate, err = submission.NewestSubmissionDate(tim)
	if err == nil {
		t.Errorf("NewestBuildDate(%v) = %v, expected error '%v'\n", tim, newBuildDate, qf.ErrMissingBuildInfo)
	}

	submission = &qf.Submission{
		BuildInfo: &score.BuildInfo{},
	}
	newBuildDate, err = submission.NewestSubmissionDate(tim)
	if err == nil {
		t.Errorf("NewestBuildDate(%v) = %v, expected error '%v'\n", tim, newBuildDate, qf.ErrMissingBuildInfo)
	}

	submission = &qf.Submission{
		BuildInfo: &score.BuildInfo{
			BuildDate: "string",
		},
	}
	newBuildDate, err = submission.NewestSubmissionDate(tim)
	if err == nil {
		t.Errorf("NewestBuildDate(%v) = %v, expected error '%v'\n", tim, newBuildDate, `parsing time "string" as "2006-01-02T15:04:05": cannot parse "string" as "2006"`)
	}

	buildDate := time.Now()
	submission = &qf.Submission{
		BuildInfo: &score.BuildInfo{
			BuildDate:      buildDate.Format(qf.TimeLayout),
			SubmissionDate: buildDate.Format(qf.TimeLayout),
		},
	}
	newBuildDate, err = submission.NewestSubmissionDate(tim)
	if err != nil {
		t.Error(err)
	}
	if newBuildDate.Before(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newBuildDate, buildDate)
	}
}

func TestByUser(t *testing.T) {
	submission := &qf.Submission{}
	if submission.ByUser(0) {
		t.Errorf("ByUser(0) = true, expected false\n")
	}

	submission = &qf.Submission{
		UserID: 1,
	}
	if !submission.ByUser(1) {
		t.Errorf("ByUser(1) = false, expected true\n")
	}

	submission = &qf.Submission{
		GroupID: 1,
	}
	if submission.ByUser(1) {
		t.Errorf("ByUser(1) = true, expected false\n")
	}

	// submissions with both user and group ID are invalid
	submission = &qf.Submission{
		UserID:  1,
		GroupID: 2,
	}
	if submission.ByUser(1) {
		t.Errorf("ByUser(1) = true, expected false\n")
	}
}

func TestByGroup(t *testing.T) {
	submission := &qf.Submission{}
	if submission.ByGroup(0) {
		t.Errorf("ByGroup(0) = true, expected false\n")
	}

	submission = &qf.Submission{
		GroupID: 1,
	}
	if !submission.ByGroup(1) {
		t.Errorf("ByGroup(1) = false, expected true\n")
	}

	submission = &qf.Submission{
		UserID: 1,
	}
	if submission.ByGroup(1) {
		t.Errorf("ByGroup(1) = true, expected false\n")
	}

	// submissions with both user and group ID are invalid
	submission = &qf.Submission{
		UserID:  1,
		GroupID: 2,
	}
	if submission.ByGroup(1) {
		t.Errorf("ByGroup(1) = true, expected false\n")
	}
}
