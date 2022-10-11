package hooks

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/assignments"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"gorm.io/gorm"
)

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
	sc, err := wh.scmMgr.GetOrCreateSCM(ctx, wh.logger, course.OrganizationName)
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
		Organization: course.GetOrganizationName(),
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

func (wh GitHubWebHook) handlePullRequestReview(payload *github.PullRequestReviewEvent) {
	wh.logger.Debugf("Received review event for pull request #%d: %q in %s",
		payload.GetPullRequest().GetNumber(), payload.GetPullRequest().GetTitle(), payload.GetRepo().GetFullName())

	// Currently, QF only needs to do something if the PR is approved
	if payload.GetReview().GetState() != "approved" {
		wh.logger.Debug("Ignoring pull request review event for non-approved review")
		return
	}
	// We make sure that the pull request is one that QF has a data record of
	pullRequest, err := wh.db.GetPullRequest(&qf.PullRequest{
		ScmRepositoryID: uint64(payload.GetRepo().GetID()),
		Number:          uint64(payload.GetPullRequest().GetNumber()),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wh.logger.Debugf("Ignoring review event for unknown pull request #%d in %s",
				payload.GetPullRequest().GetNumber(), payload.GetRepo().GetFullName())
		} else {
			wh.logger.Errorf("Failed to get pull request from database %v", err)
		}
		return
	}

	course, err := wh.db.GetCourseByOrganizationID(uint64(payload.GetOrganization().GetID()))
	if err != nil {
		wh.logger.Errorf("Failed to get course from database: %v", err)
		return
	}
	user, err := wh.db.GetUserByRemoteIdentity(&qf.RemoteIdentity{
		RemoteID: uint64(payload.GetSender().GetID()),
		Provider: "github",
	})
	if err != nil {
		wh.logger.Errorf("Failed to get user from database: %v", err)
		return
	}
	reviewer, err := wh.db.GetEnrollmentByCourseAndUser(course.GetID(), user.GetID())
	if err != nil {
		wh.logger.Errorf("Failed to get reviewer from database: %v", err)
		return
	}

	// If we reach here the pull request already has an approved state. However, only if the
	// review is from a course teacher, do we mark the pull request as approved for QuickFeed.
	if reviewer.IsTeacher() {
		pullRequest.SetApproved()
		if err := wh.db.UpdatePullRequest(pullRequest); err != nil {
			wh.logger.Errorf("Failed to update pull request in database: %v", err)
			return
		}
		wh.logger.Debugf("Pull request successfully approved for repository: %s", payload.GetRepo().GetFullName())
	}
}

func (wh GitHubWebHook) handlePullRequestOpened(payload *github.PullRequestEvent) {
	wh.logger.Debugf("Received pull request opened event for repository: %s", payload.GetRepo().GetFullName())

	repo, err := wh.getRepositoryWithIssues(payload.GetRepo().GetID())
	if err != nil {
		wh.logger.Errorf("Failed to get repository %s from database: %v", payload.GetRepo().GetFullName(), err)
		return
	}
	if !repo.IsGroupRepo() {
		wh.logger.Debugf("Ignoring pull request opened event for non-group repository: %s", payload.GetRepo().GetFullName())
		return
	}
	issue, err := findIssue(payload.GetPullRequest().GetBody(), repo.GetIssues())
	if err != nil {
		wh.logger.Errorf("Failed to find associated issue in pull request: %v", err)
		return
	}
	wh.createPullRequest(payload, issue)
}

func (wh GitHubWebHook) handlePullRequestClosed(payload *github.PullRequestEvent) {
	wh.logger.Debugf("Received pull request closed event for repository: %s", payload.GetRepo().GetFullName())

	if !payload.PullRequest.GetMerged() {
		wh.logger.Debugf("Ignoring pull request closed event for unmerged pull request #%d, in %s",
			payload.GetPullRequest().GetNumber(), payload.GetRepo().GetFullName())
		return
	}

	pullRequest, err := wh.db.GetPullRequest(&qf.PullRequest{
		ScmRepositoryID: uint64(payload.GetRepo().GetID()),
		Number:          uint64(payload.GetPullRequest().GetNumber()),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wh.logger.Debugf("Ignoring pull request closed event for non-QuickFeed managed pull request #%d, in %s",
				payload.GetPullRequest().GetNumber(), payload.GetRepo().GetFullName())
		} else {
			wh.logger.Errorf("Failed to get pull request from database: %v", err)
		}
		return
	}

	if err := wh.db.HandleMergingPR(pullRequest); err != nil {
		wh.logger.Errorf("Failed to delete pull request from database %v", err)
		return
	}
	wh.logger.Debugf("Pull request successfully closed for repository: %s", payload.GetRepo().GetFullName())
}

// createPullRequest creates a new pull request record from a pull request opened event.
// When created, it is initially in the "draft" stage, signaling that it is not yet ready for review.
func (wh GitHubWebHook) createPullRequest(payload *github.PullRequestEvent, associatedIssue *qf.Issue) {
	wh.logger.Debugf("Creating pull request (issue #%d) for repository: %s",
		associatedIssue.GetIssueNumber(), payload.GetRepo().GetFullName())

	associatedTask, err := wh.getTask(associatedIssue.GetTaskID())
	if err != nil {
		wh.logger.Errorf("Failed to get task from database: %v", err)
		return
	}

	user, err := wh.db.GetUserByRemoteIdentity(&qf.RemoteIdentity{
		RemoteID: uint64(payload.GetSender().GetID()),
		Provider: "github",
	})
	if err != nil {
		wh.logger.Errorf("Failed to get user from database: %v", err)
		return
	}

	pullRequest := &qf.PullRequest{
		ScmRepositoryID: uint64(payload.GetRepo().GetID()),
		TaskID:          associatedTask.GetID(),
		IssueID:         associatedIssue.GetID(),
		UserID:          user.GetID(),
		SourceBranch:    payload.GetPullRequest().GetHead().GetRef(),
		Number:          uint64(payload.GetNumber()),
	}
	if err = wh.db.CreatePullRequest(pullRequest); err != nil {
		wh.logger.Errorf("Failed to create pull request data-record for repository %s: %v", payload.GetRepo().GetFullName(), err)
		return
	}
	wh.logger.Debugf("Pull request successfully created for repository: %s", payload.GetRepo().GetFullName())
}

var issueRegExp = regexp.MustCompile(`(?m)((?i:fixes|closes|resolves)\s#(\d+))$`)

// findIssue returns the issue from the provided list that match the pull request body.
// Only a single issue can be linked to a pull request. The body should contain one of the
// strings "Fixes #<issue number>" or "Closes #<issue number>" or "Resolves #<issue number>".
// The issue number should not be followed by any other characters.
func findIssue(body string, issues []*qf.Issue) (*qf.Issue, error) {
	if count := strings.Count(body, "#"); count > 1 {
		return nil, errors.New("more than one '#' character in pull request body")
	}
	if issueRegExp.MatchString(body) {
		issue := issueRegExp.ReplaceAllString(body, "$2")
		// ignore error since regular expression ensure it is a positive number
		issueNum, _ := strconv.ParseUint(issue, 10, 64)
		for _, issue := range issues {
			if issue.IssueNumber == issueNum {
				return issue, nil
			}
		}
		return nil, fmt.Errorf("unknown issue #%d", issueNum)
	}
	return nil, errors.New("no issue found in pull request body")
}
