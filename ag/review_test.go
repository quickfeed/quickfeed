package ag_test

import (
	"testing"

	pb "github.com/autograde/quickfeed/ag"
)

var reviewScoreTests = []struct {
	name   string
	score  uint32
	review *pb.Review
}{
	{
		"Test 25% score",
		25,
		&pb.Review{
			ID: 1,
			GradingBenchmarks: []*pb.GradingBenchmark{
				{
					Criteria: []*pb.GradingCriterion{
						{
							Grade: pb.GradingCriterion_FAILED,
						},
						{
							Grade: pb.GradingCriterion_PASSED,
						},
					},
				},
				{
					Criteria: []*pb.GradingCriterion{
						{
							Grade: pb.GradingCriterion_FAILED,
						},
						{
							Grade: pb.GradingCriterion_FAILED,
						},
					},
				},
			},
		},
	},
	{
		"Test 75% score",
		75,
		&pb.Review{
			ID: 2,
			GradingBenchmarks: []*pb.GradingBenchmark{
				{
					Criteria: []*pb.GradingCriterion{
						{
							Grade: pb.GradingCriterion_PASSED,
						},
						{
							Grade: pb.GradingCriterion_PASSED,
						},
					},
				},
				{
					Criteria: []*pb.GradingCriterion{
						{
							Grade: pb.GradingCriterion_FAILED,
						},
						{
							Grade: pb.GradingCriterion_PASSED,
						},
					},
				},
			},
		},
	},
	{
		"Test 6/10 score",
		6,
		&pb.Review{
			ID:       3,
			Feedback: "Test 6/10 score",
			GradingBenchmarks: []*pb.GradingBenchmark{
				{
					Criteria: []*pb.GradingCriterion{
						{
							Points: 3,
							Grade:  pb.GradingCriterion_PASSED,
						},
						{
							Points: 2,
							Grade:  pb.GradingCriterion_FAILED,
						},
					},
				},
				{
					Criteria: []*pb.GradingCriterion{
						{
							Points: 2,
							Grade:  pb.GradingCriterion_FAILED,
						},
						{
							Points: 3,
							Grade:  pb.GradingCriterion_PASSED,
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
