package hooks

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/assignments"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"gorm.io/gorm"
)

// handlePullRequestPush attempts to find a pull request associated with a non-default branch push event.
// If successful, it then finds the relevant task, and uses it to retrieve the relevant task score.
// If a passing score is reached, it assigns reviewers to the pull request.
// It also uses the test results and task to generate a feedback comment for the pull request.
func (wh GitHubWebHook) handlePullRequestPush(ctx context.Context, scmClient scm.SCM, payload *github.PushEvent, results *score.Results, rd *ci.RunData) {
	logger := qlog.FromContext(ctx)
	logger.Debug("finding pull request for push")

	pullRequest, err := wh.getPullRequest(payload)
	if err != nil {
		logger.Error("failed to retrieve pull request for push", label.Error, err)
		return
	}
	task, err := wh.getTask(pullRequest.GetTaskID())
	if err != nil {
		logger.Error("failed to get pull request task", label.Error, err)
		return
	}
	taskSum := results.TaskSum(task.GetName())

	repoName := rd.Repo.Name()
	prNumber := pullRequest.GetNumber()
	// Scope the remaining records to the pull request we resolved.
	ctx, logger = qlog.WithLogger(ctx, label.PullRequest, prNumber)
	// We assign reviewers to a pull request when the tests associated with it score above the assignment score limit
	// We do not assign reviewers if the pull request has already been assigned reviewers
	scoreLimit := rd.Assignment.GetScoreLimit()
	if taskSum >= scoreLimit && !pullRequest.HasReviewers() {
		logger.Debug("assigning pull request reviewers")
		if err := assignments.AssignReviewers(ctx, scmClient, wh.db, rd.Course, rd.Repo, pullRequest); err != nil {
			logger.Error("failed to assign pull request reviewers", label.Error, err)
			return
		}
	}

	// Create a test results feedback comment on the pull request
	opt := &scm.IssueCommentOptions{
		Organization: rd.Course.GetScmOrganizationName(),
		Repository:   repoName,
		Body:         results.MarkdownComment(task.GetName(), scoreLimit),
		Number:       int(prNumber),
	}
	logger.Debug("creating pull request feedback")
	if !pullRequest.HasFeedbackComment() {
		commentID, err := scmClient.CreateIssueComment(ctx, opt)
		if err != nil {
			logger.Error("failed to create pull request feedback", label.Error, err)
			return
		}
		pullRequest.ScmCommentID = uint64(commentID)
		if err := wh.db.UpdatePullRequest(pullRequest); err != nil {
			logger.Error("failed to store pull request feedback", label.Error, err)
			return
		}
	} else {
		opt.CommentID = int64(pullRequest.GetScmCommentID())
		if err := scmClient.UpdateIssueComment(ctx, opt); err != nil {
			logger.Error("failed to update pull request feedback", label.Error, err)
			return
		}
	}
	logger.Debug("handled pull request push")
}

// getPullRequest retrieves the pull request from the database based on the push event payload.
func (wh GitHubWebHook) getPullRequest(payload *github.PushEvent) (*qf.PullRequest, error) {
	pullRequest, err := wh.db.GetPullRequest(&qf.PullRequest{
		SourceBranch:    branchName(payload.GetRef()),
		ScmRepositoryID: uint64(payload.GetRepo().GetID()),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// This can happen if someone pushes to a branch group assignment, without having a PR created for it
			// If this happens, QF should not do anything
			return nil, fmt.Errorf("no pull request found for ref: %s", payload.GetRef())
		}
		return nil, fmt.Errorf("failed to get pull request from database: %w", err)
	}
	return pullRequest, nil
}

func (wh GitHubWebHook) handlePullRequestReview(ctx context.Context, payload *github.PullRequestReviewEvent) {
	ctx, logger := qlog.WithLogger(ctx,
		label.Repository, payload.GetRepo().GetName(),
		label.PullRequest, payload.GetPullRequest().GetNumber(),
		label.User, payload.GetSender().GetLogin(),
	)
	logger.Debug("received pull request review", "title", payload.GetPullRequest().GetTitle())

	// Currently, QF only needs to do something if the PR is approved
	if payload.GetReview().GetState() != "approved" {
		logger.Debug("ignoring pull request review event for non-approved review")
		return
	}
	// We make sure that the pull request is one that QF has a data record of
	pullRequest, err := wh.db.GetPullRequest(&qf.PullRequest{
		ScmRepositoryID: uint64(payload.GetRepo().GetID()),
		Number:          uint64(payload.GetPullRequest().GetNumber()),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Debug("ignoring review for unknown pull request")
		} else {
			logger.Error("failed to get pull request", label.Error, err)
		}
		return
	}

	course, err := wh.db.GetCourseByOrganizationID(uint64(payload.GetOrganization().GetID()))
	if err != nil {
		logger.Error("failed to get course", label.Error, err)
		return
	}
	logger = logger.With(label.CourseID, course.GetID(), label.CourseCode, course.GetCode())
	user, err := wh.db.GetUserByRemoteIdentity(uint64(payload.GetSender().GetID()))
	if err != nil {
		logger.Error("failed to get review user", label.Error, err)
		return
	}
	reviewer, err := wh.db.GetEnrollmentByCourseAndUser(course.GetID(), user.GetID())
	if err != nil {
		logger.Error("failed to get reviewer enrollment", label.Error, err)
		return
	}

	// If we reach here the pull request already has an approved state. However, only if the
	// review is from a course teacher, do we mark the pull request as approved for QuickFeed.
	if reviewer.IsTeacher() {
		pullRequest.SetApproved()
		if err := wh.db.UpdatePullRequest(pullRequest); err != nil {
			logger.Error("failed to approve pull request", label.Error, err)
			return
		}
		logger.Debug("approved pull request")
	}
}

func (wh GitHubWebHook) handlePullRequestOpened(ctx context.Context, payload *github.PullRequestEvent) {
	ctx, logger := qlog.WithLogger(ctx,
		label.Repository, payload.GetRepo().GetName(),
		label.PullRequest, payload.GetPullRequest().GetNumber(),
		label.User, payload.GetSender().GetLogin(),
	)
	logger.Debug("received pull request opened event")

	repo, err := wh.getRepositoryWithIssues(payload.GetRepo().GetID())
	if err != nil {
		logger.Error("failed to get pull request repository", label.Error, err)
		return
	}
	course, err := wh.db.GetCourseByOrganizationID(repo.GetScmOrganizationID())
	if err != nil {
		logger.Error("failed to get pull request course", label.Error, err)
		return
	}
	ctx, logger = qlog.WithLogger(ctx, label.CourseID, course.GetID(), label.CourseCode, course.GetCode(), label.RepositoryType, repo.GetRepoType().String())
	if !repo.IsGroupRepo() {
		logger.Debug("ignoring pull request for non-group repository")
		return
	}
	issue, err := findIssue(payload.GetPullRequest().GetBody(), repo.GetIssues())
	if err != nil {
		logger.Error("failed to find pull request issue", label.Error, err)
		return
	}
	wh.createPullRequest(ctx, payload, issue)
}

func (wh GitHubWebHook) handlePullRequestClosed(ctx context.Context, payload *github.PullRequestEvent) {
	ctx, logger := qlog.WithLogger(ctx,
		label.Repository, payload.GetRepo().GetName(),
		label.PullRequest, payload.GetPullRequest().GetNumber(),
		label.User, payload.GetSender().GetLogin(),
	)
	logger.Debug("received pull request closed event")
	repo, err := wh.getRepository(payload.GetRepo().GetID())
	if err != nil {
		logger.Error("failed to get pull request repository", label.Error, err)
		return
	}
	course, err := wh.db.GetCourseByOrganizationID(repo.GetScmOrganizationID())
	if err != nil {
		logger.Error("failed to get pull request course", label.Error, err)
		return
	}
	logger = logger.With(label.CourseID, course.GetID(), label.CourseCode, course.GetCode(), label.RepositoryType, repo.GetRepoType().String())

	if !payload.PullRequest.GetMerged() {
		logger.Debug("ignoring unmerged pull request")
		return
	}

	pullRequest, err := wh.db.GetPullRequest(&qf.PullRequest{
		ScmRepositoryID: uint64(payload.GetRepo().GetID()),
		Number:          uint64(payload.GetPullRequest().GetNumber()),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Debug("ignoring unmanaged pull request")
		} else {
			logger.Error("failed to get pull request", label.Error, err)
		}
		return
	}

	if err := wh.db.HandleMergingPR(pullRequest); err != nil {
		logger.Error("failed to handle merged pull request", label.Error, err)
		return
	}
	logger.Debug("closed pull request")
}

// createPullRequest creates a new pull request record from a pull request opened event.
// When created, it is initially in the "draft" stage, signaling that it is not yet ready for review.
func (wh GitHubWebHook) createPullRequest(ctx context.Context, payload *github.PullRequestEvent, associatedIssue *qf.Issue) {
	logger := qlog.FromContext(ctx)
	logger.Debug("creating pull request record", "issue", associatedIssue.GetScmIssueNumber())

	associatedTask, err := wh.getTask(associatedIssue.GetTaskID())
	if err != nil {
		logger.Error("failed to get pull request task", label.Error, err)
		return
	}

	user, err := wh.db.GetUserByRemoteIdentity(uint64(payload.GetSender().GetID()))
	if err != nil {
		logger.Error("failed to get pull request user", label.Error, err)
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
		logger.Error("failed to create pull request record", label.Error, err)
		return
	}
	logger.Debug("created pull request record")
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
			if issue.GetScmIssueNumber() == issueNum {
				return issue, nil
			}
		}
		return nil, fmt.Errorf("unknown issue #%d", issueNum)
	}
	return nil, errors.New("no issue found in pull request body")
}
