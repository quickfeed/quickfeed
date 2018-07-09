package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/autograde/aguis/scm"

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
func GithubHook(logger logrus.FieldLogger, db database.Database, runner ci.Runner, buildscripts string) webhooks.ProcessPayloadFunc {
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
				course, err := db.GetCourseByDirectoryID(repo.DirectoryID)
				if err != nil {
					logger.WithError(err).Warn("Failed to get course from database")
					return
				}
				s, err := scm.NewSCMClient("github", remoteIdentity.AccessToken)
				if err != nil {
					logger.WithError(err).Warn("Failed to create SCM Client")
					return
				}
				_, err = RefreshCourseInformation(context.Background(), logger, db, course, remoteIdentity, s)
				if err != nil {
					logger.WithError(err).Error("Problem with refreshing course information")
				}
				return
			}
			RunCI(logger, repo, db, runner, p.Repository.CloneURL, p.HeadCommit.ID, remoteIdentity, buildscripts)

		default:
			logger.WithFields(logrus.Fields{
				"event":   event,
				"payload": payload,
				"header":  h,
			}).Warn("Event not implemented")
		}
	}
}

func getLatestAssignment(db database.Database, cid uint64, uid uint64, gid uint64) (*models.Assignment, error) {
	assignments, err := db.GetAssignmentsByCourse(cid)
	if err != nil {
		return nil, err
	}
	sort.Slice(assignments, func(i, j int) bool {
		return assignments[i].Order < assignments[j].Order
	})
	for _, v := range assignments {
		fmt.Println(*v)
		if uid > 0 {
			sub, err := db.GetSubmissionForUser(v.ID, uid)
			if err != nil && err != gorm.ErrRecordNotFound {
				return nil, err
			}
			if sub == nil || sub.Approved == false {
				return v, nil
			}
		} else if gid > 0 && v.IsGroupLab {
			sub, err := db.GetSubmissionForGroup(v.ID, gid)
			if err != nil && err != gorm.ErrRecordNotFound {
				return nil, err
			}
			if sub == nil || sub.Approved == false {
				return v, nil
			}
		}
	}
	return nil, nil
}

// RunCI Runs the ci from a RemoteIdentity
func RunCI(logger logrus.FieldLogger, repo *models.Repository, db database.Database, runner ci.Runner, cloneURL string, commitHash string, remoteIdentity *models.RemoteIdentity, buildscripts string) {

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

	//selectedAssignment := assignments[0]
	selectedAssignment, err := getLatestAssignment(db, course.ID, repo.UserID, repo.GroupID)
	if err != nil || selectedAssignment == nil {
		logger.WithError(err).Warn("Failed to get course from database")
		return
	}

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

	ciInfo := models.AssignmentCIInfo{
		AccessToken:    remoteIdentity.AccessToken,
		AssignmentName: selectedAssignment.Name,
		GetURL:         getURL,
		TestURL:        getURLTest,
		RawGetURL:      strings.TrimPrefix(strings.TrimSuffix(getURL, ".git"), "https://"),
		RawTestURL:     strings.TrimPrefix(strings.TrimSuffix(getURLTest, ".git"), "https://"),
	}

	result, out, err := runCIFromTMPL(runner, language, ciInfo, buildscripts)

	if err != nil {
		logger.WithError(err).Warn("Docker failed")
		return
	}

	logger.WithField("out", out).Warn("Docker success")

	if result == nil {
		logger.Error("Empty result object")
		return
	}
	bi, err := json.Marshal(result.BuildInfo)
	sc, err2 := json.Marshal(result.Scores)

	currentScore := float64(0.0)
	maxScore := float64(0.0)
	for _, v := range result.Scores {
		percent := float64(v.Score) / float64(v.Points)
		maxScore += float64(v.Weight)
		currentScore += percent * float64(v.Weight)
	}
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
		AssignmentID: selectedAssignment.ID,
		BuildInfo:    buildInfo,
		CommitHash:   commitHash,
		Score:        uint8(currentScore / maxScore * 100),
		ScoreObjects: scores,
		UserID:       repo.UserID,
		GroupID:      repo.GroupID,
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

	fmt.Println("Image:", *image)
	fmt.Println("Data:", restData)

	fmt.Println(strings.Join(restData, "\n"))

	if image == nil {
		return nil, "", fmt.Errorf("Image not specefied in template file")
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
	if err != nil {
		return nil, out, err
	}

	curDate := time.Now().Format("2006-01-02")
	totalTimeName := endTime.UnixNano() - startTime.UnixNano()
	totalms := totalTimeName / int64(time.Millisecond)
	return &models.CIResult{
		Scores: scores,
		BuildInfo: &models.BuildInfo{
			BuildID:   1,
			BuildDate: curDate,
			BuildLog:  filteredOut,
			ExecTime:  int(totalms),
		},
	}, out, nil
}

func parseCIOutput(out string) ([]*models.ScoreObject, string, error) {
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
	return scores, filteredOut, nil
}

func extractDockerImageInformation(lines []string) (data []string, image *string) {
	if len(lines) > 0 && strings.Index(lines[0], "#image") == 0 {
		firstLine := lines[0]
		rest := lines[1:]
		parts := strings.Split(firstLine, "/")
		if len(parts) > 1 {

			return rest, &parts[1]
		}
	}
	return lines, nil
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
