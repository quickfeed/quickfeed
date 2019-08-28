package web

import (
	"context"
	"fmt"

	"github.com/autograde/aguis/ag"
	pb "github.com/autograde/aguis/ag"
)

func (s *AutograderService) rebuildSubmission(ctx context.Context, submissionID uint64) error {

	fmt.Println("Received rebuild event, submission ID: ", submissionID)
	// get the latest submission from the database
	submission, err := s.db.GetSubmission(&ag.Submission{ID: submissionID})
	if err != nil {
		return err
	}

	fmt.Println("Rebuild: got submission for user ID ", submission.GetUserID())

	// get user repo for the student
	repos, err := s.db.GetRepositories(&ag.Repository{UserID: submission.GetUserID(), RepoType: pb.Repository_USER})
	if err != nil {
		return err
	}
	if len(repos) < 1 {
		s.logger.Error(len(repos), " user repositories found for user ", submission.GetUserID())
		return fmt.Errorf("Failed to get user repository for the submission")
	}

	repo := repos[0]

	fmt.Println("Starting rebuild: repo url is: ", repo.GetHTMLURL(), ", commit hach is: ", submission.GetCommitHash())

	runTests(s.logger, s.db, s.runner, repo, repo.GetHTMLURL(), submission.GetCommitHash(), "ci/scripts")

	return nil
}
