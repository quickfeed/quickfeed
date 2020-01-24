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
	"github.com/jinzhu/gorm"
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
func GithubHook(logger *zap.SugaredLogger, db database.Database, runner ci.Runner, scriptPath string) webhooks.ProcessPayloadFunc {
	return func(payload interface{}, header webhooks.Header) {
		h := http.Header(header)
		event := github.Event(h.Get("X-GitHub-Event"))

		switch event {
		case github.PushEvent:
			p := payload.(github.PushPayload)
			logger.Debug("Push event", zap.Any("payload", p))

			repo, err := db.GetRepositoryByRemoteID(uint64(p.Repository.ID))
			if err != nil {
				logger.Error("Failed to get repository from database", zap.Error(err))
				return
			}
			logger.Debugf("Push for repository %v", repo)
			course, err := db.GetCourseByOrganizationID(repo.OrganizationID)
			if err != nil {
				logger.Error("Failed to get course from database", zap.Error(err))
				return
			}
			logger.Debugf("For course(%d)=%v", course.GetID(), course.GetName())

			switch {
			case repo.IsTestsRepo():
				// the push event is for the 'tests' repo, which means that we
				// should update the course data (assignments) in the database
				refreshAssignmentsFromTestsRepo(logger, db, repo, uint64(p.Sender.ID))

			case repo.IsStudentRepo():
				// parse the lab names from the push payload
				modifiedLabs := p.HeadCommit.Modified
				var labNames []string
				for _, lab := range modifiedLabs {
					labName := strings.Split(lab, "/")[0]
					if !contains(labNames, labName) {
						labNames = append(labNames, labName)
					}
				}

				// run tests for every lab updated by student
				for _, lab := range labNames {
					// get assignment based on course id and lab name
					assignment, err := db.GetAssignment(&pb.Assignment{Name: lab, CourseID: course.GetID()})
					if err != nil {
						logger.Errorf("could not find assignment %s: %v", lab, err)
						return
					}
					// determine container tag based on user name or group name (if group lab)
					tagName := p.Sender.Login
					if assignment.IsGroupLab {
						tagName = p.Repository.Name
					}
					runTests(logger, db, runner, repo, p.Repository.CloneURL, p.HeadCommit.ID, scriptPath, assignment.GetID(), tagName)
				}

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

// refreshAssignmentFromTestsRepo updates the database record for the course assignments
func refreshAssignmentsFromTestsRepo(logger *zap.SugaredLogger, db database.Database, repo *pb.Repository, senderID uint64) {
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
		//TODO(meling) should this return?
	}
	if err = db.UpdateAssignments(assignments); err != nil {
		for _, assignment := range assignments {
			logger.Debug("Fetched assignment with ID: ", assignment.GetID())
		}
		logger.Error("Failed to update assignments in database", zap.Error(err))
	}
}

// runTests runs the ci from a RemoteIdentity
func runTests(logger *zap.SugaredLogger, db database.Database, runner ci.Runner, repo *pb.Repository,
	getURL string, commitHash string, scriptPath string, assignmentID uint64, author string) {

	course, err := db.GetCourseByOrganizationID(repo.OrganizationID)
	if err != nil {
		logger.Error("Failed to get course from database", zap.Error(err))
		return
	}

	courseCreator, err := db.GetUser(course.CourseCreatorID)
	if err != nil || len(courseCreator.RemoteIdentities) < 1 {
		logger.Error("Failed to fetch course creator", zap.Error(err))
		//TODO(meling) should this return?
	}

	var selectedAssignment *pb.Assignment

	// if assignment ID is defined, fetch the assignment by ID
	if assignmentID > 0 {
		logger.Debugf("Fetching assignment %d", assignmentID)
		selectedAssignment, err = db.GetAssignment(&pb.Assignment{ID: assignmentID})
		if err != nil {
			logger.Error("Failed to fetch assignment by ID: ", zap.Error(err))
			return
		}
		logger.Debugf("Found assignment %v", selectedAssignment)
	} else {
		//TODO(meling) this else-clause can probably be removed, since assignmentID should always be non-zero; doublecheck
		// otherwise use the last unapproved assignment for the given student/group
		logger.Debugf("Next assignment for course %d, user %d, group %d", course.ID, repo.UserID, repo.GroupID)
		selectedAssignment, err = db.GetNextAssignment(course.ID, repo.UserID, repo.GroupID)
		if err != nil {
			logger.Error("Failed to find a next unapproved assignment", zap.Error(err))
			return
		}
		logger.Debugf("Found next assignment %v", selectedAssignment)
	}

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

	secret := randomSecret()
	info := &ci.AssignmentInfo{
		CreatorAccessToken: courseCreator.RemoteIdentities[0].AccessToken,
		AssignmentName:     selectedAssignment.Name,
		Language:           selectedAssignment.Language,
		GetURL:             getURL,
		TestURL:            getURLTest,
		RawGetURL:          strings.TrimPrefix(strings.TrimSuffix(getURL, ".git"), "https://"),
		RawTestURL:         strings.TrimPrefix(strings.TrimSuffix(getURLTest, ".git"), "https://"),
		RandomSecret:       secret,
	}

	job, err := ci.ParseScriptTemplate(scriptPath, info)
	if err != nil {
		logger.Error("Failed to parse script template", zap.Error(err))
		return
	}

	authorName := author + "-" + randomSecret()[0:9]
	start := time.Now()
	logger.Debug("Job started successfully for " + author)
	out, err := runner.Run(context.Background(), job, authorName)
	if err != nil {
		logger.Error("Docker execution failed", zap.Error(err))
		return
	}
	execTime := time.Since(start)

	result, err := ci.ExtractResult(logger, out, secret, execTime)
	if err != nil {
		logger.Error("Failed to extract results from log", zap.Error(err))
		return
	}
	buildInfo, scores, err := result.Marshal()
	if err != nil {
		logger.Error("Failed to marshal build info and scores", zap.Error(err))
		//TODO(meling) should this return?
	}

	// check the approved status for the last submission
	logger.Debugf("Fetching submission for assignment %d", selectedAssignment.GetID())
	lastSubmission, err := db.GetSubmission(&pb.Submission{AssignmentID: selectedAssignment.GetID(), UserID: repo.GetUserID(), GroupID: repo.GetGroupID()})
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.Error("Failed to get submission info from the database", zap.Error(err))
		return
	}

	var approve = false
	// approve if the previous submission has already been approved
	if lastSubmission != nil {
		approve = lastSubmission.GetApproved()
		logger.Debugf("Found last submission for assignment %d: %v", selectedAssignment.GetID(), lastSubmission)
	}

	// for auto approve, use default score limit unless defined in yaml file
	lowerLimit := uint8(selectedAssignment.GetScoreLimit())
	if lowerLimit < 1 {
		lowerLimit = defaultAutoApproveScoreLimit
	}

	if selectedAssignment.AutoApprove && result.TotalScore() >= lowerLimit {
		approve = true
	}

	err = db.CreateSubmission(&pb.Submission{
		AssignmentID: selectedAssignment.ID,
		BuildInfo:    buildInfo,
		CommitHash:   commitHash,
		Score:        uint32(result.TotalScore()),
		ScoreObjects: scores,
		UserID:       repo.UserID,
		GroupID:      repo.GroupID,
		Approved:     approve,
	})
	if err != nil {
		logger.Error("Failed to add submission to database", zap.Error(err))
		return
	}
	logger.Debugf("Created submission for assignment %d in database", selectedAssignment.GetID())
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
func GitlabHook(logger *zap.SugaredLogger) webhooks.ProcessPayloadFunc {
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

func contains(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}
