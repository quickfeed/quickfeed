package qlog

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/quickfeed/quickfeed/internal/env"
)

type contextKey struct{}

// New returns a structured logger that writes debug and higher-level records to w.
func New(w io.Writer) *slog.Logger {
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: RelativeSource(env.Root()),
	})
	return slog.New(handler)
}

// RelativeSource returns a ReplaceAttr function that rewrites the slog source
// attribute's file path to be relative to root, so it stays readable outside
// the machine that produced it. Handlers that write structured records to
// storage, such as the course log sink, reuse this so their "source" field
// matches the process logger's.
func RelativeSource(root string) func(groups []string, attr slog.Attr) slog.Attr {
	return func(_ []string, attr slog.Attr) slog.Attr {
		if attr.Key != slog.SourceKey {
			return attr
		}
		source, ok := attr.Value.Any().(*slog.Source)
		if !ok {
			return attr
		}
		// Keep the recorded path for records that do not originate below
		// root, such as those from the standard library or the module
		// cache, or when the paths recorded by the compiler differ from
		// root; a relative path would then be a walk-up, not a shorthand.
		path, err := filepath.Rel(root, source.File)
		if err != nil || strings.HasPrefix(path, "..") {
			return attr
		}
		return slog.Any(slog.SourceKey, &slog.Source{
			Function: source.Function,
			File:     path,
			Line:     source.Line,
		})
	}
}

// SetDefault installs logger as the process-wide fallback logger.
func SetDefault(logger *slog.Logger) {
	slog.SetDefault(logger)
}

// WithSink returns a logger that writes every record to logger's handler and,
// in addition, to sink. Use it to fan records carrying an opt-in scope, such
// as the one WithCourse attaches, out to a secondary handler without
// affecting where records land by default.
func WithSink(logger *slog.Logger, sink slog.Handler) *slog.Logger {
	return slog.New(slog.NewMultiHandler(logger.Handler(), sink))
}

// NewContext returns a context carrying logger.
func NewContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// FromContext returns the logger carried by ctx or the default logger.
func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	if logger, ok := ctx.Value(contextKey{}).(*slog.Logger); ok && logger != nil {
		return logger
	}
	return slog.Default()
}

// With returns a context carrying a logger enriched with attrs.
func With(ctx context.Context, attrs ...any) context.Context {
	ctx, _ = WithLogger(ctx, attrs...)
	return ctx
}

// WithLogger returns a context carrying a logger enriched with attrs, along
// with that logger. Use it when the calling function both logs itself and
// passes the enriched context to downstream functions; the attrs are then
// recorded once here instead of being repeated in each log statement.
func WithLogger(ctx context.Context, attrs ...any) (context.Context, *slog.Logger) {
	logger := FromContext(ctx).With(attrs...)
	return NewContext(ctx, logger), logger
}
