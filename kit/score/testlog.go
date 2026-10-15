package score

import "strings"

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

// testLog collects output from a course run script and its Go tests.
// Output is handled according to its format:
//
//   - go test -v: test markers such as === RUN and --- FAIL identify tests
//     and their outcomes (testlog_verbose.go).
//   - go test -json: test2json events name the test for each event
//     (testlog_json.go).
//   - Both formats, plain go test, and surrounding run-script output share
//     message classification and score-line extraction (testlog_output.go).
//
// scanTestLog selects the reader for each captured line.
//
// Score lines are kept separate from student-visible output because they carry
// the session secret. Results validates the scores, attaches diagnostics and
// output, and redacts the secret before the result is stored or shown.
//
// Test markers and events identify scores, outcomes, and testing messages.
// Raw stdout from parallel tests can interleave without identifying its source,
// so attributing those writes to a test remains best effort.
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

// retainOutput keeps the output of tests that emitted no score from being lost.
// The scored map holds the runs that emitted one, keyed by Go test name.
//
// A scored test often runs unscored subtests, such as the t.Run cases of a
// table-driven test that decrements the parent's score. Their t.Errorf messages
// explain why the parent lost points, so they are appended to the nearest
// scored ancestor under a "--- name ---" header. A test with no scored
// ancestor, such as one the course never registered, is appended to the build
// log instead.
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

// testLogScanner attributes scores and output while reading a run.
// It tracks the current test from verbose markers and avoids recording attributes
// twice when JSON reports them as both attribute events and text output.
type testLogScanner struct {
	log     *testLog
	current *testRun // the test named by the most recent text marker
	// structuredAttrs records that this run reported attributes as events. Such
	// a run also prints each attribute as an "=== ATTR" line in an output event,
	// which would otherwise record the same score a second time.
	structuredAttrs bool
}

// line processes each captured line from scanTestLog as a JSON event or text.
// It handles go test with or without -v/-json, plus surrounding run-script output.
func (s *testLogScanner) line(line string) {
	if ev, ok := decodeEvent(line); ok {
		s.jsonEvent(ev)
		return
	}
	s.output(line, s.current, "")
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
