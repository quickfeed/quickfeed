package qf_test

import (
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewestSubmissionDate(t *testing.T) {
	submission := &qf.Submission{}
	tim := time.Now()
	newSubmissionDate := submission.NewestSubmissionDate(tim)
	if !newSubmissionDate.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, tim)
	}

	submission = &qf.Submission{}
	newSubmissionDate = submission.NewestSubmissionDate(tim)
	if !newSubmissionDate.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, tim)
	}

	submission = &qf.Submission{
		BuildInfo: &score.BuildInfo{},
	}
	newSubmissionDate = submission.NewestSubmissionDate(tim)
	if !newSubmissionDate.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, tim)
	}

	submission = &qf.Submission{
		BuildInfo: &score.BuildInfo{
			SubmissionDate: &timestamppb.Timestamp{},
		},
	}
	newSubmissionDate = submission.NewestSubmissionDate(tim)
	if !newSubmissionDate.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, tim)
	}

	// Seems like the conversion from time.Time to timestamppb.Timestamp is not
	// exact, so we need to add a second to make sure the build date is newer.
	buildDate := time.Now().Add(1 * time.Second)
	submission = &qf.Submission{
		BuildInfo: &score.BuildInfo{
			SubmissionDate: timestamppb.New(buildDate),
		},
	}
	newSubmissionDate = submission.NewestSubmissionDate(tim)
	if newSubmissionDate.Equal(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, buildDate)
	}
	if newSubmissionDate.Before(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, buildDate)
	}
	if newSubmissionDate.After(buildDate) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, buildDate)
	}
	if !newSubmissionDate.After(tim) {
		t.Errorf("NewestBuildDate(%v) = %v, expected '%v'\n", tim, newSubmissionDate, buildDate)
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
