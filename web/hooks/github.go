package hooks

import (
	"net/http"
	"strings"
	"time"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/assignments"
	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/log"
	"github.com/google/go-github/v35/github"
	"go.uber.org/zap"
)

// GitHubWebHook holds references and data for handling webhook events.
type GitHubWebHook struct {
	logger *zap.SugaredLogger
	db     database.Database
	runner ci.Runner
	secret string
}

// NewGitHubWebHook creates a new webhook to handle POST requests from GitHub to the QuickFeed server.
func NewGitHubWebHook(logger *zap.SugaredLogger, db database.Database, runner ci.Runner, secret string) *GitHubWebHook {
	return &GitHubWebHook{logger: logger, db: db, runner: runner, secret: secret}
}

// Handle take POST requests from GitHub, representing Push events
// associated with course repositories, which then triggers various
// actions on the QuickFeed backend.
func (wh GitHubWebHook) Handle(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(wh.secret))
	if err != nil {
		wh.logger.Errorf("Error in request body: %v", err)
		return
	}
	defer r.Body.Close()

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		wh.logger.Errorf("Could not parse github webhook: %v", err)
		return
	}
	switch e := event.(type) {
	case *github.PushEvent:
		wh.logger.Debug(log.IndentJson(e))
		wh.handlePush(e)
	default:
		wh.logger.Debugf("Ignored event type %s", github.WebHookType(r))
	}
}

func (wh GitHubWebHook) handlePush(payload *github.PushEvent) {
	wh.logger.Debugf("Received push event for branch reference: %s (user's default branch: %s)",
		payload.GetRef(), payload.GetRepo().GetDefaultBranch())
	if !strings.HasSuffix(payload.GetRef(), payload.GetRepo().GetDefaultBranch()) {
		wh.logger.Debugf("Ignoring push event for non-default branch: %s", payload.GetRef())
		return
	}

	repos, err := wh.db.GetRepositories(&pb.Repository{RepositoryID: uint64(payload.GetRepo().GetID())})
	if err != nil {
		wh.logger.Errorf("Failed to get repository by remote ID %d from database: %v", payload.GetRepo().GetID(), err)
		return
	}
	if len(repos) != 1 {
		wh.logger.Debugf("Ignoring push event for unknown repository: %s", payload.GetRepo().GetFullName())
		return
	}
	repo := repos[0]
	wh.logger.Debugf("Received push event for repository %v", repo)

	course, err := wh.db.GetCourseByOrganizationID(repo.OrganizationID)
	if err != nil {
		wh.logger.Errorf("Failed to get course from database: %v", err)
		return
	}
	wh.logger.Debugf("For course(%d)=%v", course.GetID(), course.GetName())

	switch {
	case repo.IsTestsRepo():
		// the push event is for the 'tests' repo, which means that we
		// should update the course data (assignments) in the database
		assignments.UpdateFromTestsRepo(wh.logger, wh.db, course)

	case repo.IsUserRepo():
		wh.logger.Debugf("Processing push event for user repo %s", payload.GetRepo().GetName())
		wh.updateLastActivityDate(repo.UserID, course.ID)
		assignments := wh.extractAssignments(payload, course)
		for _, assignment := range assignments {
			if !assignment.IsGroupLab {
				// only run non-group assignments
				wh.runAssignmentTests(assignment, repo, course, payload)
			} else {
				wh.logger.Debugf("Ignoring assignment: %s, pushed to user repo: %s", assignment.GetName(), payload.GetRepo().GetName())
			}
		}

	case repo.IsGroupRepo():
		wh.logger.Debugf("Processing push event for group repo %s", payload.GetRepo().GetName())
		jobOwner, _, err := wh.db.GetUserByCourse(course, payload.GetSender().GetLogin())
		if err != nil {
			wh.logger.Errorf("Failed to find user %s in course %s: %v", payload.GetSender().GetLogin(), course.GetName(), err)
			return
		}
		wh.updateLastActivityDate(jobOwner.ID, course.ID)
		assignments := wh.extractAssignments(payload, course)
		for _, assignment := range assignments {
			if assignment.IsGroupLab {
				// only run group assignments
				wh.runAssignmentTests(assignment, repo, course, payload)
			} else {
				wh.logger.Debugf("Ignoring assignment: %s, pushed to group repo: %s", assignment.GetName(), payload.GetRepo().GetName())
			}
		}

	default:
		wh.logger.Debug("Nothing to do for this push event")
	}
}

// extractAssignments extracts information from the push payload from github
// and determines the assignments that have been changed in this commit by
// querying the database based on the lab name.
func (wh GitHubWebHook) extractAssignments(payload *github.PushEvent, course *pb.Course) []*pb.Assignment {
	modifiedAssignments := make(map[string]bool)
	for _, commit := range payload.Commits {
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

// runAssignmentTests runs the tests for the given assignment pushed to repo.
func (wh GitHubWebHook) runAssignmentTests(assignment *pb.Assignment, repo *pb.Repository, course *pb.Course, payload *github.PushEvent) {
	runData := &ci.RunData{
		Course:     course,
		Assignment: assignment,
		Repo:       repo,
		CommitID:   payload.GetHeadCommit().GetID(),
		JobOwner:   payload.GetSender().GetLogin(),
	}
	if assignment.GradedManually() {
		wh.logger.Debugf("Assignment %s for course %s is manually reviewed", assignment.Name, course.Name)
		if _, err := runData.RecordResults(wh.logger, wh.db, nil); err != nil {
			wh.logger.Error(err)
		}
		return
	}
	ctx, cancel := assignment.WithTimeout(ci.DefaultContainerTimeout)
	defer cancel()
	results, err := runData.RunTests(ctx, wh.logger, wh.runner)
	if err != nil {
		wh.logger.Errorf("Failed to run tests for assignment %s for course %s: %v", assignment.Name, course.Name, err)
	}
	wh.logger.Debug("ci.RunTests", zap.Any("Results", log.IndentJson(results)))
	if _, err = runData.RecordResults(wh.logger, wh.db, results); err != nil {
		wh.logger.Error(err)
	}
}

// updateLastActivityDate sets a current date as a last activity date of the student
// on each new push to the student repository.
func (wh GitHubWebHook) updateLastActivityDate(userID, courseID uint64) {
	query := &pb.Enrollment{
		UserID:           userID,
		CourseID:         courseID,
		LastActivityDate: time.Now().Format("02 Jan"),
	}

	if err := wh.db.UpdateEnrollment(query); err != nil {
		wh.logger.Errorf("Failed to update the last activity date for user %d: %v", userID, err)
	}
}
