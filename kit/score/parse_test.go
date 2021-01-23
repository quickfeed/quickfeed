package score_test

import (
	"testing"

	"github.com/autograde/quickfeed/kit/score"
)

var theSecret = "my secret code"

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
			expectedScore = score.NewScore(t, s.max, s.weight)
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
