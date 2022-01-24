package ci

import (
	"context"
	"fmt"
	"time"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/internal/rand"
	"github.com/autograde/quickfeed/kit/score"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RunData stores CI data
type RunData struct {
	Course     *pb.Course
	Assignment *pb.Assignment
	Repo       *pb.Repository
	CommitID   string
	JobOwner   string
	Rebuild    bool
}

// String returns a string representation of the run data structure
func (r RunData) String(secret string) string {
	return fmt.Sprintf("%s-%s-%s-%s", r.Course.GetCode(), r.Assignment.GetName(), r.JobOwner, secret)
}

// RunTests runs the assignment specified in the provided RunData structure.
func (r RunData) RunTests(logger *zap.SugaredLogger, runner Runner) (*score.Results, error) {
	logger.Debugf("Running tests for %s", r.JobOwner)

	randomSecret := rand.String()
	job, err := r.parseScriptTemplate(randomSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to parse script template: %w", err)
	}

	start := time.Now()
	ctx, cancel := r.withTimeout(containerTimeout)
	defer cancel()
	out, err := runner.Run(ctx, job)
	if err != nil && out == "" {
		return nil, fmt.Errorf("test execution failed without output: %w", err)
	}
	if err != nil {
		// we may reach here with a timeout error and a non-empty output
		logger.Errorf("test execution failed with output: %v\n%v", err, out)
	}
	// return the extracted score and filtered log output
	return score.ExtractResults(out, randomSecret, time.Since(start))
}

func (r RunData) withTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	t := r.Assignment.GetContainerTimeout()
	if t > 0 {
		timeout = time.Duration(t) * time.Minute
	}
	return context.WithTimeout(context.Background(), timeout)
}

// RecordResults for the assignment given by the run data structure.
func (r RunData) RecordResults(logger *zap.SugaredLogger, db database.Database, result *score.Results) error {
	// Sanity check of the result object
	if result == nil || result.BuildInfo == nil {
		return fmt.Errorf("no build info found in results object: %v", result)
	}

	assignment := r.Assignment
	logger.Debugf("Fetching most recent submission for assignment %d", assignment.GetID())
	submissionQuery := &pb.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       r.Repo.GetUserID(),
		GroupID:      r.Repo.GetGroupID(),
	}
	newest, err := db.GetSubmission(submissionQuery)
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to get newest submission: %w", err)
	}

	// Keep the original submission's delivery date (obtained from the database (newest)) if this is a manual rebuild.
	if r.Rebuild {
		if newest != nil && newest.BuildInfo != nil {
			// Only update the build date if we found a previous submission
			result.BuildInfo.BuildDate = newest.BuildInfo.BuildDate
		} else {
			// Can happen if a previous submission failed to store to the database
			logger.Debug("Rebuild with no previous submission stored in database")
		}
	}

	score := result.Sum()
	newSubmission := &pb.Submission{
		ID:           newest.GetID(),
		AssignmentID: assignment.GetID(),
		CommitHash:   r.CommitID,
		Score:        score,
		BuildInfo:    result.BuildInfo,
		Scores:       result.Scores,
		UserID:       r.Repo.GetUserID(),
		GroupID:      r.Repo.GetGroupID(),
		Status:       assignment.IsApproved(newest, score),
	}
	err = db.CreateSubmission(newSubmission)
	if err != nil {
		return fmt.Errorf("failed to store submission %d: %w", newest.GetID(), err)
	}
	logger.Debugf("Created submission for assignment '%s' with score %d, status %s", assignment.GetName(), score, newSubmission.GetStatus())
	if !r.Rebuild {
		return r.updateSlipDays(db, newSubmission)
	}

	// TODO(meling) return newSubmission
	return nil
}

func (r RunData) updateSlipDays(db database.Database, submission *pb.Submission) error {
	buildDate := submission.GetBuildInfo().GetBuildDate()
	buildTime, err := time.Parse(pb.TimeLayout, buildDate)
	if err != nil {
		return fmt.Errorf("failed to parse time from build date (%s): %w", buildDate, err)
	}

	enrollments := make([]*pb.Enrollment, 0)
	if submission.GroupID > 0 {
		group, err := db.GetGroup(submission.GroupID)
		if err != nil {
			return fmt.Errorf("failed to get group %d: %w", submission.GroupID, err)
		}
		enrollments = append(enrollments, group.Enrollments...)
	} else {
		enrol, err := db.GetEnrollmentByCourseAndUser(r.Assignment.CourseID, submission.UserID)
		if err != nil {
			return fmt.Errorf("failed to get enrollment for user %d in course %d: %w", submission.UserID, r.Assignment.CourseID, err)
		}
		enrollments = append(enrollments, enrol)
	}

	for _, enrol := range enrollments {
		if err := enrol.UpdateSlipDays(buildTime, r.Assignment, submission); err != nil {
			return fmt.Errorf("failed to update slip days for user %d in course %d: %w", enrol.UserID, r.Assignment.CourseID, err)
		}
		if err := db.UpdateSlipDays(enrol.UsedSlipDays); err != nil {
			return fmt.Errorf("failed to update slip days for enrollment %d (user %d) (course %d): %w", enrol.ID, enrol.UserID, enrol.CourseID, err)
		}
	}
	return nil
}
