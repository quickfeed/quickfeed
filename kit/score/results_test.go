package score_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/kit/score/testdata/server"
)

// secret is the session secret the run output below is scored with.
const secret = "59fd5fe1c4f741604c1beeab875b9c789d2a7c73"

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

func TestExtractResultsParsedScores(t *testing.T) {
	const secret = "59fd5fe1c4f741604c1beeab875b9c789d2a7c73"
	expectedTests := []*score.Score{
		{TestName: "TestA", Score: 0, MaxScore: 100, Weight: 1},
	}
	tests := []struct {
		name             string
		out              string
		wantParsedScores int
	}{
		{name: "NoOutput", out: "", wantParsedScores: 0},
		{name: "NoScoreLines", out: "compile error\nsome log", wantParsedScores: 0},
		{
			name:             "OneScoreLine",
			out:              `{"Secret":"` + secret + `","TestName":"TestA","Score":50,"MaxScore":100,"Weight":1}`,
			wantParsedScores: 1,
		},
		{
			// A valid score line for a test that is not among the expected
			// tests must not count as parsed; otherwise the run would appear
			// successful even though none of the expected tests reported a score.
			name:             "UnexpectedScoreLine",
			out:              `{"Secret":"` + secret + `","TestName":"TestUnknown","Score":50,"MaxScore":100,"Weight":1}`,
			wantParsedScores: 0,
		},
		{
			// A score line with the wrong secret must not count as parsed.
			name:             "WrongSecret",
			out:              `{"Secret":"wrong","TestName":"TestA","Score":50,"MaxScore":100,"Weight":1}`,
			wantParsedScores: 0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, _ := score.ExtractResults(tc.out, secret, 10, expectedTests)
			if res.ParsedScores != tc.wantParsedScores {
				t.Errorf("ExtractResults() ParsedScores = %d, want %d", res.ParsedScores, tc.wantParsedScores)
			}
		})
	}
}

// TestExtractResultsFromNonTestCaller checks that extraction of invalid score
// lines does not panic when called from outside a test function, which is how
// the QuickFeed server calls it. Calling ExtractResults directly from a test
// would not catch this, since the call frame lookup in kit/internal/test finds
// the test function on the stack and returns without panicking.
func TestExtractResultsFromNonTestCaller(t *testing.T) {
	const secret = "59fd5fe1c4f741604c1beeab875b9c789d2a7c73"
	expectedTests := []*score.Score{
		{TestName: "TestA", Score: 0, MaxScore: 100, Weight: 1},
	}
	tests := []struct {
		name string
		out  string
	}{
		{name: "EmptyTestName", out: `{"Secret":"` + secret + `","TestName":"","Score":50,"MaxScore":100,"Weight":1}`},
		{name: "ZeroMaxScore", out: `{"Secret":"` + secret + `","TestName":"TestA","Score":50,"MaxScore":0,"Weight":1}`},
		{name: "ZeroWeight", out: `{"Secret":"` + secret + `","TestName":"TestA","Score":50,"MaxScore":100,"Weight":0}`},
		{name: "ScoreAboveMaxScore", out: `{"Secret":"` + secret + `","TestName":"TestA","Score":500,"MaxScore":100,"Weight":1}`},
		{name: "NegativeScore", out: `{"Secret":"` + secret + `","TestName":"TestA","Score":-1,"MaxScore":100,"Weight":1}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := server.ExtractResults(tc.out, secret, expectedTests)
			if err == nil {
				t.Errorf("ExtractResults(%s) = nil error, want parse error", tc.out)
			}
			if res.ParsedScores != 0 {
				t.Errorf("ExtractResults(%s) ParsedScores = %d, want 0", tc.out, res.ParsedScores)
			}
		})
	}
}

// runOutput is what a course's run script produces for one assignment: the
// script's own banners around the framing of go test -v.
const runOutput = `*** Preparing Test Execution for lab1 ***

*** Running Tests ***

=== RUN   TestStack
stack: pushing 3 elements
=== ATTR  TestStack score {"Secret":"%[1]s","TestName":"TestStack","Score":5,"MaxScore":5,"Weight":1,"TestDetails":""}
--- PASS: TestStack (0.02s)
=== RUN   TestQueue
queue: dequeue returned nothing
    queue_test.go:44: Pop() = <nil>, want: x
=== ATTR  TestQueue score {"Secret":"%[1]s","TestName":"TestQueue","Score":2,"MaxScore":5,"Weight":2,"TestDetails":""}
--- FAIL: TestQueue (0.01s)
FAIL
FAIL	lab1	0.289s

*** Finished Running Tests in 3 seconds ***
`

func extractRunOutput(t *testing.T) *score.Results {
	t.Helper()
	expectedTests := []*score.Score{
		{TestName: "TestStack", MaxScore: 5, Weight: 1},
		{TestName: "TestQueue", MaxScore: 5, Weight: 2},
	}
	res, err := score.ExtractResults(fmt.Sprintf(runOutput, secret), secret, 10, expectedTests)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestExtractResultsAttachesEachTestsOutput(t *testing.T) {
	res := extractRunOutput(t)

	byName := make(map[string]*score.Score)
	for _, sc := range res.Scores {
		byName[sc.GetTestName()] = sc
	}
	stack := byName["TestStack"]
	if got, want := stack.GetTestOutput(), "stack: pushing 3 elements"; got != want {
		t.Errorf("TestStack output = %q, want %q", got, want)
	}
	if got, want := stack.GetStatus(), score.TestStatus_PASSED; got != want {
		t.Errorf("TestStack status = %v, want %v", got, want)
	}
	if got, want := stack.GetElapsed(), 0.02; got != want {
		t.Errorf("TestStack elapsed = %v, want %v", got, want)
	}

	queue := byName["TestQueue"]
	if got, want := queue.GetStatus(), score.TestStatus_FAILED; got != want {
		t.Errorf("TestQueue status = %v, want %v", got, want)
	}
	if got, want := queue.GetTestOutput(), "queue: dequeue returned nothing"; got != want {
		t.Errorf("TestQueue output = %q, want %q", got, want)
	}
	// A failing test's diagnostics explain the failure. With no details from
	// the test itself, they are what the student has to go on.
	if got, want := queue.GetTestDetails(), "queue_test.go:44: Pop() = <nil>, want: x"; got != want {
		t.Errorf("TestQueue details = %q, want %q", got, want)
	}
}

func TestExtractResultsKeepsOnlyUnattributedOutputInBuildLog(t *testing.T) {
	res := extractRunOutput(t)

	buildLog := res.GetBuildInfo().GetBuildLog()
	for _, want := range []string{"*** Running Tests ***", "*** Finished Running Tests in 3 seconds ***", "FAIL\tlab1\t0.289s"} {
		if !strings.Contains(buildLog, want) {
			t.Errorf("build log = %q, want it to contain %q", buildLog, want)
		}
	}
	// What a test printed is shown with that test, so repeating it here would
	// only restore the wall of text this replaces.
	for _, unwanted := range []string{"stack: pushing 3 elements", "queue: dequeue returned nothing", "Pop() = <nil>", "=== RUN", "--- FAIL:"} {
		if strings.Contains(buildLog, unwanted) {
			t.Errorf("build log = %q, want it to omit %q", buildLog, unwanted)
		}
	}
}

func TestExtractResultsKeepsTheSecretOutOfEverythingShown(t *testing.T) {
	res := extractRunOutput(t)

	if strings.Contains(res.GetBuildInfo().GetBuildLog(), secret) {
		t.Error("build log leaks the session secret")
	}
	for _, sc := range res.Scores {
		if strings.Contains(sc.GetTestOutput(), secret) {
			t.Errorf("%s output leaks the session secret", sc.GetTestName())
		}
		if strings.Contains(sc.GetTestDetails(), secret) {
			t.Errorf("%s details leak the session secret", sc.GetTestName())
		}
	}
}

func TestExtractResultsPrefersDetailsReportedByTheTest(t *testing.T) {
	// A test that reports through the score object already says why it failed,
	// in its own words and with its own positions; scraped diagnostics would
	// only repeat it.
	out := fmt.Sprintf(`=== RUN   TestQueue
    queue_test.go:44: Pop() = <nil>, want: x
=== ATTR  TestQueue score {"Secret":"%s","TestName":"TestQueue","Score":0,"MaxScore":5,"Weight":1,"TestDetails":"queue_test.go:44: Pop() = <nil>, want: \"x\"\n"}
--- FAIL: TestQueue (0.01s)
`, secret)
	expectedTests := []*score.Score{{TestName: "TestQueue", MaxScore: 5, Weight: 1}}
	res, err := score.ExtractResults(out, secret, 10, expectedTests)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := res.Scores[0].GetTestDetails(), "queue_test.go:44: Pop() = <nil>, want: \"x\"\n"; got != want {
		t.Errorf("TestQueue details = %q, want %q", got, want)
	}
}
