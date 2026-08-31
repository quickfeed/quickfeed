package qlog

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/quickfeed/quickfeed/internal/qlog/label"
)

// fakeCourse implements Course without depending on qf.Course.
type fakeCourse struct {
	id   uint64
	code string
	org  string
}

func (c fakeCourse) GetID() uint64                  { return c.id }
func (c fakeCourse) GetCode() string                { return c.code }
func (c fakeCourse) GetScmOrganizationName() string { return c.org }

func TestCourseAttrs(t *testing.T) {
	course := fakeCourse{id: 7, code: "DAT320", org: "dat320-2026"}
	want := []any{
		label.CourseID, uint64(7),
		label.CourseCode, "DAT320",
		label.CourseLog, "dat320-2026",
	}
	got := CourseAttrs(course)
	if len(got) != len(want) {
		t.Fatalf("CourseAttrs() = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("CourseAttrs()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestWithCourse(t *testing.T) {
	var output bytes.Buffer
	logger := New(&output)
	course := fakeCourse{id: 42, code: "DAT520", org: "dat520-2026"}

	ctx, scoped := WithCourse(NewContext(context.Background(), logger), course, label.Repository, "student")
	if FromContext(ctx) != scoped {
		t.Fatalf("FromContext(ctx) did not return the logger WithCourse attached to ctx")
	}
	scoped.Info("scoped record")

	got := output.String()
	for _, want := range []string{"course_id=42", "course_code=DAT520", "course_log=dat520-2026", "repository=student"} {
		if !strings.Contains(got, want) {
			t.Errorf("log output %q does not contain %q", got, want)
		}
	}
}

// TestWithCourseLogOmitsCourseID guards the variant used where the RPC
// interceptors already attached course_id from the caller's claims: it must
// not repeat it, only add CourseCode and the log marker.
func TestWithCourseLogOmitsCourseID(t *testing.T) {
	var output bytes.Buffer
	logger := New(&output)
	course := fakeCourse{id: 42, code: "DAT520", org: "dat520-2026"}

	ambient := logger.With(label.CourseID, uint64(42)) // simulates enrichRequestLogger
	_, scoped := WithCourseLog(NewContext(context.Background(), ambient), course)
	scoped.Info("scoped record")

	got := output.String()
	if n := strings.Count(got, "course_id="); n != 1 {
		t.Errorf("log output %q has course_id %d times, want exactly once", got, n)
	}
	for _, want := range []string{"course_code=DAT520", "course_log=dat520-2026"} {
		if !strings.Contains(got, want) {
			t.Errorf("log output %q does not contain %q", got, want)
		}
	}
}

func TestWithSink(t *testing.T) {
	var primary, secondary bytes.Buffer
	logger := New(&primary)
	sink := slog.NewJSONHandler(&secondary, &slog.HandlerOptions{Level: slog.LevelDebug})

	combined := WithSink(logger, sink)
	combined.Info("fanned out")

	if !strings.Contains(primary.String(), "fanned out") {
		t.Errorf("primary output %q does not contain the record", primary.String())
	}
	if !strings.Contains(secondary.String(), "fanned out") {
		t.Errorf("secondary output %q does not contain the record", secondary.String())
	}
}
