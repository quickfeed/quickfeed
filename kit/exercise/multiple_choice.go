package exercise

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// compile regular expressions only once
var (
	qNumRegExp      = regexp.MustCompile(`^(\d+)\.\s.*$`)
	selectionRegExp = regexp.MustCompile(`^\s+\-\s\[(x|X)\]\s+([a-f])\)\s.*$`)
)

// ParseMarkdownAnswers returns a map of the answers found in the given answer file.
func ParseMarkdownAnswers(answerFile string) (map[int]string, error) {
	md, err := ioutil.ReadFile(answerFile)
	if err != nil {
		return nil, err
	}

	currentQ := -1
	// map: question# -> answer label
	answerMap := make(map[int]string)
	for _, line := range strings.Split(string(md), "\n") {
		if qNumRegExp.MatchString(line) {
			qNum := qNumRegExp.ReplaceAllString(line, "$1")
			// ignore error since regular expression ensure it is already a number
			currentQ, _ = strconv.Atoi(qNum)
		}
		_, found := answerMap[currentQ]
		if !found && currentQ != -1 && selectionRegExp.MatchString(line) {
			answerMap[currentQ] = selectionRegExp.ReplaceAllString(line, "$2")
		}
	}
	return answerMap, nil
}

// CheckMultipleChoice returns the result of comparing the answers to the correct maps.
// The answers and correct maps from keys representing the question number to the labels (answer value).
// The question numbers (keys) in the correct map must contain all question numbers in the range 1 - len(correct).
// The returned slices contain question numbers deemed correctly and incorrectly answered, respectively.
func CheckMultipleChoice(answers, correct map[int]string) (correctA []int, incorrectA []int) {
	for qNum, label := range correct {
		if answers[qNum] == label {
			correctA = append(correctA, qNum)
			continue
		} else {
			incorrectA = append(incorrectA, qNum)
		}
	}
	sort.Ints(correctA)
	sort.Ints(incorrectA)
	return
}

// Print returns a string representation of the given list of questions.
// The preLabel and afterLabel precedes and succeed the question number,
// and all but the last question number is preceded by the sep separator.
func Print(questions []int, preLabel, afterLabel, sep string) string {
	var b strings.Builder
	for i, q := range questions {
		fmt.Fprintf(&b, "%s%d%s", preLabel, q, afterLabel)
		if i < len(questions)-1 {
			fmt.Fprint(&b, sep)
		}
	}
	return b.String()
}
