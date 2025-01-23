package web

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/qf"
)

const maxContainers = 10

// rebuildSubmission rebuilds the given assignment and submission.
func (s *QuickFeedService) rebuildSubmission(request *qf.RebuildRequest) error {
	submission, err := s.db.GetSubmission(&qf.Submission{ID: request.GetSubmissionID()})
	if err != nil {
		return err
	}
	assignment, course, err := s.getAssignmentWithCourse(&qf.Assignment{ID: request.AssignmentID})
	if err != nil {
		return err
	}
	name := s.lookupName(submission)

	var repo *qf.Repository
	if assignment.IsGroupLab && submission.GetGroupID() > 0 {
		repo, err = s.getRepo(course, submission.GetGroupID(), qf.Repository_GROUP)
		s.logger.Debugf("Rebuilding submission %d for group(%d): %s, assignment: %+v, repo: %s",
			submission.GetID(), submission.GetGroupID(), name, assignment, repo.GetHTMLURL())
	} else {
		repo, err = s.getRepo(course, submission.GetUserID(), qf.Repository_USER)
		s.logger.Debugf("Rebuilding submission %d for user(%d): %s, assignment: %+v, repo: %s",
			submission.GetID(), submission.GetUserID(), name, assignment, repo.GetHTMLURL())
	}
	if err != nil {
		return err
	}

	runData := &ci.RunData{
		Course:     course,
		Assignment: assignment,
		Repo:       repo,
		CommitID:   submission.GetCommitHash(),
		JobOwner:   name,
		Rebuild:    true,
	}
	ctx, cancel := assignment.WithTimeout(ci.DefaultContainerTimeout)
	defer cancel()
	sc, err := s.getSCM(ctx, course.ScmOrganizationName)
	if err != nil {
		return err
	}
	results, err := runData.RunTests(ctx, s.logger, sc, s.runner)
	if err != nil {
		return err
	}
	submission, err = runData.RecordResults(s.logger, s.db, results)
	if err != nil {
		return fmt.Errorf("failed to record results for assignment %s for course %s: %w", assignment.Name, course.Name, err)
	}
	// If we fail to get owners, we ignore sending on the stream.
	if userIDs, err := runData.GetOwners(s.db); err == nil {
		// Note that streaming the submission as-is sends all grades
		// to all participants for a given group submission.
		s.streams.Submission.SendTo(submission, userIDs...)
	}
	return nil
}

func (s *QuickFeedService) rebuildSubmissions(request *qf.RebuildRequest) error {
	if _, err := s.db.GetAssignment(&qf.Assignment{ID: request.AssignmentID}); err != nil {
		return err
	}
	submissions, err := s.db.GetSubmissions(&qf.Submission{AssignmentID: request.AssignmentID})
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
		rebuildReq := &qf.RebuildRequest{
			AssignmentID: request.AssignmentID,
			SubmissionID: submission.GetID(),
		}
		// the counting semaphore limits concurrency to maxContainers
		go func() {
			sem <- struct{}{} // acquire semaphore
			err := s.rebuildSubmission(rebuildReq)
			if err != nil {
				atomic.AddInt32(&errCnt, 1)
				s.logger.Errorf("Failed to rebuild submission %d: %v\n", rebuildReq.GetSubmissionID(), err)
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

func (s *QuickFeedService) lookupName(submission *qf.Submission) string {
	if submission.GetGroupID() > 0 {
		group, _ := s.db.GetGroup(submission.GetGroupID())
		return group.GetName()
	}
	user, _ := s.db.GetUser(submission.GetUserID())
	return user.GetLogin()
}
