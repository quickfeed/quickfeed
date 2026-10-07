package testlog

import (
	"regexp"
	"strconv"
	"strings"
)

// The lines go test -v prints about a test, rather than on its behalf. The same
// lines reach the scanner through the Output field of a go test -json event on
// Go versions before 1.27, which tag no output with an OutputType.
var (
	// "=== RUN   TestStack", and the PAUSE, CONT and NAME markers that track
	// which test the following output belongs to.
	frameName = regexp.MustCompile(`^=== (RUN|PAUSE|CONT|NAME)\s+(.+?)\s*$`)
	// "=== ATTR  TestStack score {...}", printed by testing.T.Attr.
	frameAttr = regexp.MustCompile(`^=== ATTR\s+(\S+)\s+(\S+)\s(.*)$`)
	// "--- FAIL: TestQueue (0.00s)", indented once per level of subtest.
	frameResult = regexp.MustCompile(`^\s*--- (PASS|FAIL|SKIP):\s+(.+?)\s+\(([0-9.]+)s\)\s*$`)
	// The position the testing package prints before a message from Log,
	// Error, Fatal or Skip, e.g. "    queue_test.go:42: ".
	diagnosticStart = regexp.MustCompile(`^( +)\S+:\d+: `)
)

// framing reports whether line is the test framework describing a test rather
// than output from one, recording what the line says.
func (s *scanner) framing(line string) bool {
	if m := frameAttr.FindStringSubmatch(line); m != nil {
		// The attribute's own test name is authoritative even when the output
		// of parallel tests is interleaved.
		if !s.structuredAttrs && s.isScoreLine(m[3]) {
			run := s.test(m[1])
			run.Scores = append(run.Scores, m[3])
		}
		return true
	}
	if m := frameName.FindStringSubmatch(line); m != nil {
		s.current = s.test(m[2])
		return true
	}
	if m := frameResult.FindStringSubmatch(line); m != nil {
		run := s.test(m[2])
		run.Status = statusOf(strings.ToLower(m[1]))
		if elapsed, err := strconv.ParseFloat(m[3], 64); err == nil {
			run.Elapsed = elapsed
		}
		// The test has reported its result, so what follows is the package's.
		s.current = nil
		return true
	}
	return false
}

// classify returns the text to record for one line of a test's output, and
// whether the testing package printed it on the test's behalf. The indentation
// the testing package adds is framing, so it is stripped; indentation beyond it
// is the message's own and is kept.
func classify(line, outputType string, run *TestRun) (string, entryKind) {
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
func trimIndent(line string, run *TestRun) string {
	if run.indent == 0 {
		return strings.TrimLeft(line, " ")
	}
	return strings.TrimPrefix(line, strings.Repeat(" ", run.indent))
}

func lastKind(run *TestRun) entryKind {
	if len(run.entries) == 0 {
		return plain
	}
	return run.entries[len(run.entries)-1].kind
}

// CarriesScore reports whether line carries a score object, in any of the forms
// a run can print one: on its own, as the "=== ATTR" line that testing.T.Attr
// prints, or inside a go test -json event. The isScoreLine predicate recognizes
// the score object itself, as it does for Scan.
func CarriesScore(line string, isScoreLine func(string) bool) bool {
	if isScoreLine(line) {
		return true
	}
	if m := frameAttr.FindStringSubmatch(line); m != nil {
		return isScoreLine(m[3])
	}
	if ev, ok := decodeEvent(line); ok {
		switch ev.Action {
		case "attr":
			return isScoreLine(ev.Value)
		case "output":
			return isScoreLine(ev.Output)
		}
	}
	return false
}
