package score_test

import (
	"fmt"
	"testing"

	"github.com/autograde/quickfeed/kit/score"
)

type scoreData struct {
	points, max, weight int
}

var scores = []struct {
	in  []*scoreData
	out uint32
}{
	{[]*scoreData{setScore(10, 10, 1), setScore(5, 5, 1), setScore(15, 15, 1)}, 100},
	{[]*scoreData{setScore(5, 10, 1), setScore(5, 5, 1), setScore(20, 40, 1)}, 66},
	{[]*scoreData{setScore(5, 10, 1), setScore(5, 10, 1), setScore(20, 40, 1)}, 50},
	{[]*scoreData{setScore(10, 10, 2), setScore(5, 10, 1), setScore(20, 40, 1)}, 75},
	{[]*scoreData{setScore(0, 10, 2), setScore(0, 10, 1), setScore(0, 40, 1)}, 0},
	{[]*scoreData{}, 0},
}

func setScore(points, max, w int) *scoreData {
	return &scoreData{points: points, max: max, weight: w}
}

func subName(max, weight, i, j int) string {
	return fmt.Sprintf("Max%d/W%d/%d/%d", max, weight, i, j)
}

func TestSum(t *testing.T) {
	for i, s := range scores {
		// Clear the other scores before using Sum() again.
		score.Clear()
		for j, sd := range s.in {
			// AddSub is normally called from init(), but for testing Sum() this was difficult.
			score.AddSub(TestSum, subName(sd.max, sd.weight, i, j), sd.max, sd.weight)
			t.Run(subName(sd.max, sd.weight, i, j), func(t *testing.T) {
				sc := score.MinByName(t.Name())
				sc.IncBy(sd.points)
			})
		}
		tot := score.Sum()
		if tot != s.out {
			t.Errorf("Got: %d, Want: %d", tot, s.out)
		}
	}
}
