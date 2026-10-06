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

// Build check output fixtures; see buildCheckCommands.
const (
	buildCheckFailedOut = buildCheckStartMarker + `
# example/student
./student.go:8:2: undefined: missing
` + buildCheckFailedMarker
	buildCheckPassedOut = buildCheckStartMarker + "\n" + buildCheckOKMarker
	// The run script's go get and go mod tidy have not yet run, so an
	// unresolved dependency is not attributable to the submitted code.
	buildCheckUnattributableOut = buildCheckStartMarker + `
student.go:5:2: no required module provides package github.com/quickfeed/quickfeed/kit/score; to add it:
	go get github.com/quickfeed/quickfeed/kit/score
` + buildCheckFailedMarker
	// The compiler summary printed by go test for the whole module, which
	// includes the course's tests copied in by the run script.
	testCommandBuildFailed = "\nFAIL\texample/student [build failed]"
)

func TestBuildCheckCommands(t *testing.T) {
	if got := buildCheckCommands(languageDotNet); got != nil {
		t.Errorf("buildCheckCommands(%q) = %q, want nil", languageDotNet, got)
	}
	if got := buildCheckCommands(""); got != nil {
		t.Errorf("buildCheckCommands(\"\") = %q, want nil", got)
	}
	got := buildCheckCommands(languageGo)
	if len(got) != 1 {
		t.Fatalf("buildCheckCommands(%q) = %q, want a single command", languageGo, got)
	}
	for _, marker := range []string{buildCheckStartMarker, buildCheckOKMarker, buildCheckFailedMarker} {
		if !strings.Contains(got[0], marker) {
			t.Errorf("buildCheckCommands(%q) = %q, want it to contain %q", languageGo, got[0], marker)
		}
	}
}

func TestBuildCheckFailed(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want bool
	}{
		{name: "NoBuildCheck", out: "FAIL\texample/student [build failed]", want: false},
		{name: "Compiles", out: buildCheckPassedOut + testCommandBuildFailed, want: false},
		{name: "DoesNotCompile", out: buildCheckFailedOut + testCommandBuildFailed, want: true},
		{name: "MissingDependency", out: buildCheckUnattributableOut, want: false},
		{name: "MissingSumEntry", out: buildCheckStartMarker + "\nstudent.go:5:2: missing go.sum entry for module example.com/dep\n" + buildCheckFailedMarker, want: false},
		{name: "DownloadFailure", out: buildCheckStartMarker + "\ngo: example.com/dep@v1.0.0: dial tcp: i/o timeout\n" + buildCheckFailedMarker, want: false},
		{name: "MissingAssignmentFolder", out: buildCheckStartMarker + "\nbash: line 2: cd: /quickfeed/submitted/lab1: No such file or directory\n" + buildCheckFailedMarker, want: false},
		{name: "Unfinished", out: buildCheckStartMarker + "\n./student.go:8:2: undefined: missing", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := buildCheckFailed(tc.out); got != tc.want {
				t.Errorf("buildCheckFailed(%q) = %t, want %t", tc.out, got, tc.want)
			}
		})
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
		{name: "BuildCheckFailure", runErr: &ContainerExitError{Code: 1}, out: buildCheckFailedOut + testCommandBuildFailed, parsedScores: 0, want: score.RunStatus_BUILD_FAILURE},
		{name: "BuildCheckFailureWithScores", runErr: &ContainerExitError{Code: 1}, out: buildCheckFailedOut + testCommandBuildFailed, parsedScores: 2, want: score.RunStatus_SUCCESS},
		// The submitted code compiles on its own; the test command's compiler
		// summary therefore blames the course's tests or their dependencies.
		{name: "TestCompilationFailure", runErr: &ContainerExitError{Code: 1}, out: buildCheckPassedOut + testCommandBuildFailed, parsedScores: 0, want: score.RunStatus_NO_SCORES},
		{name: "UnattributableBuildCheckFailure", runErr: &ContainerExitError{Code: 1}, out: buildCheckUnattributableOut + testCommandBuildFailed, parsedScores: 0, want: score.RunStatus_NO_SCORES},
		// A language without a build check has no trustworthy zero score.
		{name: "DotNetCompilationFailure", runErr: &ContainerExitError{Code: 1}, out: "Program.cs(3,4): error CS1002: ; expected\nBuild FAILED.", parsedScores: 0, want: score.RunStatus_NO_SCORES},
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

func TestBuildCheckFailureProducesZeroScore(t *testing.T) {
	const (
		secret = "quickfeed-session-secret"
		out    = buildCheckFailedOut + testCommandBuildFailed
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
