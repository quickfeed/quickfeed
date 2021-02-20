package score_test

import (
	"strings"
	"testing"

	"github.com/autograde/quickfeed/kit/score"
)

const theSecret = "my secret code"

var nonJSONLog = []string{
	"here is some output",
	"some other output",
	"line contains " + theSecret,
	theSecret + " should not be revealed",
}

func TestParseNonJSONStrings(t *testing.T) {
	for _, s := range nonJSONLog {
		sc, err := score.Parse(s, theSecret)
		if err == nil {
			t.Errorf("Expected '%v', got '<nil>'", score.ErrScoreNotFound.Error())
		}
		if sc != nil {
			t.Errorf("Got unexpected score object '%v', wanted '<nil>'", sc)
		}
	}
}

var jsonLog = []struct {
	in          string
	max, weight int
	err         error
}{
	{
		`{"Secret":"` + theSecret + `","TestName":"TestParseJSONStrings","Score":0,"MaxScore":10,"Weight":10}`,
		10, 10,
		nil,
	},
	{
		`{"Secret":"the wrong secret","TestName":"TestParseJSONStrings","Score":0,"MaxScore":10,"Weight":10}`,
		-1, -1,
		score.ErrScoreNotFound,
	},
}

func TestParseJSONStrings(t *testing.T) {
	for _, s := range jsonLog {
		sc, err := score.Parse(s.in, theSecret)
		var expectedScore *score.Score
		if s.max > 0 {
			expectedScore = &score.Score{
				TestName: t.Name(),
				MaxScore: int32(s.max),
				Weight:   int32(s.weight),
			}
		}
		if sc != expectedScore || err != s.err {
			if !expectedScore.Equal(sc) || err != s.err {
				t.Errorf("Failed to parse:\n%v\nGot: '%v', '%v'\nExp: '%v', '%v'",
					s.in, sc, err, expectedScore, s.err)
			}
			if sc != nil && sc.Secret == theSecret {
				t.Errorf("Parse function failed to hide global secret: %v", sc.Secret)
			}
		}
	}
}

var scoreValidTests = []struct {
	name string
	in   []*score.Score
	want string
}{
	{
		name: "EmptyTestName",
		in: []*score.Score{
			{TestName: "", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 0},
		},
		want: score.ErrEmptyTestName,
	},
	{
		name: "BadWeights",
		in: []*score.Score{
			{TestName: "BadWeights", Secret: theSecret, Weight: 0, MaxScore: 100, Score: 0},
			{TestName: "BadWeights", Secret: theSecret, Weight: -10, MaxScore: 100, Score: 0},
			{TestName: "BadWeights", Secret: theSecret, Weight: -1, MaxScore: 100, Score: 0},
		},
		want: score.ErrWeight,
	},
	{
		name: "BadMaxScore",
		in: []*score.Score{
			{TestName: "BadMaxScore", Secret: theSecret, Weight: 10, MaxScore: 0, Score: 0},
			{TestName: "BadMaxScore", Secret: theSecret, Weight: 10, MaxScore: -100, Score: 0},
			{TestName: "BadMaxScore", Secret: theSecret, Weight: 10, MaxScore: -1, Score: 0},
		},
		want: score.ErrMaxScore,
	},
	{
		name: "BadScore",
		in: []*score.Score{
			{TestName: "BadScore", Secret: theSecret, Weight: 10, MaxScore: 100, Score: -1},
			{TestName: "BadScore", Secret: theSecret, Weight: 10, MaxScore: 100, Score: -20},
			{TestName: "BadScore", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 101},
			{TestName: "BadScore", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 1000},
		},
		want: score.ErrScoreInterval,
	},
	{
		name: "BadSecret",
		in: []*score.Score{
			{TestName: "BadSecret", Secret: "xyz", Weight: 10, MaxScore: 100, Score: 0},
		},
		want: score.ErrSecret,
	},
	{
		name: "GoodScore",
		in: []*score.Score{
			{TestName: "GoodScoreW", Secret: theSecret, Weight: 1, MaxScore: 100, Score: 0},
			{TestName: "GoodScoreW", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 0},
			{TestName: "GoodScoreW", Secret: theSecret, Weight: 100, MaxScore: 100, Score: 0},
			{TestName: "GoodScoreM", Secret: theSecret, Weight: 10, MaxScore: 1, Score: 0},
			{TestName: "GoodScoreM", Secret: theSecret, Weight: 10, MaxScore: 10, Score: 0},
			{TestName: "GoodScoreM", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 0},
			{TestName: "GoodScoreS", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 10},
			{TestName: "GoodScoreS", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 50},
			{TestName: "GoodScoreS", Secret: theSecret, Weight: 10, MaxScore: 100, Score: 100},
		},
		want: "nil",
	},
}

func TestScoreIsValid(t *testing.T) {
	for _, test := range scoreValidTests {
		t.Run(test.name, func(t *testing.T) {
			for _, sc := range test.in {
				err := sc.IsValid(theSecret)
				if err != nil {
					if !strings.Contains(err.Error(), test.want) {
						t.Errorf("IsValid(%q) = %v, expected = %v", sc, err, test.want)
					}
				} else if test.want != "nil" {
					t.Errorf("IsValid(%q) = %v, expected = %v", sc, err, test.want)
				}
			}
		})
	}
}
