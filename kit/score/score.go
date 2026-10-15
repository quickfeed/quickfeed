package score

import (
	"encoding/json"
	"fmt"
	"math"
	"runtime/debug"
	"strings"
)

// fail sets Score to zero.
func (s *Score) fail() {
	s.Score = 0
}

// inc increments score if score is less than MaxScore.
func (s *Score) inc() {
	if s.GetScore() < s.GetMaxScore() {
		s.Score++
	}
}

// incBy increments score n times or until score equals MaxScore.
func (s *Score) incBy(n int) {
	m := int32(n)
	if s.GetScore()+m < s.GetMaxScore() {
		s.Score += m
	} else {
		s.Score = s.GetMaxScore()
	}
}

// dec decrements score if score is greater than zero.
func (s *Score) dec() {
	if s.GetScore() > 0 {
		s.Score--
	}
}

// decBy decrements score n times or until Score equals zero.
func (s *Score) decBy(n int) {
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

// scoreAttrKey is the attribute key that a score is reported under.
const scoreAttrKey = "score"

// emit reports the score to the test framework, which attributes it to the test
// that reported it even when the output of parallel tests is interleaved. The
// framework emits attributes only in chatty mode, so a run without -v or -json
// falls back to printing the score.
func (s *Score) emit(t interface{ Attr(key, value string) }, chatty bool) {
	if chatty {
		// The score is marshalled JSON, whose newlines are escaped, so it
		// satisfies Attr's requirement that a value hold none.
		t.Attr(scoreAttrKey, s.json())
		return
	}
	// We rely on JSON score objects to start on a new line, since otherwise
	// scanning long student generated output lines can be costly.
	fmt.Println()
	// print JSON score object: {"Secret":"my secret code","TestName": ...}
	fmt.Println(s.json())
}

// json returns a JSON string for the score object.
func (s *Score) json() string {
	b, err := json.Marshal(s)
	if err != nil {
		return fmt.Sprintf("json.Marshal error: %v\n", err)
	}
	return string(b)
}

func printPanicMessage(testName, msg string, recoverVal any) {
	var s strings.Builder
	s.WriteString("******************\n")
	s.WriteString(testName)
	s.WriteString(" panicked: ")
	fmt.Fprintf(&s, "%v", recoverVal)
	if msg != "" {
		s.WriteString("\n\nMessage:\n")
		s.WriteString(msg)
	}
	s.WriteString("\n\nStack trace from panic:\n")
	s.WriteString(string(debug.Stack()))
	s.WriteString("******************\n")
	fmt.Println(s.String())
}

func firstMessage(messages []string) string {
	if len(messages) > 0 {
		return messages[0]
	}
	return ""
}
