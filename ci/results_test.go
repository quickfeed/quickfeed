package ci

import (
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/autograde/quickfeed/kit/score"
	"github.com/autograde/quickfeed/log"
	"go.uber.org/zap"
)

func TestExtractResult(t *testing.T) {
	out := `here is some output in the log.

{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":100,"MaxScore":100,"Weight":1}

Here are some more logs for the student.
`

	res, err := ExtractResult(zap.NewNop().Sugar(), out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res.BuildInfo.BuildLog, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73") {
		t.Fatal("build log contains secret")
		t.Logf("res %+v", res.BuildInfo)
	}
}

func TestExtractResultWithWhitespace(t *testing.T) {
	out := `here is some output in the log with whitespace before the JSON string below.

    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":100,"MaxScore":100,"Weight":1}

Here are some more logs for the student.
`

	res, err := ExtractResult(zap.NewNop().Sugar(), out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res.BuildInfo.BuildLog, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73") {
		t.Fatal("build log contains secret")
		t.Logf("res %+v", res.BuildInfo)
	}
}

func TestExtractResultWithTwoScoreLines(t *testing.T) {
	out := `here is some output in the log with whitespace before the JSON string below.

    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":0,"MaxScore":100,"Weight":1}

Here are some more logs for the student.
    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":100,"MaxScore":100,"Weight":1}

	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"JoGo","Score":0,"MaxScore":100,"Weight":1}

Here are some more logs for the student.
`

	res, err := ExtractResult(zap.NewNop().Sugar(), out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Scores) != 2 {
		t.Fatalf("ExtractResult() expected 2 Score entries, got %d: %+v", len(res.Scores), res.Scores)
	}
}

func TestExtractResultWithPanicedAndMaliciousScoreLines(t *testing.T) {
	out := `
    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"GoodTest1","Score":0,"MaxScore":100,"Weight":1}
    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"GoodTest1","Score":100,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"GoodTest2","Score":0,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"GoodTest2","Score":50,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"PanicedTest1","Score":0,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"PanicedTest2","Score":0,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"PanicedTest3","Score":0,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"MaliciousTest","Score":100,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"MaliciousTest","Score":100,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"MaliciousTest","Score":100,"MaxScore":100,"Weight":1}
`

	res, err := ExtractResult(zap.NewNop().Sugar(), out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10)
	if err != nil {
		t.Fatal(err)
	}
	const expected = 6
	if len(res.Scores) != expected {
		t.Fatalf("ExtractResult() expected %d Score entries, got %d: %+v", expected, len(res.Scores), res.Scores)
	}

	testOrder := []string{
		"GoodTest1",
		"GoodTest2",
		"PanicedTest1",
		"PanicedTest2",
		"PanicedTest3",
		"MaliciousTest",
	}
	for i, score := range res.Scores {
		if score.TestName != testOrder[i] {
			t.Errorf("ExtractResult() returned unexpected order of tests: expected %s, got %s", testOrder[i], score.TestName)
		}
	}
}

// scoreObjects is obtained using this query (dat320-2020/lab4):
// select score_objects from submissions where user_id='19' and assignment_id='8';
var scoreObjects = `
[{"Secret":"hidden","TestName":"TestLintAG","Score":3,"MaxScore":3,"Weight":5},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/No_jobs","Score":0,"MaxScore":0,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/Two_jobs","Score":2,"MaxScore":2,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/Three_jobs","Score":3,"MaxScore":3,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/Five_jobs","Score":5,"MaxScore":5,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/Six_jobs","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/Six_jobs_unordered","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/No_jobs","Score":0,"MaxScore":0,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/Two_jobs","Score":10,"MaxScore":10,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/Three_jobs","Score":15,"MaxScore":15,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/Five_jobs","Score":25,"MaxScore":25,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/Six_jobs","Score":28,"MaxScore":28,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/Six_jobs_unordered","Score":28,"MaxScore":28,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/No_jobs","Score":0,"MaxScore":0,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/Two_jobs","Score":4,"MaxScore":4,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/Three_jobs","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/Five_jobs","Score":10,"MaxScore":10,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/Six_jobs","Score":12,"MaxScore":12,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/Six_jobs_unordered","Score":12,"MaxScore":12,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/No_jobs","Score":0,"MaxScore":0,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/Two_jobs","Score":2,"MaxScore":2,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/Three_jobs","Score":3,"MaxScore":3,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/Five_jobs","Score":5,"MaxScore":5,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/Six_jobs","Score":8,"MaxScore":8,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/Six_jobs_unordered","Score":8,"MaxScore":8,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/No_jobs","Score":0,"MaxScore":0,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Two_jobs","Score":2,"MaxScore":2,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Three_jobs","Score":3,"MaxScore":3,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Five_jobs","Score":5,"MaxScore":5,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Six_jobs","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Six_jobs_unordered","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Six_jobs_different_unordered","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SS(5)/No_jobs","Score":0,"MaxScore":0,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SS(5)/ABC_jobs","Score":12,"MaxScore":12,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SS(5)/ABC_jobs_long","Score":60,"MaxScore":60,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SS(5)/Varying_length_ABC_jobs","Score":32,"MaxScore":32,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SS(5)/ABCDE_jobs","Score":84,"MaxScore":84,"Weight":2}]
`

func TestTotalScore(t *testing.T) {
	result := &Result{}
	result.Scores = make([]*score.Score, 0)
	dec := json.NewDecoder(strings.NewReader(scoreObjects))
	for {
		if err := dec.Decode(&result.Scores); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
	}
	got := result.TotalScore()
	const want = 100
	if got != want {
		t.Errorf("TotalScore() = '%d', want '%d'", got, want)
		t.Logf("scores: %+v", log.IndentJson(result.Scores))
	}
}

func TestExecTime(t *testing.T) {
	tests := []struct {
		id   string
		in   time.Duration
		want int64
	}{
		{"1", 1_000_000_000, 1000},
		{"2", 2_000_000_000, 2000},
		{"3", 2_550_000_000, 2550},
		{"4", 2_800_000_000, 2800},
		{"5", 3_888_900_000, 3888},
	}
	logger := zap.NewNop().Sugar()
	for _, tt := range tests {
		t.Run("ExecTime#"+tt.id, func(t *testing.T) {
			res, err := ExtractResult(logger, "", "", tt.in)
			if err != nil {
				t.Fatal(err)
			}
			got := res.BuildInfo.ExecTime
			if got != tt.want {
				t.Errorf("ExtractResult(..., %q) = '%v', want '%v'", tt.in, got, tt.want)
			}
		})
	}
}
