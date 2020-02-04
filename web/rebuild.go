package web

import (
	"context"
	"fmt"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/ci"
)

// rebuildSubmission rebuilds the given assignment and submission.
func (s *AutograderService) rebuildSubmission(ctx context.Context, request *pb.LabRequest) error {
	submission, err := s.db.GetSubmission(&pb.Submission{ID: request.GetSubmissionID()})
	if err != nil {
		return err
	}
	assignment, err := s.db.GetAssignment(&pb.Assignment{ID: request.GetAssignmentID()})
	if err != nil {
		return err
	}
	course, err := s.db.GetCourse(assignment.GetCourseID(), false)
	if err != nil {
		return err
	}
	name := s.lookupName(submission)

	repoQuery := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		UserID:         submission.GetUserID(),
		RepoType:       pb.Repository_USER,
	}
	if assignment.IsGroupLab {
		repoQuery.GroupID = submission.GetGroupID()
		repoQuery.RepoType = pb.Repository_GROUP
	}
	repos, err := s.db.GetRepositories(repoQuery)
	if err != nil || len(repos) < 1 {
		return fmt.Errorf("could not find repository for user/group: %s, course: %s: %w", name, course.GetCode(), err)
	}
	repo := repos[0]

	s.logger.Debugf("Rebuilding submission %d for user(%d)/group(%d): %s, assignment: %+v, repo: %s",
		submission.GetID(), submission.GetUserID(), submission.GetGroupID(), name, assignment, repo.GetHTMLURL())
	runData := &ci.RunData{
		Course:     course,
		Assignment: assignment,
		Repo:       repo,
		CloneURL:   repo.GetHTMLURL(),
		CommitID:   submission.GetCommitHash(),
		JobOwner:   name,
	}
	ci.RunTests(s.logger, s.db, s.runner, runData)
	return nil
}

func (s *AutograderService) lookupName(submission *pb.Submission) string {
	if submission.GetGroupID() > 0 {
		group, _ := s.db.GetGroup(submission.GetGroupID())
		return group.GetName()
	}
	user, _ := s.db.GetUser(submission.GetUserID())
	return user.GetLogin()
}
