package web

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
)

const maxContainers = 10

const (
	// ownerLabel is the attribute holding the job owner: the student's or group's name.
	ownerLabel = "owner"
	// repositoryURLLabel is the attribute holding the repository's web URL.
	repositoryURLLabel = "repository_url"
)

// internalRebuildSubmission rebuilds the given assignment and submission.
func (s *QuickFeedService) internalRebuildSubmission(ctx context.Context, request *qf.RebuildRequest) error {
	submission, err := s.db.GetSubmission(&qf.Submission{ID: request.GetSubmissionID()})
	if err != nil {
		return err
	}
	assignment, err := s.db.GetAssignment(&qf.Assignment{ID: request.GetAssignmentID()})
	if err != nil {
		return err
	}
	course, err := s.db.GetCourse(assignment.GetCourseID())
	if err != nil {
		return err
	}
	name := s.lookupName(submission)

	var repo *qf.Repository
	if assignment.GetIsGroupLab() && submission.GetGroupID() > 0 {
		repo, err = s.getRepo(course, submission.GetGroupID(), qf.Repository_GROUP)
	} else {
		repo, err = s.getRepo(course, submission.GetUserID(), qf.Repository_USER)
	}
	if err != nil {
		return err
	}
	// Scope the rebuild; RunTests and RecordResults log under the same context.
	// The course ID comes from the request logger; see enrichRequestLogger.
	ctx, logger := qlog.WithCourseLog(ctx,
		course,
		label.Repository, repo.Name(),
		label.RepositoryType, repo.GetRepoType().String(),
		label.Assignment, assignment.GetName(),
		label.Commit, submission.GetCommitHash(),
		label.SubmissionID, submission.GetID(),
	)
	logger.Debug("rebuilding submission", ownerLabel, name, repositoryURLLabel, repo.GetHTMLURL())

	runData := &ci.RunData{
		Course:     course,
		Assignment: assignment,
		Repo:       repo,
		CommitID:   submission.GetCommitHash(),
		JobOwner:   name,
		Rebuild:    true,
	}
	ctx, cancel := assignment.WithTimeout(ctx, ci.DefaultContainerTimeout)
	defer cancel()
	sc, err := s.getSCM(ctx, course.GetScmOrganizationName())
	if err != nil {
		return err
	}
	results, err := runData.RunTests(ctx, sc, s.runner)
	if err != nil {
		return err
	}
	submission, err = runData.RecordResults(ctx, s.db, results)
	if err != nil {
		return fmt.Errorf("recording results for assignment %s for course %s: %w", assignment.GetName(), course.GetName(), err)
	}
	// If we fail to get owners, we ignore sending on the stream.
	if userIDs, err := runData.GetOwners(s.db); err == nil {
		// Note that streaming the submission as-is sends all grades
		// to all participants for a given group submission.
		s.streams.Submission.SendTo(submission, userIDs...)
	}
	return nil
}

func (s *QuickFeedService) internalRebuildAllSubmissions(ctx context.Context, request *qf.RebuildRequest) error {
	submissions, err := s.db.GetSubmissions(&qf.Submission{AssignmentID: request.GetAssignmentID()})
	if err != nil {
		return err
	}
	assignment, err := s.db.GetAssignment(&qf.Assignment{ID: request.GetAssignmentID()})
	if err != nil {
		return err
	}
	course, err := s.db.GetCourse(assignment.GetCourseID())
	if err != nil {
		return err
	}
	// logger is scoped to the course so this batch's own summary and failure
	// records land in the course log too, but ctx is left unscoped: each
	// rebuild below derives and attaches this same course scope for itself
	// (see internalRebuildSubmission), and attaching it here as well would
	// duplicate it in every one of their records.
	_, logger := qlog.WithCourseLog(ctx, course, label.AssignmentID, request.GetAssignmentID())
	logger.Debug("rebuilding all submissions")
	start := time.Now()

	// counting semaphore: limit concurrent rebuilding to maxContainers
	sem := make(chan struct{}, maxContainers)
	errCnt := int32(0)
	var wg sync.WaitGroup
	wg.Add(len(submissions))
	// The rebuilds must complete even if the client disconnects, so detach
	// them from the request's cancellation; each rebuild is bounded by the
	// assignment's container timeout.
	workerCtx := context.WithoutCancel(ctx)
	for _, submission := range submissions {
		rebuildReq := &qf.RebuildRequest{
			AssignmentID: request.GetAssignmentID(),
			SubmissionID: submission.GetID(),
		}
		// the counting semaphore limits concurrency to maxContainers
		go func() {
			sem <- struct{}{} // acquire semaphore
			err := s.internalRebuildSubmission(workerCtx, rebuildReq)
			if err != nil {
				atomic.AddInt32(&errCnt, 1)
				logger.Error("failed to rebuild submission", label.SubmissionID, rebuildReq.GetSubmissionID(), label.Error, err)
			}
			<-sem // release semaphore
			wg.Done()
		}()
	}
	// wait for all submissions to finish rebuilding
	wg.Wait()
	close(sem)

	logger.Debug("rebuilt submissions", "count", len(submissions), label.Duration, time.Since(start), "failed", errCnt)
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
