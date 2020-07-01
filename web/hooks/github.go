package hooks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/assignments"
	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/database"
	"github.com/google/go-github/v30/github"
	"go.uber.org/zap"
)

// GitHubWebHook holds references and data for handling webhook events.
type GitHubWebHook struct {
	logger *zap.SugaredLogger
	db     database.Database
	runner ci.Runner
	secret string
}

// NewGitHubWebHook creates a new webhook to handle POST requests from GitHub to the Autograder server.
func NewGitHubWebHook(logger *zap.SugaredLogger, db database.Database, runner ci.Runner, secret string) *GitHubWebHook {
	return &GitHubWebHook{logger: logger, db: db, runner: runner, secret: secret}
}

// Handle take POST requests from GitHub, representing Push events
// associated with course repositories, which then triggers various
// actions on the Autograder backend.
func (wh GitHubWebHook) Handle(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(wh.secret))
	if err != nil {
		wh.logger.Errorf("Error in request body: %w", err)
		return
	}
	defer r.Body.Close()

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		wh.logger.Errorf("Could not parse github webhook: %w", err)
		return
	}
	switch e := event.(type) {
	case *github.PushEvent:
		wh.logger.Debug(jsonString(e))
		wh.handlePush(e)
	default:
		wh.logger.Debugf("Ignored event type %s", github.WebHookType(r))
	}
}

func (wh GitHubWebHook) handlePush(payload *github.PushEvent) {
	if payload.GetRef() != "refs/heads/master" {
		wh.logger.Debugf("Ignoring push event for non-master branch: %s", payload.GetRef())
		return
	}

	repo, err := wh.db.GetRepositoryByRemoteID(uint64(payload.GetRepo().GetID()))
	if err != nil {
		wh.logger.Errorf("Failed to get repository from database: %w", err)
		return
	}
	wh.logger.Debugf("Received Push Event for repository %v", repo)

	course, err := wh.db.GetCourseByOrganizationID(repo.OrganizationID)
	if err != nil {
		wh.logger.Errorf("Failed to get course from database: %w", err)
		return
	}
	wh.logger.Debugf("For course(%d)=%v", course.GetID(), course.GetName())

	switch {
	case repo.IsTestsRepo():
		// the push event is for the 'tests' repo, which means that we
		// should update the course data (assignments) in the database
		assignments.UpdateFromTestsRepo(wh.logger, wh.db, repo, course)

	case repo.IsStudentRepo():
		wh.logger.Debugf("Processing push event for %s", payload.GetRepo().GetName())
		assignments := wh.extractAssignments(payload, course)
		for _, assignment := range assignments {
			runData := &ci.RunData{
				Course:     course,
				Assignment: assignment,
				Repo:       repo,
				CloneURL:   payload.GetRepo().GetCloneURL(),
				CommitID:   payload.GetHeadCommit().GetID(),
				JobOwner:   payload.GetSender().GetLogin(),
			}
			ci.RunTests(wh.logger, wh.db, wh.runner, runData)
		}

	default:
		wh.logger.Debug("Nothing to do for this push event")
	}
}

// jsonString returns a JSON formatted string
// with structured indents and line breaks.
func jsonString(event interface{}) string {
	prettyJSON, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return fmt.Sprintf("JSON error: %v", err)
	}
	return string(prettyJSON)
}

// extractAssignments extracts information from the push payload from github
// and determines the assignments that have been changed in this commit by
// querying the database based on the lab name.
func (wh GitHubWebHook) extractAssignments(payload *github.PushEvent, course *pb.Course) []*pb.Assignment {
	modifiedAssignments := make(map[string]bool)
	for _, commit := range payload.Commits {
		wh.logger.Debugf("Examining commit (%s) for modifications/additions/removals", commit.GetID())
		extractChanges(commit.Modified, modifiedAssignments)
		extractChanges(commit.Added, modifiedAssignments)
		extractChanges(commit.Removed, modifiedAssignments)
	}

	var assignments []*pb.Assignment
	for name := range modifiedAssignments {
		// get assignment based on course id and assignment name
		assignment, err := wh.db.GetAssignment(&pb.Assignment{Name: name, CourseID: course.GetID()})
		if err != nil {
			wh.logger.Errorf("Could not find assignment '%s' for course %d in database: %v", name, course.GetID(), err)
			continue
		}
		assignments = append(assignments, assignment)
	}
	return assignments
}

func extractChanges(changes []string, modifiedAssignments map[string]bool) {
	for _, changedFile := range changes {
		index := strings.Index(changedFile, "/")
		if index == -1 {
			// ignore root-level files
			continue
		}
		// we assume the first path component holds the assignment name
		name := changedFile[:index]
		if name == "" {
			// ignore names that start with "/" or empty names
			continue
		}
		modifiedAssignments[name] = true
	}
}
