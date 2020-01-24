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
	scriptPath                   = "ci/scripts"
	defaultAutoApproveScoreLimit = 80
)

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
	job, err := ParseScriptTemplate(scriptPath, info) //TODO(meling) func can be made private
	if err != nil {
		logger.Errorf("Failed to parse script template: %w", err)
		return
	}

	jobName := rData.String(info.RandomSecret[:6])
	logger.Debugf("Running tests for %s", jobName)
	start := time.Now()
	out, err := runner.Run(context.Background(), job, jobName)
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
	courseCreator, err := db.GetUser(course.GetCourseCreatorID())
	if err != nil {
		return nil, fmt.Errorf("failed to get course creator: %w", err)
	}
	accessToken, err := courseCreator.GetAccessToken(course.GetProvider())
	if err != nil {
		return nil, fmt.Errorf("failed to get access token for course creator: %w", err)
	}

	repoQuery := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		RepoType:       pb.Repository_TESTS,
	}
	testRepos, err := db.GetRepositories(repoQuery)
	if err != nil || len(testRepos) < 1 {
		return nil, fmt.Errorf("failed to find a test repository for %s: %w", course.GetName(), err)
	}
	getURLTest := testRepos[0].HTMLURL

	return &AssignmentInfo{
		AssignmentName:     assignment.GetName(),
		Language:           assignment.GetLanguage(),
		CreatorAccessToken: accessToken,
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
	approved := newest.GetApproved()
	if approved {
		logger.Debugf("Assignment %d already approved in %v", rData.Assignment.GetID(), newest)
	}

	// for auto approve, use default score limit unless defined in yaml file
	//TODO(meling) this logic should be done in assignment_parser.go
	minimumScore := uint8(rData.Assignment.GetScoreLimit())
	if minimumScore < 1 {
		minimumScore = defaultAutoApproveScoreLimit
	}

	if rData.Assignment.AutoApprove && result.TotalScore() >= minimumScore {
		approved = true
	}

	err = db.CreateSubmission(&pb.Submission{
		AssignmentID: rData.Assignment.ID,
		BuildInfo:    buildInfo,
		CommitHash:   rData.CommitID,
		Score:        uint32(result.TotalScore()),
		ScoreObjects: scores,
		UserID:       rData.Repo.UserID,
		GroupID:      rData.Repo.GroupID,
		Approved:     approved,
	})
	if err != nil {
		logger.Errorf("Failed to add submission to database: %w", err)
		return
	}
	logger.Debugf("Created submission for assignment %d in database with approve=%t", rData.Assignment.GetID(), approved)
}

func randomSecret() string {
	randomness := make([]byte, 10)
	_, err := rand.Read(randomness)
	if err != nil {
		panic("couldn't generate randomness")
	}
	return fmt.Sprintf("%x", sha1.Sum(randomness))
}
