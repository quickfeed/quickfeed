package hooks

import (
	"context"
	"strings"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/assignments"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

func (wh GitHubWebHook) handlePush(payload *github.PushEvent) {
	wh.logger.Debugf("Received push event for branch reference: %s (user's default branch: %s)",
		payload.GetRef(), payload.GetRepo().GetDefaultBranch())

	repo, err := wh.getRepository(payload.GetRepo().GetID())
	if err != nil {
		wh.logger.Errorf("Failed to get repository %s from database: %v", payload.GetRepo().GetFullName(), err)
		return
	}
	wh.logger.Debugf("Received push event for repository %v", repo)

	if !isDefaultBranch(payload) && !repo.IsGroupRepo() {
		wh.logger.Debugf("Ignoring push event for non-default branch: %s", payload.GetRef())
		return
	}

	course, err := wh.db.GetCourseByOrganizationID(repo.OrganizationID)
	if err != nil {
		wh.logger.Errorf("Failed to get course from database: %v", err)
		return
	}
	wh.logger.Debugf("For course(%d)=%v", course.GetID(), course.GetName())

	if repo.IsStudentRepo() {
		wh.updateLastActivityDate(course, repo, payload.GetSender().GetLogin())
	}

	ctx := context.Background()
	sc, err := wh.scmMgr.GetOrCreateSCM(ctx, wh.logger, course.GetOrganizationName())
	if err != nil {
		wh.logger.Errorf("Failed to get or create SCM Client: %v", err)
		return
	}

	switch {
	case repo.IsTestsRepo():
		// the push event is for the 'tests' repo, which means that we
		// should update the course data (assignments) in the database
		assignments.UpdateFromTestsRepo(wh.logger, wh.db, sc, course)

	case repo.IsAssignmentsRepo():
		// the push event is for the 'assignments' repo; we need to update the local working copy
		clonedAssignmentsRepo, err := sc.Clone(ctx, &scm.CloneOptions{
			Organization: course.GetOrganizationName(),
			Repository:   qf.AssignmentsRepo,
			DestDir:      course.CloneDir(),
		})
		if err != nil {
			wh.logger.Errorf("Failed to clone '%s' repository: %v", qf.AssignmentsRepo, err)
			return
		}
		wh.logger.Debugf("Successfully cloned assignments repository to: %s", clonedAssignmentsRepo)

	case repo.IsStudentRepo():
		wh.logger.Debugf("Processing push event for repo %s", payload.GetRepo().GetName())
		assignments := wh.extractAssignments(payload, course)
		for _, assignment := range assignments {
			wh.runAssignmentTests(sc, assignment, repo, course, payload)
		}

	default:
		wh.logger.Debug("Nothing to do for this push event")
	}
}

// extractAssignments extracts information from the push payload from github
// and determines the assignments that have been changed in this commit by
// querying the database based on the lab name.
func (wh GitHubWebHook) extractAssignments(payload *github.PushEvent, course *qf.Course) []*qf.Assignment {
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
			wh.logger.Errorf("Could not find assignment '%s' for course %d in database: %v", name, course.GetID(), err)
			continue
		}
		assignments = append(assignments, assignment)
	}
	return assignments
}

// runAssignmentTests runs the tests for the given assignment pushed to repo.
func (wh GitHubWebHook) runAssignmentTests(sc scm.SCM, assignment *qf.Assignment, repo *qf.Repository, course *qf.Course, payload *github.PushEvent) {
	runData := &ci.RunData{
		Course:     course,
		Assignment: assignment,
		Repo:       repo,
		BranchName: branchName(payload.GetRef()),
		CommitID:   payload.GetHeadCommit().GetID(),
		JobOwner:   payload.GetSender().GetLogin(),
	}
	if assignment.GradedManually() {
		wh.logger.Debugf("Assignment %s for course %s is manually reviewed", assignment.Name, course.Name)
		if _, err := runData.RecordResults(wh.logger, wh.db, nil); err != nil {
			wh.logger.Error(err)
		}
		return
	}
	ctx, cancel := assignment.WithTimeout(ci.DefaultContainerTimeout)
	defer cancel()
	results, err := runData.RunTests(ctx, wh.logger, sc, wh.runner)
	if err != nil {
		wh.logger.Error(err)
		return
	}
	if _, err = runData.RecordResults(wh.logger, wh.db, results); err != nil {
		wh.logger.Error(err)
		return
	}
	// Non-default branch indicates push to a group repo.
	if !isDefaultBranch(payload) {
		// Attempt to find the pull request for the branch, if it exists,
		// and then assign reviewers to it, if the branch task score is higher than the assignment score limit
		wh.handlePullRequestPush(payload, results, assignment, course, repo)
	}
}

// updateLastActivityDate sets a current date as a last activity date of the student
// on each new push to the student repository.
func (wh GitHubWebHook) updateLastActivityDate(course *qf.Course, repo *qf.Repository, login string) {
	query := &qf.Enrollment{
		UserID:           repo.UserID,
		CourseID:         course.ID,
		LastActivityDate: time.Now().Format("02 Jan"),
	}

	if repo.IsGroupRepo() {
		user, err := wh.db.GetUserByCourse(course, login)
		if err != nil {
			wh.logger.Errorf("Failed to find user %s in course %s: %v", login, course.GetName(), err)
			return
		}
		query.UserID = user.ID
	}

	if err := wh.db.UpdateEnrollment(query); err != nil {
		wh.logger.Errorf("Failed to update the last activity date for user %d: %v", query.UserID, err)
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
