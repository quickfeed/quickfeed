package ci

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/quickfeed/quickfeed/internal/fileop"
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
	EnvVarsFn  func(secret, homeDir string) []string
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

// RunTests runs the tests for the assignment specified in the provided RunData structure,
// and returns the score results or an error.
// The method is idempotent and can be called concurrently on multiple RunData objects.
// The method clones the student or group repository from GitHub as specified in RunData,
// and copies the course's tests and assignment repositories from the host machine's file system.
// runs the tests and returns the score results.
//
// The os.MkdirTemp() function ensures that any concurrent calls to this method will always
// use distinct temp directories. Specifically, the method creates a temporary directory on
// the host machine running the quickfeed server that holds the cloned/copied repositories
// and will be mounted as '/quickfeed' inside the container. This allows the docker container
// to run the tests on the student code and manipulate the folders as needed for a particular
// lab assignment's test requirements. The temporary directory is deleted when the container
// exits at the end of this method.
func (r *RunData) RunTests(ctx context.Context, logger *zap.SugaredLogger, sc scm.SCM, runner Runner) (*score.Results, error) {
	testsStartedCounter.WithLabelValues(r.JobOwner, r.Course.Code).Inc()

	dstDir, err := os.MkdirTemp("", quickfeedTestsPath)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dstDir)

	logger.Debugf("Cloning repository for %s", r)
	if err = r.clone(ctx, sc, dstDir); err != nil {
		return nil, err
	}
	logger.Debugf("Successfully cloned student repository to: %s", dstDir)

	if err := scanStudentRepo(filepath.Join(dstDir, r.Repo.Name()), r.Course.GetCode(), r.JobOwner); err != nil {
		return nil, err
	}

	randomSecret := rand.String()
	job, err := r.parseTestRunnerScript(randomSecret, dstDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse run script for assignment %s in %s: %w", r.Assignment.GetName(), r.Repo.GetTestURL(), err)
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
		testsFailedWithOutputCounter.WithLabelValues(r.JobOwner, r.Course.Code).Inc()
		logger.Errorf("Test execution failed with output: %v\n%v", err, out)
	}

	results, err := score.ExtractResults(out, randomSecret, time.Since(start))
	if err != nil {
		// Log the errors from the extraction process
		testsFailedExtractResultsCounter.WithLabelValues(r.JobOwner, r.Course.Code).Inc()
		logger.Debugf("Session secret: %s", randomSecret)
		logger.Errorf("Failed to extract (some) results for assignment %s for course %s: %v", r.Assignment.Name, r.Course.Name, err)
		// don't return here; we still want partial results!
	}

	testsSucceededCounter.WithLabelValues(r.JobOwner, r.Course.Code).Inc()
	logger.Debug("ci.RunTests", zap.Any("Results", qlog.IndentJson(results)))
	// return the extracted score and filtered log output
	return results, nil
}

func (r RunData) clone(ctx context.Context, sc scm.SCM, dstDir string) error {
	defer timer(r.JobOwner, r.Course.GetCode(), cloneTimeGauge)()

	clonedStudentRepo, err := sc.Clone(ctx, &scm.CloneOptions{
		Organization: r.Course.GetOrganizationName(),
		Repository:   r.Repo.Name(),
		DestDir:      dstDir,
		Branch:       r.BranchName,
	})
	if err != nil {
		return fmt.Errorf("failed to clone %q repository: %w", r.Repo.Name(), err)
	}

	// Clone the course's tests and assignments repositories if they are missing.
	// Cloning is only needed when the quickfeed server has not yet received a push event
	// for a course's tests or assignments repositories or an UpdateAssignment request.
	if err := cloneMissingRepositories(ctx, sc, r.Course); err != nil {
		return err
	}

	// Check that all repositories contains the current assignment
	currentAssignment := r.Assignment.GetName()
	testsDir := filepath.Join(r.Course.CloneDir(), qf.TestsRepo)
	assignmentDir := filepath.Join(r.Course.CloneDir(), qf.AssignmentRepo)
	for _, repoDir := range []string{clonedStudentRepo, testsDir, assignmentDir} {
		if err := hasAssignment(repoDir, currentAssignment); err != nil {
			return err
		}
	}
	// Copy the tests and assignment repos to the destination directory
	if err = fileop.CopyDir(testsDir, dstDir); err != nil {
		return err
	}
	return fileop.CopyDir(assignmentDir, dstDir)
}
