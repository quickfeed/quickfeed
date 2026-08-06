package qlog

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/quickfeed/quickfeed/internal/env"
)

type contextKey struct{}

// New returns a structured logger that writes debug and higher-level records to w.
func New(w io.Writer) *slog.Logger {
	root := env.Root()
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key != slog.SourceKey {
				return attr
			}
			source, ok := attr.Value.Any().(*slog.Source)
			if !ok {
				return attr
			}
			if path, err := filepath.Rel(root, source.File); err == nil {
				source.File = path
			}
			return slog.Any(slog.SourceKey, source)
		},
	})
	return slog.New(handler)
}

// SetDefault installs logger as the process-wide fallback logger.
func SetDefault(logger *slog.Logger) {
	slog.SetDefault(logger)
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
