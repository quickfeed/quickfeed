package exercise

import (
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/autograde/quickfeed/kit/score"
)

// Choices are the set of correct choices for the questions.
type Choices []struct {
	Number int
	Want   rune
}

// MultipleChoice computes the score of a multiple choice exercise
// with student answers provided in fileName, and the answers provided
// in the answerKey object. The function requires a Score object, and
// will produce both string output and JSON output.
func MultipleChoice(t *testing.T, sc *score.Score, fileName string, answers Choices) {
	t.Helper()
	defer sc.WriteString(os.Stdout)
	defer sc.WriteJSON(os.Stdout)

	// Read the whole file
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		sc.Score = 0
		t.Fatalf("%v: error reading the file: %v", fileName, err)
		return
	}

	for i := range answers {
		// find the user's answer to the corresponding question number
		regexStr := "\n" + strconv.Itoa(answers[i].Number) + "[.)]*[ \t\v\r\n\f]*([A-Za-z]*)"
		regex := regexp.MustCompile(regexStr)
		ans := regex.FindStringSubmatch(string(bytes))
		if len(ans) < 1 {
			t.Errorf("%v %d: Answer not found.\n", sc.TestName, answers[i].Number)
			sc.Dec()
			continue
		}
		match := ans[1]
		if len(match) == 0 {
			t.Errorf("%v %d: Answer not found.\n", sc.TestName, answers[i].Number)
			sc.Dec()
			continue
		}
		if len(match) > 1 {
			t.Errorf("%v %d: Multiple answers for question: %s\n", sc.TestName, answers[i].Number, match)
			sc.Dec()
			continue
		}
		got := strings.ToUpper(match)
		if !strings.ContainsRune(got, answers[i].Want) {
			t.Errorf("%v %d: %q is incorrect.\n", sc.TestName, answers[i].Number, got)
			sc.Dec()
		}
	}
}
