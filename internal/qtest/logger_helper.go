package qtest

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/quickfeed/quickfeed/internal/qlog"
)

// Context returns a background context containing the test logger.
func Context(t *testing.T) context.Context {
	t.Helper()
	return qlog.NewContext(context.Background(), Logger(t))
}

// Logger returns a logger that discards its output, unless the LOG environment
// variable is set, in which case it writes to stderr; see doc/gorm-issues.md.
func Logger(t *testing.T) *slog.Logger {
	t.Helper()
	if os.Getenv("LOG") == "" {
		return qlog.New(io.Discard)
	}
	return qlog.New(os.Stderr)
}
