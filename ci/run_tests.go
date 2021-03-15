package ci

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"time"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

const (
	scriptPath = "ci/scripts"
	layout     = "2006-01-02T15:04:05"
)

const submissionMapSize = 32

var SubmissionsMap = make(map[string]chan *pb.Submission, submissionMapSize)

// RunData stores CI data
type RunData struct {
	Course     *pb.Course
	Assignment *pb.Assignment
	Repo       *pb.Repository
	CommitID   string
	JobOwner   string
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
		logger.Errorf("Failed to run tests: %w", err)
		if ed == nil {
			return
		}
		// we only get here if err was a timeout, so that we can log 'out' to the user
	}
	result, err := ExtractResult(logger, ed.out, info.RandomSecret, ed.execTime)
	if err != nil {
		logger.Errorf("Failed to extract results from log: %w", err)
		return
	}
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
func recordResults(logger *zap.SugaredLogger, db database.Database, rData *RunData, result *Result) {
	buildInfo, scores, err := result.Marshal()
	if err != nil {
		logger.Errorf("Failed to marshal build info and scores: %w", err)
		return
	}

	logger.Debugf("Fetching most recent submission for assignment %d", rData.Assignment.GetID())
	submissionQuery := &pb.Submission{
		AssignmentID: rData.Assignment.GetID(),
		UserID:       rData.Repo.GetUserID(),
		GroupID:      rData.Repo.GetGroupID(),
	}
	newest, err := db.GetSubmission(submissionQuery)
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.Errorf("Failed to get submission data from database: %w", err)
		return
	}
	// keep approved status if already approved
	approvedStatus := newest.GetStatus()
	if rData.Assignment.AutoApprove && result.TotalScore() >= rData.Assignment.GetScoreLimit() {
		approvedStatus = pb.Submission_APPROVED
	}

	score := result.TotalScore()
	newSubmission := &pb.Submission{
		AssignmentID: rData.Assignment.ID,
		BuildInfo:    buildInfo,
		CommitHash:   rData.CommitID,
		Score:        score,
		ScoreObjects: scores,
		UserID:       rData.Repo.GetUserID(),
		GroupID:      rData.Repo.GetGroupID(),
		Status:       approvedStatus,
	}
	err = db.CreateSubmission(newSubmission)
	if err != nil {
		logger.Errorf("Failed to add submission to database: %w", err)
		return
	}
	if submissionChan, found := SubmissionsMap[rData.JobOwner]; found {
		// only send on channels created in GetSubmissionStream
		submissionChan <- newSubmission
	}
	logger.Debugf("Created submission for assignment '%s' with status %s", rData.Assignment.GetName(), approvedStatus)
	updateSlipDays(logger, db, rData.Assignment, newSubmission, result.BuildInfo.BuildDate)
}

func randomSecret() string {
	randomness := make([]byte, 10)
	_, err := rand.Read(randomness)
	if err != nil {
		panic("couldn't generate randomness")
	}
	return fmt.Sprintf("%x", sha1.Sum(randomness))
}

func updateSlipDays(logger *zap.SugaredLogger, db database.Database, assignment *pb.Assignment, submission *pb.Submission, buildDate string) {
	buildTime, err := time.Parse(layout, buildDate)
	if err != nil {
		logger.Errorf("Failed to parse time from string (%s)", buildDate)
	}

	enrollments := make([]*pb.Enrollment, 0)
	if submission.GroupID > 0 {
		group, err := db.GetGroup(submission.GroupID)
		if err != nil {
			logger.Errorf("Failed to get group %d: %w", submission.GroupID, err)
			return
		}
		enrollments = append(enrollments, group.Enrollments...)
	} else {
		enrol, err := db.GetEnrollmentByCourseAndUser(assignment.CourseID, submission.UserID)
		if err != nil {
			logger.Errorf("Failed to get enrollment for user %d: %w", submission.UserID, err)
			return
		}
		enrollments = append(enrollments, enrol)
	}

	for _, enrol := range enrollments {
		if err := enrol.UpdateSlipDays(buildTime, assignment, submission); err != nil {
			logger.Errorf("Failed updating slip days for submission ID (%d): %w", submission.ID, err)
			return
		}
		if err := db.UpdateSlipDays(enrol.UsedSlipDays); err != nil {
			logger.Errorf("Failed to update slip days (enrollment ID %d): %w", enrol.GetID(), err)
			return
		}
	}
}
