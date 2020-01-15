package web

import (
	"context"
	"fmt"

	pb "github.com/autograde/aguis/ag"
)

// rebuildSubmission rebuilds the given lab assignment and submission.
func (s *AutograderService) rebuildSubmission(ctx context.Context, request *pb.LabRequest) error {
	lab, err := s.db.GetAssignment(&pb.Assignment{ID: request.GetAssignmentID()})
	if err != nil {
		return err
	}
	course, err := s.db.GetCourse(lab.GetCourseID(), false)
	if err != nil {
		return err
	}
	submission, err := s.db.GetSubmission(&pb.Submission{
		ID:           request.GetSubmissionID(),
		AssignmentID: request.GetAssignmentID(),
	})
	if err != nil {
		return err
	}

	repoQuery := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		UserID:         submission.GetUserID(), // defaults to 0 if not set
		RepoType:       pb.Repository_USER,
	}
	if lab.IsGroupLab {
		repoQuery.GroupID = submission.GetGroupID()
		repoQuery.RepoType = pb.Repository_GROUP
	}
	repos, err := s.db.GetRepositories(repoQuery)
	if err != nil {
		return err
	}

	// TODO(vera): debugging, to be removed
	s.logger.Debugf("Starting rebuild for user %d or group %d, lab %+v", submission.GetUserID(), submission.GetGroupID(), lab)

	// it is possible to have duplicate records for the same user repo because there were no database constraints
	// it is fixed for new records, but can be relevant for older database records
	// that's why we allow len(repos) be > 1 and just use the first found record
	if len(repos) < 1 {
		return fmt.Errorf("failed to get user repository for the submission")
	}
	repo := repos[0]

	s.logger.Info("Rebuilding user submission: repo url is: ", repo.GetHTMLURL())

	nameTag := s.makeContainerTag(submission)

	runTests(s.logger, s.db, s.runner, repo, repo.GetHTMLURL(), submission.GetCommitHash(), "ci/scripts", lab.GetID(), nameTag)
	return nil
}

func (s *AutograderService) makeContainerTag(submission *pb.Submission) string {
	if submission.GetGroupID() > 0 {
		group, _ := s.db.GetGroup(submission.GetGroupID())
		return group.GetName()
	}
	user, _ := s.db.GetUser(submission.GetUserID())
	return user.GetLogin()
}
