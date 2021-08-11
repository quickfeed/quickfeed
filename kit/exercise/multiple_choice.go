package exercise

import (
	"io/ioutil"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/autograde/quickfeed/kit/score"
	"github.com/google/go-cmp/cmp"
)

// compile regular expressions only once
var (
	qNumRegExp      = regexp.MustCompile(`^(\d+)\.\s.*$`)
	selectionRegExp = regexp.MustCompile(`^\s+\-\s\[(x|X)\]\s+([a-f])\)\s.*$`)
)

func parseMCAnswers(mdFile string) (map[string]string, error) {
	md, err := ioutil.ReadFile(mdFile)
	if err != nil {
		return nil, err
	}

	var curQ string
	// map: question# -> answer label
	qaMap := make(map[string]string)
	for _, line := range strings.Split(string(md), "\n") {
		if qNumRegExp.MatchString(line) {
			curQ = qNumRegExp.ReplaceAllString(line, "$1")
		}
		_, found := qaMap[curQ]
		if !found && curQ != "" && selectionRegExp.MatchString(line) {
			qaMap[curQ] = selectionRegExp.ReplaceAllString(line, "$2")
		}
	}
	return qaMap, nil
}

// MultipleChoiceWithDesc computes the score of a multiple choice exercise
// with student providing answers in the mdFile, where the correct map is
// expected to contain the correct answers. The function emits a JSON Score
// object and a corresponding message for x/y test cases passed.
func MultipleChoiceWithDesc(t *testing.T, mdFile string, correct map[int]string) {
	t.Helper()
	sc := score.NewScoreMax(t, len(correct), 1)
	defer sc.Print(t)

	qaMap, err := parseMCAnswers(mdFile)
	if err != nil {
		sc.Score = 0
		t.Fatal(err)
	}
	// sort map keys: question numbers
	qNumbers := make([]int, 0, len(correct))
	for k := range correct {
		qNumbers = append(qNumbers, k)
	}
	sort.Ints(qNumbers)

	for _, qNum := range qNumbers {
		ans, found := qaMap[strconv.Itoa(qNum)]
		if !found || !cmp.Equal(correct[qNum], ans) {
			t.Errorf("%v: Question %d: Answer not found or incorrect.\n", sc.TestName, qNum)
			sc.Dec()
			continue
		}
	}
}
