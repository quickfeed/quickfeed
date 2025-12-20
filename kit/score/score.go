package score

import (
	"encoding/json"
	"fmt"
	"math"
	"runtime"
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
	if s.GetScore() < s.GetMaxScore() {
		s.Score++
	}
}

// IncBy increments score n times or until score equals MaxScore.
func (s *Score) IncBy(n int) {
	m := int32(n)
	if s.GetScore()+m < s.GetMaxScore() {
		s.Score += m
	} else {
		s.Score = s.GetMaxScore()
	}
}

// Dec decrements score if score is greater than zero.
func (s *Score) Dec() {
	if s.GetScore() > 0 {
		s.Score--
	}
}

// DecBy decrements score n times or until Score equals zero.
func (s *Score) DecBy(n int) {
	m := int32(n)
	if s.GetScore()-m > 0 {
		s.Score -= m
	} else {
		s.Score = 0
	}
}

// Normalize the score to the given maxScore.
func (s *Score) Normalize(maxScore int) {
	f := float64(maxScore) / float64(s.GetMaxScore())
	normScore := float64(s.GetScore()) * f
	s.Score = int32(math.Round(normScore))
	s.MaxScore = int32(maxScore)
}

// weightedScore returns the weighted score for this test score.
func (s *Score) weightedScore(totalWeight float64) float64 {
	return weightedScore(float64(s.GetScore()), float64(s.GetMaxScore()), float64(s.GetWeight()), totalWeight)
}

// Equal returns true if s equals other. Ignores the Secret field.
func (s *Score) Equal(other *Score) bool {
	return other != nil &&
		s.GetTestName() == other.GetTestName() &&
		s.GetScore() == other.GetScore() &&
		s.GetMaxScore() == other.GetMaxScore() &&
		s.GetWeight() == other.GetWeight()
}

// RelativeScore returns a string with the following format:
// "TestName: score = x/y = s".
func (s *Score) RelativeScore() string {
	return fmt.Sprintf("%s: score = %d/%d = %.1f", s.GetTestName(), s.GetScore(), s.GetMaxScore(), float32(s.GetScore())/float32(s.GetMaxScore()))
}

// Print prints a JSON representation of the score that can be picked up by QuickFeed.
// To ensure that panic message and stack trace is printed, this method must be called via defer.
// If a test panics, the score will be set to zero, and a panic message will be emitted.
// Note that, if subtests are used, each subtest must defer call the PanicHandler method
// to ensure that panics are caught and handled appropriately.
// The msg parameter is optional, and will be printed in case of a panic.
func (s *Score) Print(t *testing.T, msg ...string) {
	if r := recover(); r != nil {
		s.internalFail(t)
		printPanicMessage(s.GetTestName(), msg[0], r)
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
//
//	defer s.PanicHandler(t)
//
// The msg parameter is optional, and will be printed in case of a panic.
func (s *Score) PanicHandler(t *testing.T, msg ...string) {
	if r := recover(); r != nil {
		s.internalFail(t)
		printPanicMessage(t.Name(), msg[0], r)
	}
}

// internalFail resets the score to zero and fails the provided test.
func (s *Score) internalFail(t *testing.T) {
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

// Error prints an error message and sets the TestDetails field.
// Calling this method will fail the test.
func (s *Score) Error(t *testing.T, args ...any) {
	t.Helper()
	s.addDetails(t, "%v", args...)
	t.Error(args...)
}

// Errorf prints an error message and sets the TestDetails field.
// Calling this method will fail the test.
func (s *Score) Errorf(t *testing.T, format string, args ...any) {
	t.Helper()
	s.addDetails(t, format, args...)
	t.Errorf(format, args...)
}

// Fatal prints an error message and sets the TestDetails field.
// Calling this method will fail the test and stop the execution of the test.
func (s *Score) Fatal(t *testing.T, args ...any) {
	t.Helper()
	s.addDetails(t, "%v", args...)
	t.Fatal(args...)
}

// Fatalf prints an error message and sets the TestDetails field.
// Calling this method will fail the test and stop the execution of the test.
func (s *Score) Fatalf(t *testing.T, format string, args ...any) {
	t.Helper()
	s.addDetails(t, format, args...)
	t.Fatalf(format, args...)
}

// addDetails updates the TestDetails field with the provided error message.
func (s *Score) addDetails(t *testing.T, format string, args ...any) {
	t.Helper()
	// this function is called from Errorf and Error, which are called from tests
	// we want to get the file and line number of the test function that called Errorf or Error
	_, file, line, ok := runtime.Caller(2)
	if ok {
		// decorate the error message with file and line number
		// to make it easier to locate the source of the error
		s.TestDetails += fmt.Sprintf("%s:%d: %s\n", file, line, fmt.Sprintf(format, args...))
	} else {
		s.TestDetails += fmt.Sprintln(fmt.Sprintf(format, args...))
	}
}

func printPanicMessage(testName, msg string, recoverVal any) {
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
