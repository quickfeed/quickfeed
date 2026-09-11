package hooks

import (
	"context"
	"errors"
	"strings"
	"time"

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
// The returned names are folder names, not assignment names; runCanary matches
// them against the course's assignments and ignores those that match nothing.
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
// the tests; see canaryTargets and qf.Assignment.GradedManually.
func (wh GitHubWebHook) runCanary(ctx context.Context, scmClient scm.SCM, course *qf.Course, payload *github.PushEvent) {
	logger := qlog.FromContext(ctx)
	names, all := canaryTargets(payload)

	courseAssignments, err := wh.db.GetAssignmentsByCourse(course.GetID())
	if err != nil {
		logger.Error("failed to get course assignments", label.Error, err)
		return
	}
	known := make(map[string]bool, len(courseAssignments))
	var selected []*qf.Assignment
	for _, assignment := range courseAssignments {
		known[assignment.GetName()] = true
		if assignment.GradedManually() {
			// Nothing to run: the assignment has no tests to go wrong.
			continue
		}
		if !all && !names[assignment.GetName()] {
			continue
		}
		selected = append(selected, assignment)
	}
	for name := range names {
		if !known[name] {
			logger.Debug("changed folder is not a known assignment", label.Assignment, name)
		}
	}
	if len(selected) == 0 {
		logger.Debug("no assignments to check with the test environment canary", "all_assignments", all)
		return
	}

	start := time.Now()
	problems := 0
	for _, assignment := range selected {
		if wh.runCanaryFor(ctx, scmClient, course, assignment, payload.GetHeadCommit().GetID()) {
			problems++
		}
	}
	logger.Info("test environment canary finished",
		"assignments", len(selected), "problems", problems, label.Duration, time.Since(start))
}

// runCanaryFor runs the canary for a single assignment and reports whether the
// run found a problem. The run is bounded by the assignment's container
// timeout, as a student's run would be.
func (wh GitHubWebHook) runCanaryFor(ctx context.Context, scmClient scm.SCM, course *qf.Course, assignment *qf.Assignment, commitID string) bool {
	ctx, logger := qlog.WithLogger(ctx, label.Assignment, assignment.GetName())
	runData := ci.NewSkeletonRun(course, assignment, commitID)
	ctx, cancel := assignment.WithTimeout(ctx, ci.DefaultContainerTimeout)
	defer cancel()

	results, err := runData.RunTests(ctx, scmClient, wh.runner)
	if err != nil {
		if errors.Is(err, ci.ErrConflict) {
			// The same canary is already running; the running one reports.
			logger.Debug("canary already running")
			return false
		}
		logger.Warn("test environment canary could not run", label.Error, err)
		return true
	}
	if problem := ci.SkeletonProblem(results, ci.MaxSkeletonScore); problem != "" {
		logger.Warn("test environment canary failed", "problem", problem,
			"score", results.Sum(), "run_status", results.GetBuildInfo().GetStatus().String())
		return true
	}
	logger.Info("test environment canary passed", "score", results.Sum())
	return false
}
