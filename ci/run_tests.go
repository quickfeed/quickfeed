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

// RunData stores CI data
type RunData struct {
	Course     *pb.Course
	Assignment *pb.Assignment
	Repo       *pb.Repository
	CloneURL   string
	CommitID   string
	JobOwner   string
}

// String returns a string representation of the run data structure
func (r RunData) String(secret string) string {
	return fmt.Sprintf("%s-%s-%s-%s", r.Course.GetCode(), r.Assignment.GetName(), r.JobOwner, secret)
}

// RunTests runs the assignment specified in the provided RunData structure.
func RunTests(logger *zap.SugaredLogger, db database.Database, runner Runner, rData *RunData) {
	info, err := createAssignmentInfo(db, rData.Course, rData.Assignment, rData.CloneURL)
	if err != nil {
		logger.Errorf("Failed to construct assignment info: %w", err)
		return
	}
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

// createAssignmentInfo creates a struct with data to be supplied to
// the template script files.
func createAssignmentInfo(db database.Database, course *pb.Course, assignment *pb.Assignment, cloneURL string) (*AssignmentInfo, error) {
	repoQuery := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		RepoType:       pb.Repository_TESTS,
	}
	testRepos, err := db.GetRepositories(repoQuery)
	if err != nil || len(testRepos) < 1 {
		return nil, fmt.Errorf("failed to find a test repository for %s: %w", course.GetName(), err)
	}
	testURL := testRepos[0].GetHTMLURL()
	return newAssignmentInfo(course, assignment, cloneURL, testURL), nil
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
		UserID:       rData.Repo.UserID,
		GroupID:      rData.Repo.GroupID,
		Status:       approvedStatus,
	}
	err = db.CreateSubmission(newSubmission)
	if err != nil {
		logger.Errorf("Failed to add submission to database: %w", err)
		return
	}

	logger.Debugf("Created submission for assignment %d in database with status=%t", rData.Assignment.GetID(), approvedStatus)
	updateSlipDays(logger, db, rData.Repo, rData.Assignment, newSubmission, result.BuildInfo.BuildDate)
}

func randomSecret() string {
	randomness := make([]byte, 10)
	_, err := rand.Read(randomness)
	if err != nil {
		panic("couldn't generate randomness")
	}
	return fmt.Sprintf("%x", sha1.Sum(randomness))
}

func updateSlipDays(logger *zap.SugaredLogger, db database.Database, repo *pb.Repository, assignment *pb.Assignment, submission *pb.Submission, buildDate string) {
	buildTime, err := time.Parse(layout, buildDate)
	if err != nil {
		logger.Errorf("Failed to parse time from string (%s)", buildDate)
	}

	enrollments := make([]*pb.Enrollment, 0)
	if repo.GroupID > 0 {
		group, err := db.GetGroup(repo.GroupID)
		if err != nil {
			logger.Errorf("Failed to get group %d: %w", repo.GroupID, err)
			return
		}
		enrollments = append(enrollments, group.Enrollments...)
	} else {
		enrol, err := db.GetEnrollmentByCourseAndUser(assignment.CourseID, repo.UserID)
		if err != nil {
			logger.Errorf("Failed to get enrollment for user %d: %w", repo.UserID, err)
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
