package qf_test

import (
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewestSubmissionDate(t *testing.T) {
	submission := &qf.Submission{}
	tim := time.Now()
	newBuildDate, err := submission.NewestBuildDate(tim)
	if err == nil {
		t.Errorf("NewestBuildDate(%v) = %v, expected error '%v'\n", tim, newBuildDate, qf.ErrMissingBuildInfo)
	}

	submission = &qf.Submission{}
	newBuildDate, err = submission.NewestBuildDate(tim)
	if err == nil {
		t.Errorf("NewestBuildDate(%v) = %v, expected error '%v'\n", tim, newBuildDate, qf.ErrMissingBuildInfo)
	}

	submission = &qf.Submission{
		BuildInfo: &score.BuildInfo{},
	}
	newBuildDate, err = submission.NewestBuildDate(tim)
	if err == nil {
		t.Errorf("NewestBuildDate(%v) = %v, expected error '%v'\n", tim, newBuildDate, qf.ErrMissingBuildInfo)
	}

	submission = &qf.Submission{
		BuildInfo: &score.BuildInfo{
			BuildDate: &timestamppb.Timestamp{},
		},
	}
	newBuildDate, err = submission.NewestBuildDate(tim)
	if err != nil {
		t.Error(err)
	}
	if !newBuildDate.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v' = '%v'\n", tim, newBuildDate, tim, newBuildDate)
	}
	if newBuildDate.Before(submission.BuildInfo.BuildDate.AsTime()) {
		t.Errorf("NewestBuildDate(%v) = %v, expected tim '%v' to be after submission.BuildDate '%v'\n", tim, newBuildDate, tim, submission.BuildInfo.BuildDate.AsTime())
	}

	submission = &qf.Submission{
		BuildInfo: &score.BuildInfo{
			BuildDate: timestamppb.Now(),
		},
	}
	newBuildDate, err = submission.NewestBuildDate(tim)
	if err != nil {
		t.Error(err)
	}
	if newBuildDate.Before(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newBuildDate, submission.BuildInfo.BuildDate)
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
