package qf_test

import (
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// farPast and farFuture never bind Interval's [minFrom, maxTo] clamp in the
// tests below, isolating the defaulting behavior under test.
var (
	farPast   = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	farFuture = time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)
)

func TestCourseLogRequestIntervalDefaults(t *testing.T) {
	before := time.Now()
	from, to := (&qf.CourseLogRequest{}).Interval(farPast, time.Now())
	after := time.Now()

	if to.Before(before) || to.After(after) {
		t.Errorf("to = %v, want within [%v, %v]", to, before, after)
	}
	if got, want := to.Sub(from), 24*time.Hour; got != want {
		t.Errorf("to.Sub(from) = %v, want the default interval %v", got, want)
	}
}

// TestCourseLogRequestIntervalHonorsExplicitBounds guards that Interval
// passes explicit From/To through unclamped when they fall within
// [minFrom, maxTo].
func TestCourseLogRequestIntervalHonorsExplicitBounds(t *testing.T) {
	to := time.Now().Add(-time.Hour)
	from := to.Add(-time.Hour)
	gotFrom, gotTo := (&qf.CourseLogRequest{
		From: timestamppb.New(from),
		To:   timestamppb.New(to),
	}).Interval(farPast, farFuture)
	if !gotFrom.Equal(from) || !gotTo.Equal(to) {
		t.Errorf("Interval() = (%v, %v), want (%v, %v)", gotFrom, gotTo, from, to)
	}
}

// TestCourseLogRequestIntervalClampsToBounds guards that an explicit From
// before minFrom, or a To after maxTo (including one defaulted from a To
// beyond maxTo), is clamped rather than passed through.
func TestCourseLogRequestIntervalClampsToBounds(t *testing.T) {
	minFrom := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	maxTo := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

	from, to := (&qf.CourseLogRequest{
		From: timestamppb.New(minFrom.Add(-time.Hour)),
		To:   timestamppb.New(maxTo.Add(time.Hour)),
	}).Interval(minFrom, maxTo)
	if !from.Equal(minFrom) || !to.Equal(maxTo) {
		t.Errorf("Interval() = (%v, %v), want (%v, %v)", from, to, minFrom, maxTo)
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
