package hooks

import (
	"context"
	"errors"
	"strings"
	"time"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/assignments"
	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/kit/score"
	"github.com/autograde/quickfeed/log"
	"github.com/autograde/quickfeed/scm"
	"github.com/google/go-github/v35/github"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (wh GitHubWebHook) handlePush(payload *github.PushEvent) {
	wh.logger.Debugf("Received push event for branch reference: %s (user's default branch: %s)",
		payload.GetRef(), payload.GetRepo().GetDefaultBranch())

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
		if !isDefaultBranch(payload) {
			wh.logger.Debugf("Ignoring push event for non-default branch: %s", payload.GetRef())
			return
		}
		// the push event is for the 'tests' repo, which means that we
		// should update the course data (assignments) in the database
		assignments.UpdateFromTestsRepo(wh.logger, wh.db, course)

	case repo.IsUserRepo():
		wh.logger.Debugf("Processing push event for user repo %s", payload.GetRepo().GetName())
		if !isDefaultBranch(payload) {
			wh.logger.Debugf("Ignoring push event for non-default branch: %s", payload.GetRef())
			return
		}
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
				results := wh.runAssignmentTests(assignment, repo, course, payload)
				if !isDefaultBranch(payload) && !assignment.GradedManually() {
					// Attempt to find the pull request for the branch, if it exists,
					// and then assign reviewers to it, if the branch task score is higher than the assignment score limit
					wh.handlePullRequestPush(payload, results, assignment, course, repo)
				}
			} else {
				wh.logger.Debugf("Ignoring assignment: %s, pushed to group repo: %s", assignment.GetName(), payload.GetRepo().GetName())
			}
		}
	default:
		wh.logger.Debug("Nothing to do for this push event")
	}
}

// handlePullRequestPush attempts to find a pull requested associated with the non-default branch.
// If successfull, it then finds the relevant task, and uses it to receive the relevant task score.
// If a passing score is reached, it assigns reviewers to the pull request.
// It also uses the test results and task to generate a feedback comment for the pull request.
func (wh GitHubWebHook) handlePullRequestPush(payload *github.PushEvent, results *score.Results, assignment *pb.Assignment, course *pb.Course, repo *pb.Repository) {
	wh.logger.Debugf("Attempting to find pull request for ref: %s, in repository: %s",
		payload.GetRef(), payload.GetRepo().GetFullName())

	pullRequest, err := wh.db.GetPullRequest(&pb.PullRequest{
		SourceBranch:    branchName(payload.GetRef()),
		ScmRepositoryID: uint64(payload.GetRepo().GetID()),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// This can happen if someone pushes to a branch group assignment, without having a PR created for it
			// If this happens, QF should not do anything
			wh.logger.Debugf("No pull request found for ref: %s", payload.GetRef())
			return
		}
		wh.logger.Errorf("Failed to get pull request from database: %v", err)
		return
	}
	tasks, err := wh.db.GetTasks(&pb.Task{ID: pullRequest.GetTaskID()})
	if err != nil {
		// A pull request should always have a task association
		// If not, something must have gone wrong elsewhere
		wh.logger.Errorf("Failed to get task from the database: %v", err)
		return
	}
	if len(tasks) != 1 {
		// This should never happen
		wh.logger.Errorf("Got an unexpected number of tasks: %d", len(tasks))
		return
	}
	task := tasks[0]
	// TODO(meling): My idea is that when a teacher wants to assign a test to a specific task they will use the 'local' task name.
	// For example, if a teacher has created the markdown file task-hello_world.md,
	// they would do scores.AddWithTaskName(TestHelloWorld, "hello_world", max, weight).
	// Do you concur with this approach?
	// We could of course simply make teachers assign the global task name, but that seems somewhat counter-intuitive to me.

	// TODO(espeland): Revise this when score task name format has been decided.
	// Should also write some documentation on how a task in essence has two names.
	// One global and one local. Where the score package only uses the local name.
	taskSum := results.TaskSum(task.LocalName())

	// TODO(espeland): When the project is finished. Create a GitHub issue that states all places where
	// we need to update for GitHub apps. As it is, this would create comments as the course creator.
	sc, err := scm.NewSCMClient(wh.logger, course.GetProvider(), course.GetAccessToken())
	if err != nil {
		wh.logger.Errorf("Failed to create SCM Client: %v", err)
		return
	}
	ctx := context.Background()
	// We assign reviewers to a pull request when the tests associated with it score above the assignment score limit
	// We do not assign reviewers if the pull request has already been assigned reviewers
	if taskSum >= assignment.GetScoreLimit() && !pullRequest.HasReviewers() {
		wh.logger.Debugf("Assigning reviewers to pull request #%d, in repository: %s", pullRequest.GetNumber(), repo.Name())
		if err := assignments.AssignReviewers(ctx, sc, wh.db, course, repo, pullRequest); err != nil {
			wh.logger.Errorf("Failed to assign reviewers to pull request: %v", err)
			return
		}
	}

	// Create a test results feedback comment on the pull request
	opt := &scm.IssueCommentOptions{
		Organization: course.GetOrganizationPath(),
		Repository:   repo.Name(),
		Body:         assignments.CreateFeedbackComment(results, task, assignment),
	}
	wh.logger.Debugf("Creating feedback comment on pull request #%d, in repository: %s", pullRequest.GetNumber(), repo.Name())
	if !pullRequest.HasFeedbackComment() {
		commentID, err := sc.CreateIssueComment(ctx, int(pullRequest.Number), opt)
		if err != nil {
			wh.logger.Errorf("Failed to create feedback comment for pull request #%d, in repository", pullRequest.GetNumber(), repo.Name())
			return
		}
		pullRequest.ScmCommentID = commentID
		if err := wh.db.UpdatePullRequest(pullRequest); err != nil {
			wh.logger.Errorf("Failed to update pull request: %v", err)
			return
		}
	} else {
		if err := sc.EditIssueComment(ctx, int64(pullRequest.GetScmCommentID()), opt); err != nil {
			wh.logger.Errorf("Failed to update feedback comment for pull request #%d, in repository", pullRequest.GetNumber(), repo.Name())
			return
		}
	}
	wh.logger.Debugf("Successfully handled push to pull request #%d, in repository: %s", pullRequest.GetNumber(), repo.Name())
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
func (wh GitHubWebHook) runAssignmentTests(assignment *pb.Assignment, repo *pb.Repository, course *pb.Course, payload *github.PushEvent) *score.Results {
	runData := &ci.RunData{
		Course:     course,
		Assignment: assignment,
		Repo:       repo,
		// TODO(Espeland): If a student for some reason pushes to a remote branch with a different name than
		// the local one, this will fail.
		// E.g. if a student pushes to remote branch "branch1", but their code is on the local branch "branch2",
		// QF will try to checkout the wrong local branch.
		// The payload contains no information on the local branch, and it therefore seems like there is no good workaround.
		// We must make sure that if this scenario happens, QF can handle it.
		BranchName: branchName(payload.GetRef()),
		CommitID:   payload.GetHeadCommit().GetID(),
		JobOwner:   payload.GetSender().GetLogin(),
	}
	if assignment.GradedManually() {
		wh.logger.Debugf("Assignment %s for course %s is manually reviewed", assignment.Name, course.Name)
		if _, err := runData.RecordResults(wh.logger, wh.db, nil); err != nil {
			wh.logger.Error(err)
		}
		return nil
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
	return results
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

// branchName returns the branch name from a push event ref.
func branchName(ref string) string {
	components := strings.Split(ref, "/")
	return components[len(components)-1]
}

// isDefaultBranch returns true if a push event is for a repository's default branch.
func isDefaultBranch(payload *github.PushEvent) bool {
	return strings.HasSuffix(payload.GetRef(), payload.GetRepo().GetDefaultBranch())
}
