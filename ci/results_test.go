package ci

import (
	"strings"
	"testing"
	"time"

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
