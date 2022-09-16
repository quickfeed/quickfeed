package ci

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/quickfeed/quickfeed/internal/rand"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"go.uber.org/zap"
)

// pattern to prefix the tmp folder for quickfeed tests
const quickfeedTestsPath = "quickfeed-tests"

// RunData stores CI data
type RunData struct {
	Course     *qf.Course
	Assignment *qf.Assignment
	Repo       *qf.Repository
	BranchName string
	CommitID   string
	JobOwner   string
	Rebuild    bool
}

// String returns a string representation of the run data structure.
func (r RunData) String() string {
	commitID := r.CommitID
	if len(commitID) > 7 {
		commitID = r.CommitID[:6]
	}
	return fmt.Sprintf("%s-%s-%s-%s", strings.ToLower(r.Course.GetCode()), r.Assignment.GetName(), r.JobOwner, commitID)
}

// RunTests runs the assignment specified in the provided RunData structure.
// This function can be called concurrently on different RunData objects;
// the function is idempotent. That is, it only clones repositories from GitHub,
// runs the tests and returns the score results. The os.MkdirTemp() function ensures that
// any concurrent calls to this function will always use distinct temp directories.
//
// Note that this function creates a temporary directory on the host machine running
// the quickfeed server. This directory holds the cloned repositories (student and tests repos)
// and will be mounted as '/quickfeed' inside the container, allowing the docker container
// to run the tests on the student code. The temporary directory is deleted when the container
// exits at the end of this function.
func (r RunData) RunTests(ctx context.Context, logger *zap.SugaredLogger, sc scm.SCM, runner Runner) (*score.Results, error) {
	testsStartedCounter.WithLabelValues(r.JobOwner, r.Course.Code).Inc()

	dstDir, err := os.MkdirTemp("", quickfeedTestsPath)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dstDir)

	if err = r.cloneRepositories(ctx, logger, sc, dstDir); err != nil {
		return nil, err
	}

	randomSecret := rand.String()
	job, err := r.parseTestRunnerScript(randomSecret, dstDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse run script: %w", err)
	}

	defer timer(r.JobOwner, r.Course.Code, testExecutionTimeGauge)()
	logger.Debugf("Running tests for %s", r)
	start := time.Now()
	out, err := runner.Run(ctx, job)
	if err != nil && out == "" {
		testsFailedCounter.WithLabelValues(r.JobOwner, r.Course.Code).Inc()
		return nil, fmt.Errorf("test execution failed without output: %w", err)
	}
	if err != nil {
		// We may reach here with a timeout error and a non-empty output
		logger.Errorf("Test execution failed with output: %v\n%v", err, out)
	}

	results, err := score.ExtractResults(out, randomSecret, time.Since(start))
	if err != nil {
		// Log the errors from the extraction process
		logger.Debugf("Session secret: %s", randomSecret)
		logger.Errorf("Failed to extract (some) results for assignment %s for course %s: %v", r.Assignment.Name, r.Course.Name, err)
		// don't return here; we still want partial results!
	}

	testsSucceededCounter.WithLabelValues(r.JobOwner, r.Course.Code).Inc()
	logger.Debug("ci.RunTests", zap.Any("Results", qlog.IndentJson(results)))
	// return the extracted score and filtered log output
	return results, nil
}
