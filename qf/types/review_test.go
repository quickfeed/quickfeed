package types_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/qf/types"
)

var reviewScoreTests = []struct {
	name   string
	score  uint32
	review *types.Review
}{
	{
		"Test 25% score",
		25,
		&types.Review{
			ID: 1,
			GradingBenchmarks: []*types.GradingBenchmark{
				{
					Criteria: []*types.GradingCriterion{
						{
							Grade: types.GradingCriterion_FAILED,
						},
						{
							Grade: types.GradingCriterion_PASSED,
						},
					},
				},
				{
					Criteria: []*types.GradingCriterion{
						{
							Grade: types.GradingCriterion_FAILED,
						},
						{
							Grade: types.GradingCriterion_FAILED,
						},
					},
				},
			},
		},
	},
	{
		"Test 75% score",
		75,
		&types.Review{
			ID: 2,
			GradingBenchmarks: []*types.GradingBenchmark{
				{
					Criteria: []*types.GradingCriterion{
						{
							Grade: types.GradingCriterion_PASSED,
						},
						{
							Grade: types.GradingCriterion_PASSED,
						},
					},
				},
				{
					Criteria: []*types.GradingCriterion{
						{
							Grade: types.GradingCriterion_FAILED,
						},
						{
							Grade: types.GradingCriterion_PASSED,
						},
					},
				},
			},
		},
	},
	{
		"Test 6/10 score",
		6,
		&types.Review{
			ID:       3,
			Feedback: "Test 6/10 score",
			GradingBenchmarks: []*types.GradingBenchmark{
				{
					Criteria: []*types.GradingCriterion{
						{
							Points: 3,
							Grade:  types.GradingCriterion_PASSED,
						},
						{
							Points: 2,
							Grade:  types.GradingCriterion_FAILED,
						},
					},
				},
				{
					Criteria: []*types.GradingCriterion{
						{
							Points: 2,
							Grade:  types.GradingCriterion_FAILED,
						},
						{
							Points: 3,
							Grade:  types.GradingCriterion_PASSED,
						},
					},
				},
			},
		},
	},
}

func TestComputeScore(t *testing.T) {
	for _, reviewTest := range reviewScoreTests {
		reviewTest.review.ComputeScore()
		if reviewTest.review.Score != reviewTest.score {
			t.Fatalf("Computed wrong review score: expected %d, got %d", reviewTest.score, reviewTest.review.Score)
		}
	}
}
