package courselog

import (
	"context"
	"log/slog"

	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
)

// maxFieldBytes is the largest message or string attribute value kept in a
// course log record; anything longer is cut down to size and the record is
// marked label.Truncated.
const maxFieldBytes = 64 * 1024

// handler is a slog.Handler that copies records carrying label.CourseLog to a
// Store, and drops every other record. Attach it alongside the operator's
// handler with slog.NewMultiHandler; see qlog.WithSink.
//
// A handler is immutable: WithAttrs and WithGroup return a new value, as
// slog.Handler requires. Until a scope's WithAttrs call carries
// label.CourseLog, its attributes are held pending; the call that carries it
// selects the course's file and replays the pending attributes into it.
type handler struct {
	store *Store

	// inner is nil until label.CourseLog has been observed in this scope.
	inner   slog.Handler
	pending []slog.Attr
}

// NewHandler returns a Handler that writes through store.
func NewHandler(store *Store) slog.Handler {
	return &handler{store: store}
}

func (h *handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelDebug
}

func (h *handler) Handle(ctx context.Context, r slog.Record) error {
	if h.inner == nil {
		// No scope up to this record has attached label.CourseLog.
		return nil
	}
	return h.inner.Handle(ctx, truncate(r))
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	if h.inner != nil {
		return &handler{store: h.store, inner: h.inner.WithAttrs(attrs)}
	}
	org, rest, found := extractCourseLog(attrs)
	pending := make([]slog.Attr, 0, len(h.pending)+len(rest))
	pending = append(pending, h.pending...)
	pending = append(pending, rest...)
	if !found {
		return &handler{store: h.store, pending: pending}
	}
	inner := slog.NewJSONHandler(h.store.Writer(org), &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: qlog.RelativeSource(env.Root()),
	})
	return &handler{store: h.store, inner: inner.WithAttrs(pending)}
}

func (h *handler) WithGroup(name string) slog.Handler {
	if h.inner != nil {
		return &handler{store: h.store, inner: h.inner.WithGroup(name)}
	}
	// No call site scopes a course logger with a group before attaching
	// label.CourseLog, so there is nothing meaningful to carry forward here.
	return h
}

// extractCourseLog reports whether attrs carries label.CourseLog, returning
// its string value and the remaining attributes with it removed. The marker
// itself is not written to the course log: it is a routing detail internal
// to the sink, not a field a teacher needs to see.
func extractCourseLog(attrs []slog.Attr) (org string, rest []slog.Attr, found bool) {
	rest = make([]slog.Attr, 0, len(attrs))
	for _, a := range attrs {
		if !found && a.Key == label.CourseLog {
			org = a.Value.String()
			found = true
			continue
		}
		rest = append(rest, a)
	}
	return org, rest, found
}

// truncate returns a copy of r with its message, and any string attribute
// value, longer than maxFieldBytes cut down to size. The record is marked
// label.Truncated when anything was cut.
func truncate(r slog.Record) slog.Record {
	msg := r.Message
	truncated := false
	if len(msg) > maxFieldBytes {
		msg = msg[:maxFieldBytes]
		truncated = true
	}
	out := slog.NewRecord(r.Time, r.Level, msg, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		if a.Value.Kind() == slog.KindString {
			if s := a.Value.String(); len(s) > maxFieldBytes {
				a = slog.String(a.Key, s[:maxFieldBytes])
				truncated = true
			}
		}
		out.AddAttrs(a)
		return true
	})
	if truncated {
		out.AddAttrs(slog.Bool(label.Truncated, true))
	}
	return out
}
