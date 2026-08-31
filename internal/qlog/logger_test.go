package qlog

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/quickfeed/quickfeed/internal/qlog/label"
)

func TestFromContext(t *testing.T) {
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })

	var fallbackOutput bytes.Buffer
	fallback := New(&fallbackOutput)
	SetDefault(fallback)
	if got := FromContext(context.Background()); got != fallback {
		t.Fatalf("FromContext() = %p, want fallback %p", got, fallback)
	}
	// A nil context variable, rather than a nil literal, to exercise the
	// nil guard in FromContext without tripping the SA1012 check.
	var noContext context.Context
	if got := FromContext(noContext); got != fallback {
		t.Fatalf("FromContext(nil) = %p, want fallback %p", got, fallback)
	}

	var output bytes.Buffer
	logger := New(&output)
	ctx := With(NewContext(context.Background(), logger), label.CourseID, uint64(42), label.Repository, "student")
	FromContext(ctx).Info("scoped record")
	got := output.String()
	for _, want := range []string{`msg="scoped record"`, "course_id=42", "repository=student", "source=internal/qlog/logger_test.go:"} {
		if !strings.Contains(got, want) {
			t.Errorf("log output %q does not contain %q", got, want)
		}
	}
}
