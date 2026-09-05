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
		{name: "NoOutput", runErr: errors.New("failed to create container"), out: "", parsedScores: 0, want: score.RunStatus_NO_SCORES},
		{name: "NoOutputNonZeroExit", runErr: &ContainerExitError{Code: 2}, out: "", parsedScores: 0, want: score.RunStatus_NO_SCORES},
		{name: "GoCompilationFailure", runErr: &ContainerExitError{Code: 1}, out: "FAIL\texample/student [build failed]", parsedScores: 0, want: score.RunStatus_BUILD_FAILURE},
		{name: "DotNetCompilationFailure", runErr: &ContainerExitError{Code: 1}, out: "Program.cs(3,4): error CS1002: ; expected\nBuild FAILED.", parsedScores: 0, want: score.RunStatus_BUILD_FAILURE},
		{name: "CompilationFailureWithScores", runErr: &ContainerExitError{Code: 1}, out: "FAIL\texample/student [build failed]", parsedScores: 2, want: score.RunStatus_SUCCESS},
		{name: "NoScores", runErr: nil, out: "test command produced no score output", parsedScores: 0, want: score.RunStatus_NO_SCORES},
		{name: "NoScoresNonZeroExit", runErr: &ContainerExitError{Code: 2}, out: "dependency download failed", parsedScores: 0, want: score.RunStatus_NO_SCORES},
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

func TestCompilationFailureProducesZeroScore(t *testing.T) {
	const (
		secret = "quickfeed-session-secret"
		out    = "# example/student\n./student.go:8:2: undefined: missing\nFAIL\texample/student [build failed]"
	)
	results, err := score.ExtractResults(out, secret, 10, []*score.Score{{TestName: "TestStudent", MaxScore: 10, Weight: 1}})
	if err != nil {
		t.Fatal(err)
	}
	status := classifyRun(&ContainerExitError{Code: 1}, out, results.ParsedScores)
	if status != score.RunStatus_BUILD_FAILURE {
		t.Fatalf("classifyRun() = %s, want BUILD_FAILURE", status)
	}
	failed := failedRunResults(status, results)
	if !failed.ScoresValid() {
		t.Fatal("compilation failure scores must be valid")
	}
	if got := failed.Sum(); got != 0 {
		t.Errorf("compilation failure score = %d, want 0", got)
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
