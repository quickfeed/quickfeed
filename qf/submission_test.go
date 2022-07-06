package qf_test

import (
	"testing"
	"time"

	score "github.com/quickfeed/quickfeed/kit/score"
	pb "github.com/quickfeed/quickfeed/qf"
)

func TestNewestSubmissionDate(t *testing.T) {
	submission := &pb.Submission{}
	tim := time.Now()
	new, err := submission.NewestBuildDate(tim)
	if err == nil {
		t.Errorf("NewestBuildDate(%v) = %v, expected error '%v'\n", tim, new, pb.ErrMissingBuildInfo)
	}

	submission = &pb.Submission{}
	new, err = submission.NewestBuildDate(tim)
	if err == nil {
		t.Errorf("NewestBuildDate(%v) = %v, expected error '%v'\n", tim, new, pb.ErrMissingBuildInfo)
	}

	submission = &pb.Submission{
		BuildInfo: &score.BuildInfo{},
	}
	new, err = submission.NewestBuildDate(tim)
	if err == nil {
		t.Errorf("NewestBuildDate(%v) = %v, expected error '%v'\n", tim, new, pb.ErrMissingBuildInfo)
	}

	submission = &pb.Submission{
		BuildInfo: &score.BuildInfo{
			BuildDate: "string",
		},
	}
	new, err = submission.NewestBuildDate(tim)
	if err == nil {
		t.Errorf("NewestBuildDate(%v) = %v, expected error '%v'\n", tim, new, `parsing time "string" as "2006-01-02T15:04:05": cannot parse "string" as "2006"`)
	}

	buildDate := time.Now()
	submission = &pb.Submission{
		BuildInfo: &score.BuildInfo{
			BuildDate: buildDate.Format(pb.TimeLayout),
		},
	}
	new, err = submission.NewestBuildDate(tim)
	if err != nil {
		t.Error(err)
	}
	if new.Before(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, new, buildDate)
	}
}

func TestByUser(t *testing.T) {
	submission := &pb.Submission{}
	if submission.ByUser(0) {
		t.Errorf("ByUser(0) = true, expected false\n")
	}

	submission = &pb.Submission{
		UserID: 1,
	}
	if !submission.ByUser(1) {
		t.Errorf("ByUser(1) = false, expected true\n")
	}

	submission = &pb.Submission{
		GroupID: 1,
	}
	if submission.ByUser(1) {
		t.Errorf("ByUser(1) = true, expected false\n")
	}

	// submissions with both user and group ID are invalid
	submission = &pb.Submission{
		UserID:  1,
		GroupID: 2,
	}
	if submission.ByUser(1) {
		t.Errorf("ByUser(1) = true, expected false\n")
	}
}

func TestByGroup(t *testing.T) {
	submission := &pb.Submission{}
	if submission.ByGroup(0) {
		t.Errorf("ByGroup(0) = true, expected false\n")
	}

	submission = &pb.Submission{
		GroupID: 1,
	}
	if !submission.ByGroup(1) {
		t.Errorf("ByGroup(1) = false, expected true\n")
	}

	submission = &pb.Submission{
		UserID: 1,
	}
	if submission.ByGroup(1) {
		t.Errorf("ByGroup(1) = true, expected false\n")
	}

	// submissions with both user and group ID are invalid
	submission = &pb.Submission{
		UserID:  1,
		GroupID: 2,
	}
	if submission.ByGroup(1) {
		t.Errorf("ByGroup(1) = true, expected false\n")
	}
}
