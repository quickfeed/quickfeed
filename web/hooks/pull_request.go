package hooks

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/google/go-github/v35/github"
	"gorm.io/gorm"
)

func (wh GitHubWebHook) handlePullRequestReview(payload *github.PullRequestReviewEvent) {
	wh.logger.Debugf("Received review event for pull request #%d: %q in %s",
		payload.GetPullRequest().GetNumber(), payload.GetPullRequest().GetTitle(), payload.GetRepo().GetFullName())

	// Currently, QF only needs to do something if the PR is approved
	if payload.GetReview().GetState() != "approved" {
		wh.logger.Debug("Ignoring pull request review event for non-approved review")
		return
	}
	// We make sure that the pull request is one that QF has a data record of
	pullRequest, err := wh.db.GetPullRequest(&pb.PullRequest{
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
	user, err := wh.db.GetUserByRemoteIdentity(&pb.RemoteIdentity{
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

	// Only if the review is from a course teacher, do we set the pull request to approved
	// We do not check whether the review itself is "approved" here,
	// given that we earlier in the method discard all events that are not "approved".
	// If this method is to handle different types states, this check must be moved to here.
	if reviewer.IsTeacher() {
		pullRequest.SetApproved()
		wh.db.UpdatePullRequest(pullRequest)
		wh.logger.Debugf("Pull request successfully approved for repository: %s", payload.GetRepo().GetFullName())
	}
}

func (wh GitHubWebHook) handlePullRequestOpened(payload *github.PullRequestEvent) {
	wh.logger.Debugf("Received pull request opened event for repository: %s, in organization: %s",
		payload.GetRepo().GetName(), payload.GetOrganization().GetLogin())

	repos, err := wh.db.GetRepositoriesWithIssues(&pb.Repository{RepositoryID: uint64(payload.GetRepo().GetID())})
	if err != nil {
		wh.logger.Errorf("Failed to get repository by remote ID %d from database: %v", payload.GetRepo().GetID(), err)
		return
	}
	if len(repos) != 1 {
		wh.logger.Debugf("Ignoring pull request opened event for unknown repository: %s", payload.GetRepo().GetFullName())
		return
	}
	repo := repos[0]
	if !repo.IsGroupRepo() {
		wh.logger.Debugf("Ignoring pull request opened event for non-group repository: %s", payload.GetRepo().GetFullName())
		return
	}
	wh.createPullRequest(payload, repo)
}

func (wh GitHubWebHook) handlePullRequestClosed(payload *github.PullRequestEvent) {
	wh.logger.Debugf("Received pull request closed event for repository: %s, in organization: %s",
		payload.GetRepo().GetName(), payload.GetOrganization().GetLogin())

	if !payload.PullRequest.GetMerged() {
		wh.logger.Debugf("Ignoring pull request closed event for non-merged pull request #%d, in %s",
			payload.GetPullRequest().GetNumber(), payload.GetRepo().GetFullName())
		return
	}

	pullRequest, err := wh.db.GetPullRequest(&pb.PullRequest{
		ScmRepositoryID: uint64(payload.GetRepo().GetID()),
		Number:          uint64(payload.GetPullRequest().GetNumber()),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wh.logger.Debugf("Ignoring pull request closed event for non-managed pull request #%d, in %s",
				payload.GetPullRequest().GetNumber(), payload.GetRepo().GetFullName())
		} else {
			wh.logger.Errorf("Failed to get pull request from database: %v", err)
		}
		return
	}

	if err := wh.db.HandleMergingPR(pullRequest); err != nil {
		wh.logger.Errorf("Failed to delete pull request from database %v", err)
	}
	wh.logger.Debugf("Pull request successfully closed for repository: %s", payload.GetRepo().GetFullName())
}

// createPullRequest creates a new pull request record from a pull request opened event.
// When created, it is initially in the "draft" stage, signaling that it is not yet ready for review.
func (wh GitHubWebHook) createPullRequest(payload *github.PullRequestEvent, repo *pb.Repository) {
	wh.logger.Debugf("Attempting to create pull request for repository: %s", payload.GetRepo().GetFullName())
	issueNumber, err := getLinkedIssue(payload.GetPullRequest().GetBody())
	if err != nil {
		wh.logger.Debugf("Failed to get issue number from pull request body: %v, in repository %s", err, payload.GetRepo().GetFullName())
		return
	}
	var associatedIssue *pb.Issue = nil
	for _, issue := range repo.Issues {
		if issue.IssueNumber == issueNumber {
			associatedIssue = issue
			break
		}
	}
	if associatedIssue == nil {
		wh.logger.Debugf("Ignoring pull request opened event for: %s, found no repository issue with number: %d", payload.GetRepo().GetFullName(), issueNumber)
		return
	}

	tasks, err := wh.db.GetTasks(&pb.Task{ID: associatedIssue.GetTaskID()})
	if err != nil {
		wh.logger.Errorf("Failed to get task from the database: %v", err)
		return
	}
	if len(tasks) != 1 {
		// This should never happen
		wh.logger.Errorf("Got an unexpected number of tasks: %d", len(tasks))
		return
	}
	associatedTask := tasks[0]

	user, err := wh.db.GetUserByRemoteIdentity(&pb.RemoteIdentity{
		RemoteID: uint64(payload.GetSender().GetID()),
		Provider: "github",
	})
	if err != nil {
		wh.logger.Errorf("Failed to get user from database: %v", err)
		return
	}

	pullRequest := &pb.PullRequest{
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

// TODO(Espeland): This function would probably be best implemented using a regular expression search.
// See: https://docs.github.com/es/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue for patterns.
// GitHub also supports linking multiple issues. I do not think this is a feature we can/need to support atm.
// Currently, creating a pull request in a QF context certainly relies entirely on there only being one linked issue.

// getLinkedIssue returns the issue number from a pull requests body.
// E.g. 30, from the body "Fixes #30".
// It expects only one '#' character, that should be followed only by number characters.
// I.e. it would return an error for the body "Fixes #30 task-hello_world".
func getLinkedIssue(body string) (uint64, error) {
	if count := strings.Count(body, "#"); count != 1 {
		return 0, errors.New("pull request body does not contain exactly one '#' character")
	}
	subStrings := strings.Split(body, "#")
	issueNumber, err := strconv.Atoi(subStrings[len(subStrings)-1])
	if err != nil {
		return 0, fmt.Errorf("failed to parse issue number from pull request body: %w", err)
	}
	return uint64(issueNumber), nil
}