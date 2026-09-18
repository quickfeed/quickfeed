package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/quickfeed/quickfeed/assignments"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

const runSynopsis = `Usage: qcm run -course ORG -lab LAB (-submission DIR | -user LOGIN | -group NAME) [flags]

Run the course's tests for one assignment and print the resulting scores.
The tests run in Docker, exactly as they do on the QuickFeed server, and read
the course's tests and assignments repositories from <dir>/<org>.

The code to test comes from one of:

	-submission DIR   a local directory, laid out like a student repository
	-user LOGIN       the student repository of the given GitHub login
	-group NAME       the repository of the given group

Use -submission to iterate on an assignment without pushing anything.`

func runCmd(args []string, stdout, stderr io.Writer) error {
	fs := newFlagSet("run", stderr, runSynopsis)
	var c commonFlags
	c.register(fs)
	lab := fs.String("lab", "", "`name` of the assignment to test, e.g., lab1 (required)")
	submission := fs.String("submission", "", "local `directory` holding the code to test")
	user := fs.String("user", "", "GitHub `login` of the student whose repository to test")
	group := fs.String("group", "", "`name` of the group whose repository to test")
	timeout := fs.Duration("timeout", 0, "container `timeout`; overrides the assignment's own timeout")
	build := fs.Bool("build", true, "build the course's Docker image before running the tests")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := c.validate(); err != nil {
		return err
	}
	if *lab == "" {
		return errors.New("missing required flag -lab")
	}
	if err := exactlyOneSource(*submission, *user, *group); err != nil {
		return err
	}
	if *submission != "" {
		if err := checkSubmissionDir(*submission, *lab); err != nil {
			return err
		}
	}

	logger := c.logger(stderr)
	course := c.course()
	parsed, buildContext, issues, err := assignments.ReadTestsRepository(c.testsDir(), 0)
	if err != nil {
		return fmt.Errorf("reading tests repository %s: %w", c.testsDir(), err)
	}
	printIssues(stderr, issues)
	assignment, err := findAssignment(parsed, *lab)
	if err != nil {
		return err
	}
	dockerfile, err := courseDockerfile(buildContext)
	if err != nil {
		return err
	}
	course.UpdateDockerfile(dockerfile)

	ctx := qlog.NewContext(context.Background(), logger)
	runner, err := newDockerRunner(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = runner.Close() }()

	if dockerfile != "" && *build {
		fmt.Fprintf(stdout, "Building the %s image from the course's Dockerfile\n", course.DockerImage())
		if err := assignments.BuildDockerImage(ctx, runner, course, buildContext); err != nil {
			return fmt.Errorf("building course image: %w", err)
		}
	}

	runData, sc, err := newRunData(course, assignment, *submission, *user, *group, logger, &c)
	if err != nil {
		return err
	}
	return runAndReport(ctx, runData, sc, runner, *timeout, stdout)
}

// runAndReport runs one test job and prints the resulting scores. A test run
// that failed is reported as an error, so that the command exits non-zero.
func runAndReport(ctx context.Context, runData *ci.RunData, sc scm.SCM, runner ci.Runner, timeout time.Duration, stdout io.Writer) error {
	fmt.Fprintf(stdout, "Running the tests for %s\n", runData.Assignment.GetName())
	results, err := runTests(ctx, runData, sc, runner, timeout)
	if err != nil {
		return fmt.Errorf("running tests for %s: %w", runData.Assignment.GetName(), err)
	}
	printResults(stdout, results)
	if results.Failed() {
		return fmt.Errorf("test run failed with status %s", results.GetBuildInfo().GetStatus())
	}
	return nil
}

// runTests runs one test job, bounded by the given timeout or else by the
// assignment's own; see withTimeout. sc is only needed when the code to test
// has to be cloned, and is nil for a run of a local directory. A run cannot
// start while a container of the same name exists, which is the case when the
// same code is already being tested, or when an earlier run of it was
// interrupted before removing its container; the error says so.
func runTests(ctx context.Context, runData *ci.RunData, sc scm.SCM, runner ci.Runner, timeout time.Duration) (*score.Results, error) {
	ctx, cancel := withTimeout(ctx, runData.Assignment, timeout)
	defer cancel()
	results, err := runData.RunTests(ctx, sc, runner)
	if errors.Is(err, ci.ErrConflict) {
		return nil, fmt.Errorf("%w: wait for the running job %s to finish, or remove its container if the job was interrupted", err, runData)
	}
	return results, err
}

// newRunData returns the run data for the requested submission source, along
// with the SCM client needed to fetch it. A local submission directory needs
// no SCM client, and the returned client is then nil.
func newRunData(course *qf.Course, assignment *qf.Assignment, submission, user, group string, logger *slog.Logger, c *commonFlags) (*ci.RunData, scm.SCM, error) {
	if submission != "" {
		return ci.NewLocalRun(course, assignment, localOwner, submission, localCommitID), nil, nil
	}
	repo := qf.RepoURL{ProviderURL: "github.com", Organization: c.org}
	runData := &ci.RunData{
		Course:     course,
		Assignment: assignment,
		CommitID:   localCommitID,
	}
	if user != "" {
		runData.Repo = &qf.Repository{HTMLURL: repo.StudentRepoURL(user), RepoType: qf.Repository_USER}
		runData.JobOwner = user
	} else {
		runData.Repo = &qf.Repository{HTMLURL: repo.GroupRepoURL(group), RepoType: qf.Repository_GROUP}
		runData.JobOwner = group
	}
	sc, err := c.scmClient(logger)
	if err != nil {
		return nil, nil, err
	}
	return runData, sc, nil
}

// exactlyOneSource returns an error unless exactly one submission source is given.
func exactlyOneSource(submission, user, group string) error {
	given := 0
	for _, source := range []string{submission, user, group} {
		if source != "" {
			given++
		}
	}
	switch {
	case given == 0:
		return errors.New("missing the code to test: give one of -submission, -user, or -group")
	case given > 1:
		return errors.New("conflicting sources for the code to test: give only one of -submission, -user, or -group")
	}
	return nil
}

// checkSubmissionDir fails if the local directory holding the code to test is
// missing or has no folder for the assignment. The run would fail on either,
// but only after the course's image has been built, which can take minutes.
func checkSubmissionDir(dir, lab string) error {
	if !isDir(dir) {
		return fmt.Errorf("submission directory %q does not exist", dir)
	}
	if !isDir(filepath.Join(dir, lab)) {
		return fmt.Errorf("submission directory %q has no %s folder; it must be laid out like a student repository", dir, lab)
	}
	return nil
}

// findAssignment returns the parsed assignment with the given name, or an error
// naming the assignments that the tests repository does define.
func findAssignment(parsed []*qf.Assignment, name string) (*qf.Assignment, error) {
	for _, assignment := range parsed {
		if assignment.GetName() == name {
			return assignment, nil
		}
	}
	names := make([]string, 0, len(parsed))
	for _, assignment := range parsed {
		names = append(names, assignment.GetName())
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("assignment %q not found; the tests repository defines no assignments", name)
	}
	return nil, fmt.Errorf("assignment %q not found; available assignments: %s", name, strings.Join(names, ", "))
}

// withTimeout bounds a test run. An explicit timeout overrides the assignment's
// own container timeout, which in turn overrides the server's default.
func withTimeout(ctx context.Context, assignment *qf.Assignment, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(ctx, timeout)
	}
	return assignment.WithTimeout(ctx, ci.DefaultContainerTimeout)
}
