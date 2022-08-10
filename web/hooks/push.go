package hooks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/assignments"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"go.uber.org/zap"
	"gorm.io/gorm"
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

	course, err := wh.db.GetCourseByOrganizationID(repo.OrganizationID)
	if err != nil {
		wh.logger.Errorf("Failed to get course from database: %v", err)
		return
	}
	wh.logger.Debugf("For course(%d)=%v", course.GetID(), course.GetName())

	switch {
	case repo.IsTestsRepo():
		if !isDefaultBranch(payload) {
			wh.logger.Debugf("Ignoring push event for non-default branch: %s", payload.GetRef())
			return
		}
		// the push event is for the 'tests' repo, which means that we
		// should update the course data (assignments) in the database
		assignments.UpdateFromTestsRepo(wh.logger, wh.db, wh.scms, course)

	case repo.IsUserRepo():
		wh.logger.Debugf("Processing push event for user repo %s", payload.GetRepo().GetName())
		if !isDefaultBranch(payload) {
			wh.logger.Debugf("Ignoring push event for non-default branch: %s", payload.GetRef())
			return
		}
		wh.updateLastActivityDate(repo.UserID, course.ID)
		assignments := wh.extractAssignments(payload, course)
		for _, assignment := range assignments {
			if !assignment.IsGroupLab {
				// only run non-group assignments
				wh.runAssignmentTests(assignment, repo, course, payload)
			} else {
				wh.logger.Debugf("Ignoring assignment: %s, pushed to user repo: %s", assignment.GetName(), payload.GetRepo().GetName())
			}
		}

	case repo.IsGroupRepo():
		wh.logger.Debugf("Processing push event for group repo %s", payload.GetRepo().GetName())
		jobOwner, _, err := wh.db.GetUserByCourse(course, payload.GetSender().GetLogin())
		if err != nil {
			wh.logger.Errorf("Failed to find user %s in course %s: %v", payload.GetSender().GetLogin(), course.GetName(), err)
			return
		}
		wh.updateLastActivityDate(jobOwner.ID, course.ID)
		assignments := wh.extractAssignments(payload, course)
		for _, assignment := range assignments {
			if assignment.IsGroupLab {
				// only run group assignments
				results := wh.runAssignmentTests(assignment, repo, course, payload)
				if !isDefaultBranch(payload) && !assignment.GradedManually() {
					// Attempt to find the pull request for the branch, if it exists,
					// and then assign reviewers to it, if the branch task score is higher than the assignment score limit
					wh.handlePullRequestPush(payload, results, assignment, course, repo)
				}
			} else {
				wh.logger.Debugf("Ignoring assignment: %s, pushed to group repo: %s", assignment.GetName(), payload.GetRepo().GetName())
			}
		}
	default:
		wh.logger.Debug("Nothing to do for this push event")
	}
}

// handlePullRequestPush attempts to find a pull request associated with a non-default branch push event.
// If successful, it then finds the relevant task, and uses it to retrieve the relevant task score.
// If a passing score is reached, it assigns reviewers to the pull request.
// It also uses the test results and task to generate a feedback comment for the pull request.
func (wh GitHubWebHook) handlePullRequestPush(payload *github.PushEvent, results *score.Results, assignment *qf.Assignment, course *qf.Course, repo *qf.Repository) {
	wh.logger.Debugf("Attempting to find pull request for ref: %s, in repository: %s",
		payload.GetRef(), payload.GetRepo().GetFullName())

	pullRequest, taskName, err := wh.handlePullRequestPushPayload(payload)
	if err != nil {
		wh.logger.Errorf("Failed to retrieve pull request data from push payload: %v", err)
		return
	}
	taskSum := results.TaskSum(taskName)

	ctx := context.Background()

	sc, err := wh.scms.GetOrCreateSCM(ctx, wh.logger, course.OrganizationPath)
	if err != nil {
		wh.logger.Errorf("Failed to create SCM Client: %v", err)
		return
	}

	// We assign reviewers to a pull request when the tests associated with it score above the assignment score limit
	// We do not assign reviewers if the pull request has already been assigned reviewers
	scoreLimit := assignment.GetScoreLimit()
	if taskSum >= scoreLimit && !pullRequest.HasReviewers() {
		wh.logger.Debugf("Assigning reviewers to pull request #%d, in repository: %s", pullRequest.GetNumber(), repo.Name())
		if err := assignments.AssignReviewers(ctx, sc, wh.db, course, repo, pullRequest); err != nil {
			wh.logger.Errorf("Failed to assign reviewers to pull request: %v", err)
			return
		}
	}

	// Create a test results feedback comment on the pull request
	opt := &scm.IssueCommentOptions{
		Organization: course.GetOrganizationPath(),
		Repository:   repo.Name(),
		Body:         results.MarkdownComment(taskName, scoreLimit),
		Number:       int(pullRequest.GetNumber()),
	}
	wh.logger.Debugf("Creating feedback comment on pull request #%d, in repository: %s", pullRequest.GetNumber(), repo.Name())
	if !pullRequest.HasFeedbackComment() {
		commentID, err := sc.CreateIssueComment(ctx, opt)
		if err != nil {
			wh.logger.Errorf("Failed to create feedback comment for pull request #%d, in repository", pullRequest.GetNumber(), repo.Name())
			return
		}
		pullRequest.ScmCommentID = uint64(commentID)
		if err := wh.db.UpdatePullRequest(pullRequest); err != nil {
			wh.logger.Errorf("Failed to update pull request: %v", err)
			return
		}
	} else {
		opt.CommentID = int64(pullRequest.GetScmCommentID())
		if err := sc.UpdateIssueComment(ctx, opt); err != nil {
			wh.logger.Errorf("Failed to update feedback comment for pull request #%d, in repository", pullRequest.GetNumber(), repo.Name())
			return
		}
	}
	wh.logger.Debugf("Successfully handled push to pull request #%d, in repository: %s", pullRequest.GetNumber(), repo.Name())
}

// handlePullRequestPushPayload retrieves the pull request and task name associated with it from an event payload.
func (wh GitHubWebHook) handlePullRequestPushPayload(payload *github.PushEvent) (*qf.PullRequest, string, error) {
	pullRequest, err := wh.db.GetPullRequest(&qf.PullRequest{
		SourceBranch:    branchName(payload.GetRef()),
		ScmRepositoryID: uint64(payload.GetRepo().GetID()),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// This can happen if someone pushes to a branch group assignment, without having a PR created for it
			// If this happens, QF should not do anything
			return nil, "", fmt.Errorf("no pull request found for ref: %s", payload.GetRef())
		}
		return nil, "", fmt.Errorf("failed to get pull request from database: %v", err)
	}
	associatedTask, err := wh.getTask(pullRequest.GetTaskID())
	if err != nil {
		// A pull request should always have a task association
		// If not, something must have gone wrong elsewhere
		return nil, "", fmt.Errorf("failed to get task from the database: %w", err)
	}
	return pullRequest, associatedTask.GetName(), nil
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
func (wh GitHubWebHook) runAssignmentTests(assignment *qf.Assignment, repo *qf.Repository, course *qf.Course, payload *github.PushEvent) *score.Results {
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
		return nil
	}
	ctx, cancel := assignment.WithTimeout(ci.DefaultContainerTimeout)
	defer cancel()
	sc, err := wh.scms.SCMWithToken(ctx, wh.logger, course.OrganizationPath)
	if err != nil {
		wh.logger.Errorf("Failed to create scm client: %v", err)
		return nil
	}
	results, err := runData.RunTests(ctx, wh.logger, sc, wh.runner)
	if err != nil {
		wh.logger.Errorf("Failed to run tests for assignment %s for course %s: %v", assignment.Name, course.Name, err)
	}
	wh.logger.Debug("ci.RunTests", zap.Any("Results", qlog.IndentJson(results)))
	if _, err = runData.RecordResults(wh.logger, wh.db, results); err != nil {
		wh.logger.Error(err)
	}
	return results
}

// updateLastActivityDate sets a current date as a last activity date of the student
// on each new push to the student repository.
func (wh GitHubWebHook) updateLastActivityDate(userID, courseID uint64) {
	query := &qf.Enrollment{
		UserID:           userID,
		CourseID:         courseID,
		LastActivityDate: time.Now().Format("02 Jan"),
	}

	if err := wh.db.UpdateEnrollment(query); err != nil {
		wh.logger.Errorf("Failed to update the last activity date for user %d: %v", userID, err)
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
