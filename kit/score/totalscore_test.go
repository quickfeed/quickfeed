package score

import "testing"

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

func TestTotal(t *testing.T) {
	for _, s := range scores {
		allScores := make([]*Score, 0)
		for _, sd := range s.in {
			sc := NewScore(t, sd.max, sd.weight)
			sc.IncBy(sd.points)
			allScores = append(allScores, sc)
		}
		tot := Total(allScores)
		if tot != s.out {
			t.Errorf("Got: %d, Want: %d", tot, s.out)
		}
	}
}
