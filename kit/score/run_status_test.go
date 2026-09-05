package score_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/kit/score"
)

func TestRunStatusFailed(t *testing.T) {
	tests := []struct {
		name        string
		status      score.RunStatus
		failed      bool
		scoresValid bool
	}{
		{name: "Success", status: score.RunStatus_SUCCESS, scoresValid: true},
		{name: "BuildFailure", status: score.RunStatus_BUILD_FAILURE, failed: true, scoresValid: true},
		{name: "Timeout", status: score.RunStatus_TIMEOUT, failed: true},
		{name: "NoScores", status: score.RunStatus_NO_SCORES, failed: true},
		{name: "TestPanic", status: score.RunStatus_TEST_PANIC, failed: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.status.Failed(); got != test.failed {
				t.Errorf("RunStatus.Failed() = %t, want %t", got, test.failed)
			}
			if got := test.status.ScoresValid(); got != test.scoresValid {
				t.Errorf("RunStatus.ScoresValid() = %t, want %t", got, test.scoresValid)
			}
			buildInfo := &score.BuildInfo{Status: test.status}
			if got := buildInfo.Failed(); got != test.failed {
				t.Errorf("BuildInfo.Failed() = %t, want %t", got, test.failed)
			}
			if got := buildInfo.ScoresValid(); got != test.scoresValid {
				t.Errorf("BuildInfo.ScoresValid() = %t, want %t", got, test.scoresValid)
			}
			results := &score.Results{BuildInfo: buildInfo}
			if got := results.Failed(); got != test.failed {
				t.Errorf("Results.Failed() = %t, want %t", got, test.failed)
			}
			if got := results.ScoresValid(); got != test.scoresValid {
				t.Errorf("Results.ScoresValid() = %t, want %t", got, test.scoresValid)
			}
		})
	}
}
