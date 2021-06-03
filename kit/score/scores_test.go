package score_test

import (
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
