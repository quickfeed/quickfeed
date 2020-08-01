package exercise_test

import (
	"path/filepath"
	"testing"

	"github.com/autograde/quickfeed/kit/exercise"
	"github.com/autograde/quickfeed/kit/score"
)

var answers = exercise.Choices{
	{1, 'C'},
	{2, 'B'},
	{3, 'C'},
	{4, 'A'},
	{5, 'B'},
	{6, 'D'},
	{7, 'D'},
	{8, 'D'},
}

var expectToFail = []int{
	3, 4, 6, 8,
}

func TestMultipleChoice(t *testing.T) {
	t.Skip("This is expected to fail, so we skip it when running normally (see comment).")
	// This currently fails, since it tests what students might write.
	// TODO(meling) In the future we may decouple it better so that we can
	// check if specific tests are expected to fail, and reorganizing it
	// as a table-driven test.

	oldStyleMC := filepath.Join("..", "testdata", "old-style-answers.md")
	sc := score.NewScoreMax(len(answers), 1)
	exercise.MultipleChoice(t, sc, oldStyleMC, answers)
}

func TestMultipleChoiceWithDesc(t *testing.T) {
	tests := []struct {
		name  string
		file  string
		qaMap map[string]string
	}{
		{
			name:  "BlankAnswers",
			file:  "c-prog-questions-blank-answers.md",
			qaMap: map[string]string{},
		},
		{
			name:  "PartialAnswers",
			file:  "c-prog-questions-partial-answers.md",
			qaMap: map[string]string{"1": "c", "4": "c", "6": "a"},
		},
		{
			name:  "AllAnswersSomeVCheckMark",
			file:  "c-prog-questions-all-answers-v-mark.md",
			qaMap: map[string]string{"1": "a", "3": "c", "5": "b", "6": "b", "7": "d"},
		},
		{
			name:  "AllAnswers",
			file:  "c-prog-questions-all-answers.md",
			qaMap: map[string]string{"1": "a", "2": "b", "3": "c", "4": "a", "5": "b", "6": "b", "7": "d"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exercise.MultipleChoiceWithDesc(t, filepath.Join("..", "testdata", test.file), test.qaMap)
		})
	}
}
