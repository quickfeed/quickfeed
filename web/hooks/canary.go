package hooks

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

// scriptsFolder is the top-level folder in the tests repository holding the
// course-wide run script and Dockerfile; see assignments.scriptsDir.
const scriptsFolder = "scripts"

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

// canaryTargets returns the assignment folders whose tests the canary should
// run for the given push, and whether every assignment must be checked.
//
// The push may touch either course repository, whose top-level folders are the
// assignment folders. Every path in the push is therefore classified by its
// first path component:
//
//   - a path with no separator names a file in the repository root, such as
//     go.mod, and widens the run to all assignments;
//   - the scripts folder holds the course-wide run script and Dockerfile, and
//     likewise widens the run to all assignments;
//   - a folder whose name begins with a dot, such as .github, is repository
//     metadata and triggers nothing;
//   - an empty first component, from an empty path or a leading separator,
//     names no folder and triggers nothing;
//   - anything else names an assignment folder.
//
// A change outside any assignment folder widens the run because such a change
// can break every assignment at once: the run script, the Dockerfile, and the
// root-level module files are shared by all of them, and nothing in the push
// says which assignments they affect.
//
// The returned names are folder names, not assignment names. A name matching
// no assignment names a folder shared by the assignments and widens the run as
// well, but that is decided by canaryAssignments, which the course's
// assignments are available to.
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
				case name == scriptsFolder:
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

// skipCanary reports whether the canary must be skipped because the local
// clone of the pushed repository is not at the pushed commit, logging why.
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

// runCanary runs the course's own tests against the skeleton code in the
// course's assignments repository, and reports a suspect test environment to
// the teaching staff through the course log.
//
// The canary is a diagnostic, not a gate: it is called only after the database
// has been updated from the course repositories, it never returns an error, it
// writes nothing to the database, and its only effect is the log records below.
// Callers must keep it last, so that a broken test environment cannot prevent
// or delay the assignment update or the student repository sync.
//
// Only the assignments the push touched are checked, and only those graded by
// the tests; see canaryTargets and canaryAssignments.
//
// pushedRepo names the pushed course repository, tests or assignments. The
// canary is skipped unless the local clone of that repository is at the pushed
// commit: assignments.UpdateFromCourseRepositories reports success even when
// refreshing a clone failed, and testing a stale clone would report problems
// that the pushed commit does not have; see skipCanary.
//
// The runs are made on a context derived with context.WithoutCancel, keeping
// the values, hence the scoped logger, but not the handler's webhook deadline.
// A push to the scripts folder or to the repository root selects every
// auto-graded assignment, and the runs are sequential, so a later run can
// start after that deadline has passed; inheriting it would report such a run
// as failed instead of running it. Each run is instead bounded by its own
// container timeout, and the canary as a whole by the number of selected
// assignments. The handler holds its semaphore slot throughout, which is what
// bounds the number of concurrent runs; see maxConcurrentTestRuns.
func (wh GitHubWebHook) runCanary(ctx context.Context, scmClient scm.SCM, course *qf.Course, pushedRepo string, payload *github.PushEvent) {
	logger := qlog.FromContext(ctx)
	if skipCanary(ctx, course, pushedRepo, payload.GetHeadCommit().GetID()) {
		return
	}

	courseAssignments, err := wh.db.GetAssignmentsByCourse(course.GetID())
	if err != nil {
		logger.Error("failed to get course assignments", label.Error, err)
		return
	}
	names, all := canaryTargets(payload)
	selected, all := canaryAssignments(ctx, courseAssignments, names, all)
	if len(selected) == 0 {
		logger.Debug("no assignments to check with the test environment canary", "all_assignments", all)
		return
	}

	// Detach from the webhook deadline; see the note above.
	runCtx := context.WithoutCancel(ctx)
	start := time.Now()
	problems, skipped := 0, 0
	for _, assignment := range selected {
		switch wh.runCanaryFor(runCtx, scmClient, course, assignment, payload.GetHeadCommit().GetID()) {
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
// The run is bounded by the assignment's container timeout, as a student's run
// would be.
func (wh GitHubWebHook) runCanaryFor(ctx context.Context, scmClient scm.SCM, course *qf.Course, assignment *qf.Assignment, commitID string) canaryOutcome {
	ctx, logger := qlog.WithLogger(ctx, label.Assignment, assignment.GetName())
	runData := ci.NewSkeletonRun(course, assignment, commitID)
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
