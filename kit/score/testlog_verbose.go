package score

import (
	"regexp"
	"strconv"
	"strings"
)

// go test -v emits text markers such as === RUN and --- FAIL to identify tests
// and their outcomes. This reader tracks those markers to attribute text output.
// Older go test -json events also carry the markers in their Output field.
var (
	// "=== RUN   TestStack", and the PAUSE, CONT and NAME markers that track
	// which test the following output belongs to.
	testNameMarker = regexp.MustCompile(`^=== (RUN|PAUSE|CONT|NAME)\s+(.+?)\s*$`)
	// "=== ATTR  TestStack score {...}", printed by testing.T.Attr.
	testAttrMarker = regexp.MustCompile(`^=== ATTR\s+(\S+)\s+(\S+)\s(.*)$`)
	// "--- FAIL: TestQueue (0.00s)", indented once per level of subtest.
	testResultMarker = regexp.MustCompile(`^\s*--- (PASS|FAIL|SKIP):\s+(.+?)\s+\(([0-9.]+)s\)\s*$`)
)

// textMarker reports whether line is the test framework describing a test rather
// than output from one, recording what the line says.
func (s *testLogScanner) textMarker(line string) bool {
	if m := testAttrMarker.FindStringSubmatch(line); m != nil {
		// The attribute's own test name is authoritative even when the output
		// of parallel tests is interleaved.
		if !s.structuredAttrs && HasPrefix(m[3]) {
			run := s.test(m[1])
			run.Scores = append(run.Scores, m[3])
		}
		return true
	}
	if m := testNameMarker.FindStringSubmatch(line); m != nil {
		s.current = s.test(m[2])
		return true
	}
	if m := testResultMarker.FindStringSubmatch(line); m != nil {
		run := s.test(m[2])
		run.Status = testStatusOf(strings.ToLower(m[1]))
		if elapsed, err := strconv.ParseFloat(m[3], 64); err == nil {
			run.Elapsed = elapsed
		}
		// The test has reported its result, so what follows is the package's.
		s.current = nil
		return true
	}
	return false
}
