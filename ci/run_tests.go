package ci

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"time"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/kit/score"
	"github.com/autograde/quickfeed/log"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	scriptPath = "ci/scripts"
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
func RunTests(logger *zap.SugaredLogger, db database.Database, runner Runner, rData *RunData) {
	info := newAssignmentInfo(rData.Course, rData.Assignment, rData.Repo.GetHTMLURL(), rData.Repo.GetTestURL())
	logger.Debugf("Running tests for %s", rData.JobOwner)
	ed, err := runTests(scriptPath, runner, info, rData)
	if err != nil {
		logger.Errorf("Failed to run tests: %v", err)
		if ed == nil {
			return
		}
		// we only get here if err was a timeout, so that we can log 'out' to the user
	}
	result, err := score.ExtractResults(ed.out, info.RandomSecret, ed.execTime)
	if err != nil {
		logger.Errorf("Failed to extract results from log: %v", err)
		return
	}
	logger.Debug("ci.ExtractResults",
		zap.Any("results", log.IndentJson(result)),
	)
	recordResults(logger, db, rData, result)
}

type execData struct {
	out      string
	execTime time.Duration
}

// runTests returns execData struct.
// An error is returned if the execution fails, or times out.
// If a timeout is the cause of the error, we also return an output string to the user.
func runTests(path string, runner Runner, info *AssignmentInfo, rData *RunData) (*execData, error) {
	job, err := parseScriptTemplate(path, info)
	if err != nil {
		return nil, fmt.Errorf("failed to parse script template: %w", err)
	}

	job.Name = rData.String(info.RandomSecret[:6])
	start := time.Now()

	timeout := containerTimeout
	t := rData.Assignment.GetContainerTimeout()
	if t > 0 {
		timeout = time.Duration(t) * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	out, err := runner.Run(ctx, job)
	if err != nil && out == "" {
		return nil, fmt.Errorf("test execution failed: %w", err)
	}
	// this may return a timeout error as well
	return &execData{out: out, execTime: time.Since(start)}, err
}

// recordResults for the assignment given by the run data structure.
func recordResults(logger *zap.SugaredLogger, db database.Database, rData *RunData, result *score.Results) {
	assignment := rData.Assignment
	logger.Debugf("Fetching most recent submission for assignment %d", assignment.GetID())
	submissionQuery := &pb.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       rData.Repo.GetUserID(),
		GroupID:      rData.Repo.GetGroupID(),
	}
	newest, err := db.GetSubmission(submissionQuery)
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.Errorf("Failed to get submission data from database: %v", err)
		return
	}

	// Keep the original submission's delivery date (obtained from the database (newest)) if this is a manual rebuild.
	if rData.Rebuild {
		result.BuildInfo.BuildDate = newest.BuildInfo.BuildDate
	}

	score := result.Sum()
	newSubmission := &pb.Submission{
		ID:           newest.GetID(),
		AssignmentID: assignment.GetID(),
		CommitHash:   rData.CommitID,
		Score:        score,
		BuildInfo:    result.BuildInfo,
		Scores:       result.Scores,
		UserID:       rData.Repo.GetUserID(),
		GroupID:      rData.Repo.GetGroupID(),
		Status:       assignment.IsApproved(newest, score),
	}
	err = db.CreateSubmission(newSubmission)
	if err != nil {
		logger.Errorf("Failed to add submission to database: %v", err)
		return
	}
	logger.Debugf("Created submission for assignment '%s' with score %d, status %s", assignment.GetName(), score, newSubmission.GetStatus())
	if !rData.Rebuild {
		updateSlipDays(logger, db, rData.Assignment, newSubmission)
	}
}

func randomSecret() string {
	randomness := make([]byte, 10)
	if _, err := rand.Read(randomness); err != nil {
		panic("couldn't generate randomness")
	}
	return fmt.Sprintf("%x", sha1.Sum(randomness))
}

func updateSlipDays(logger *zap.SugaredLogger, db database.Database, assignment *pb.Assignment, submission *pb.Submission) {
	enrollments := make([]*pb.Enrollment, 0)
	if submission.GroupID > 0 {
		group, err := db.GetGroup(submission.GroupID)
		if err != nil {
			logger.Errorf("Failed to get group %d: %v", submission.GroupID, err)
			return
		}
		enrollments = append(enrollments, group.Enrollments...)
	} else {
		enrol, err := db.GetEnrollmentByCourseAndUser(assignment.CourseID, submission.UserID)
		if err != nil {
			logger.Errorf("Failed to get enrollment for user %d: %v", submission.UserID, err)
			return
		}
		enrollments = append(enrollments, enrol)
	}

	for _, enrol := range enrollments {
		if err := enrol.UpdateSlipDays(assignment, submission); err != nil {
			logger.Errorf("Failed to update slip days for submission %d: %v", submission.ID, err)
			return
		}
		if err := db.UpdateSlipDays(enrol.UsedSlipDays); err != nil {
			logger.Errorf("Failed to update slip days for enrollment %d: %v", enrol.ID, err)
			return
		}
	}
}
