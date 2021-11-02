package web

import (
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/ci"
	"github.com/gosimple/slug"
)

const maxContainers = 10

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
	if _, err := s.db.GetAssignment(&pb.Assignment{ID: request.AssignmentID}); err != nil {
		return err
	}
	submissions, err := s.db.GetSubmissions(&pb.Submission{AssignmentID: request.AssignmentID})
	if err != nil {
		return err
	}
	s.logger.Debugf("Rebuilding all submissions for assignment %d for course %d\n", request.GetAssignmentID(), request.GetCourseID())
	start := time.Now()

	// counting semaphore: limit concurrent rebuilding to maxContainers
	sem := make(chan struct{}, maxContainers)
	errCnt := int32(0)
	var wg sync.WaitGroup
	wg.Add(len(submissions))
	for _, submission := range submissions {
		rebuildReq := &pb.RebuildRequest{
			AssignmentID: request.AssignmentID,
			SubmissionID: submission.ID,
		}
		// the counting semaphore limits concurrency to maxContainers
		go func() {
			sem <- struct{}{} // acquire semaphore
			_, err := s.rebuildSubmission(rebuildReq)
			if err != nil {
				atomic.AddInt32(&errCnt, 1)
				s.logger.Errorf("Failed to rebuild submission ID %d: %v\n", rebuildReq.GetSubmissionID(), err)
			}
			<-sem // release semaphore
			wg.Done()
		}()
	}
	// wait for all submissions to finish rebuilding
	wg.Wait()
	close(sem)

	s.logger.Debugf("Rebuilt %d submissions in %v (failed: %d)",
		len(submissions), time.Since(start), errCnt)
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
