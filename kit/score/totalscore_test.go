package score_test

import (
	"strings"
	"testing"

	"github.com/autograde/quickfeed/kit/score"
)

var scoreGrades = []struct {
	in        []*score.Score
	out       uint32
	wantGrade string
}{
	{
		in: []*score.Score{
			{TestName: "A", Score: 10, MaxScore: 10, Weight: 1},
			{TestName: "B", Score: 05, MaxScore: 05, Weight: 1},
			{TestName: "C", Score: 15, MaxScore: 15, Weight: 1},
		},
		out:       100,
		wantGrade: "A",
	},
	{
		in: []*score.Score{
			{TestName: "A", Score: 05, MaxScore: 10, Weight: 1},
			{TestName: "B", Score: 05, MaxScore: 05, Weight: 1},
			{TestName: "C", Score: 20, MaxScore: 40, Weight: 1},
		},
		out:       67,
		wantGrade: "C",
	},
	{
		in: []*score.Score{
			{TestName: "A", Score: 05, MaxScore: 10, Weight: 1},
			{TestName: "B", Score: 05, MaxScore: 10, Weight: 1},
			{TestName: "C", Score: 20, MaxScore: 40, Weight: 1},
		},
		out:       50,
		wantGrade: "D",
	},
	{
		in: []*score.Score{
			{TestName: "A", Score: 10, MaxScore: 10, Weight: 2},
			{TestName: "B", Score: 05, MaxScore: 10, Weight: 1},
			{TestName: "C", Score: 20, MaxScore: 40, Weight: 1},
		},
		out:       75,
		wantGrade: "C",
	},
	{
		in: []*score.Score{
			{TestName: "A", Score: 00, MaxScore: 10, Weight: 2},
			{TestName: "B", Score: 00, MaxScore: 10, Weight: 1},
			{TestName: "C", Score: 00, MaxScore: 40, Weight: 1},
		},
		out:       0,
		wantGrade: "F",
	},
}

func TestSumGrade(t *testing.T) {
	g := score.GradingScheme{
		Name:        "C Bias (UiS Scheme)",
		GradePoints: []uint32{90, 80, 60, 50, 40, 0},
		GradeNames:  []string{"A", "B", "C", "D", "E", "F"},
	}

	for _, s := range scoreGrades {
		scoreTable := score.NewResults()
		for _, sc := range s.in {
			scoreTable.AddScore(sc)
		}
		results := &score.Results{Scores: scoreTable.ToScoreSlice()}
		if err := results.Validate(""); err != nil {
			t.Error(err)
		}
		tot := results.Sum()
		grade := g.Grade(tot)
		if grade != s.wantGrade {
			t.Errorf("Grade(%d) = %s, expected %s", tot, grade, s.wantGrade)
		}
		if tot != s.out {
			t.Errorf("Sum() = %d, expected %d", tot, s.out)
		}
	}
}

var valScores = []struct {
	desc string
	in   []*score.Score
	err  error
}{
	{
		in: []*score.Score{
			{TestName: "A", Score: 10, MaxScore: 10, Weight: 1},
			{TestName: "B", Score: 05, MaxScore: 05, Weight: 1},
			{TestName: "C", Score: 15, MaxScore: 15, Weight: 1},
		},
		err: nil,
	},
	{
		in: []*score.Score{
			{TestName: "A", Score: 05, MaxScore: 10, Weight: 1},
			{TestName: "B", Score: 05, MaxScore: 05, Weight: 1},
			{TestName: "C", Score: 20, MaxScore: 40, Weight: 1},
		},
		err: nil,
	},
	{
		in: []*score.Score{
			{TestName: "A", Score: 05, MaxScore: 10, Weight: 1},
			{TestName: "B", Score: 05, MaxScore: 10, Weight: 1},
			{TestName: "C", Score: 20, MaxScore: 40, Weight: 1},
		},
		err: nil,
	},
	{
		in: []*score.Score{
			{TestName: "A", Score: 10, MaxScore: 10, Weight: 2},
			{TestName: "B", Score: 05, MaxScore: 10, Weight: 1},
			{TestName: "C", Score: 20, MaxScore: 40, Weight: 1},
		},
		err: nil,
	},
	{
		in: []*score.Score{
			{TestName: "A", Score: 00, MaxScore: 10, Weight: 2},
			{TestName: "B", Score: 00, MaxScore: 10, Weight: 1},
			{TestName: "C", Score: 00, MaxScore: 40, Weight: 1},
		},
		err: nil,
	},
	{
		in:  nil,
		err: nil,
	},
	{
		in:  []*score.Score{},
		err: nil,
	},
	{
		in: []*score.Score{
			{TestName: "A", Score: -10, MaxScore: 10, Weight: 1},
			{TestName: "B", Score: 005, MaxScore: 05, Weight: 1},
			{TestName: "C", Score: 015, MaxScore: 15, Weight: 1},
		},
		err: score.ErrScoreInterval,
	},
	{
		in: []*score.Score{
			{TestName: "A", Score: 10, MaxScore: 10, Weight: 1},
			{TestName: "B", Score: 05, MaxScore: 05, Weight: 1},
			{TestName: "C", Score: -1, MaxScore: 15, Weight: 1},
		},
		err: score.ErrScoreInterval,
	},
	{
		desc: "score = 0",
		in: []*score.Score{
			{TestName: "A", Score: 00, MaxScore: 10, Weight: 1},
		},
		err: nil,
	},
	{
		desc: "score = maxScore",
		in: []*score.Score{
			{TestName: "A", Score: 10, MaxScore: 10, Weight: 1},
		},
		err: nil,
	},
	{
		desc: "large maxScore",
		in: []*score.Score{
			{TestName: "A", Score: 10, MaxScore: 1000, Weight: 10},
		},
		err: nil,
	},
	{
		desc: "large weight",
		in: []*score.Score{
			{TestName: "A", Score: 10, MaxScore: 10, Weight: 1000},
		},
		err: nil,
	},
	{
		desc: "score > maxScore",
		in: []*score.Score{
			{TestName: "A", Score: 10, MaxScore: 1, Weight: 1},
		},
		err: score.ErrScoreInterval,
	},
	{
		desc: "score < 0",
		in: []*score.Score{
			{TestName: "A", Score: -1, MaxScore: 1, Weight: 1},
		},
		err: score.ErrScoreInterval,
	},
	{
		desc: "maxScore = 0 (would normally panic during Add)",
		in: []*score.Score{
			{TestName: "A", Score: 10, MaxScore: 0, Weight: 1},
		},
		err: score.ErrMaxScore,
	},
	{
		desc: "maxScore < 0 (would normally panic during Add)",
		in: []*score.Score{
			{TestName: "A", Score: 10, MaxScore: -1, Weight: 1},
		},
		err: score.ErrMaxScore,
	},
	{
		desc: "weight = 0 (would normally panic during Add)",
		in: []*score.Score{
			{TestName: "A", Score: 10, MaxScore: 1, Weight: 0},
		},
		err: score.ErrWeight,
	},
	{
		desc: "weight < 0 (would normally panic during Add)",
		in: []*score.Score{
			{TestName: "A", Score: 10, MaxScore: 1, Weight: -1},
		},
		err: score.ErrWeight,
	},
}

func TestValidate(t *testing.T) {
	for _, s := range valScores {
		scoreTable := score.NewResults()
		for _, sc := range s.in {
			scoreTable.AddScore(sc)
		}
		results := &score.Results{Scores: scoreTable.ToScoreSlice()}
		if err := results.Validate(""); err != s.err {
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
