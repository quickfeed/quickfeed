package web

import (
	"context"
	"fmt"

	"github.com/autograde/aguis/ag"
	pb "github.com/autograde/aguis/ag"
)

func (s *AutograderService) rebuildSubmission(ctx context.Context, submissionID uint64) error {

	// get the latest submission from the database
	submission, err := s.db.GetSubmission(&ag.Submission{ID: submissionID})
	if err != nil {
		return err
	}

	// get user repo for the student
	repos, err := s.db.GetRepositories(&ag.Repository{UserID: submission.GetUserID(), RepoType: pb.Repository_USER})
	if err != nil {
		return err
	}

	// it is possible to have duplicate records for the same user repo because there were no database constraints
	// it is fixed for new records, but can be relevant for older database records
	// that's why we allow len(repos) be > 1 and just use the first found record
	if len(repos) < 1 {
		return fmt.Errorf("Failed to get user repository for the submission")
	}
	repo := repos[0]

	s.logger.Info("Rebuilding user submission: repo url is: ", repo.GetHTMLURL())

	runTests(s.logger, s.db, s.runner, repo, repo.GetHTMLURL(), submission.GetCommitHash(), "ci/scripts")

	return nil
}
