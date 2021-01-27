package score_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/autograde/quickfeed/kit/score"
)

type scoreData struct {
	points, max, weight int
}

var scores = []struct {
	in    []*scoreData
	out   uint32
	grade string
}{
	{[]*scoreData{setScore(10, 10, 1), setScore(5, 5, 1), setScore(15, 15, 1)}, 100, "A"},
	{[]*scoreData{setScore(5, 10, 1), setScore(5, 5, 1), setScore(20, 40, 1)}, 66, "C"},
	{[]*scoreData{setScore(5, 10, 1), setScore(5, 10, 1), setScore(20, 40, 1)}, 50, "D"},
	{[]*scoreData{setScore(10, 10, 2), setScore(5, 10, 1), setScore(20, 40, 1)}, 75, "C"},
	{[]*scoreData{setScore(0, 10, 2), setScore(0, 10, 1), setScore(0, 40, 1)}, 0, "F"},
	{[]*scoreData{}, 0, "F"},
}

func setScore(points, max, w int) *scoreData {
	return &scoreData{points: points, max: max, weight: w}
}

func subName(points, max, weight, i, j int) string {
	return fmt.Sprintf("P%02d/M%02d/W%d/%d/%d", points, max, weight, i, j)
}

func TestSum(t *testing.T) {
	for i, s := range scores {
		// Clear other scores before using Sum() again.
		score.Clear()
		for j, sd := range s.in {
			// AddSub is normally called from init(), but here we are testing Sum().
			score.AddSub(TestSum, subName(sd.points, sd.max, sd.weight, i, j), sd.max, sd.weight)
			t.Run(subName(sd.points, sd.max, sd.weight, i, j), func(t *testing.T) {
				sc := score.MinByName(t.Name())
				sc.IncBy(sd.points)
			})
		}
		tot := score.Sum()
		if tot != s.out {
			t.Errorf("Sum() = %d, expected %d", tot, s.out)
		}
	}
}

func TestSumGrade(t *testing.T) {
	g := score.GradingScheme{
		Name:        "C Bias (UiS Scheme)",
		GradePoints: []uint32{90, 80, 60, 50, 40, 0},
		GradeNames:  []string{"A", "B", "C", "D", "E", "F"},
	}

	for i, s := range scores {
		// Clear other scores before using Sum() again.
		score.Clear()
		for j, sd := range s.in {
			// AddSub is normally called from init(), but here we are testing Sum() and Grade().
			score.AddSub(TestSumGrade, subName(sd.points, sd.max, sd.weight, i, j), sd.max, sd.weight)
			t.Run(subName(sd.points, sd.max, sd.weight, i, j), func(t *testing.T) {
				sc := score.MinByName(t.Name())
				sc.IncBy(sd.points)
			})
		}
		tot := score.Sum()
		grade := g.Grade(tot)
		if grade != s.grade {
			t.Errorf("Grade(%d) = %s, expected %s", tot, grade, s.grade)
		}
	}
}

var (
	// copied from string constants in score package
	errScoreInterval = errors.New("Score must be in the interval [0, MaxScore]")
	errMaxScore      = errors.New("MaxScore must be greater than 0")
	errWeight        = errors.New("Weight must be greater than 0")
)

var validateScores = []struct {
	in  []*scoreData
	err error
}{
	{[]*scoreData{setScore(10, 10, 1), setScore(5, 5, 1), setScore(15, 15, 1)}, nil},
	{[]*scoreData{setScore(5, 10, 1), setScore(5, 5, 1), setScore(20, 40, 1)}, nil},
	{[]*scoreData{setScore(5, 10, 1), setScore(5, 10, 1), setScore(20, 40, 1)}, nil},
	{[]*scoreData{setScore(10, 10, 2), setScore(5, 10, 1), setScore(20, 40, 1)}, nil},
	{[]*scoreData{setScore(0, 10, 2), setScore(0, 10, 1), setScore(0, 40, 1)}, nil},
	{[]*scoreData{}, nil},
	{[]*scoreData{setScore(-10, 10, 1), setScore(5, 5, 1), setScore(15, 15, 1)}, errScoreInterval},
	{[]*scoreData{setScore(10, 10, 1), setScore(5, 5, 1), setScore(-1, 15, 1)}, errScoreInterval},
	{[]*scoreData{setScore(0, 10, 1)}, nil},                             // score = 0
	{[]*scoreData{setScore(10, 10, 1)}, nil},                            // score = maxScore
	{[]*scoreData{setScore(10, 1000, 10)}, nil},                         // large maxScore
	{[]*scoreData{setScore(10, 10, 1000)}, nil},                         // large weight
	{[]*scoreData{setScore(10, 1, 1)}, errScoreInterval},                // score > maxScore
	{[]*scoreData{setScore(-1, 1, 1)}, errScoreInterval},                // score < 0
	{[]*scoreData{setScore(10, 0, 1)}, errMaxScore},                     // maxScore = 0 (would normally panic during Add)
	{[]*scoreData{setScore(10, -1, 1)}, errMaxScore},                    // maxScore < 0 (would normally panic during Add)
	{[]*scoreData{setScore(10, 1, 0)}, errWeight},                       // weight = 0 (would normally panic during Add)
	{[]*scoreData{setScore(10, 10, 1), setScore(10, 1, -1)}, errWeight}, // weight < 0 (would normally panic during Add)
}

func TestValidate(t *testing.T) {
	for i, s := range validateScores {
		// Clear other scores before using Validate() again.
		score.Clear()
		for j, sd := range s.in {
			// AddSub is normally called from init(), but here we are testing Validate().
			// Hack to avoid panic due to invalid max and weight initialization.
			score.AddSub(TestValidate, subName(sd.points, sd.max, sd.weight, i, j), 1, 1)
		}
		for j, sd := range s.in {
			t.Run(subName(sd.points, sd.max, sd.weight, i, j), func(t *testing.T) {
				sc := score.MinByName(t.Name())
				sc.Score = int32(sd.points)
				sc.MaxScore = int32(sd.max)
				sc.Weight = int32(sd.weight)
			})
		}
		err := score.Validate()
		if err != s.err {
			var e, se string
			if err != nil {
				e = err.Error()
			}
			if s.err != nil {
				se = s.err.Error()
			}
			if !(len(se) > 0 && strings.Contains(e, se)) {
				t.Errorf("Validate() = %q, expected %v", err, s.err)
			}
		}
	}
}
