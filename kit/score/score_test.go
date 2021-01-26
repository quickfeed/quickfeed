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
		Weight:   10,
	}
	rs := sc.RelativeScore()
	expectedRS := t.Name() + ": score = 3/10 = 0.3"
	if rs != expectedRS {
		t.Errorf("RelativeScore()='%s', expected '%s'", rs, expectedRS)
	}
}
