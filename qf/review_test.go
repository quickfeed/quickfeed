package qf_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/qf"
)

var reviewScoreTests = []struct {
	name   string
	score  uint32
	review *qf.Review
}{
	{
		"Test 25% score",
		25,
		&qf.Review{
			ID: 1,
			GradingBenchmarks: []*qf.GradingBenchmark{
				{
					Criteria: []*qf.GradingCriterion{
						{
							Grade: qf.GradingCriterion_FAILED,
						},
						{
							Grade: qf.GradingCriterion_PASSED,
						},
					},
				},
				{
					Criteria: []*qf.GradingCriterion{
						{
							Grade: qf.GradingCriterion_FAILED,
						},
						{
							Grade: qf.GradingCriterion_FAILED,
						},
					},
				},
			},
		},
	},
	{
		"Test 75% score",
		75,
		&qf.Review{
			ID: 2,
			GradingBenchmarks: []*qf.GradingBenchmark{
				{
					Criteria: []*qf.GradingCriterion{
						{
							Grade: qf.GradingCriterion_PASSED,
						},
						{
							Grade: qf.GradingCriterion_PASSED,
						},
					},
				},
				{
					Criteria: []*qf.GradingCriterion{
						{
							Grade: qf.GradingCriterion_FAILED,
						},
						{
							Grade: qf.GradingCriterion_PASSED,
						},
					},
				},
			},
		},
	},
	{
		"Test 6/10 score",
		6,
		&qf.Review{
			ID:       3,
			Feedback: "Test 6/10 score",
			GradingBenchmarks: []*qf.GradingBenchmark{
				{
					Criteria: []*qf.GradingCriterion{
						{
							Points: 3,
							Grade:  qf.GradingCriterion_PASSED,
						},
						{
							Points: 2,
							Grade:  qf.GradingCriterion_FAILED,
						},
					},
				},
				{
					Criteria: []*qf.GradingCriterion{
						{
							Points: 2,
							Grade:  qf.GradingCriterion_FAILED,
						},
						{
							Points: 3,
							Grade:  qf.GradingCriterion_PASSED,
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
