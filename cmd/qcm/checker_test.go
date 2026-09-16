package main

import (
	"strings"
	"testing"
)

// TestCheckerProgress checks that a slow check is announced before it starts,
// that its result completes the announcement's line, and that every other
// result is reported as soon as it is known, so that the command does not
// appear to hang while the Docker-backed checks run.
func TestCheckerProgress(t *testing.T) {
	var progress strings.Builder
	ck := newChecker(&progress)
	ck.add(checkResult{name: "run scripts", result: pass, details: []string{"parsed scripts/run.sh"}})
	ck.start("skeleton lab1", "running the tests against the skeleton code")
	ck.add(checkResult{name: "skeleton lab1", result: fail, details: []string{"skeleton code scored 100%"}})

	want := "run scripts: PASS\n" +
		"skeleton lab1: running the tests against the skeleton code... FAIL\n"
	if got := progress.String(); got != want {
		t.Errorf("checker progress =\n%s\nwant\n%s", got, want)
	}
	if len(ck.results) != 2 || ck.results[1].details[0] != "skeleton code scored 100%" {
		t.Errorf("checker results = %+v, want both results with their details", ck.results)
	}
}

// TestCheckerProgressWithLogOutput checks that a log record written while a
// progress line is open goes on a line of its own, and that the result is
// then reported on a full line rather than glued to the record.
func TestCheckerProgressWithLogOutput(t *testing.T) {
	var progress strings.Builder
	ck := newChecker(&progress)
	ck.start("skeleton lab1", "running the tests against the skeleton code")
	if _, err := ck.progress.Write([]byte("level=ERROR msg=\"test run failed\"\n")); err != nil {
		t.Fatal(err)
	}
	ck.add(checkResult{name: "skeleton lab1", result: fail, details: []string{"the skeleton run failed"}})

	want := "skeleton lab1: running the tests against the skeleton code...\n" +
		"level=ERROR msg=\"test run failed\"\n" +
		"skeleton lab1: FAIL\n"
	if got := progress.String(); got != want {
		t.Errorf("checker progress =\n%s\nwant\n%s", got, want)
	}
}
