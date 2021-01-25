package score

import (
	"fmt"
	"io"
	"testing"
)

// NewScore returns a new Score object with the given max and weight.
// The Score is initially 0, and Inc() and IncBy() can be called
// on the returned Score object, for each test that passes.
// The TestName is initialized to the name of the provided t.Name().
// This function also prints a JSON representation of the Score object
// to ensure that the test is recorded by Quickfeed.
//
// Deprecated: Do not use this function; it will be removed in the future.
// Use Add() in conjunction with score.Min() instead for regular tests,
// and AddSub() in conjunction with score.MinByName() instead for subtests.
//
func NewScore(t *testing.T, max, weight int) *Score {
	sc := &Score{
		Secret:   sessionSecret,
		TestName: t.Name(),
		MaxScore: int32(max),
		Weight:   int32(weight),
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
//
// Deprecated: Do not use this function; it will be removed in the future.
// Use Add() in conjunction with score.Max() instead for regular tests,
// and AddSub() in conjunction with score.MaxByName() instead for subtests.
//
func NewScoreMax(t *testing.T, max, weight int) *Score {
	sc := NewScore(t, max, weight)
	sc.Score = int32(max)
	return sc
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
