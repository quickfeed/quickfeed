package database

import (
	"encoding/json"
	"fmt"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/kit/score"
)

// This file is mean to temporarily transform BuildInfo and ScoreObjects strings to new score.Results type
// so that we can continue to use the old database for testing.
// TODO(meling) Remove this file and its uses when migration to new format is done.

func transform(submissions ...*pb.Submission) {
	for _, sub := range submissions {
		buildInfo := unmarshalBuildInfo(sub)
		scores := unmarshalScores(sub)
		if buildInfo != nil {
			sub.BuildInfo = buildInfo
			sub.OldBuildInfo = ""
		}
		if scores != nil {
			sub.Scores = scores
			sub.ScoreObjects = ""
		}
	}
}

func unmarshalBuildInfo(submission *pb.Submission) *score.BuildInfo {
	if submission.OldBuildInfo == "" {
		return nil
	}
	var buildInfo score.BuildInfo
	if err := json.Unmarshal([]byte(submission.OldBuildInfo), &buildInfo); err != nil {
		fmt.Printf("Failed to unmarshal JSON BuildInfo string: (%v): %v\n", submission.OldBuildInfo, err)
	}
	return &buildInfo
}

func unmarshalScores(submission *pb.Submission) []*score.Score {
	if submission.ScoreObjects == "" {
		return nil
	}
	var scores []*score.Score
	if err := json.Unmarshal([]byte(submission.ScoreObjects), &scores); err != nil {
		fmt.Printf("Failed to unmarshal JSON ScoreObjects string: (%v): %v\n", submission.ScoreObjects, err)
	}
	return scores
}
