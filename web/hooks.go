package web

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"
	"go.uber.org/zap"

	webhooks "gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"
	"gopkg.in/go-playground/webhooks.v3/gitlab"

	pb "github.com/autograde/aguis/ag"
)

// BaseHookOptions contains options shared among all webhooks.
type BaseHookOptions struct {
	BaseURL string
	// Secret is used to verify that the event received is legit. GitHub
	// sends back a signature of the payload, while GitLab just sends back
	// the secret. This is all handled by the
	// gopkg.in/go-playground/webhooks.v3 package.
	Secret string
}

// GithubHook handles webhook events from GitHub.
func GithubHook(logger *zap.Logger, db database.Database, runner ci.Runner, scriptPath string) webhooks.ProcessPayloadFunc {
	return func(payload interface{}, header webhooks.Header) {
		h := http.Header(header)
		event := github.Event(h.Get("X-GitHub-Event"))

		switch event {
		case github.PushEvent:
			p := payload.(github.PushPayload)
			logger.Debug("Push event", zap.Any("payload", p))

			repo, err := db.GetRepository(uint64(p.Repository.ID))
			if err != nil {
				logger.Error("Failed to get repository from database", zap.Error(err))
				return
			}
			logger.Debug("Found repository, moving on", zap.Any("repo", repo))

			switch {
			case repo.IsTestsRepo():
				// the push event is for the 'tests' repo, which means that we
				// should update the course data (assignments) in the database
				refreshAssignmentsFromTestsRepo(logger, db, repo, uint64(p.Sender.ID))

			case repo.IsStudentRepo():
				// the push event is from a student or group repo; run the tests
				runTests(logger, db, runner, repo, p.Repository.CloneURL, p.HeadCommit.ID, scriptPath)

			default:
				logger.Debug("Nothing to do for this push event")
			}

		default:
			logger.Debug("Event not implemented",
				zap.Any("event", event),
				zap.Any("payload", payload),
				zap.Any("header", h),
			)
		}
	}
}

func refreshAssignmentsFromTestsRepo(logger *zap.Logger, db database.Database, repo *pb.Repository, senderID uint64) {
	logger.Debug("Refreshing course informaton in database")
	provider := "github"

	remoteIdentity, err := db.GetRemoteIdentity(provider, senderID)
	if err != nil {
		logger.Error("Failed to get sender's remote identity", zap.Error(err))
		return
	}
	logger.Debug("Found sender's remote identity", zap.String("remote identity", remoteIdentity.String()))

	s, err := scm.NewSCMClient(logger, provider, remoteIdentity.AccessToken)
	if err != nil {
		logger.Error("Failed to create SCM Client", zap.Error(err))
		return
	}

	course, err := db.GetCourseByOrganizationID(repo.OrganizationID)
	if err != nil {
		logger.Error("Failed to get course from database", zap.Error(err))
		return
	}

	assignments, err := fetchAssignments(context.Background(), s, course)
	if err != nil {
		logger.Error("Failed to fetch assignments from 'tests' repository", zap.Error(err))
	}
	if err = db.UpdateAssignments(assignments); err != nil {
		for _, assignment := range assignments {
			logger.Sugar().Debug("Fetched assignment with ID: ", assignment.GetID())
		}
		logger.Error("Failed to update assignments in database", zap.Error(err))
	}
}

// runTests runs the ci from a RemoteIdentity
func runTests(logger *zap.Logger, db database.Database, runner ci.Runner, repo *pb.Repository,
	getURL string, commitHash string, scriptPath string) {

	course, err := db.GetCourseByOrganizationID(repo.OrganizationID)
	if err != nil {
		logger.Error("Failed to get course from database", zap.Error(err))
		return
	}

	courseCreator, err := db.GetUser(course.CourseCreatorID)
	if err != nil || len(courseCreator.RemoteIdentities) < 1 {
		logger.Error("Failed to fetch course creator", zap.Error(err))
	}

	selectedAssignment, err := db.GetNextAssignment(course.ID, repo.UserID, repo.GroupID)
	if err != nil {
		logger.Error("Failed to find a next unapproved assignment", zap.Error(err))
		return
	}
	logger.Debug("Found assignment", zap.String("assignment", selectedAssignment.String()))

	testsRepoQuery := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		RepoType:       pb.Repository_TESTS,
	}
	testRepos, err := db.GetRepositories(testsRepoQuery)
	if err != nil || len(testRepos) < 1 {
		logger.Error("Failed to find test repository in database", zap.Error(err))
		return
	}
	getURLTest := testRepos[0].HTMLURL
	logger.Debug("Code Repository", zap.String("url", getURL))
	logger.Debug("Test Repository", zap.String("url", getURLTest))

	randomSecret := randomSecret()
	info := ci.AssignmentInfo{
		CreatorAccessToken: courseCreator.RemoteIdentities[0].AccessToken,
		AssignmentName:     selectedAssignment.Name,
		Language:           selectedAssignment.Language,
		GetURL:             getURL,
		TestURL:            getURLTest,
		RawGetURL:          strings.TrimPrefix(strings.TrimSuffix(getURL, ".git"), "https://"),
		RawTestURL:         strings.TrimPrefix(strings.TrimSuffix(getURLTest, ".git"), "https://"),
		RandomSecret:       randomSecret,
	}

	job, err := ci.ParseScriptTemplate(scriptPath, info)
	if err != nil {
		logger.Error("Failed to parse script template", zap.Error(err))
		return
	}

	start := time.Now()
	logger.Debug("Job started successfully")
	out, err := runner.Run(context.Background(), job)
	if err != nil {
		logger.Error("Docker execution failed", zap.Error(err))
		return
	}
	logger.Sugar().Debugf("Job output: ", out)
	execTime := time.Since(start)

	result, err := ci.ExtractResult(logger, out, randomSecret, execTime)
	if err != nil {
		logger.Error("Failed to extract results from log", zap.Error(err))
		return
	}
	buildInfo, scores, err := result.Marshal()
	if err != nil {
		logger.Error("Failed to marshal build info and scores", zap.Error(err))
	}
	logger.Debug("Extracted results",
		zap.Any("result", result),
		zap.Uint8("total score", result.TotalScore()),
		zap.String("marshalled build info", buildInfo),
		zap.String("marshalled scores", scores),
	)
	err = db.CreateSubmission(&pb.Submission{
		AssignmentID: selectedAssignment.ID,
		BuildInfo:    buildInfo,
		CommitHash:   commitHash,
		Score:        uint32(result.TotalScore()),
		ScoreObjects: scores,
		UserID:       repo.UserID,
		GroupID:      repo.GroupID,
	})
	if err != nil {
		logger.Error("Failed to add submission to database", zap.Error(err))
		return
	}
}

func randomSecret() string {
	randomness := make([]byte, 10)
	_, err := rand.Read(randomness)
	if err != nil {
		panic("couldn't generate randomness")
	}
	return fmt.Sprintf("%x", sha1.Sum(randomness))
}

// GitlabHook handles events from Gitlab.
func GitlabHook(logger *zap.Logger) webhooks.ProcessPayloadFunc {
	return func(payload interface{}, header webhooks.Header) {
		h := http.Header(header)
		event := gitlab.Event(h.Get("X-Gitlab-Event"))

		switch event {
		case gitlab.PushEvents:
			p := payload.(gitlab.PushEventPayload)
			logger.Debug("Push event", zap.Any("payload", p))
		default:
			logger.Debug("Event not implemented",
				zap.Any("event", event),
				zap.Any("payload", payload),
				zap.Any("header", h),
			)
		}
	}
}
