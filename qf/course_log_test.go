package qf_test

import (
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCourseLogRequestIntervalDefaults(t *testing.T) {
	before := time.Now()
	from, to := (&qf.CourseLogRequest{}).Interval()
	after := time.Now()

	if to.Before(before) || to.After(after) {
		t.Errorf("to = %v, want within [%v, %v]", to, before, after)
	}
	if got, want := to.Sub(from), 24*time.Hour; got != want {
		t.Errorf("to.Sub(from) = %v, want the default interval %v", got, want)
	}
}

// TestCourseLogRequestIntervalHonorsExplicitBounds guards that Interval
// passes explicit From/To through unclamped, including a From beyond the
// store's retention window or a To in the future.
func TestCourseLogRequestIntervalHonorsExplicitBounds(t *testing.T) {
	to := time.Now().Add(-time.Hour)
	from := to.Add(-time.Hour)
	gotFrom, gotTo := (&qf.CourseLogRequest{
		From: timestamppb.New(from),
		To:   timestamppb.New(to),
	}).Interval()
	if !gotFrom.Equal(from) || !gotTo.Equal(to) {
		t.Errorf("Interval() = (%v, %v), want (%v, %v)", gotFrom, gotTo, from, to)
	}
}

func TestCourseLogRequestEffectiveLimit(t *testing.T) {
	const maxCourseLogLimit = 5000
	tests := []struct {
		requested uint32
		want      int
	}{
		{requested: 0, want: 2000},
		{requested: 10, want: 10},
		{requested: maxCourseLogLimit, want: maxCourseLogLimit},
		{requested: maxCourseLogLimit + 1000, want: maxCourseLogLimit},
	}
	for _, test := range tests {
		req := &qf.CourseLogRequest{Limit: test.requested}
		if got := req.EffectiveLimit(); got != test.want {
			t.Errorf("EffectiveLimit(%d) = %d, want %d", test.requested, got, test.want)
		}
	}
}
