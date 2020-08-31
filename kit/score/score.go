package score

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
)

// GlobalSecret represents the unique course identifier that will be used in
// the Score constructors. Users of this package must set this variable
// appropriately (for example in func init) before using any exported
// function in this package. The value of the global secret is available from
// the teachers panel after a course has been created.
var GlobalSecret = "NOT SET"

// Score encodes the score of a test or a group of tests. When a test passes in
// Autograder, a JSON object representing this struct is emitted to the output
// stream.
//
// The JSON object emitted on the output stream contains a Secret hash value
// which is a unique course identifier that can be obtained from the teachers
// panel in Autograder. This Secret is used by Autograder to extract Score
// objects from the output stream. All other output is ignored when computing
// the score.
//
// The Autograder computes the score according to the formula below, providing
// a percentage score for a test or a group of tests. The Weight parameter can
// be used to give more/less value to some Score objects (representing
// different test sets). For example, a Weight of 2 on test A and a Weight of 1
// on all other tests will give twice the score for test A compare to the
// other tests.
//
// If you want to only give a score for completing a test, then you can simply
// use NewScoreMax(1, 1), without using any API methods to decrement the score,
// giving a result of Score = MaxScore = 1 (and Weight = 1).
//
// The Autograder computes the final score as follows:
// TotalWeight     = sum(Weight)
// TaskScore[i]    = Score[i] / MaxScore[i], gives {0 < TaskScore < 1}
// TaskWeight[i]   = Weight[i] / TotalWeight
// Score           = sum(TaskScore[i]*TaskWeight[i]), gives {0 < Score < 1}
type Score struct {
	Secret   string // the unique identifier for the course
	TestName string // name of the test
	Score    int    // the score obtained
	MaxScore int    // max score possible to get on this specific test
	Weight   int    // the weight of this test; used to compute final grade
}

// NewScore returns a new Score object with the given max and weight.
// The Score.Score field is 0 initially, and so Score.Inc() and IncBy() can
// be called on the returned Score object.
// The TestName is initialized as the name of the provided t.Name().
func NewScore(t *testing.T, max, weight int) *Score {
	sc := &Score{
		Secret:   GlobalSecret,
		TestName: t.Name(),
		MaxScore: max,
		Weight:   weight,
	}
	sc.WriteJSON(os.Stdout)
	return sc
}

// NewScoreMax returns a new Score object with the given max and weight.
// The Score.Score field is max initially, and so Score.Dec() and DecBy() can
// be called on the returned Score object.
// The TestName is initialized as the name of the provided t.Name().
func NewScoreMax(t *testing.T, max, weight int) *Score {
	sc := NewScore(t, max, weight)
	sc.Score = max
	return sc
}

// IncBy increments score n times or until score equals MaxScore.
func (s *Score) IncBy(n int) {
	if s.Score+n < s.MaxScore {
		s.Score += n
	} else {
		s.Score = s.MaxScore
	}
}

// Inc increments score if score is less than MaxScore.
func (s *Score) Inc() {
	if s.Score < s.MaxScore {
		s.Score++
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

// String returns a string representation of score s.
// Format: "TestName: 2/10 cases passed".
func (s *Score) String() string {
	return fmt.Sprintf("%s: %d/%d cases passed", s.TestName, s.Score, s.MaxScore)
}

// WriteString writes the string representation of s to w.
func (s *Score) WriteString(w io.Writer) {
	// check if calling func panicked before calling this
	if r := recover(); r != nil {
		// reset score for panicked functions
		s.Score = 0
	}
	fmt.Fprintf(w, "%v\n", s)
}

// WriteJSON writes the JSON representation of s to w.
func (s *Score) WriteJSON(w io.Writer) {
	// check if calling func panicked before calling this
	if r := recover(); r != nil {
		// reset score for panicked functions
		s.Score = 0
	}
	b, err := json.Marshal(s)
	if err != nil {
		fmt.Fprintf(w, "json.Marshal error: \n%v\n", err)
	}
	fmt.Fprintf(w, "\n%s\n", b)
}
