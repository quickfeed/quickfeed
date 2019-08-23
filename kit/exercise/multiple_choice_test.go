package exercise_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/autograde/kit/exercise"
	"github.com/autograde/kit/score"
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

var markDownLines = `
## The Coolest Answers (don't remove this line)

1. C
2.   b
3. ACABD
4. AbC
5. B
6. 
7. D dd (maybe this should fail but currently doesn't)
8. C   
9. ABCD
10. 
11.
`

var expectToFail = []int{
	3, 4, 6, 8,
}

func TestMultipleChoice(t *testing.T) {
	t.Skip("This is expected to fail, so we skip it when running normally (see comment).")
	// This currently fails, since it tests what students might write.
	// TODO(meling) In the future we may decouple it better so that we can
	// check if specific tests are expected to fail, and reorganizing it
	// as a table-driven test.
	answerFile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatal(err)
	}
	// clean up
	defer os.Remove(answerFile.Name())

	if _, err := answerFile.Write([]byte(markDownLines)); err != nil {
		t.Fatal(err)
	}
	if err := answerFile.Close(); err != nil {
		t.Fatal(err)
	}

	sc := score.NewScoreMax(len(answers), 1)
	exercise.MultipleChoice(t, sc, answerFile.Name(), answers)
}
