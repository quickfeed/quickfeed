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
	"github.com/sirupsen/logrus"

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
func GithubHook(logger logrus.FieldLogger, db database.Database, runner ci.Runner, scriptPath string) webhooks.ProcessPayloadFunc {
	return func(payload interface{}, header webhooks.Header) {
		h := http.Header(header)
		event := github.Event(h.Get("X-GitHub-Event"))

		switch event {
		case github.PushEvent:
			p := payload.(github.PushPayload)
			logger.WithField("payload", p).Println("Push event")

			repo, err := db.GetRepository(uint64(p.Repository.ID))
			if err != nil {
				logger.WithError(err).Error("Failed to get repository from database")
				return
			}
			logger.WithField("repo", repo).Info("Found repository, moving on")

			switch {
			case repo.IsTestsRepo():
				// the push event is for the 'tests' repo, which means that we
				// should update the course data (assignments) in the database
				refreshAssignmentsFromTestsRepo(logger, db, repo, uint64(p.Sender.ID))

			case repo.IsStudentRepo():
				// the push event is from a student or group repo; run the tests
				runTests(logger, db, runner, repo, p.Repository.CloneURL, p.HeadCommit.ID, scriptPath)

			default:
				logger.Info("Nothing to do for this push event")
			}

		default:
			logger.WithFields(logrus.Fields{
				"event":   event,
				"payload": payload,
				"header":  h,
			}).Warn("Event not implemented")
		}
	}
}

func refreshAssignmentsFromTestsRepo(logger logrus.FieldLogger, db database.Database, repo *pb.Repository, senderID uint64) {
	logger.Info("Refreshing course informaton in database")

	remoteIdentity, err := db.GetRemoteIdentity("github", senderID)
	if err != nil {
		logger.WithError(err).Error("Failed to get sender's remote identity")
		return
	}
	logger.WithField("identity", remoteIdentity).Info("Found sender's remote identity")

	s, err := scm.NewSCMClient("github", remoteIdentity.AccessToken)
	if err != nil {
		logger.WithError(err).Error("Failed to create SCM Client")
		return
	}

	course, err := db.GetCourseByDirectoryID(repo.DirectoryId)
	if err != nil {
		logger.WithError(err).Error("Failed to get course from database")
		return
	}

	assignments, err := FetchAssignments(context.Background(), s, course)
	if err != nil {
		logger.WithError(err).Error("Failed to fetch assignments from 'tests' repository")
	}
	if err = db.UpdateAssignments(assignments); err != nil {
		logger.WithError(err).Error("Failed to update assignments in database")
	}
}

// runTests runs the ci from a RemoteIdentity
func runTests(logger logrus.FieldLogger, db database.Database, runner ci.Runner, repo *pb.Repository,
	getURL string, commitHash string, scriptPath string) {

	course, err := db.GetCourseByDirectoryID(repo.DirectoryId)
	if err != nil {
		logger.WithError(err).Error("Failed to get course from database")
		return
	}

	courseCreator, err := db.GetUser(course.CoursecreatorId)
	if err != nil || len(courseCreator.RemoteIdentities) < 1 {
		logger.WithError(err).Error("Failed to fetch course creator")
	}

	selectedAssignment, err := db.GetNextAssignment(course.Id, repo.UserId, repo.GroupId)
	if err != nil {
		logger.WithError(err).Error("Failed to find a next unapproved assignment")
		return
	}
	logger.WithField("Assignment", selectedAssignment).Info("Found assignment")

	testRepos, err := db.GetRepositoriesByCourseAndType(course.Id, pb.Repository_TESTS)
	if err != nil || len(testRepos) < 1 {
		logger.WithError(err).Error("Failed to find test repository in database")
		return
	}
	getURLTest := testRepos[0].HtmlUrl
	logger.WithField("url", getURL).Info("Code Repository")
	logger.WithField("url", getURLTest).Info("Test repository")

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
		logger.WithError(err).Error("Failed to parse script template")
		return
	}

	start := time.Now()
	out, err := runner.Run(context.Background(), job)
	if err != nil {
		logger.WithError(err).Error("Docker execution failed")
		return
	}
	execTime := time.Since(start)
	logger.WithField("out", out).WithField("execTime", execTime).Info("Docker execution successful")

	result, err := ci.ExtractResult(out, randomSecret, execTime)
	if err != nil {
		logger.WithError(err).Error("Failed to extract results from log")
		return
	}
	buildInfo, scores, err := result.Marshal()
	if err != nil {
		logger.WithError(err).Error("Failed to marshal build info and scores")
	}
	logger.WithField("result", result).Info("Extracted results")

	err = db.CreateSubmission(&pb.Submission{
		AssignmentId: selectedAssignment.Id,
		BuildInfo:    buildInfo,
		CommitHash:   commitHash,
		Score:        uint32(result.TotalScore()),
		ScoreObjects: scores,
		UserId:       repo.UserId,
		GroupId:      repo.GroupId,
	})
	if err != nil {
		logger.WithError(err).Error("Failed to add submission to database")
		return
	}
}

//TODO(Vera): not needed anymore
/*
func getTestRepoCloneURL(logger logrus.FieldLogger, db database.Database, remoteIdentity *pb.RemoteIdentity, repo *pb.Repository) (string, error) {
	// Add repository url to repository table in database to prevent requestion the data every time we need it.
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: remoteIdentity.AccessToken})
	client := gh.NewClient(oauth2.NewClient(context.Background(), ts))
	allRepos, err := db.GetRepositoriesByDirectory(repo.DirectoryId)
	if err != nil {
		logger.WithError(err).Error("Problem with requesting repositories")
		return "", err
	}
	var testRepo *ag.Repository
	for _, v := range allRepos {
		if v.RepoType == ag.Repository_TESTS {
			testRepo = v
			break
		}
	}
	if testRepo == nil {
		logger.Error("Test Repo does not exists")
	}
	testRepos, _, err := client.Repositories.GetByID(context.Background(), int(testRepo.RepositoryId))
	if err != nil {
		logger.WithError(err).Error("Got error while requesting repository")
		return "", err
	}
	return *testRepos.CloneURL, nil

}

func runCIFromTMPL(runner ci.Runner, language string, ciInfo models.AssignmentCIInfo, buildscripts string) (*models.CIResult, string, error) {
	bPath := path.Join(buildscripts, language+".tmpl")
	if _, err := os.Stat(bPath); err != nil {
		return nil, "", err
	}
	t, err := template.ParseFiles(bPath)
	if err != nil {
		return nil, "", err
	}

	buffer := bytes.NewBufferString("")

	t.Execute(buffer, ciInfo)

	lines := strings.Split(buffer.String(), "\n")
	restData, image := extractDockerImageInformation(lines)

	if image == nil {
		return nil, "", fmt.Errorf("image not specified in template file")
	}

	startTime := time.Now()
	out, err := runner.Run(context.Background(), &ci.Job{
		Image:    *image,
		Commands: restData,
	})
	endTime := time.Now()

	if err != nil {
		return nil, out, err
	}

	scores, filteredOut, err := parseCIOutput(out)

*/
func randomSecret() string {
	randomness := make([]byte, 10)
	_, err := rand.Read(randomness)
	if err != nil {
		panic("couldn't generate randomness")
	}
	return fmt.Sprintf("%x", sha1.Sum(randomness))
}

// GitlabHook handles events from Gitlab.
func GitlabHook(logger logrus.FieldLogger) webhooks.ProcessPayloadFunc {
	return func(payload interface{}, header webhooks.Header) {
		h := http.Header(header)
		event := gitlab.Event(h.Get("X-Gitlab-Event"))

		switch event {
		case gitlab.PushEvents:
			p := payload.(gitlab.PushEventPayload)
			logger.WithField("payload", p).Println("Push event")
		default:
			logger.WithFields(logrus.Fields{
				"event":   event,
				"payload": payload,
				"header":  h,
			}).Warn("Event not implemented")
		}
	}
}
