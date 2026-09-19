package hooks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/assignments"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/fileop"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

// canaryOutcome is the outcome of the canary for a single assignment.
type canaryOutcome int

const (
	// canaryPassed means the run completed and found nothing wrong.
	canaryPassed canaryOutcome = iota
	// canaryFailed means the run found a problem, or could not run at all.
	canaryFailed
	// canarySkipped means nothing ran because an identical run was already in
	// progress; that run reports for this one.
	canarySkipped
)

// canaryTargets returns the top-level folders whose tests the canary should
// run for the given push event, and whether all assignments must be checked.
//
// A change outside a single assignment folder can impact every assignment and
// widens the run to all of them: a file in the repository root, or any file in
// the scripts folder, which holds the course-wide run script and Dockerfile. A
// top-level dot folder, such as .github, holds repository metadata and triggers
// nothing. A top-level folder that is not an assignment widens the run for the
// same reason, but the push event does not say which folders those are, so
// canaryAssignments matches the returned names against the course's assignments.
func canaryTargets(payload *github.PushEvent) (names map[string]bool, all bool) {
	names = make(map[string]bool)
	for _, commit := range payload.Commits {
		for _, changes := range [][]string{commit.Added, commit.Modified, commit.Removed} {
			for _, changedFile := range changes {
				name, _, nested := strings.Cut(changedFile, "/")
				switch {
				case name == "":
					// Empty path or a path with a leading separator.
				case !nested:
					// A file in the repository root.
					all = true
				case name == ci.ScriptsDir:
					all = true
				case strings.HasPrefix(name, "."):
					// Repository metadata, such as .github.
				default:
					names[name] = true
				}
			}
		}
	}
	return names, all
}

// cloneHead returns the commit that the git clone in dir is at.
//
// A dir that is not a git repository is reported as the empty string and no
// error, which callers must treat as a match: only local test fixtures build a
// course clone directory out of plain directories; on the server the course
// repositories are always go-git clones by the time the canary runs.
func cloneHead(dir string) (string, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			return "", nil
		}
		return "", err
	}
	head, err := repo.Head()
	if err != nil {
		return "", err
	}
	return head.Hash().String(), nil
}

// skipCanary reports whether the canary must be skipped because the local clone
// of the pushed repository is not at the pushed commit, logging why.
//
// assignments.UpdateFromCourseRepositories reports success even when refreshing
// a clone failed, and testing a stale clone would report problems that the
// pushed commit does not have. Only the pushed repository is checked; a
// concurrent push to the other course repository can still leave the two clones
// at commits from different pushes, and the canary then reports on the pushed
// commit against the current skeleton.
func skipCanary(ctx context.Context, course *qf.Course, pushedRepo, commitID string) bool {
	head, err := cloneHead(filepath.Join(course.CloneDir(), pushedRepo))
	if err == nil && (head == "" || head == commitID) {
		return false
	}
	// The push scope already names the repository and the pushed commit.
	attrs := []any{"clone_head", head}
	if err != nil {
		attrs = append(attrs, label.Error, err)
	}
	qlog.FromContext(ctx).Warn("skipping test environment canary: local clone is not at the pushed commit", attrs...)
	return true
}

// canaryAssignments returns the assignments the canary must check for a push
// touching the given top-level folders, and whether every assignment is
// checked; names and all are what canaryTargets returned for the push.
//
// Only assignments graded by the tests are checked, since a manually graded
// assignment has no tests to go wrong. A folder name matching no assignment
// widens the run to every assignment: such a folder holds code shared by the
// assignments, such as one listed in .quickfeedignore, and nothing in the push
// says which of them use it.
func canaryAssignments(ctx context.Context, courseAssignments []*qf.Assignment, names map[string]bool, all bool) ([]*qf.Assignment, bool) {
	logger := qlog.FromContext(ctx)
	known := make(map[string]bool, len(courseAssignments))
	for _, assignment := range courseAssignments {
		known[assignment.GetName()] = true
	}
	for name := range names {
		if !known[name] {
			all = true
			logger.Debug("changed folder is not an assignment; checking every assignment", label.Assignment, name)
		}
	}
	var selected []*qf.Assignment
	for _, assignment := range courseAssignments {
		if !all && !names[assignment.GetName()] {
			continue
		}
		if assignment.GradedManually() {
			continue
		}
		selected = append(selected, assignment)
	}
	return selected, all
}

// snapshotCanaryRepositories copies the course's tests and assignments
// repositories into a new temporary directory and returns it.
//
// The canary reads these repositories for as long as its runs take, while
// assignments.UpdateFromCourseRepositories refreshes the course's clones in
// place; running against a private copy keeps a concurrent update from changing
// files under a running container, which the canary would report as a broken
// test environment. The caller must hold the course lock, so that the copy is
// of a single complete state, and must remove the directory when the runs are
// done. The runs themselves need no lock, since the copy cannot change.
func snapshotCanaryRepositories(course *qf.Course) (string, error) {
	snapshotDir, err := os.MkdirTemp("", "quickfeed-canary-")
	if err != nil {
		return "", err
	}
	for _, repo := range []string{qf.TestsRepo, qf.AssignmentsRepo} {
		if err := fileop.CopyDir(filepath.Join(course.CloneDir(), repo), filepath.Join(snapshotDir, repo)); err != nil {
			_ = os.RemoveAll(snapshotDir)
			return "", fmt.Errorf("snapshotting %s repository: %w", repo, err)
		}
	}
	return snapshotDir, nil
}

// prepareCanary returns the assignments to check for the given push, and the
// snapshot directory for the canary to run against. It returns false if there
// is nothing to do.
//
// The course lock is held throughout, so that the snapshot is taken from the
// same state skipCanary verified and no update can interleave. It is released
// before the canary runs, which take far longer and must never block an update;
// see snapshotCanaryRepositories.
func (wh GitHubWebHook) prepareCanary(ctx context.Context, course *qf.Course, pushedRepo string, payload *github.PushEvent) ([]*qf.Assignment, string, bool) {
	logger := qlog.FromContext(ctx)
	names, all := canaryTargets(payload)

	unlock := course.Lock()
	defer unlock()

	if skipCanary(ctx, course, pushedRepo, payload.GetHeadCommit().GetID()) {
		return nil, "", false
	}
	courseAssignments, err := wh.db.GetAssignmentsByCourse(course.GetID())
	if err != nil {
		logger.Error("failed to get course assignments", label.Error, err)
		return nil, "", false
	}
	selected, all := canaryAssignments(ctx, courseAssignments, names, all)
	if len(selected) == 0 {
		logger.Debug("no assignments to check with the test environment canary", "all_assignments", all)
		return nil, "", false
	}
	snapshotDir, err := snapshotCanaryRepositories(course)
	if err != nil {
		logger.Warn("skipping test environment canary: failed to snapshot course repositories", label.Error, err)
		return nil, "", false
	}
	return selected, snapshotDir, true
}

// runCanaryAfterUpdate runs the canary for a push whose assignment update
// succeeded. The canary is skipped when the update could not refresh the
// assignments clone, since the skeleton code the canary tests would then be
// missing or stale; see assignments.UpdateResult.
func (wh GitHubWebHook) runCanaryAfterUpdate(ctx context.Context, scmClient scm.SCM, course *qf.Course, pushedRepo string, payload *github.PushEvent, update assignments.UpdateResult) {
	if !update.AssignmentsCloneReady {
		qlog.FromContext(ctx).Warn("skipping test environment canary: assignments clone was not refreshed")
		return
	}
	wh.runCanary(ctx, scmClient, course, pushedRepo, payload)
}

// runCanary runs the course's own tests against the skeleton code in the
// course's assignments repository, and reports a suspect test environment to
// the teaching staff through the course log.
//
// The canary is a diagnostic: it is called only after the database has been
// updated from the course repositories, it never returns an error, it writes
// nothing to the database, and its only effect is the log records below.
// Callers must keep it last, so that a broken test environment does not prevent
// updating the assignment information or synchronizing the student repositories.
//
// Only the assignments the push touched are checked, and only those graded by
// the tests; see canaryTargets and canaryAssignments. The canary run uses a
// snapshot of the course repositories; see prepareCanary.
//
// The runs are made on a context derived with context.WithoutCancel, keeping
// the values, hence the scoped logger, but not the handler's webhook deadline.
// A push to the scripts folder or to the repository root selects every
// auto-graded assignment, and the runs are sequential, so a later run can start
// after that deadline has passed; inheriting it would report such a run as
// failed instead of running it. Each run is instead bounded by its own
// container timeout, and the canary as a whole by the number of selected
// assignments. The handler holds its semaphore slot throughout, which is what
// bounds the number of concurrent runs; see maxConcurrentTestRuns.
func (wh GitHubWebHook) runCanary(ctx context.Context, scmClient scm.SCM, course *qf.Course, pushedRepo string, payload *github.PushEvent) {
	logger := qlog.FromContext(ctx)
	selected, snapshotDir, ok := wh.prepareCanary(ctx, course, pushedRepo, payload)
	if !ok {
		return
	}
	defer func() {
		// A leaked snapshot is a full copy of both course repositories.
		if err := os.RemoveAll(snapshotDir); err != nil {
			logger.Warn("failed to remove test environment canary snapshot",
				label.Path, snapshotDir, label.Error, err)
		}
	}()

	// Detach from the webhook deadline; see the note above.
	runCtx := context.WithoutCancel(ctx)
	commitID := payload.GetHeadCommit().GetID()
	start := time.Now()
	problems, skipped := 0, 0
	for _, assignment := range selected {
		switch wh.runCanaryFor(runCtx, scmClient, course, assignment, commitID, snapshotDir) {
		case canaryFailed:
			problems++
		case canarySkipped:
			skipped++
		}
	}
	logger.Info("test environment canary finished", "assignments", len(selected),
		"problems", problems, "skipped", skipped, label.Duration, time.Since(start))
}

// runCanaryFor runs the canary for a single assignment and reports its outcome.
// Both course repositories are read from snapshotDir, and the run is bounded by
// the assignment's container timeout, as a student's run would be.
func (wh GitHubWebHook) runCanaryFor(ctx context.Context, scmClient scm.SCM, course *qf.Course, assignment *qf.Assignment, commitID, snapshotDir string) canaryOutcome {
	ctx, logger := qlog.WithLogger(ctx, label.Assignment, assignment.GetName())
	runData := ci.NewSkeletonRunFromSnapshot(course, assignment, commitID, snapshotDir)
	ctx, cancel := assignment.WithTimeout(ctx, ci.DefaultContainerTimeout)
	defer cancel()

	results, err := runData.RunTests(ctx, scmClient, wh.runner)
	if err != nil {
		if errors.Is(err, ci.ErrConflict) {
			// The same canary is already running; the running one reports.
			logger.Debug("canary already running")
			return canarySkipped
		}
		logger.Warn("test environment canary could not run", label.Error, err)
		return canaryFailed
	}
	if problem := ci.SkeletonProblem(results, ci.MaxSkeletonScore); problem != "" {
		logger.Warn("test environment canary failed", "problem", problem,
			"score", results.Sum(), "run_status", results.GetBuildInfo().GetStatus().String())
		return canaryFailed
	}
	logger.Info("test environment canary passed", "score", results.Sum())
	return canaryPassed
}
