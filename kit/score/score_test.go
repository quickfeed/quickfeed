package score_test

import (
	"testing"

	"github.com/autograde/quickfeed/kit/score"
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
