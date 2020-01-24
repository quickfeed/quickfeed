package hooks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/database"
	"github.com/google/go-github/v29/github"
	"go.uber.org/zap"
)

type GitHubWebHook struct {
	logger *zap.SugaredLogger
	db     database.Database
	runner ci.Runner
	secret string
}

func NewGitHubWebHook(logger *zap.SugaredLogger, db database.Database, runner ci.Runner, secret string) *GitHubWebHook {
	return &GitHubWebHook{logger: logger, db: db, runner: runner, secret: secret}
}

func (wh GitHubWebHook) Handle(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(wh.secret))
	if err != nil {
		wh.logger.Errorf("error in request body: %w", err)
		return
	}
	defer r.Body.Close()

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		wh.logger.Errorf("could not parse github webhook: %w", err)
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
	repo, err := wh.db.GetRepositoryByRemoteID(uint64(payload.GetRepo().GetID()))
	if err != nil {
		wh.logger.Error("Failed to get repository from database", zap.Error(err))
		return
	}
	wh.logger.Debugf("Received Push Event for repository %v", repo)
	course, err := wh.db.GetCourseByOrganizationID(repo.OrganizationID)
	if err != nil {
		wh.logger.Error("Failed to get course from database", zap.Error(err))
		return
	}
	wh.logger.Debugf("For course(%d)=%v", course.GetID(), course.GetName())
	switch {
	case repo.IsTestsRepo():
		// the push event is for the 'tests' repo, which means that we
		// should update the course data (assignments) in the database
		// refreshAssignmentsFromTestsRepo(logger, db, repo, uint64(p.Sender.ID))

	case repo.IsStudentRepo():
		assignments, err := wh.extractAssignments(payload, course)
		if err != nil {
			wh.logger.Error(err)
			return
		}
		for _, assignment := range assignments {
			// TODO(meling) actually run tests
			wh.logger.Debugf("Running tests for %v", assignment.GetName())
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
// TODO(meling) consider to call runTests() in parallel?
// TODO(meling) implement test cases for this function.
// TODO(meling) Thinking: One complication with this approach is that we depend on the YAML's 'name' field
//   being the same as the assignment name in the folder structure in the assignments repository.
//   This is perhaps fine, but could be problematic if someone uses a name like "Lab assignment 1"
//   and the folder is named only "lab1". We should make this more robust; can we add a field to the
//   pb.Assignment type to hold the directory name, which should not be parsed from YAML, but computed
//   in assignment_parser.go, based on parent directory of the YAML. Issue is that we may need to add it to the DB.
func (wh GitHubWebHook) extractAssignments(payload *github.PushEvent, course *pb.Course) ([]*pb.Assignment, error) {
	modifiedAssignments := make(map[string]bool)
	for c, commit := range payload.Commits {
		for i, modifiedFile := range commit.Modified {
			// we assume the first path component holds the assignment name
			name := strings.Split(modifiedFile, "/")[0]
			modifiedAssignments[name] = true
			wh.logger.Debugf("commit %d (%s), file %d: %s", c, commit.GetID(), i, modifiedFile)
		}
	}

	var assignments []*pb.Assignment
	for name := range modifiedAssignments {
		// get assignment based on course id and assignment name
		assignment, err := wh.db.GetAssignment(&pb.Assignment{Name: name, CourseID: course.GetID()})
		if err != nil {
			return nil, fmt.Errorf("could not find assignment '%s' for course %d in database: %v", name, course.GetID(), err)
		}
		assignments = append(assignments, assignment)
	}
	return assignments, nil
}
