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
		if reviewTest.review.GetScore() != reviewTest.score {
			t.Fatalf("Computed wrong review score: expected %d, got %d", reviewTest.score, reviewTest.review.GetScore())
		}
	}
}

func TestReviewPercentScore(t *testing.T) {
	criterion := func(points uint64, grade qf.GradingCriterion_Grade) *qf.GradingCriterion {
		return &qf.GradingCriterion{Points: points, Grade: grade}
	}
	review := func(criteria ...*qf.GradingCriterion) *qf.Review {
		return &qf.Review{GradingBenchmarks: []*qf.GradingBenchmark{{Criteria: criteria}}}
	}
	tests := []struct {
		name   string
		review *qf.Review
		want   uint32
	}{
		{name: "NoCriteria", review: &qf.Review{}, want: 0},
		{name: "NoPoints", review: review(criterion(0, qf.GradingCriterion_PASSED), criterion(0, qf.GradingCriterion_FAILED), criterion(0, qf.GradingCriterion_PASSED), criterion(0, qf.GradingCriterion_NONE)), want: 50},
		{name: "PointsNotSummingTo100", review: review(criterion(3, qf.GradingCriterion_PASSED), criterion(2, qf.GradingCriterion_FAILED)), want: 60},
		{name: "PointsAllFailed", review: review(criterion(3, qf.GradingCriterion_FAILED), criterion(2, qf.GradingCriterion_FAILED)), want: 0},
		{name: "PointsAllPassed", review: review(criterion(30, qf.GradingCriterion_PASSED), criterion(20, qf.GradingCriterion_PASSED)), want: 100},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.review.PercentScore(); got != tc.want {
				t.Errorf("PercentScore() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestReviewComplete(t *testing.T) {
	review := func(grades ...qf.GradingCriterion_Grade) *qf.Review {
		criteria := make([]*qf.GradingCriterion, len(grades))
		for i, grade := range grades {
			criteria[i] = &qf.GradingCriterion{Grade: grade}
		}
		return &qf.Review{GradingBenchmarks: []*qf.GradingBenchmark{{Criteria: criteria}}}
	}
	tests := []struct {
		name   string
		review *qf.Review
		want   bool
	}{
		{name: "NoCriteria", review: &qf.Review{}, want: true},
		{name: "AllGraded", review: review(qf.GradingCriterion_PASSED, qf.GradingCriterion_FAILED), want: true},
		{name: "OneUngraded", review: review(qf.GradingCriterion_PASSED, qf.GradingCriterion_NONE), want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.review.Complete(); got != tc.want {
				t.Errorf("Complete() = %t, want %t", got, tc.want)
			}
		})
	}
}
