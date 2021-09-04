package web

import (
	"time"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/ci"
	"github.com/gosimple/slug"
	"golang.org/x/sync/errgroup"
)

var maxContainers = 10

// rebuildSubmission rebuilds the given assignment and submission.
func (s *AutograderService) rebuildSubmission(request *pb.RebuildRequest) (*pb.Submission, error) {
	submission, err := s.db.GetSubmission(&pb.Submission{ID: request.GetSubmissionID()})
	if err != nil {
		return nil, err
	}
	assignment, course, err := s.getAssignmentWithCourse(&pb.Assignment{ID: request.AssignmentID}, false)
	if err != nil {
		return nil, err
	}
	name := s.lookupName(submission)

	var repo *pb.Repository
	if assignment.IsGroupLab {
		s.logger.Debugf("Rebuilding submission %d for group(%d): %s, assignment: %+v, repo: %s",
			submission.GetID(), submission.GetGroupID(), name, assignment, repo.GetHTMLURL())
		repo, err = s.getGroupRepo(course, submission.GetGroupID())
	} else {
		s.logger.Debugf("Rebuilding submission %d for user(%d): %s, assignment: %+v, repo: %s",
			submission.GetID(), submission.GetUserID(), name, assignment, repo.GetHTMLURL())
		repo, err = s.getUserRepo(course, submission.GetUserID())
	}
	if err != nil {
		return nil, err
	}

	runData := &ci.RunData{
		Course:     course,
		Assignment: assignment,
		Repo:       repo,
		CommitID:   submission.GetCommitHash(),
		JobOwner:   slug.Make(name),
		Rebuild:    true,
	}
	ci.RunTests(s.logger, s.db, s.runner, runData)
	return s.db.GetSubmission(&pb.Submission{ID: request.GetSubmissionID()})
}

func (s *AutograderService) rebuildSubmissions(request *pb.AssignmentRequest) error {
	s.logger.Debugf("Running tests for all submissions for assignment ID %d of course ID %d\n", request.GetAssignmentID(), request.GetCourseID())
	start := time.Now()
	if _, err := s.db.GetAssignment(&pb.Assignment{ID: request.AssignmentID}); err != nil {
		return err
	}
	submissions, err := s.db.GetSubmissions(&pb.Submission{AssignmentID: request.AssignmentID})
	if err != nil {
		return err
	}

	limit := maxContainers
	for i := 0; i < len(submissions); i += maxContainers {
		var errgrp errgroup.Group

		if i+maxContainers > len(submissions) {
			limit = len(submissions) - i
		}
		for j := 0; j < limit; j++ {
			submission := submissions[i+j]
			rebuildRequest := &pb.RebuildRequest{
				AssignmentID: request.AssignmentID,
				SubmissionID: submission.ID}
			errgrp.Go(func() error {
				_, err := s.rebuildSubmission(rebuildRequest)
				return err
			})
		}
		err = errgrp.Wait()
		if err != nil {
			return err
		}
	}
	total := time.Since(start)
	s.logger.Debug("Finished running all tests, took ", total)
	return err
}

func (s *AutograderService) lookupName(submission *pb.Submission) string {
	if submission.GetGroupID() > 0 {
		group, _ := s.db.GetGroup(submission.GetGroupID())
		return group.GetName()
	}
	user, _ := s.db.GetUser(submission.GetUserID())
	return user.GetLogin()
}
