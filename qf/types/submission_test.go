package types_test

import (
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf/types"
)

func TestNewestSubmissionDate(t *testing.T) {
	submission := &types.Submission{}
	tim := time.Now()
	newBuildDate, err := submission.NewestBuildDate(tim)
	if err == nil {
		t.Errorf("NewestBuildDate(%v) = %v, expected error '%v'\n", tim, newBuildDate, types.ErrMissingBuildInfo)
	}

	submission = &types.Submission{}
	newBuildDate, err = submission.NewestBuildDate(tim)
	if err == nil {
		t.Errorf("NewestBuildDate(%v) = %v, expected error '%v'\n", tim, newBuildDate, types.ErrMissingBuildInfo)
	}

	submission = &types.Submission{
		BuildInfo: &score.BuildInfo{},
	}
	newBuildDate, err = submission.NewestBuildDate(tim)
	if err == nil {
		t.Errorf("NewestBuildDate(%v) = %v, expected error '%v'\n", tim, newBuildDate, types.ErrMissingBuildInfo)
	}

	submission = &types.Submission{
		BuildInfo: &score.BuildInfo{
			BuildDate: "string",
		},
	}
	newBuildDate, err = submission.NewestBuildDate(tim)
	if err == nil {
		t.Errorf("NewestBuildDate(%v) = %v, expected error '%v'\n", tim, newBuildDate, `parsing time "string" as "2006-01-02T15:04:05": cannot parse "string" as "2006"`)
	}

	buildDate := time.Now()
	submission = &types.Submission{
		BuildInfo: &score.BuildInfo{
			BuildDate: buildDate.Format(types.TimeLayout),
		},
	}
	newBuildDate, err = submission.NewestBuildDate(tim)
	if err != nil {
		t.Error(err)
	}
	if newBuildDate.Before(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newBuildDate, buildDate)
	}
}

func TestByUser(t *testing.T) {
	submission := &types.Submission{}
	if submission.ByUser(0) {
		t.Errorf("ByUser(0) = true, expected false\n")
	}

	submission = &types.Submission{
		UserID: 1,
	}
	if !submission.ByUser(1) {
		t.Errorf("ByUser(1) = false, expected true\n")
	}

	submission = &types.Submission{
		GroupID: 1,
	}
	if submission.ByUser(1) {
		t.Errorf("ByUser(1) = true, expected false\n")
	}

	// submissions with both user and group ID are invalid
	submission = &types.Submission{
		UserID:  1,
		GroupID: 2,
	}
	if submission.ByUser(1) {
		t.Errorf("ByUser(1) = true, expected false\n")
	}
}

func TestByGroup(t *testing.T) {
	submission := &types.Submission{}
	if submission.ByGroup(0) {
		t.Errorf("ByGroup(0) = true, expected false\n")
	}

	submission = &types.Submission{
		GroupID: 1,
	}
	if !submission.ByGroup(1) {
		t.Errorf("ByGroup(1) = false, expected true\n")
	}

	submission = &types.Submission{
		UserID: 1,
	}
	if submission.ByGroup(1) {
		t.Errorf("ByGroup(1) = true, expected false\n")
	}

	// submissions with both user and group ID are invalid
	submission = &types.Submission{
		UserID:  1,
		GroupID: 2,
	}
	if submission.ByGroup(1) {
		t.Errorf("ByGroup(1) = true, expected false\n")
	}
}
