package web

import (
	"log/slog"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog/courselog"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCourseLogIntervalDefaults(t *testing.T) {
	before := time.Now()
	from, to := courseLogInterval(&qf.CourseLogRequest{})
	after := time.Now()

	if to.Before(before) || to.After(after) {
		t.Errorf("to = %v, want within [%v, %v]", to, before, after)
	}
	if got, want := to.Sub(from), defaultCourseLogInterval; got != want {
		t.Errorf("to.Sub(from) = %v, want the default interval %v", got, want)
	}
}

func TestCourseLogIntervalHonorsExplicitBounds(t *testing.T) {
	// Within the retention window relative to the real clock, so the
	// retention clamp (tested separately below) does not also kick in here.
	to := time.Now().Add(-time.Hour)
	from := to.Add(-time.Hour)
	gotFrom, gotTo := courseLogInterval(&qf.CourseLogRequest{
		From: timestamppb.New(from),
		To:   timestamppb.New(to),
	})
	if !gotFrom.Equal(from) || !gotTo.Equal(to) {
		t.Errorf("courseLogInterval() = (%v, %v), want (%v, %v)", gotFrom, gotTo, from, to)
	}
}

// TestCourseLogIntervalClampsToRetention guards that a From beyond the
// retention window is clamped rather than rejected, consistent with how a
// result that still exceeds Limit is reported truncated instead of erroring.
func TestCourseLogIntervalClampsToRetention(t *testing.T) {
	to := time.Now()
	from := to.Add(-2 * courselog.Retention)
	gotFrom, gotTo := courseLogInterval(&qf.CourseLogRequest{
		From: timestamppb.New(from),
		To:   timestamppb.New(to),
	})
	if !gotTo.Equal(to) {
		t.Errorf("to = %v, want %v unchanged", gotTo, to)
	}
	if cutoff := to.Add(-courselog.Retention); gotFrom.Before(cutoff) {
		t.Errorf("from = %v, want clamped to no earlier than %v", gotFrom, cutoff)
	}
	if gotFrom.Equal(from) {
		t.Errorf("from = %v, want it clamped rather than left at the requested value", gotFrom)
	}
}

func TestCourseLogLimit(t *testing.T) {
	tests := []struct {
		requested uint32
		want      int
	}{
		{requested: 0, want: defaultCourseLogLimit},
		{requested: 10, want: 10},
		{requested: maxCourseLogLimit, want: maxCourseLogLimit},
		{requested: maxCourseLogLimit + 1000, want: maxCourseLogLimit},
	}
	for _, test := range tests {
		if got := courseLogLimit(test.requested); got != test.want {
			t.Errorf("courseLogLimit(%d) = %d, want %d", test.requested, got, test.want)
		}
	}
}

func TestSlogLevelAndCourseLogEntryLevelRoundTrip(t *testing.T) {
	levels := []qf.CourseLogEntry_Level{
		qf.CourseLogEntry_DEBUG,
		qf.CourseLogEntry_INFO,
		qf.CourseLogEntry_WARN,
		qf.CourseLogEntry_ERROR,
	}
	for _, level := range levels {
		if got := courseLogEntryLevel(slogLevel(level)); got != level {
			t.Errorf("courseLogEntryLevel(slogLevel(%v)) = %v, want %v", level, got, level)
		}
	}
}

func TestToCourseLogEntries(t *testing.T) {
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

	got := toCourseLogEntries(entries)
	if len(got) != 1 {
		t.Fatalf("len(toCourseLogEntries()) = %d, want 1", len(got))
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
