package ci

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"strings"
	"time"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
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
	job, err := parseScriptTemplate(scriptPath, info)
	if err != nil {
		logger.Errorf("Failed to parse script template: %w", err)
		return
	}

	jobName := rData.String(info.RandomSecret[:6])
	logger.Debugf("Running tests for %s", jobName)
	start := time.Now()
	out, err := runner.Run(context.Background(), job, jobName, time.Duration(rData.Assignment.ContainerTimeout))
	if err != nil {
		logger.Errorf("Test execution failed: %w", err)
		return
	}
	execTime := time.Since(start)

	result, err := ExtractResult(logger, out, info.RandomSecret, execTime)
	if err != nil {
		logger.Errorf("Failed to extract results from log: %w", err)
		return
	}
	recordResults(logger, db, rData, result)
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
	getURLTest := testRepos[0].GetHTMLURL()

	return &AssignmentInfo{
		AssignmentName:     assignment.GetName(),
		Language:           assignment.GetLanguage(),
		CreatorAccessToken: course.GetAccessToken(),
		GetURL:             cloneURL,
		TestURL:            getURLTest,
		RawGetURL:          strings.TrimPrefix(strings.TrimSuffix(cloneURL, ".git"), "https://"),
		RawTestURL:         strings.TrimPrefix(strings.TrimSuffix(getURLTest, ".git"), "https://"),
		RandomSecret:       randomSecret(),
	}, nil
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
