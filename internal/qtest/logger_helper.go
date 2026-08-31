package qtest

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/quickfeed/quickfeed/internal/qlog"
)

// Context returns the test's context containing the test logger.
func Context(t *testing.T) context.Context {
	t.Helper()
	return qlog.NewContext(t.Context(), Logger(t))
}

// Logger returns a logger that discards its output, unless the LOG environment
// variable is set, in which case it writes to stderr; see doc/gorm-issues.md.
func Logger(t *testing.T) *slog.Logger {
	t.Helper()
	if os.Getenv("LOG") == "" {
		return slog.New(slog.DiscardHandler)
	}
	return qlog.New(os.Stderr)
}
