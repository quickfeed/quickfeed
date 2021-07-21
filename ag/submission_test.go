package ag_test

import (
	"testing"
	"time"

	pb "github.com/autograde/quickfeed/ag"
	score "github.com/autograde/quickfeed/kit/score"
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
