package score

import (
	"encoding/json"
	"fmt"
	"math"
	"runtime/debug"
	"strings"
	"testing"
)

// Fail sets Score to zero.
func (s *Score) Fail() {
	s.Score = 0
}

// Inc increments score if score is less than MaxScore.
func (s *Score) Inc() {
	if s.Score < s.MaxScore {
		s.Score++
	}
}

// IncBy increments score n times or until score equals MaxScore.
func (s *Score) IncBy(n int) {
	m := int32(n)
	if s.Score+m < s.MaxScore {
		s.Score += m
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
	m := int32(n)
	if s.Score-m > 0 {
		s.Score -= m
	} else {
		s.Score = 0
	}
}

// Normalize the score to the given maxScore.
func (s *Score) Normalize(maxScore int) {
	f := float64(maxScore) / float64(s.MaxScore)
	normScore := float64(s.Score) * f
	s.Score = int32(math.Round(normScore))
	s.MaxScore = int32(maxScore)
}

// Equal returns true if s equals other. Ignores the Secret field.
func (s *Score) Equal(other *Score) bool {
	return other != nil &&
		s.TestName == other.TestName &&
		s.Score == other.Score &&
		s.MaxScore == other.MaxScore &&
		s.Weight == other.Weight
}

// RelativeScore returns a string with the following format:
// "TestName: score = x/y = s".
func (s *Score) RelativeScore() string {
	return fmt.Sprintf("%s: score = %d/%d = %.1f", s.TestName, s.Score, s.MaxScore, float32(s.Score)/float32(s.MaxScore))
}

// Print prints a JSON representation of the score that can be picked up by QuickFeed.
// To ensure that panic message and stack trace is printed, this method must be called via defer.
// If a test panics, the score will be set to zero, and a panic message will be emitted.
// Note that, if subtests are used, each subtest must defer call the PanicHandler method
// to ensure that panics are caught and handled appropriately.
// The msg parameter is optional, and will be printed in case of a panic.
func (s *Score) Print(t *testing.T, msg ...string) {
	if r := recover(); r != nil {
		s.fail(t)
		printPanicMessage(s.TestName, msg[0], r)
	}
	// We rely on JSON score objects to start on a new line, since otherwise
	// scanning long student generated output lines can be costly.
	fmt.Println()
	// print JSON score object: {"Secret":"my secret code","TestName": ...}
	fmt.Println(s.json())
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
// The msg parameter is optional, and will be printed in case of a panic.
func (s *Score) PanicHandler(t *testing.T, msg ...string) {
	if r := recover(); r != nil {
		s.fail(t)
		printPanicMessage(t.Name(), msg[0], r)
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
func (s *Score) json() string {
	b, err := json.Marshal(s)
	if err != nil {
		return fmt.Sprintf("json.Marshal error: %v\n", err)
	}
	return string(b)
}

func printPanicMessage(testName, msg string, recoverVal interface{}) {
	var s strings.Builder
	s.WriteString("******************\n")
	s.WriteString(testName)
	s.WriteString(" panicked: ")
	s.WriteString(fmt.Sprintf("%v", recoverVal))
	if msg != "" {
		s.WriteString("\n\nMessage:\n")
		s.WriteString(msg)
	}
	s.WriteString("\n\nStack trace from panic:\n")
	s.WriteString(string(debug.Stack()))
	s.WriteString("******************\n")
	fmt.Println(s.String())
}
