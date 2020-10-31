package score

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"testing"
)

const (
	secretEnvName = "QUICKFEED_SESSION_SECRET"
)

var sessionSecret string

func init() {
	sessionSecret = os.Getenv(secretEnvName)
	// remove variable as soon as it has been read
	_ = os.Setenv(secretEnvName, "")
}

// Score encodes the score of a test or a group of tests.
type Score struct {
	Secret   string // the unique identifier for a scoring session
	TestName string // name of the test
	Score    int    // the score obtained
	MaxScore int    // max score possible to get on this specific test
	Weight   int    // the weight of this test; used to compute final grade
}

// NewScore returns a new Score object with the given max and weight.
// The Score is initially 0, and Inc() and IncBy() can be called
// on the returned Score object, for each test that passes.
// The TestName is initialized to the name of the provided t.Name().
// This function also prints a JSON representation of the Score object
// to ensure that the test is recorded by Quickfeed.
func NewScore(t *testing.T, max, weight int) *Score {
	sc := &Score{
		Secret:   sessionSecret,
		TestName: t.Name(),
		MaxScore: max,
		Weight:   weight,
	}
	// prints JSON score object with zero score, e.g.:
	// {"Secret":"my secret code","TestName":"TestPanicHandler","Score":0,"MaxScore":8,"Weight":5}
	// This registers the test, in case a panic occurs that prevents printing the score object.
	fmt.Println(sc.json())
	return sc
}

// NewScoreMax returns a new Score object with the given max and weight.
// The Score is initially set to max, and Dec() and DecBy() can be called
// on the returned Score object, for each test that fails.
// The TestName is initialized as the name of the provided t.Name().
// This function also prints a JSON representation of the Score object
// to ensure that the test is recorded by Quickfeed.
func NewScoreMax(t *testing.T, max, weight int) *Score {
	sc := NewScore(t, max, weight)
	sc.Score = max
	return sc
}

// Inc increments score if score is less than MaxScore.
func (s *Score) Inc() {
	if s.Score < s.MaxScore {
		s.Score++
	}
}

// IncBy increments score n times or until score equals MaxScore.
func (s *Score) IncBy(n int) {
	if s.Score+n < s.MaxScore {
		s.Score += n
	} else {
		s.Score = s.MaxScore
	}
}

// Dec decrements score if score is greater than zero.
func (s *Score) Dec() {
	if s.Score > 0 {
		s.Score--
	}
}

// DecBy decrements score n times or until Score equals zero.
func (s *Score) DecBy(n int) {
	if s.Score-n > 0 {
		s.Score -= n
	} else {
		s.Score = 0
	}
}

// String returns a string representation of the score.
// Format: "TestName: 2/10 test cases passed".
func (s Score) String() string {
	return fmt.Sprintf("%s: %d/%d test cases passed", s.TestName, s.Score, s.MaxScore)
}

// WriteString writes the string representation of s to w.
// Deprecated: Do not use this function; it will be removed in the future.
// Use Print() instead to replace both WriteString() and WriteJSON().
func (s *Score) WriteString(w io.Writer) {
	if r := recover(); r != nil {
		s.Score = 0
		fmt.Fprintf(w, "******************\n%s panicked:\n%s\n******************\n", s.TestName, r)
	}
	fmt.Fprintf(w, "%v\n", s)
}

// WriteJSON writes the JSON representation of s to w.
// Deprecated: Do not use this function; it will be removed in the future.
// Use Print() instead to replace both WriteString() and WriteJSON().
func (s *Score) WriteJSON(w io.Writer) {
	if r := recover(); r != nil {
		s.Score = 0
		fmt.Fprintf(w, "******************\n%s panicked:\n%s\n******************\n", s.TestName, r)
	}
	fmt.Fprintf(w, "\n%s\n", s.json())
}

// Print prints both the JSON secret string and emits the number of test cases passed.
// If a test panics, the score will be set to zero, and a panic message will be emitted.
// Note that, if subtests are used, each subtest must defer call the PanicHandler method
// to ensure that panics are caught and handled appropriately.
func (s *Score) Print(t *testing.T) {
	if r := recover(); r != nil {
		s.fail(t)
		printPanicMessage(s.TestName, r)
	}
	// print JSON score object: {"Secret":"my secret code","TestName": ...}
	fmt.Println(s.json())
	// print: TestName: x/y test cases passed
	fmt.Println(s)
}

// PanicHandler recovers from a panicking test, resets the score to zero and
// emits an error message. This is only needed when using a single score object
// for multiple subtests each of which may panic, which would prevent the deferred
// Print() call from executing its recovery handler.
//
// This must be called as a deferred function from within a subtest, that is
// within a t.Run() function:
//   defer s.PanicHandler(t)
//
func (s *Score) PanicHandler(t *testing.T) {
	if r := recover(); r != nil {
		s.fail(t)
		printPanicMessage(s.TestName, r)
	}
}

// fail resets the score to zero and fails the provided test.
func (s *Score) fail(t *testing.T) {
	// reset score for panicked test functions
	s.Score = 0
	// fail the test
	t.Fail()
}

// json returns a JSON string for the score object.
func (s Score) json() string {
	b, err := json.Marshal(s)
	if err != nil {
		return fmt.Sprintf("json.Marshal error: %v\n", err)
	}
	return string(b)
}

func printPanicMessage(testName string, recoverVal interface{}) {
	var s strings.Builder
	s.WriteString("******************\n")
	s.WriteString(testName)
	s.WriteString(" panicked: ")
	s.WriteString(fmt.Sprintf("%v", recoverVal))
	s.WriteString("\n\nStack trace from panic:\n")
	s.WriteString(string(debug.Stack()))
	s.WriteString("******************\n")
	fmt.Println(s.String())
}
