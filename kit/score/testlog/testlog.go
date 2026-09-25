// Package testlog attributes a test run's output to the tests that produced it,
// so that a student can read one failing test's output instead of the whole run.
//
// It reads both of the forms a course's run script can produce: the framing that
// go test -v prints, and the event stream that go test -json prints. The two are
// the same information, since go test -json is go test -v routed through
// cmd/internal/test2json, and the choice is made per line, so that the output a
// run script prints around the test command is handled either way.
//
// Attribution is exact for a test's scores, its status, and the messages it
// reported through Error, Fatal, and Skip, because those pass through the
// testing package's per-test framing. It is best effort for what a test writes
// to stdout itself: go test interleaves the raw writes of parallel tests, and
// neither form records which test made which write.
package testlog

import (
	"encoding/json"
	"strings"
)

// Status is the outcome a test framework reported for a single test.
type Status int

const (
	StatusUnknown Status = iota // the run reported no outcome
	StatusPassed
	StatusFailed
	StatusSkipped
)

// String returns the lowercase name of the status.
func (s Status) String() string {
	switch s {
	case StatusPassed:
		return "passed"
	case StatusFailed:
		return "failed"
	case StatusSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// TestRun is one test's output, as the test framework attributed it.
type TestRun struct {
	Name     string
	Status   Status
	Elapsed  float64  // seconds
	Failures []string // messages reported through Error or Fatal
	Output   []string // everything else the test printed
	Scores   []string // score lines the test emitted

	entries []entry
	indent  int // width of the indentation the testing package adds to messages
}

// Log is a run's output, attributed to the tests that produced it.
type Log struct {
	Tests        []*TestRun // in the order the tests started
	Unattributed []string   // setup, compile, and package-level output
	Scores       []string   // score lines that belong to no test

	byName map[string]*TestRun
}

// Test returns the run of the named test, or nil if the run has no such test.
func (l *Log) Test(name string) *TestRun {
	return l.byName[name]
}

// Scan attributes the output of a test run to the tests that produced it.
// The isScoreLine predicate reports whether a line is a score line; it is a
// parameter so that this package need not depend on the score package, and so
// that a score line never reaches a caller as ordinary output.
func Scan(out string, isScoreLine func(string) bool) *Log {
	s := &scanner{
		log:         &Log{byName: make(map[string]*TestRun)},
		isScoreLine: isScoreLine,
	}
	for line := range strings.SplitSeq(out, "\n") {
		s.line(strings.TrimSuffix(line, "\r"))
	}
	s.finish()
	return s.log
}

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

type scanner struct {
	log         *Log
	isScoreLine func(string) bool
	current     *TestRun // the test the text framing is inside of
	// structuredAttrs records that this run reported attributes as events. Such
	// a run also prints each attribute as an "=== ATTR" line in an output event,
	// which would otherwise record the same score a second time.
	structuredAttrs bool
}

// event is the subset of a test2json event that attribution needs.
type event struct {
	Action     string
	Test       string
	Elapsed    float64
	Output     string
	OutputType string
	Key        string
	Value      string
}

func (s *scanner) line(line string) {
	if ev, ok := decodeEvent(line); ok {
		s.event(ev)
		return
	}
	s.output(line, s.current, "")
}

// event records a test2json event. The event names the test it belongs to, so
// it does not rely on the framing that the text form tracks.
func (s *scanner) event(ev event) {
	switch ev.Action {
	case "attr":
		s.structuredAttrs = true
		// Only score attributes are kept; any other attribute is metadata about
		// the test rather than output from it.
		if s.isScoreLine(ev.Value) {
			s.test(ev.Test).Scores = append(s.test(ev.Test).Scores, ev.Value)
		}
	case "run":
		s.test(ev.Test)
	case "pass", "fail", "skip":
		if ev.Test == "" {
			return // the package's own result, not a test's
		}
		run := s.test(ev.Test)
		run.Status = statusOf(ev.Action)
		run.Elapsed = ev.Elapsed
	case "output":
		var run *TestRun
		if ev.Test != "" {
			run = s.test(ev.Test)
		}
		for line := range strings.SplitSeq(strings.TrimSuffix(ev.Output, "\n"), "\n") {
			s.output(line, run, ev.OutputType)
		}
	}
}

// output records one line of a run's output. Framing is recognized here rather
// than by the caller, because the events of Go 1.25 and 1.26 carry the framing
// in their Output field with no OutputType to tell it apart.
func (s *scanner) output(line string, run *TestRun, outputType string) {
	if outputType == "frame" {
		return
	}
	if s.framing(line) {
		return
	}
	if strings.TrimSpace(line) == "" {
		return
	}
	if s.isScoreLine(line) {
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
func (s *scanner) finish() {
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
			case e.kind == diagnostic && !tagged && run.Status == StatusFailed:
				run.Failures = append(run.Failures, e.text)
			default:
				run.Output = append(run.Output, e.text)
			}
		}
		run.entries = nil
	}
}

// test returns the named test's run, recording it in start order if new.
func (s *scanner) test(name string) *TestRun {
	if run, ok := s.log.byName[name]; ok {
		return run
	}
	run := &TestRun{Name: name}
	s.log.byName[name] = run
	s.log.Tests = append(s.log.Tests, run)
	return run
}

func statusOf(action string) Status {
	switch action {
	case "pass":
		return StatusPassed
	case "fail":
		return StatusFailed
	case "skip":
		return StatusSkipped
	}
	return StatusUnknown
}

// decodeEvent reports whether line is a test2json event, and decodes it.
func decodeEvent(line string) (event, bool) {
	if !strings.HasPrefix(line, `{"Time":`) && !strings.HasPrefix(line, `{"Action":`) {
		return event{}, false
	}
	var ev event
	if err := json.Unmarshal([]byte(line), &ev); err != nil || ev.Action == "" {
		return event{}, false
	}
	return ev, true
}
