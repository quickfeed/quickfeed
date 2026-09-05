package hooks

import (
	"context"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/assignments"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (wh GitHubWebHook) handlePush(ctx context.Context, payload *github.PushEvent) {
	// The commit is already in scope; see the webhook Handle method.
	ctx, logger := qlog.WithLogger(ctx,
		branchRefLabel, payload.GetRef(),
		label.User, payload.GetSender().GetLogin(),
	)
	logger.Debug("received push event", "default_branch", payload.GetRepo().GetDefaultBranch())

	repo, err := wh.getRepository(payload.GetRepo().GetID())
	if err != nil {
		logger.Error("failed to get repository from database", label.Repository, payload.GetRepo().GetFullName(), label.Error, err)
		return
	}
	ctx, logger = qlog.WithLogger(ctx,
		label.Repository, repo.Name(),
		label.RepositoryType, repo.GetRepoType().String(),
	)
	logger.Debug("resolved push repository")

	if ignorePush(payload) {
		logger.Debug("ignoring push event for non-default branch")
		return
	}

	course, err := wh.db.GetCourseByOrganizationID(repo.GetScmOrganizationID())
	if err != nil {
		logger.Error("failed to get course from database", label.Error, err)
		return
	}
	ctx, logger = qlog.WithCourse(ctx, course)
	logger.Debug("resolved push course")

	if repo.IsStudentRepo() {
		wh.updateLastActivityDate(ctx, course, repo, payload.GetSender().GetLogin())
	}

	scmClient, err := wh.scmMgr.GetOrCreateSCM(ctx, course.GetScmOrganizationName())
	if err != nil {
		logger.Error("failed to create SCM client", label.Error, err)
		return
	}

	switch {
	case repo.IsTestsRepo():
		// The push event is for the tests repository, so update the course
		// assignments in the database.
		if _, err := assignments.UpdateFromTestsRepo(ctx, wh.runner, wh.db, scmClient, course); err != nil {
			logger.Error("failed to update assignments from tests repository", label.Error, err)
		}

	case repo.IsAssignmentsRepo():
		if err := validateAssignmentsPush(ctx, scmClient, course); err != nil {
			logger.Error("failed to clone repository", label.Error, err)
			return
		}
		if isDefaultBranch(payload) {
			// Sync all student repositories (forks) with the updated assignments repo
			wh.syncStudentRepos(ctx, scmClient, course, payload.GetRepo().GetDefaultBranch())
		}

	case repo.IsStudentRepo():
		if payload.GetSender().GetType() == "Bot" {
			logger.Debug("ignoring push event from bot")
			return
		}
		logger.Debug("processing student push")
		assignments := wh.extractAssignments(ctx, payload, course)
		for _, assignment := range assignments {
			wh.runAssignmentTests(ctx, scmClient, assignment, repo, course, payload)
		}

	default:
		logger.Debug("nothing to do for push event")
	}
}

// validateAssignmentsPush updates both local course repositories and checks
// their content and assignment-folder alignment. A tests repository failure is
// logged but does not prevent syncing the assignments repository to students.
func validateAssignmentsPush(ctx context.Context, scmClient scm.SCM, course *qf.Course) error {
	unlock := course.Lock()
	defer unlock()

	logger := qlog.FromContext(ctx)
	clonedAssignmentsRepo, err := scmClient.Clone(ctx, &scm.CloneOptions{
		Organization: course.GetScmOrganizationName(),
		Repository:   qf.AssignmentsRepo,
		DestDir:      course.CloneDir(),
	})
	if err != nil {
		return err
	}
	logger.Debug("cloned assignments repository", label.Path, clonedAssignmentsRepo)

	clonedTestsRepo, err := scmClient.Clone(ctx, &scm.CloneOptions{
		Organization: course.GetScmOrganizationName(),
		Repository:   qf.TestsRepo,
		DestDir:      course.CloneDir(),
	})
	if err != nil {
		logger.Error("failed to clone tests repository for validation", label.Error, err)
		return nil
	}
	logger.Debug("cloned tests repository for validation", label.Path, clonedTestsRepo)
	if _, err := assignments.ValidateCourseRepositories(ctx, clonedTestsRepo, clonedAssignmentsRepo, course.GetID()); err != nil {
		logger.Error("failed to validate course repositories", label.Error, err)
	}
	return nil
}

// ignorePush returns true if the push event should be ignored.
// Only pushes to the default branch are processed.
func ignorePush(payload *github.PushEvent) bool {
	return !isDefaultBranch(payload)
}

// extractAssignments extracts information from the push payload from github
// and determines the assignments that have been changed in this commit by
// querying the database based on the lab name.
func (wh GitHubWebHook) extractAssignments(ctx context.Context, payload *github.PushEvent, course *qf.Course) []*qf.Assignment {
	logger := qlog.FromContext(ctx)
	modifiedAssignments := make(map[string]bool)
	for _, commit := range payload.Commits {
		extractChanges(commit.Modified, modifiedAssignments)
		extractChanges(commit.Added, modifiedAssignments)
		extractChanges(commit.Removed, modifiedAssignments)
	}

	var assignments []*qf.Assignment
	for name := range modifiedAssignments {
		// get assignment based on course id and assignment name
		assignment, err := wh.db.GetAssignment(&qf.Assignment{Name: name, CourseID: course.GetID()})
		if err != nil {
			logger.Error("failed to find assignment", label.Assignment, name, label.Error, err)
			continue
		}
		assignments = append(assignments, assignment)
	}
	return assignments
}

// runAssignmentTests runs the tests for the given assignment pushed to repo.
func (wh GitHubWebHook) runAssignmentTests(ctx context.Context, scmClient scm.SCM, assignment *qf.Assignment, repo *qf.Repository, course *qf.Course, payload *github.PushEvent) {
	ctx, logger := qlog.WithLogger(ctx, label.Assignment, assignment.GetName())
	runData := &ci.RunData{
		Course:     course,
		Assignment: assignment,
		Repo:       repo,
		BranchName: branchName(payload.GetRef()),
		CommitID:   payload.GetHeadCommit().GetID(),
		JobOwner:   payload.GetSender().GetLogin(),
	}
	if assignment.GradedManually() {
		logger.Debug("assignment is manually reviewed")
		if _, err := runData.RecordResults(ctx, wh.db, nil); err != nil {
			logger.Error("failed to record manual assignment result", label.Error, err)
		}
		return
	}
	ctx, cancel := assignment.WithTimeout(ctx, ci.DefaultContainerTimeout)
	defer cancel()
	results, err := runData.RunTests(ctx, scmClient, wh.runner)
	if err != nil {
		logger.Error("failed to run assignment tests", label.Error, err)
		return
	}
	submission, err := runData.RecordResults(ctx, wh.db, results)
	if err != nil {
		logger.Error("failed to record assignment result", label.Error, err)
		return
	}
	// If we fail to get owners, we ignore sending on the stream.
	if userIDs, err := runData.GetOwners(wh.db); err == nil {
		// Note that streaming the submission as-is will send all grades
		// to all participants for a given group submission.
		wh.streams.Submission.SendTo(submission, userIDs...)
	}
}

// updateLastActivityDate sets a current date as a last activity date of the student
// on each new push to the student repository.
func (wh GitHubWebHook) updateLastActivityDate(ctx context.Context, course *qf.Course, repo *qf.Repository, login string) {
	logger := qlog.FromContext(ctx)
	userID := repo.GetUserID()
	if userID < 1 && repo.IsGroupRepo() {
		user, err := wh.db.GetUserByCourse(course, login)
		if err != nil {
			logger.Error("failed to find user", label.Error, err)
			return
		}
		userID = user.GetID()
	}
	// We want to fetch the original enrollment to ensure all Enrollment fields are set to correct values
	// to ensure gorm Select.Updates behave correctly.
	enrol, err := wh.db.GetEnrollmentByCourseAndUser(course.GetID(), userID)
	if err != nil {
		logger.Error("failed to find enrollment", label.Error, err)
		return
	}
	enrol.LastActivityDate = timestamppb.Now()

	if err := wh.db.UpdateEnrollment(enrol); err != nil {
		logger.Error("failed to update last activity", label.TargetUserID, userID, label.Error, err)
	}
}

// branchName returns the branch name from a push event ref.
func branchName(ref string) string {
	components := strings.Split(ref, "/")
	return components[len(components)-1]
}

// isDefaultBranch returns true if a push event is for a repository's default branch.
func isDefaultBranch(payload *github.PushEvent) bool {
	return strings.HasSuffix(payload.GetRef(), payload.GetRepo().GetDefaultBranch())
}
