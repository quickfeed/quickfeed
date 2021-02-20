package score_test

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/autograde/quickfeed/kit/score"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var scoreTests = []struct {
	name string
	desc string
	in   []*score.Score
	want *score.Scores
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
		want: &score.Scores{
			TestNames: []string{"A", "B", "C"},
			ScoreMap: map[string]*score.Score{
				"A": {TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
				"B": {TestName: "B", Secret: theSecret, Weight: 20, MaxScore: 100, Score: 60},
				"C": {TestName: "C", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 70},
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
		want: &score.Scores{
			TestNames: []string{"A", "B", "C", "D"},
			ScoreMap: map[string]*score.Score{
				"A": {TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
				"B": {TestName: "B", Secret: theSecret, Weight: 20, MaxScore: 100, Score: 60},
				"C": {TestName: "C", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 70},
				"D": {TestName: "D", Secret: theSecret, Weight: 30, MaxScore: 100, Score: 0},
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
		want: &score.Scores{
			TestNames: []string{"A"},
			ScoreMap: map[string]*score.Score{
				"A": {TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: -1},
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
		want: &score.Scores{
			TestNames: []string{"A"},
			ScoreMap: map[string]*score.Score{
				"A": {TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: -1},
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
		want: &score.Scores{
			TestNames: []string{"A"},
			ScoreMap: map[string]*score.Score{
				"A": {TestName: "A", Secret: theSecret, Weight: 10, MaxScore: 100, Score: -1},
			},
		},
	},
}

func TestAddScore(t *testing.T) {
	for _, test := range scoreTests {
		t.Run(test.name, func(t *testing.T) {
			scores := score.NewScores()
			for _, sc := range test.in {
				scores.AddScore(sc)
			}
			if diff := cmp.Diff(test.want, scores, cmpopts.IgnoreUnexported(score.Scores{})); diff != "" {
				t.Errorf("\nDescription: %s\nScores are different (-want +got):\n%s", test.desc, diff)
			}
		})
	}
}

func extract(out, secret string) (string, *score.Scores, error) {
	var filteredLog []string
	scores := score.NewScores()
	for _, line := range strings.Split(out, "\n") {
		// check if line has expected JSON score string
		if score.HasPrefix(line) {
			sc, err := score.Parse(line, secret)
			if err != nil {
				return "", nil, err
			}
			scores.AddScore(sc)
		} else if line != "" { // include only non-empty lines
			// the filtered log without JSON score strings
			filteredLog = append(filteredLog, line)
		}
	}
	return strings.Join(filteredLog, "\n"), scores, nil
}

func TestExtractResultWithTwoScoreLines(t *testing.T) {
	out := `here is some output in the log with whitespace before the JSON string below.

    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":0,"MaxScore":100,"Weight":1}

Here are some more logs for the student.
    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":100,"MaxScore":100,"Weight":1}

	{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"JoGo","Score":0,"MaxScore":100,"Weight":1}

Here are some more logs for the student.
`

	_, scores, err := extract(out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73")
	if err != nil {
		t.Fatal(err)
	}
	if len(scores.ScoreMap) != 2 {
		t.Errorf("Extract(out, secret) = %v, expected 2 entries", scores)
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

	_, scores, err := extract(out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73")
	if err != nil {
		t.Fatal(err)
	}
	const expectedTests = 6
	if len(scores.ScoreMap) != expectedTests {
		t.Fatalf("Extract() expected %d Score entries, got %d: %+v", expectedTests, len(scores.ScoreMap), scores)
	}
	if len(scores.TestNames) != expectedTests {
		t.Fatalf("Extract() expected %d Test entries, got %d: %+v", expectedTests, len(scores.TestNames), scores)
	}

	testOrder := []string{
		"GoodTest1",
		"GoodTest2",
		"PanicedTest1",
		"PanicedTest2",
		"PanicedTest3",
		"MaliciousTest",
	}
	for i, testName := range scores.TestNames {
		if testName != testOrder[i] {
			t.Errorf("Extract() returned unexpected order of tests: expected %s, got %s", testOrder[i], testName)
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
	s := score.NewScores()
	const hiddenSecret = "hidden"
	for _, sc := range scores {
		// The scoreObjects was extracted when we allowed Weight=0
		// We now return an error for when Weight=0.
		// Hence, we only add scores with non-zero weights.
		if err := sc.IsValid(hiddenSecret); err == nil {
			s.AddScore(sc)
		}
	}
	err := s.Validate()
	if err != nil {
		t.Errorf("Validate() = %v, expected <nil>", err)
	}
	got := s.Sum()
	const want = 100
	if got != want {
		t.Errorf("Sum() = '%d', want '%d'", got, want)
	}
}
