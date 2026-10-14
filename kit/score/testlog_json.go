package score

import (
	"encoding/json"
	"strings"
)

// go test -json emits test2json events that name the test for each event.
// This reader uses those names to attribute output without tracking text markers.

// event is the subset of a test2json event that attribution needs.
type event struct {
	Action     string
	Test       string
	Elapsed    float64
	Output     string
	OutputType string
	Value      string
}

// jsonEvent records a test2json event. The event names the test it belongs to, so
// it does not rely on the test markers that the text form tracks.
func (s *testLogScanner) jsonEvent(ev event) {
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

// decodeEvent reports whether line is a test2json event, and decodes it.
func decodeEvent(line string) (event, bool) {
	// Time is omitted for cached results and by test2json without -t.
	if !strings.HasPrefix(line, `{"Time":`) && !strings.HasPrefix(line, `{"Action":`) {
		return event{}, false
	}
	var ev event
	if err := json.Unmarshal([]byte(line), &ev); err != nil || ev.Action == "" {
		return event{}, false
	}
	return ev, true
}
