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
// when the submitted code fails tests. The reliable environment-failure
// signal is that the run produced no parsable score output at all, since
// even non-compiling submissions emit the zero-score initialization lines.
func classifyRun(runErr error, out string, parsedScores int) score.RunStatus {
	switch {
	case runErr != nil && errors.Is(runErr, context.DeadlineExceeded):
		return score.RunStatus_TIMEOUT
	case runErr != nil && out == "":
		return score.RunStatus_BUILD_FAILURE
	case parsedScores > 0:
		return score.RunStatus_SUCCESS
	case strings.Contains(out, "panic:"):
		return score.RunStatus_TEST_PANIC
	default:
		return score.RunStatus_NO_SCORES
	}
}

// studentFailureMessage returns the first line of the build log shown to the
// student for a failed run.
func studentFailureMessage(status score.RunStatus) string {
	switch status {
	case score.RunStatus_BUILD_FAILURE:
		return "The test environment failed before your code could be tested. This is not a problem with your submission; please notify the teaching staff."
	case score.RunStatus_TIMEOUT:
		return "The test run timed out. Please check for infinite loops or other slowness in your code; otherwise notify the teaching staff."
	case score.RunStatus_TEST_PANIC:
		return "The test run panicked before producing test results. If your code runs locally, this may be a problem with the test environment; please notify the teaching staff."
	case score.RunStatus_NO_SCORES:
		return "The test run produced no test results. If your code compiles locally, this may be a problem with the test environment; please notify the teaching staff."
	}
	return ""
}

// failedRunResults returns a Results object for a failed run, carrying only
// build info: the failure status and a build log starting with a student-facing
// explanation. The nil Scores let RecordResults keep the previous submission's
// scores instead of overwriting them with zeros.
func failedRunResults(status score.RunStatus, results *score.Results) *score.Results {
	buildInfo := results.GetBuildInfo()
	buildLog := studentFailureMessage(status)
	if log := buildInfo.GetBuildLog(); log != "" {
		buildLog += "\n\n" + log
	}
	buildInfo.BuildLog = buildLog
	buildInfo.Status = status
	return &score.Results{BuildInfo: buildInfo}
}
