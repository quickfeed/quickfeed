package qf_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog/courselog"
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

func TestCourseLogLevelSlogLevelRoundTrip(t *testing.T) {
	levels := []qf.CourseLogEntry_Level{
		qf.CourseLogEntry_DEBUG,
		qf.CourseLogEntry_INFO,
		qf.CourseLogEntry_WARN,
		qf.CourseLogEntry_ERROR,
	}
	for _, level := range levels {
		if got := qf.CourseLogLevelFromSlog(level.SlogLevel()); got != level {
			t.Errorf("CourseLogLevelFromSlog(%v.SlogLevel()) = %v, want %v", level, got, level)
		}
	}
}

func TestCourseLogEntriesFrom(t *testing.T) {
	at := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	entries := []courselog.Entry{{
		Time:           at,
		Level:          slog.LevelWarn,
		Message:        "test output",
		Source:         "ci/run_tests.go:42",
		Repository:     "student-repo",
		RepositoryType: "USER",
		Truncated:      true,
		Fields:         map[string]string{"assignment": "lab1"},
	}}

	got := qf.CourseLogEntriesFrom(entries)
	if len(got) != 1 {
		t.Fatalf("len(CourseLogEntriesFrom()) = %d, want 1", len(got))
	}
	e := got[0]
	if !e.GetTime().AsTime().Equal(at) {
		t.Errorf("Time = %v, want %v", e.GetTime().AsTime(), at)
	}
	if e.GetLevel() != qf.CourseLogEntry_WARN {
		t.Errorf("Level = %v, want WARN", e.GetLevel())
	}
	if e.GetMessage() != "test output" || e.GetSource() != "ci/run_tests.go:42" {
		t.Errorf("Message/Source = %q/%q, want %q/%q", e.GetMessage(), e.GetSource(), "test output", "ci/run_tests.go:42")
	}
	if e.GetRepository() != "student-repo" || e.GetRepositoryType() != "USER" || !e.GetTruncated() {
		t.Errorf("entry = %+v, want dedicated fields carried through", e)
	}
	if e.GetFields()["assignment"] != "lab1" {
		t.Errorf(`Fields["assignment"] = %q, want "lab1"`, e.GetFields()["assignment"])
	}
}
