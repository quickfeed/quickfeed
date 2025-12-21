package score_test

import (
	"strings"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/kit/score"
)

func TestExtractResults(t *testing.T) {
	out := `here is some output in the log.

{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":100,"MaxScore":100,"Weight":1}

Here are some more logs for the student.
`

	expectedTests := []*score.Score{
		{TestName: "Gradle", Score: 0, MaxScore: 100, Weight: 1},
	}
	res, err := score.ExtractResults(out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10, expectedTests)
	if err != nil {
		// err may contain multiple errors
		t.Fatal(err)
	}
	if strings.Contains(res.GetBuildInfo().GetBuildLog(), "59fd5fe1c4f741604c1beeab875b9c789d2a7c73") {
		t.Fatal("build log contains secret")
		t.Logf("res %+v", res.GetBuildInfo())
	}
}

func TestExtractResultsWithWhitespace(t *testing.T) {
	out := `here is some output in the log with whitespace before the JSON string below.

    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":100,"MaxScore":100,"Weight":1}

Here are some more logs for the student.
`

	expectedTests := []*score.Score{
		{TestName: "Gradle", Score: 0, MaxScore: 100, Weight: 1},
	}
	res, err := score.ExtractResults(out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10, expectedTests)
	if err != nil {
		// err may contain multiple errors
		t.Fatal(err)
	}
	if strings.Contains(res.GetBuildInfo().GetBuildLog(), "59fd5fe1c4f741604c1beeab875b9c789d2a7c73") {
		t.Fatal("build log contains secret")
		t.Logf("res %+v", res.GetBuildInfo())
	}
}

func TestExtractResultsWithTwoScoreLines(t *testing.T) {
	out := `here is some output in the log with whitespace before the JSON string below.

    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":0,"MaxScore":100,"Weight":1}

Here are some more logs for the student.
    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":100,"MaxScore":100,"Weight":1}

	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"JoGo","Score":0,"MaxScore":100,"Weight":1}

Here are some more logs for the student.
`

	expectedTests := []*score.Score{
		{TestName: "Gradle", Score: 0, MaxScore: 100, Weight: 1},
		{TestName: "JoGo", Score: 0, MaxScore: 100, Weight: 1},
	}
	res, err := score.ExtractResults(out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10, expectedTests)
	if err != nil {
		// err may contain multiple errors
		t.Fatal(err)
	}
	if len(res.Scores) != 2 {
		t.Fatalf("ExtractResult() expected 2 Score entries, got %d: %+v", len(res.Scores), res.Scores)
	}
}

func TestExtractResultsWithMultipleZeroScoreLines(t *testing.T) {
	out := `
    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":0,"MaxScore":100,"Weight":1}
    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":0,"MaxScore":100,"Weight":1}
    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":50,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"JoGo","Score":0,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"JoGo","Score":0,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"JoGo","Score":0,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"JoGo","Score":50,"MaxScore":100,"Weight":1}
`

	expectedTests := []*score.Score{
		{TestName: "Gradle", Score: 0, MaxScore: 100, Weight: 1},
		{TestName: "JoGo", Score: 0, MaxScore: 100, Weight: 1},
	}
	res, err := score.ExtractResults(out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10, expectedTests)
	if err != nil {
		// err may contain multiple errors
		t.Fatal(err)
	}
	if len(res.Scores) != 2 {
		t.Fatalf("ExtractResult() expected 2 Score entries, got %d: %+v", len(res.Scores), res.Scores)
	}
	for _, score := range res.Scores {
		if score.GetScore() != 50 {
			t.Errorf("ExtractResult() expected 50, got %d", score.GetScore())
		}
	}
}

func TestExtractResultsWithMultipleNonZeroScoreLines(t *testing.T) {
	out := `
    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":0,"MaxScore":100,"Weight":1}
    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":0,"MaxScore":100,"Weight":1}
    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":50,"MaxScore":100,"Weight":1}
    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":100,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"JoGo","Score":0,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"JoGo","Score":20,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"JoGo","Score":0,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"JoGo","Score":30,"MaxScore":100,"Weight":1}
`

	expectedTests := []*score.Score{
		{TestName: "Gradle", Score: 0, MaxScore: 100, Weight: 1},
		{TestName: "JoGo", Score: 0, MaxScore: 100, Weight: 1},
	}
	res, err := score.ExtractResults(out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10, expectedTests)
	if err != nil {
		// err may contain multiple errors
		t.Fatal(err)
	}
	if len(res.Scores) != 2 {
		t.Fatalf("ExtractResult() expected 2 Score entries, got %d: %+v", len(res.Scores), res.Scores)
	}
	for _, score := range res.Scores {
		if score.GetScore() != -1 {
			t.Errorf("ExtractResult() expected -1, got %d", score.GetScore())
		}
	}
}

func TestExtractResultsWithPanickedAndMaliciousScoreLines(t *testing.T) {
	out := `
    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"GoodTest1","Score":0,"MaxScore":100,"Weight":1}
    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"GoodTest1","Score":100,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"GoodTest2","Score":0,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"GoodTest2","Score":50,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"PanickedTest1","Score":0,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"PanickedTest2","Score":0,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"PanickedTest3","Score":0,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"MaliciousTest","Score":100,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"MaliciousTest","Score":100,"MaxScore":100,"Weight":1}
	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"MaliciousTest","Score":100,"MaxScore":100,"Weight":1}
`

	expectedTests := []*score.Score{
		{TestName: "GoodTest1", Score: 0, MaxScore: 100, Weight: 1},
		{TestName: "GoodTest2", Score: 0, MaxScore: 100, Weight: 1},
		{TestName: "PanickedTest1", Score: 0, MaxScore: 100, Weight: 1},
		{TestName: "PanickedTest2", Score: 0, MaxScore: 100, Weight: 1},
		{TestName: "PanickedTest3", Score: 0, MaxScore: 100, Weight: 1},
		{TestName: "MaliciousTest", Score: 0, MaxScore: 100, Weight: 1},
	}
	res, err := score.ExtractResults(out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10, expectedTests)
	if err != nil {
		// err may contain multiple errors
		t.Fatal(err)
	}
	const expectedTestCount = 6
	if len(res.Scores) != expectedTestCount {
		t.Fatalf("ExtractResult() expected %d Score entries, got %d: %+v", expectedTestCount, len(res.Scores), res.Scores)
	}

	testOrder := []string{
		"GoodTest1",
		"GoodTest2",
		"PanickedTest1",
		"PanickedTest2",
		"PanickedTest3",
		"MaliciousTest",
	}
	for i, sc := range res.Scores {
		if sc.TestName != testOrder[i] {
			t.Errorf("ExtractResult() returned unexpected order of tests: expected %s, got %s", testOrder[i], sc.TestName)
		}
	}
}

func TestExtractResultsExecTime(t *testing.T) {
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
	for _, tt := range tests {
		t.Run("ExecTime#"+tt.id, func(t *testing.T) {
			res, err := score.ExtractResults("", "", tt.in, nil)
			if err != nil {
				// err may contain multiple errors
				t.Fatal(err)
			}
			got := res.GetBuildInfo().GetExecTime()
			if got != tt.want {
				t.Errorf("ExtractResult(..., %q) = '%v', want '%v'", tt.in, got, tt.want)
			}
		})
	}
}

func TestExtractResultsWithExpectedTests(t *testing.T) {
	tests := []struct {
		name          string
		out           string
		secret        string
		expectedTests []*score.Score
		wantTestNames []string
		wantScores    []int32
	}{
		{
			name:          "NilExpectedTests",
			out:           `{"Secret":"secret","TestName":"TestA","Score":80,"MaxScore":100,"Weight":1}`,
			secret:        "secret",
			expectedTests: nil,
			wantTestNames: []string{},
			wantScores:    []int32{},
		},
		{
			name:          "EmptyExpectedTests",
			out:           `{"Secret":"secret","TestName":"TestA","Score":80,"MaxScore":100,"Weight":1}`,
			secret:        "secret",
			expectedTests: []*score.Score{},
			wantTestNames: []string{},
			wantScores:    []int32{},
		},
		{
			name:          "AllPresent",
			out:           `{"Secret":"secret","TestName":"TestA","Score":80,"MaxScore":100,"Weight":1}` + "\n" + `{"Secret":"secret","TestName":"TestB","Score":40,"MaxScore":50,"Weight":2}`,
			secret:        "secret",
			expectedTests: []*score.Score{{TestName: "TestA", MaxScore: 100, Weight: 1}, {TestName: "TestB", MaxScore: 50, Weight: 2}},
			wantTestNames: []string{"TestA", "TestB"},
			wantScores:    []int32{80, 40},
		},
		{
			name:          "MissingTest",
			out:           `{"Secret":"secret","TestName":"TestA","Score":80,"MaxScore":100,"Weight":1}`,
			secret:        "secret",
			expectedTests: []*score.Score{{TestName: "TestA", MaxScore: 100, Weight: 1}, {TestName: "TestB", MaxScore: 50, Weight: 2}},
			wantTestNames: []string{"TestA", "TestB"},
			wantScores:    []int32{80, 0}, // TestB should have score 0
		},
		{
			name:          "UnexpectedTestFiltered",
			out:           `{"Secret":"secret","TestName":"TestA","Score":80,"MaxScore":100,"Weight":1}` + "\n" + `{"Secret":"secret","TestName":"TestX","Score":90,"MaxScore":100,"Weight":1}`,
			secret:        "secret",
			expectedTests: []*score.Score{{TestName: "TestA", MaxScore: 100, Weight: 1}},
			wantTestNames: []string{"TestA"},
			wantScores:    []int32{80}, // TestX should be filtered out
		},
		{
			name:          "EmptyOutput",
			out:           "",
			secret:        "secret",
			expectedTests: []*score.Score{{TestName: "TestA", MaxScore: 100, Weight: 1}, {TestName: "TestB", MaxScore: 50, Weight: 2}},
			wantTestNames: []string{"TestA", "TestB"},
			wantScores:    []int32{0, 0}, // All tests should have score 0
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			results, err := score.ExtractResults(test.out, test.secret, 10*time.Millisecond, test.expectedTests)
			if err != nil {
				t.Fatal(err)
			}

			if len(results.Scores) != len(test.wantTestNames) {
				t.Errorf("Expected %d scores, got %d", len(test.wantTestNames), len(results.Scores))
			}

			// Check test names and scores
			scoreMap := make(map[string]int32)
			for _, score := range results.Scores {
				scoreMap[score.GetTestName()] = score.GetScore()
			}

			for i, wantTestName := range test.wantTestNames {
				gotScore, found := scoreMap[wantTestName]
				if !found {
					t.Errorf("Expected test %s not found in results", wantTestName)
					continue
				}
				if gotScore != test.wantScores[i] {
					t.Errorf("Test %s: expected score %d, got %d", wantTestName, test.wantScores[i], gotScore)
				}
			}
		})
	}
}
