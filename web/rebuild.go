package web

import (
	"context"
	"fmt"

	"github.com/autograde/aguis/ag"
	pb "github.com/autograde/aguis/ag"
)

func (s *AutograderService) rebuildSubmission(ctx context.Context, request *pb.LabRequest) error {

	lab, err := s.db.GetAssignment(&pb.Assignment{ID: request.GetAssignmentID()})
	if err != nil {
		return err
	}

	course, err := s.db.GetCourse(lab.GetCourseID(), false)
	if err != nil {
		return err
	}

	submission, err := s.db.GetSubmission(&pb.Submission{ID: request.GetAssignmentID(), AssignmentID: request.GetAssignmentID()})
	if err != nil {
		return err
	}

	repos := make([]*pb.Repository, 0)
	if lab.IsGroupLab {
		repos, err = s.db.GetRepositories(&ag.Repository{
			OrganizationID: course.GetOrganizationID(),
			GroupID:        submission.GetGroupID(),
			RepoType:       pb.Repository_GROUP})
	} else {
		repos, err = s.db.GetRepositories(&ag.Repository{
			OrganizationID: course.GetOrganizationID(),
			UserID:         submission.GetUserID(),
			RepoType:       pb.Repository_USER})
	}
	if err != nil {
		return err
	}

	// TODO(vera): debugging, to be removed
	s.logger.Debugf("Starting rebuild for user %d or group %d, lab %+v", submission.GetUserID(), submission.GetGroupID(), lab)

	// it is possible to have duplicate records for the same user repo because there were no database constraints
	// it is fixed for new records, but can be relevant for older database records
	// that's why we allow len(repos) be > 1 and just use the first found record
	if len(repos) < 1 {
		return fmt.Errorf("Failed to get user repository for the submission")
	}
	repo := repos[0]

	s.logger.Info("Rebuilding user submission: repo url is: ", repo.GetHTMLURL())

	runTests(s.logger, s.db, s.runner, repo, repo.GetHTMLURL(), submission.GetCommitHash(), "ci/scripts", lab.GetID())

	return nil
}
