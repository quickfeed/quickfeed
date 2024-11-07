package score_test

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/quickfeed/quickfeed/kit/score"
)

func TestRelativeScore(t *testing.T) {
	sc := &score.Score{
		TestName: t.Name(),
		MaxScore: 10,
		Score:    3,
	}
	rs := sc.RelativeScore()
	expectedRS := t.Name() + ": score = 3/10 = 0.3"
	if rs != expectedRS {
		t.Errorf(`RelativeScore() = %q, expected %q`, rs, expectedRS)
	}
}

func TestNormalize(t *testing.T) {
	sc := &score.Score{
		TestName: t.Name(),
		MaxScore: 100,
		Score:    33,
	}
	newMaxScore := 50
	sc.Normalize(newMaxScore)
	expectedScore := int32(17)
	if sc.Score != expectedScore {
		t.Errorf("Normalize(%d) = %d, expected %d", newMaxScore, sc.Score, expectedScore)
	}
}

func TestScoreDetails(t *testing.T) {
	sc := &score.Score{}

	messages := []string{"first", "second", "third"}
	expectedMessages := messages[:2]

	wg := &sync.WaitGroup{}
	wg.Add(1)
	// run the test in a goroutine to avoid Fatalf
	// from exiting the test via t.FailNow (runtime.Goexit)
	mockT := &testing.T{}
	go func() {
		defer wg.Done()
		for i, m := range messages {
			if i == 1 {
				sc.Fatalf(mockT, m)
			} else {
				sc.Errorf(mockT, m)
			}
		}
	}()
	wg.Wait()
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	sc.Print(t)
	w.Close()
	os.Stdout = originalStdout

	out := make([]byte, 1024)
	n, _ := r.Read(out)

	parsedScore := &score.Score{}
	json.Unmarshal(out[:n], parsedScore)

	for i, m := range expectedMessages {
		if !strings.Contains(parsedScore.TestDetails, m) {
			t.Errorf("TestDetails does not contain error message %d: %s", i, m)
		}
	}

	// the third message should not be in the TestDetails
	// as test execution should have stopped after the second message
	// due to the call to sc.Fatalf
	if strings.Contains(parsedScore.TestDetails, messages[2]) {
		t.Errorf("TestDetails contains unexpected error message: %s", messages[2])
	}

	if !mockT.Failed() {
		t.Error("Test did not fail")
	}
}
