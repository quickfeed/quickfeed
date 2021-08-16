package exercise_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autograde/quickfeed/kit/exercise"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func init() {
	scores.Add(TestMultipleChoice0, len(tests[0].correct), 1)
	scores.Add(TestMultipleChoice1, len(tests[1].correct), 1)
	scores.Add(TestMultipleChoice2, len(tests[2].correct), 1)
	scores.Add(TestMultipleChoice3, len(tests[3].correct), 1)
	for _, test := range tests {
		scores.AddSub(TestMultipleChoiceWithFail, test.name, len(test.correct), 1)
	}
}

var tests = []struct {
	name          string
	file          string
	answers       map[int]string
	correct       map[int]string
	wantCorrect   []int
	wantIncorrect []int
}{
	{
		name:          "BlankAnswers",
		file:          "c-prog-questions-blank-answers.md",
		answers:       map[int]string{},
		correct:       map[int]string{1: "a", 2: "b", 3: "c", 4: "a", 5: "b", 6: "b", 7: "d"},
		wantCorrect:   []int{},
		wantIncorrect: []int{1, 2, 3, 4, 5, 6, 7},
	},
	{
		name:          "PartialAnswers",
		file:          "c-prog-questions-partial-answers.md",
		answers:       map[int]string{1: "c", 4: "c", 6: "a"},
		correct:       map[int]string{1: "a", 2: "b", 3: "c", 4: "a", 5: "b", 6: "b", 7: "d"},
		wantCorrect:   []int{},
		wantIncorrect: []int{1, 2, 3, 4, 5, 6, 7},
	},
	{
		name:          "AllAnswersSomeVCheckMark",
		file:          "c-prog-questions-all-answers-v-mark.md",
		answers:       map[int]string{1: "a", 3: "c", 5: "b", 6: "b", 7: "d"},
		correct:       map[int]string{1: "a", 2: "b", 3: "c", 4: "a", 5: "b", 6: "b", 7: "d"},
		wantCorrect:   []int{1, 3, 5, 6, 7},
		wantIncorrect: []int{2, 4},
	},
	{
		name:          "AllAnswers",
		file:          "c-prog-questions-all-answers.md",
		answers:       map[int]string{1: "a", 2: "b", 3: "c", 4: "a", 5: "b", 6: "b", 7: "d"},
		correct:       map[int]string{1: "a", 2: "b", 3: "c", 4: "a", 5: "b", 6: "b", 7: "d"},
		wantCorrect:   []int{1, 2, 3, 4, 5, 6, 7},
		wantIncorrect: []int{},
	},
}

func TestParseMarkdownAnswers(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			answerFile := filepath.Join("..", "testdata", test.file)
			gotAnswers, err := exercise.ParseMarkdownAnswers(answerFile)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(test.answers, gotAnswers); diff != "" {
				t.Errorf("TestParseMarkdownAnswers() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCheckMultipleChoice(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotCorrect, gotIncorrect := exercise.CheckMultipleChoice(test.answers, test.correct)
			// Example code for use of Print function
			if len(gotCorrect) > 0 {
				t.Log("Correct answers:\n", exercise.Print(gotCorrect, "\tQuestion ", " is correct", "\n"))
			}
			if len(gotIncorrect) > 0 {
				t.Log("Incorrect or missing answers:", exercise.Print(gotIncorrect, "Q", "=W", ", "))
			}
			if diff := cmp.Diff(test.wantCorrect, gotCorrect, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("TestCheckMultipleChoice():correct mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.wantIncorrect, gotIncorrect, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("TestCheckMultipleChoice():incorrect mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// To run these tests use:
//   QUICKFEED_FAIL_TEST=1 go test -v -run TestMultipleChoice
//

const (
	failTestEnvName = "QUICKFEED_FAIL_TEST"
)

func TestMultipleChoice0(t *testing.T) {
	// This test aims to emulate that student submitted code may result in a panic,
	// and thus a test failure along with a stack trace would be expected.
	// Hence, we do not run this as part of the CI tests. To run, see instructions below.
	failTest := os.Getenv(failTestEnvName)
	if failTest == "" {
		t.Skipf("Skipping; expected to fail. Run with: %s=1 go test -v -run %s", failTestEnvName, t.Name())
	}

	test := tests[0]
	sc := scores.Max()
	defer sc.Print(t)
	answerFile := filepath.Join("..", "testdata", test.file)
	answers, err := exercise.ParseMarkdownAnswers(answerFile)
	if err != nil {
		t.Error(err)
	}
	_, gotIncorrect := exercise.CheckMultipleChoice(answers, test.correct)
	for _, incorrect := range gotIncorrect {
		t.Errorf("%v: Question %d: Answer not found or incorrect.\n", sc.TestName, incorrect)
		sc.Dec()
	}
}

func TestMultipleChoice1(t *testing.T) {
	// This test aims to emulate that student submitted code may result in a panic,
	// and thus a test failure along with a stack trace would be expected.
	// Hence, we do not run this as part of the CI tests. To run, see instructions below.
	failTest := os.Getenv(failTestEnvName)
	if failTest == "" {
		t.Skipf("Skipping; expected to fail. Run with: %s=1 go test -v -run %s", failTestEnvName, t.Name())
	}

	test := tests[1]
	sc := scores.Max()
	defer sc.Print(t)
	answerFile := filepath.Join("..", "testdata", test.file)
	answers, err := exercise.ParseMarkdownAnswers(answerFile)
	if err != nil {
		t.Error(err)
	}
	_, gotIncorrect := exercise.CheckMultipleChoice(answers, test.correct)
	for _, incorrect := range gotIncorrect {
		t.Errorf("%v: Question %d: Answer not found or incorrect.\n", sc.TestName, incorrect)
		sc.Dec()
	}
}

func TestMultipleChoice2(t *testing.T) {
	// This test aims to emulate that student submitted code may result in a panic,
	// and thus a test failure along with a stack trace would be expected.
	// Hence, we do not run this as part of the CI tests. To run, see instructions below.
	failTest := os.Getenv(failTestEnvName)
	if failTest == "" {
		t.Skipf("Skipping; expected to fail. Run with: %s=1 go test -v -run %s", failTestEnvName, t.Name())
	}

	test := tests[2]
	sc := scores.Max()
	defer sc.Print(t)
	answerFile := filepath.Join("..", "testdata", test.file)
	answers, err := exercise.ParseMarkdownAnswers(answerFile)
	if err != nil {
		t.Error(err)
	}
	_, gotIncorrect := exercise.CheckMultipleChoice(answers, test.correct)
	for _, incorrect := range gotIncorrect {
		t.Errorf("%v: Question %d: Answer not found or incorrect.\n", sc.TestName, incorrect)
		sc.Dec()
	}
}

func TestMultipleChoice3(t *testing.T) {
	// This test aims to emulate that student submitted code may result in a panic,
	// and thus a test failure along with a stack trace would be expected.
	// Hence, we do not run this as part of the CI tests. To run, see instructions below.
	failTest := os.Getenv(failTestEnvName)
	if failTest == "" {
		t.Skipf("Skipping; expected to fail. Run with: %s=1 go test -v -run %s", failTestEnvName, t.Name())
	}

	test := tests[3]
	sc := scores.Max()
	defer sc.Print(t)
	answerFile := filepath.Join("..", "testdata", test.file)
	answers, err := exercise.ParseMarkdownAnswers(answerFile)
	if err != nil {
		t.Error(err)
	}
	_, gotIncorrect := exercise.CheckMultipleChoice(answers, test.correct)
	for _, incorrect := range gotIncorrect {
		t.Errorf("%v: Question %d: Answer not found or incorrect.\n", sc.TestName, incorrect)
		sc.Dec()
	}
}

func TestMultipleChoiceWithFail(t *testing.T) {
	// This test aims to emulate that student submitted code may result in a panic,
	// and thus a test failure along with a stack trace would be expected.
	// Hence, we do not run this as part of the CI tests. To run, see instructions below.
	failTest := os.Getenv(failTestEnvName)
	if failTest == "" {
		t.Skipf("Skipping; expected to fail. Run with: %s=1 go test -v -run %s", failTestEnvName, t.Name())
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sc := scores.MaxByName(t.Name())
			defer sc.Print(t)
			answerFile := filepath.Join("..", "testdata", test.file)
			answers, err := exercise.ParseMarkdownAnswers(answerFile)
			if err != nil {
				t.Error(err)
			}
			_, gotIncorrect := exercise.CheckMultipleChoice(answers, test.correct)
			for _, incorrect := range gotIncorrect {
				t.Errorf("%v: Question %d: Answer not found or incorrect.\n", sc.TestName, incorrect)
				sc.Dec()
			}
		})
	}
}
