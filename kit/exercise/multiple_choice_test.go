package exercise_test

import (
	"path/filepath"
	"testing"

	"github.com/autograde/quickfeed/kit/exercise"
)

func TestMultipleChoiceWithDesc(t *testing.T) {
	tests := []struct {
		name  string
		file  string
		qaMap map[int]string
	}{
		{
			name:  "BlankAnswers",
			file:  "c-prog-questions-blank-answers.md",
			qaMap: map[int]string{},
		},
		{
			name:  "PartialAnswers",
			file:  "c-prog-questions-partial-answers.md",
			qaMap: map[int]string{1: "c", 4: "c", 6: "a"},
		},
		{
			name:  "AllAnswersSomeVCheckMark",
			file:  "c-prog-questions-all-answers-v-mark.md",
			qaMap: map[int]string{1: "a", 3: "c", 5: "b", 6: "b", 7: "d"},
		},
		{
			name:  "AllAnswers",
			file:  "c-prog-questions-all-answers.md",
			qaMap: map[int]string{1: "a", 2: "b", 3: "c", 4: "a", 5: "b", 6: "b", 7: "d"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exercise.MultipleChoiceWithDesc(t, filepath.Join("..", "testdata", test.file), test.qaMap)
		})
	}
}
