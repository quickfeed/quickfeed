package score

import (
	"regexp"
	"strings"
)

// Output from go test -v and go test -json shares message classification and
// score-line extraction here. Plain go test and surrounding run-script output
// use the same handling, with output kept in the build log when no test is named.

// entryKind distinguishes a test's own writes from the messages the testing
// package printed on its behalf.
type entryKind int

const (
	plain      entryKind = iota // the test wrote this itself
	diagnostic                  // the testing package printed this for Log, Error, Fatal or Skip
	reported                    // go test -json tagged this as Error or Fatal output
)

type entry struct {
	text string
	kind entryKind
}

// diagnosticStart matches the source position before a testing message.
var diagnosticStart = regexp.MustCompile(`^( +)\S+:\d+: `)

// output records one line of a run's output. Test markers are recognized here rather
// than by the caller, because the events of Go 1.25 and 1.26 carry the markers
// in their Output field with no OutputType to tell it apart.
func (s *testLogScanner) output(line string, run *testRun, outputType string) {
	if outputType == "frame" {
		return
	}
	if s.textMarker(line) {
		return
	}
	if strings.TrimSpace(line) == "" {
		return
	}
	if HasPrefix(line) {
		if run == nil {
			s.log.Scores = append(s.log.Scores, line)
		} else {
			run.Scores = append(run.Scores, line)
		}
		return
	}
	if run == nil {
		s.log.Unattributed = append(s.log.Unattributed, line)
		return
	}
	text, kind := classify(line, outputType, run)
	run.entries = append(run.entries, entry{text: text, kind: kind})
}

// finish sorts each test's entries into its failures and its output. A failing
// test's diagnostics explain the failure, which is what a student needs first;
// the same lines under a passing or skipped test are ordinary output.
func (s *testLogScanner) finish() {
	for _, run := range s.log.Tests {
		tagged := false
		for _, e := range run.entries {
			if e.kind == reported {
				tagged = true
				break
			}
		}
		for _, e := range run.entries {
			switch {
			case e.kind == reported:
				run.Failures = append(run.Failures, e.text)
			case e.kind == diagnostic && !tagged && run.Status == TestStatus_FAILED:
				run.Failures = append(run.Failures, e.text)
			default:
				run.Output = append(run.Output, e.text)
			}
		}
		run.entries = nil
	}
}

// classify returns the text to record for one line of a test's output, and
// whether the testing package printed it on the test's behalf. The indentation
// the testing package adds is stripped; indentation beyond it
// is the message's own and is kept.
func classify(line, outputType string, run *testRun) (string, entryKind) {
	if outputType == "error" || outputType == "error-continue" {
		return trimIndent(line, run), reported
	}
	if m := diagnosticStart.FindStringSubmatch(line); m != nil {
		run.indent = len(m[1])
		return trimIndent(line, run), diagnostic
	}
	// A message spanning several lines indents its continuations past the
	// position of the first, which carries no position of its own to match.
	if last := lastKind(run); last == diagnostic || last == reported {
		if strings.HasPrefix(line, " ") && len(line)-len(strings.TrimLeft(line, " ")) > run.indent {
			return trimIndent(line, run), last
		}
	}
	return line, plain
}

// trimIndent removes the indentation the testing package added to a message.
func trimIndent(line string, run *testRun) string {
	if run.indent == 0 {
		return strings.TrimLeft(line, " ")
	}
	return strings.TrimPrefix(line, strings.Repeat(" ", run.indent))
}

func lastKind(run *testRun) entryKind {
	if len(run.entries) == 0 {
		return plain
	}
	return run.entries[len(run.entries)-1].kind
}
