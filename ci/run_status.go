package ci

import (
	"context"
	"errors"
	"strings"

	"github.com/quickfeed/quickfeed/kit/score"
)

// classifyRun determines the outcome of a test run from the run error, the
// container output, and the number of valid score lines parsed from it.
//
// A non-zero container exit status alone does not classify as a failure:
// a course's run script may end with the test command, which exits non-zero
// when the submitted code fails tests. A compilation failure is also a
// trustworthy zero-score result even though the test binary never started
// and therefore could not print its zero-score initialization lines.
func classifyRun(runErr error, out string, parsedScores int) score.RunStatus {
	switch {
	case runErr != nil && errors.Is(runErr, context.DeadlineExceeded):
		return score.RunStatus_TIMEOUT
	case parsedScores > 0:
		return score.RunStatus_SUCCESS
	case compilationFailed(out):
		return score.RunStatus_BUILD_FAILURE
	case runErr != nil && out == "":
		return score.RunStatus_NO_SCORES
	case strings.Contains(out, "panic:"):
		return score.RunStatus_TEST_PANIC
	default:
		return score.RunStatus_NO_SCORES
	}
}

// compilationFailed recognizes the compiler summaries emitted by the Go and
// .NET test commands used by the bundled course templates. These failures are
// attributable to the submitted code and produce a trustworthy zero score.
func compilationFailed(out string) bool {
	lower := strings.ToLower(out)
	if strings.Contains(lower, "[build failed]") {
		return true
	}
	return strings.Contains(lower, "build failed.") && strings.Contains(lower, ": error ")
}

// studentFailureMessage returns the first line of the build log shown to the
// student for a failed run.
func studentFailureMessage(status score.RunStatus) string {
	switch status {
	case score.RunStatus_BUILD_FAILURE:
		return "The submitted code did not compile. The run was recorded with a zero score."
	case score.RunStatus_TIMEOUT:
		return "The test run timed out. Please check for infinite loops or other slowness in your code; otherwise notify the teaching staff."
	case score.RunStatus_TEST_PANIC:
		return "The test run panicked before producing test results. If your code runs locally, this may be a problem with the test environment; please notify the teaching staff."
	case score.RunStatus_NO_SCORES:
		return "The test run produced no test results. If your code compiles locally, this may be a problem with the test environment; please notify the teaching staff."
	}
	return ""
}

// failedRunResults annotates results with the failure status and a build log
// starting with a student-facing explanation. Failures without trustworthy
// scores clear the score list so RecordResults keeps the previous scores.
func failedRunResults(status score.RunStatus, results *score.Results) *score.Results {
	buildInfo := results.GetBuildInfo()
	buildLog := studentFailureMessage(status)
	if log := buildInfo.GetBuildLog(); log != "" {
		buildLog += "\n\n" + log
	}
	buildInfo.BuildLog = buildLog
	buildInfo.Status = status
	results.BuildInfo = buildInfo
	if !status.ScoresValid() {
		results.Scores = nil
	}
	return results
}
