package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/models"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	webhooks "gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"
	"gopkg.in/go-playground/webhooks.v3/gitlab"

	gh "github.com/google/go-github/github"
)

// GithubHook handles events from GitHub.
func GithubHook(logger logrus.FieldLogger, db database.Database, runner ci.Runner) webhooks.ProcessPayloadFunc {
	return func(payload interface{}, header webhooks.Header) {
		h := http.Header(header)
		event := github.Event(h.Get("X-GitHub-Event"))

		switch event {
		case github.PushEvent:
			p := payload.(github.PushPayload)
			logger.WithField("payload", p).Println("Push event")

			remoteIdentity, err := db.GetRemoteIdentity("github", uint64(p.Sender.ID))
			if err != nil {
				logger.WithError(err).Warn("Failed to get sender's remote identity")
				return
			}
			logger.WithField("identity", remoteIdentity).Warn("Found sender's remote identity")

			id := p.Repository.ID
			logger.Infof("fetching repo with id: %d\n", id)
			repo, err := db.GetRepository(uint64(p.Repository.ID))
			if err != nil {
				logger.WithError(err).Warn("Failed to get repository from database")
				return
			}
			logger.WithField("repo", repo).Info("Found repository, continuing on")

			if repo.Type > 0 {
				logger.Info("Should refresh database course informaton")
				// Here should we do a refresh of the courses since this would be a repo with a type
				return
			}
			RunCI(logger, repo, db, runner, p.Repository.CloneURL, p.HeadCommit.ID, remoteIdentity)

		default:
			logger.WithFields(logrus.Fields{
				"event":   event,
				"payload": payload,
				"header":  h,
			}).Warn("Event not implemented")
		}
	}
}

// RunCI Runs the ci from a RemoteIdentity
func RunCI(logger logrus.FieldLogger, repo *models.Repository, db database.Database, runner ci.Runner, cloneURL string, commitHash string, remoteIdentity *models.RemoteIdentity) {

	course, err := db.GetCourseByDirectoryID(repo.DirectoryID)
	if err != nil {
		logger.WithError(err).Warn("Failed to get course from database")
		return
	}

	assignments, err := db.GetAssignmentsByCourse(course.ID)
	if err != nil {
		logger.WithError(err).Warn("Failed to get course from database")
		return
	} else if len(assignments) < 1 {
		logger.Warn("No assignments in database")
		return
	}

	selectedAssignment := assignments[0]

	language := selectedAssignment.Language

	logger.WithField("Assignemnt", selectedAssignment).Info("Found assignment")

	testCloneURL, err := getTestRepoCloneURL(logger, db, remoteIdentity, repo)
	if err != nil {
		return
	}

	getURL := cloneURL
	getURLTest := testCloneURL

	logger.WithField("url", getURL).Warn("Repository's go get URL")
	logger.WithField("url", getURLTest).Warn("Repository's go get test URL")
	var result *models.CIResult
	switch language {
	case "java":
		logger.Println("Starting java build")
	case "go":
		logger.Println("Starting go build")
		var out string

		result, out, err = runGoCI(runner, getURL, getURLTest, remoteIdentity.AccessToken, selectedAssignment.Name)

		if err != nil {
			logger.WithError(err).Warn("Docker failed")
			return
		}

		logger.WithField("out", out).Warn("Docker success")
	}
	if result == nil {
		logger.Error("Empty result object")
		return
	}
	bi, err := json.Marshal(result.BuildInfo)
	sc, err2 := json.Marshal(result.Scores)
	if err != nil {
		logger.WithError(err).Error("Problems with marshaling the build object")
		return
	}
	if err2 != nil {
		logger.WithError(err2).Error("Problems with marshaling the score object")
		return
	}
	buildInfo := string(bi)
	scores := string(sc)

	err = db.CreateSubmission(&models.Submission{
		AssignmentID: assignments[0].ID,
		BuildInfo:    buildInfo,
		CommitHash:   commitHash,
		Score:        0,
		ScoreObjects: scores,
		UserID:       repo.UserID,
	})
	if err != nil {
		logger.WithError(err).Error("Problems inserting the submission into the database")
		return
	}
}

func getTestRepoCloneURL(logger logrus.FieldLogger, db database.Database, remoteIdentity *models.RemoteIdentity, repo *models.Repository) (string, error) {
	// Add repository url to repository table in database to prevent requestion the data every time we need it.
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: remoteIdentity.AccessToken})
	client := gh.NewClient(oauth2.NewClient(context.Background(), ts))
	allRepos, err := db.GetRepositoriesByDirectory(repo.DirectoryID)
	if err != nil {
		logger.WithError(err).Error("Problem with requesting repositories")
		return "", err
	}
	var testRepo *models.Repository
	for _, v := range allRepos {
		if v.Type == models.TestsRepo {
			testRepo = v
			break
		}
	}
	if testRepo == nil {
		logger.Error("Test Repo does not exists")
	}
	testRepos, _, err := client.Repositories.GetByID(context.Background(), int(testRepo.RepositoryID))
	if err != nil {
		logger.WithError(err).Error("Got error while requesting repository")
		return "", err
	}
	return *testRepos.CloneURL, nil

}

func runGoCI(runner ci.Runner, getURL string, testURL string, accessToken string, assignmentName string) (*models.CIResult, string, error) {
	// getURL = strings.TrimPrefix(getURL, "https://")
	// getURL = strings.TrimSuffix(getURL, ".git")
	// getURL = strings.TrimPrefix(getURL, "https://")
	// getURL = strings.TrimSuffix(getURL, ".git")
	startTime := time.Now()
	out, err := runner.Run(context.Background(), &ci.Job{
		Image: "golang:1.8.3",
		Commands: []string{
			`echo "\n\n==START_CI==\n"`,
			`git config --global url."https://` + accessToken + `:x-oauth-basic@github.com/".insteadOf "https://github.com/"`,
			//`go get "` + getURL + `"`,
			//`cd "$GOPATH/src/` + getURL + `"`,
			//`go test -v`,
			`MD="merge"     # Merged Dir`,
			`UD="user-dir"  # User Dir`,
			`TD="test-dir"  # Test Dir`,
			`rm -rf $MD`,
			`mkdir $MD`,
			`git clone ` + getURL + ` $UD`,
			`git clone ` + testURL + ` $TD`,
			`cp -r $UD/* $MD`,
			`cp -r $TD/* $MD`,
			`cd $MD`,
			`cd "` + assignmentName + `"`,
			// `find`, // Could be readded for debugging the Internal of the CI files
			`echo "Setup done"`,
			`go test -v`,
			`echo "\n==DONE_CI==\n"`,
		},
	})
	endTime := time.Now()

	if err != nil {
		return nil, out, err
	}
	parts := strings.Split(out, "\n")
	var scores []*models.ScoreObject
	var filteredOutLines []string

	for _, v := range parts {
		if strings.Contains(v, "---|||---|||---|||---") {
			score := &models.CIOutput{}
			err := json.Unmarshal([]byte(v), score)
			if err != nil {
				return nil, out, err
			}
			scores = append(scores, &models.ScoreObject{Name: score.TestName, Points: score.MaxScore, Score: score.Score, Weight: score.Weight})
		} else {
			filteredOutLines = append(filteredOutLines, v)
		}
	}
	filteredOut := strings.Join(filteredOutLines, "\n")
	curDate := time.Now().Format("2006-01-02")
	totalTimeName := endTime.UnixNano() - startTime.UnixNano()
	totalms := totalTimeName / int64(time.Millisecond)
	return &models.CIResult{
		Scores: scores,
		BuildInfo: &models.BuildInfo{
			BuildID:   1,
			BuildDate: curDate,
			BuildLog:  "Not fully implemented\n" + filteredOut,
			ExecTime:  int(totalms),
		},
	}, out, nil
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
