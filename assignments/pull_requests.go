package assignments

import (
	"context"
	"errors"
	"strconv"
	"strings"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/scm"
	"github.com/google/go-github/v35/github"
	"go.uber.org/zap"
)

var ErrInvalidBody = errors.New("invalid pull request body")

// getLinkedIssue returns the issue number from a pull requests body.
// E.g. 30, from the body "Fixes #30".
func getLinkedIssue(body string) (uint64, error) {
	if count := strings.Count(body, "#"); count != 1 {
		return 0, ErrInvalidBody
	}
	_, numString, _ := strings.Cut(body, "#")
	issueNumber, err := strconv.Atoi(numString)
	if err != nil {
		return 0, ErrInvalidBody
	}
	return uint64(issueNumber), nil
}

func assignReviewers(ctx context.Context, sc scm.SCM, db database.Database, course *pb.Course, repo *pb.Repository, pullRequestNumber int) error {
	teachers, err := db.GetCourseTeachers(course)
	if err != nil {
		return err
	}
	reviewers := []string{}
	for _, teacher := range teachers {
		reviewers = append(reviewers, teacher.GetLogin())
	}

	opt := &scm.RequestReviewersOptions{
		Organization: course.GetOrganizationPath(),
		Repository:   repo.Name(),
		Number:       pullRequestNumber,
		Reviewers:    reviewers,
	}

	if err := sc.RequestReviewers(ctx, opt); err != nil {
		return err
	}
	return nil
}

func CreatePullRequest(db database.Database, logger *zap.SugaredLogger, course *pb.Course, repo *pb.Repository, payload *github.PullRequestEvent) {
	scm, err := scm.NewSCMClient(logger, course.GetProvider(), course.GetAccessToken())
	if err != nil {
		logger.Errorf("Failed to create SCM Client: %v", err)
		return
	}
	ctx := context.Background()

	issueNumber, err := getLinkedIssue(payload.GetPullRequest().GetBody())
	if err != nil {
		logger.Debugf("Failed to get issue number from pull request body: %v, in repository %s", err, payload.GetRepo().GetFullName())
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
		logger.Debugf("Ignoring pull request opened event for: %s, since linked issue is not managed", payload.GetRepo().GetFullName())
		return
	}

	pullRequest := &pb.PullRequest{
		PullRequestID: uint64(payload.GetPullRequest().GetID()),
		IssueID:       associatedIssue.GetID(),
		// TODO(Espeland): Probably a way of approved being automatically set to false
		Approved: false,
	}

	if err = db.CreatePullRequest(pullRequest); err != nil {
		logger.Errorf("Failed to create pull request record for repository %s: %v", payload.GetRepo().GetFullName(), err)
		return
	}

	if err = assignReviewers(ctx, scm, db, course, repo, payload.GetNumber()); err != nil {
		logger.Errorf("Failed to assign reviewers to pull request: %v", err)
		return
	}

	logger.Debugf("Pull request successfully created for repository: %s", payload.GetRepo().GetFullName())
}
