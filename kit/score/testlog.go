package score

import (
	"encoding/json"
	"strings"
)

// testRun is one test's output, as the test framework attributed it.
type testRun struct {
	Name     string
	Status   TestStatus
	Elapsed  float64  // seconds
	Failures []string // messages reported through Error or Fatal
	Output   []string // everything else the test printed
	Scores   []string // score lines the test emitted

	entries []entry
	indent  int // width of the indentation the testing package adds to messages
}

// testLog is a run's output, attributed to the tests that produced it.
type testLog struct {
	Tests        []*testRun // in the order the tests started
	Unattributed []string   // setup, compile, and package-level output
	Scores       []string   // score lines that belong to no test

	byName map[string]*testRun
}

// Test returns the run of the named test, or nil if the run has no such test.
func (l *testLog) Test(name string) *testRun {
	return l.byName[name]
}

// retainOutput assigns unscored subtests to their nearest scored ancestor.
// Tests without a scored ancestor keep their output in the build log.
func (l *testLog) retainOutput(scored map[string]*testRun) {
	for _, run := range l.Tests {
		if scored[run.Name] != nil {
			continue
		}
		if parent := scoredAncestor(run.Name, scored); parent != nil {
			parent.Output = append(parent.Output, namedOutput(run.Name, run.Output)...)
			parent.Failures = append(parent.Failures, namedOutput(run.Name, run.Failures)...)
		} else {
			l.Unattributed = append(l.Unattributed, namedOutput(run.Name, run.Failures)...)
			l.Unattributed = append(l.Unattributed, namedOutput(run.Name, run.Output)...)
		}
	}
}

func scoredAncestor(name string, scored map[string]*testRun) *testRun {
	for {
		index := strings.LastIndex(name, "/")
		if index < 0 {
			return nil
		}
		name = name[:index]
		if run := scored[name]; run != nil {
			return run
		}
	}
}

func namedOutput(name string, lines []string) []string {
	if len(lines) == 0 {
		return nil
	}
	return append([]string{"--- " + name + " ---"}, lines...)
}

// scanTestLog attributes the output of a test run to the tests that produced it.
// Score lines are kept separate from ordinary output.
func scanTestLog(out string) *testLog {
	s := &testLogScanner{
		log: &testLog{byName: make(map[string]*testRun)},
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

type testLogScanner struct {
	log     *testLog
	current *testRun // the test the text framing is inside of
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

func (s *testLogScanner) line(line string) {
	if ev, ok := decodeEvent(line); ok {
		s.event(ev)
		return
	}
	s.output(line, s.current, "")
}

// event records a test2json event. The event names the test it belongs to, so
// it does not rely on the framing that the text form tracks.
func (s *testLogScanner) event(ev event) {
	switch ev.Action {
	case "attr":
		s.structuredAttrs = true
		// Only score attributes are kept; any other attribute is metadata about
		// the test rather than output from it.
		if HasPrefix(ev.Value) {
			s.test(ev.Test).Scores = append(s.test(ev.Test).Scores, ev.Value)
		}
	case "run":
		s.test(ev.Test)
	case "pass", "fail", "skip":
		if ev.Test == "" {
			return // the package's own result, not a test's
		}
		run := s.test(ev.Test)
		run.Status = testStatusOf(ev.Action)
		run.Elapsed = ev.Elapsed
	case "output":
		var run *testRun
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
func (s *testLogScanner) output(line string, run *testRun, outputType string) {
	if outputType == "frame" {
		return
	}
	if s.framing(line) {
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

// test returns the named test's run, recording it in start order if new.
func (s *testLogScanner) test(name string) *testRun {
	if run, ok := s.log.byName[name]; ok {
		return run
	}
	run := &testRun{Name: name}
	s.log.byName[name] = run
	s.log.Tests = append(s.log.Tests, run)
	return run
}

func testStatusOf(action string) TestStatus {
	switch action {
	case "pass":
		return TestStatus_PASSED
	case "fail":
		return TestStatus_FAILED
	case "skip":
		return TestStatus_SKIPPED
	}
	return TestStatus_NOT_RUN
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
