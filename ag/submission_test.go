package ag_test

import (
	"testing"
	"time"

	pb "github.com/autograde/quickfeed/ag"
	score "github.com/autograde/quickfeed/kit/score"
	"google.golang.org/protobuf/types/known/timestamppb"
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
			BuildDate: &timestamppb.Timestamp{},
		},
	}
	new, err = submission.NewestBuildDate(tim)
	if err != nil {
		t.Error(err)
	}
	if !new.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v' = '%v'\n", tim, new, tim, new)
	}
	if new.Before(submission.BuildInfo.BuildDate.AsTime()) {
		t.Errorf("NewestBuildDate(%v) = %v, expected tim '%v' to be after submission.BuildDate '%v'\n", tim, new, tim, submission.BuildInfo.BuildDate.AsTime())
	}

	submission = &pb.Submission{
		BuildInfo: &score.BuildInfo{
			BuildDate: timestamppb.Now(),
		},
	}
	new, err = submission.NewestBuildDate(tim)
	if err != nil {
		t.Error(err)
	}
	if new.Before(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, new, submission.BuildInfo.BuildDate)
	}
}
