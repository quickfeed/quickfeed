package score_test

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/autograde/quickfeed/kit/score"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestExtractResult(t *testing.T) {
	out := `here is some output in the log.

{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":100,"MaxScore":100,"Weight":1}

Here are some more logs for the student.
`

	res, err := score.ExtractResults(out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10)
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

	res, err := score.ExtractResults(out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10)
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

	res, err := score.ExtractResults(out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10)
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

	res, err := score.ExtractResults(out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10)
	if err != nil {
		t.Fatal(err)
	}
	const expectedTests = 6
	if len(res.Scores) != expectedTests {
		t.Fatalf("ExtractResult() expected %d Score entries, got %d: %+v", expectedTests, len(res.Scores), res.Scores)
	}

	testOrder := []string{
		"GoodTest1",
		"GoodTest2",
		"PanicedTest1",
		"PanicedTest2",
		"PanicedTest3",
		"MaliciousTest",
	}
	for i, sc := range res.Scores {
		if sc.TestName != testOrder[i] {
			t.Errorf("ExtractResult() returned unexpected order of tests: expected %s, got %s", testOrder[i], sc.TestName)
		}
	}
}

// scoreObjects is obtained using this query (dat320-2020/lab4):
// select score_objects from submissions where user_id='19' and assignment_id='8';
var scoreObjects = `
[{"Secret":"hidden","TestName":"TestLintAG","Score":3,"MaxScore":3,"Weight":5},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/No_jobs","Score":0,"MaxScore":0,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/Two_jobs","Score":2,"MaxScore":2,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/Three_jobs","Score":3,"MaxScore":3,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/Five_jobs","Score":5,"MaxScore":5,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/Six_jobs","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/Six_jobs_unordered","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/No_jobs","Score":0,"MaxScore":0,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/Two_jobs","Score":10,"MaxScore":10,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/Three_jobs","Score":15,"MaxScore":15,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/Five_jobs","Score":25,"MaxScore":25,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/Six_jobs","Score":28,"MaxScore":28,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(2)/Six_jobs_unordered","Score":28,"MaxScore":28,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/No_jobs","Score":0,"MaxScore":0,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/Two_jobs","Score":4,"MaxScore":4,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/Three_jobs","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/Five_jobs","Score":10,"MaxScore":10,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/Six_jobs","Score":12,"MaxScore":12,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(5)/Six_jobs_unordered","Score":12,"MaxScore":12,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/No_jobs","Score":0,"MaxScore":0,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/Two_jobs","Score":2,"MaxScore":2,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/Three_jobs","Score":3,"MaxScore":3,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/Five_jobs","Score":5,"MaxScore":5,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/Six_jobs","Score":8,"MaxScore":8,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/RR(10)/Six_jobs_unordered","Score":8,"MaxScore":8,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/No_jobs","Score":0,"MaxScore":0,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Two_jobs","Score":2,"MaxScore":2,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Three_jobs","Score":3,"MaxScore":3,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Five_jobs","Score":5,"MaxScore":5,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Six_jobs","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Six_jobs_unordered","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SJF/Six_jobs_different_unordered","Score":6,"MaxScore":6,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SS(5)/No_jobs","Score":0,"MaxScore":0,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SS(5)/ABC_jobs","Score":12,"MaxScore":12,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SS(5)/ABC_jobs_long","Score":60,"MaxScore":60,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SS(5)/Varying_length_ABC_jobs","Score":32,"MaxScore":32,"Weight":2},{"Secret":"hidden","TestName":"TestSchedulersAG/SS(5)/ABCDE_jobs","Score":84,"MaxScore":84,"Weight":2}]
`

func TestScoresSum(t *testing.T) {
	scores := make([]*score.Score, 0)
	dec := json.NewDecoder(strings.NewReader(scoreObjects))
	for {
		if err := dec.Decode(&scores); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
	}
	s := score.NewResults()
	const secret = "hidden"
	for _, sc := range scores {
		// The scoreObjects was extracted when we allowed Weight=0
		// We now return an error for when Weight=0.
		// Hence, we only add scores with non-zero weights.
		if err := sc.IsValid(secret); err == nil {
			s.AddScore(sc)
		}
	}
	results := &score.Results{Scores: s.ToScoreSlice()}
	// IsValid above redacts the Secret field with the empty string.
	// Hence, we call Validate with the empty string.
	if err := results.Validate(""); err != nil {
		t.Errorf("Validate() = %v, expected <nil>", err)
	}
	got := results.Sum()
	const want = 100
	if got != want {
		t.Errorf("Sum() = %d, want %d", got, want)
	}
}

// RegExp patterns to use to extract from JSON output.
//    Search: \{\W+"Secret": "hidden",\W+"(\w+)"(:.*)\W+"(\w+)"(:.*)\W+"(\w+)"(:.*)\W+"(\w+)"(:\W+\d+)\n(.*)
//   Replace: {$1$2$3$4$5$6$7$8$9
// To use, copy the JSON string start on the line after: "Scores": [
// And stop on the line before the corresponding ].
// You will need to add the final comma for the last element.

var score100 = []*score.Score{
	{TestName: "TestVetCheckAG", Score: 1, MaxScore: 1, Weight: 5},
	{TestName: "TestFormattingAG", Score: 1, MaxScore: 1, Weight: 5},
	{TestName: "TestTODOItemsAG", Score: 1, MaxScore: 1, Weight: 5},
	{TestName: "TestLintAG", Score: 1, MaxScore: 1, Weight: 5},
	{TestName: "TestAverageMetrics/fifo/book_schedule1", Score: 4, MaxScore: 4, Weight: 4},
	{TestName: "TestAverageMetrics/fifo/book_schedule2", Score: 4, MaxScore: 4, Weight: 4},
	{TestName: "TestAverageMetrics/fifo/book_schedule3", Score: 4, MaxScore: 4, Weight: 4},
	{TestName: "TestAverageMetrics/rr/book_schedule1/q=1ms", Score: 4, MaxScore: 4, Weight: 4},
	{TestName: "TestRoundRobin", Score: 169, MaxScore: 169, Weight: 30},
	{TestName: "TestSingleJobMetrics/rr/book_schedule3/q=1ms", Score: 2, MaxScore: 2, Weight: 2},
	{TestName: "TestAverageMetrics/rr/book_schedule2/q=1ms", Score: 4, MaxScore: 4, Weight: 4},
	{TestName: "TestAverageMetrics/rr/book_schedule3/q=1ms", Score: 4, MaxScore: 4, Weight: 4},
	{TestName: "TestShortestJobFirst", Score: 163, MaxScore: 163, Weight: 20},
	{TestName: "TestStride", Score: 248, MaxScore: 248, Weight: 30},
	{TestName: "TestMinPass", Score: 5, MaxScore: 5, Weight: 5},
	{TestName: "TestStrideNewJob", Score: 2, MaxScore: 2, Weight: 2},
	{TestName: "TestSingleJobMetrics/fifo/book_schedule1", Score: 2, MaxScore: 2, Weight: 2},
	{TestName: "TestSingleJobMetrics/fifo/book_schedule2", Score: 2, MaxScore: 2, Weight: 2},
	{TestName: "TestSingleJobMetrics/fifo/book_schedule3", Score: 2, MaxScore: 2, Weight: 2},
	{TestName: "TestSingleJobMetrics/rr/book_schedule1/q=1ms", Score: 2, MaxScore: 2, Weight: 2},
	{TestName: "TestSingleJobMetrics/rr/book_schedule2/q=1ms", Score: 2, MaxScore: 2, Weight: 2},
}

var score100v2 = []*score.Score{
	{TestName: "TestTODOItemsAG", Score: 1, MaxScore: 1, Weight: 5},
	{TestName: "TestAllocAG", Score: 14, MaxScore: 14, Weight: 20},
	{TestName: "TestAllocMultipleAG", Score: 63, MaxScore: 63, Weight: 10},
	{TestName: "TestFreeAG", Score: 40, MaxScore: 40, Weight: 20},
	{TestName: "TestPTLookupAG", Score: 12, MaxScore: 12, Weight: 10},
	{TestName: "TestNewMMUAG", Score: 12, MaxScore: 12, Weight: 10},
	{TestName: "TestReadAG", Score: 13, MaxScore: 13, Weight: 30},
	{TestName: "TestPTAppendAG", Score: 4, MaxScore: 4, Weight: 10},
	{TestName: "TestFormattingAG", Score: 1, MaxScore: 1, Weight: 5},
	{TestName: "TestLintAG", Score: 1, MaxScore: 1, Weight: 5},
	{TestName: "TestVetCheckAG", Score: 1, MaxScore: 1, Weight: 5},
	{TestName: "TestExtractAG", Score: 20, MaxScore: 20, Weight: 10},
	{TestName: "TestWriteAG", Score: 48, MaxScore: 48, Weight: 10},
	{TestName: "TestSequencesAG", Score: 16, MaxScore: 16, Weight: 40},
	{TestName: "TestMemoryManagementMultipleChoiceAG", Score: 3, MaxScore: 3, Weight: 5},
	{TestName: "TestPTFreeAG", Score: 18, MaxScore: 18, Weight: 10},
}

func TestScore100(t *testing.T) {
	const want = 100
	for i, sc100 := range [][]*score.Score{score100, score100v2} {
		t.Run(fmt.Sprintf("Sample%d", i), func(t *testing.T) {
			scoreTable := score.NewResults()
			for _, sc := range sc100 {
				scoreTable.AddScore(sc)
				if sc.Score != sc.MaxScore {
					// sanity check; all scores must be max
					t.Errorf("%s Score=%d, expected %d", sc.TestName, sc.Score, sc.MaxScore)
				}
			}
			results := &score.Results{Scores: scoreTable.ToScoreSlice()}
			if err := results.Validate(""); err != nil {
				t.Error(err)
			}
			got := results.Sum()
			if got != want {
				t.Errorf("Sum() = %d, want %d", got, want)
			}
		})
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
	for _, tt := range tests {
		t.Run("ExecTime#"+tt.id, func(t *testing.T) {
			res, err := score.ExtractResults("", "", tt.in)
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

var scoreTests = []struct {
	name string
	desc string
	in   []*score.Score
	want *score.Results
}{
	{
		name: "Record the score of the second emitted score object",
		desc: "First score is registration of the test, second score is the actual score.",
		in: []*score.Score{
			{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 0},
			{TestName: "B", Secret: theSecret, Weight: 20, MaxScore: 100, Score: 0},
			{TestName: "C", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 0},
			{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
			{TestName: "B", Secret: theSecret, Weight: 20, MaxScore: 100, Score: 60},
			{TestName: "C", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 70},
		},
		want: &score.Results{
			Scores: []*score.Score{
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
				{TestName: "B", Secret: theSecret, Weight: 20, MaxScore: 100, Score: 60},
				{TestName: "C", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 70},
			},
		},
	},
	{
		name: "TestName D is missing score",
		desc: "Can be due to test D panicking or some other reason for not emitting a score object",
		in: []*score.Score{
			{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 0},
			{TestName: "B", Secret: theSecret, Weight: 20, MaxScore: 100, Score: 0},
			{TestName: "C", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 0},
			{TestName: "D", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 0},
			{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
			{TestName: "B", Secret: theSecret, Weight: 20, MaxScore: 100, Score: 60},
			{TestName: "C", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 70},
		},
		want: &score.Results{
			Scores: []*score.Score{
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
				{TestName: "B", Secret: theSecret, Weight: 20, MaxScore: 100, Score: 60},
				{TestName: "C", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 70},
				{TestName: "D", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 0},
			},
		},
	},
	{
		name: "Test A recorded 3 times",
		desc: "We only allow the same test to be recorded two times",
		in: []*score.Score{
			{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 0},
			{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
			{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 100},
		},
		want: &score.Results{
			Scores: []*score.Score{
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: -1},
			},
		},
	},
	{
		name: "Test A with non-zero score recorded 3 times",
		desc: "We only allow the same test to be recorded two times",
		in: []*score.Score{
			{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 40},
			{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
			{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 100},
		},
		want: &score.Results{
			Scores: []*score.Score{
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: -1},
			},
		},
	},
	{
		name: "Test A with non-zero score recorded 5 times",
		desc: "We only allow the same test to be recorded two times",
		in: []*score.Score{
			{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 40},
			{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
			{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 100},
			{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 100},
			{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 100},
		},
		want: &score.Results{
			Scores: []*score.Score{
				{TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: -1},
			},
		},
	},
}

func TestAddScore(t *testing.T) {
	for _, test := range scoreTests {
		t.Run(test.name, func(t *testing.T) {
			scores := score.NewResults()
			for _, sc := range test.in {
				scores.AddScore(sc)
			}
			results := &score.Results{Scores: scores.ToScoreSlice()}
			if diff := cmp.Diff(test.want, results, cmpopts.IgnoreUnexported(score.Results{})); diff != "" {
				t.Errorf("\nDescription: %s\nScores are different (-want +got):\n%s", test.desc, diff)
			}
		})
	}
}
