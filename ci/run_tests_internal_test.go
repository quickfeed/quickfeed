package ci

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/quickfeed/quickfeed/kit/score"
)

func TestRedactOutput(t *testing.T) {
	const secret = "quickfeed-session-secret"
	got := redactOutput("failure: "+secret+" repeated "+secret, secret)
	if want := "failure: [REDACTED] repeated [REDACTED]"; got != want {
		t.Errorf("redactOutput() = %q, want %q", got, want)
	}
}

func TestClassifyRun(t *testing.T) {
	wrappedTimeout := fmt.Errorf("waiting for container: %w", context.DeadlineExceeded)
	tests := []struct {
		name         string
		runErr       error
		out          string
		parsedScores int
		want         score.RunStatus
	}{
		{name: "Success", runErr: nil, out: "some output", parsedScores: 3, want: score.RunStatus_SUCCESS},
		{name: "SuccessNonZeroExit", runErr: &ContainerExitError{Code: 1}, out: "go test failed", parsedScores: 3, want: score.RunStatus_SUCCESS},
		{name: "Timeout", runErr: context.DeadlineExceeded, out: "Container timeout.", parsedScores: 0, want: score.RunStatus_TIMEOUT},
		{name: "TimeoutWrapped", runErr: wrappedTimeout, out: "Container timeout.", parsedScores: 5, want: score.RunStatus_TIMEOUT},
		{name: "BuildFailure", runErr: errors.New("failed to create container"), out: "", parsedScores: 0, want: score.RunStatus_BUILD_FAILURE},
		{name: "BuildFailureNonZeroExit", runErr: &ContainerExitError{Code: 2}, out: "", parsedScores: 0, want: score.RunStatus_BUILD_FAILURE},
		{name: "NoScores", runErr: nil, out: "compile error: missing file", parsedScores: 0, want: score.RunStatus_NO_SCORES},
		{name: "NoScoresNonZeroExit", runErr: &ContainerExitError{Code: 2}, out: "compile error", parsedScores: 0, want: score.RunStatus_NO_SCORES},
		{name: "Panic", runErr: nil, out: "goroutine 1 [running]:\npanic: runtime error", parsedScores: 0, want: score.RunStatus_TEST_PANIC},
		{name: "PanicWithScores", runErr: nil, out: "panic: runtime error", parsedScores: 2, want: score.RunStatus_SUCCESS},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyRun(tc.runErr, tc.out, tc.parsedScores); got != tc.want {
				t.Errorf("classifyRun(%v, %q, %d) = %s, want %s", tc.runErr, tc.out, tc.parsedScores, got, tc.want)
			}
		})
	}
}

func TestFailedRunResults(t *testing.T) {
	results := &score.Results{
		BuildInfo: &score.BuildInfo{BuildLog: "compile error"},
		Scores:    []*score.Score{{TestName: "Test", Score: 5, MaxScore: 10, Weight: 1}},
	}
	failed := failedRunResults(score.RunStatus_NO_SCORES, results)
	if len(failed.Scores) != 0 {
		t.Errorf("failedRunResults() has %d scores, want 0", len(failed.Scores))
	}
	buildInfo := failed.GetBuildInfo()
	if buildInfo.GetStatus() != score.RunStatus_NO_SCORES {
		t.Errorf("failedRunResults() status = %s, want NO_SCORES", buildInfo.GetStatus())
	}
	wantPrefix := studentFailureMessage(score.RunStatus_NO_SCORES)
	if !strings.HasPrefix(buildInfo.GetBuildLog(), wantPrefix) {
		t.Errorf("failedRunResults() build log = %q, want prefix %q", buildInfo.GetBuildLog(), wantPrefix)
	}
	if !strings.HasSuffix(buildInfo.GetBuildLog(), "compile error") {
		t.Errorf("failedRunResults() build log = %q, want suffix %q", buildInfo.GetBuildLog(), "compile error")
	}
}
