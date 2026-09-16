package main

import (
	"strings"
	"testing"
)

// TestLoggerOutsideRepository checks that the command's logger can be created
// from a working directory that is not inside a QuickFeed checkout, which is
// where teachers run qcm. The server's logger resolves the repository root
// from the working directory and panics when it cannot; qcm must not use it.
func TestLoggerOutsideRepository(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("QUICKFEED", "")

	var quiet strings.Builder
	c := commonFlags{}
	logger := c.logger(&quiet)
	logger.Debug("debug record")
	logger.Warn("warning record")
	if got := quiet.String(); strings.Contains(got, "debug record") || !strings.Contains(got, "warning record") {
		t.Errorf("logger() without -v printed:\n%s\nwant only the warning", got)
	}

	var verbose strings.Builder
	c.verbose = true
	c.logger(&verbose).Debug("debug record")
	if !strings.Contains(verbose.String(), "debug record") {
		t.Errorf("logger() with -v printed:\n%s\nwant the debug record", verbose.String())
	}
}
