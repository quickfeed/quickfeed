package assignments

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/scm"
	"go.uber.org/zap"
)

// newTask returns a task from markdown contents and associates it with the given assignment.
// The provided markdown contents must contain a title specified on the first line,
// starting with the "# " character sequence, followed by two new line characters.
func newTask(contents []byte, assignment *pb.Assignment) (*pb.Task, error) {
	if !bytes.HasPrefix(contents, []byte("# ")) {
		return nil, fmt.Errorf("task for assignment %s does not start with a # title marker", assignment.Name)
	}
	bodyIndex := bytes.Index(contents, []byte("\n\n"))
	if bodyIndex == -1 {
		return nil, fmt.Errorf("failed to find task body in %s", assignment.Name)
	}
	return &pb.Task{
		AssignmentID: assignment.ID,
		Title:        string(contents[2:bodyIndex]),
		Body:         string(contents[bodyIndex+2:]),
	}, nil
}

// TODO(meling) consider to move this as method on pb.Task??
func isExists(gitIssues []*scm.Issue, task *pb.Task) (gitIssue *scm.Issue) {
	for _, gitIssue := range gitIssues {
		if gitIssue.ID == task.GitIssueID {
			return gitIssue
		}
	}
	return nil
}

func UpdateIssue(c context.Context, sc scm.SCM, course *pb.Course, repo *scm.Repository, task *pb.Task, gitIssue *scm.Issue) (issue *scm.Issue, err error) {
	newIssue := &scm.CreateIssueOptions{
		Organization: course.Name,
		Repository:   repo.Path,
		Title:        task.Title,
		Body:         task.Body,
	}
	updateIssue := &scm.IssueOptions{
		Organization: course.Name,
		Repository:   repo.Path,
		IssueNumber:  int(gitIssue.IssueNumber),
	}
	return sc.EditRepoIssue(c, updateIssue, newIssue)
}

func CreateIssue(c context.Context, sc scm.SCM, course *pb.Course, repo *scm.Repository, task *pb.Task) (issue *scm.Issue, err error) {
	newIssue := &scm.CreateIssueOptions{
		Organization: course.Name,
		Repository:   repo.Path,
		Title:        task.Title,
		Body:         task.Body,
	}
	issue, err = sc.CreateIssue(c, newIssue)
	if err != nil {
		return nil, err
	}
	// TODO(meling) maybe these need to be recorded in the database; in which case, maybe this should be done outside this function?
	task.GitIssueID = issue.ID
	task.IssueNumber = uint32(issue.IssueNumber)
	task.Status = issue.Status
	return issue, nil
}

// SyncTasks will create Issues in all the git repositories with in an Organization
// It will exclude only  repository with suffix -info
// It will also update all the existing issues with in all the repositories
func SyncTasks(c context.Context, logger *zap.SugaredLogger, sc scm.SCM, course *pb.Course, assignments []*pb.Assignment) error {
	logger.Debugf("SyncTasks: Syncing tasks  for all the assignments for Course: %s", course.Name)
	org, err := sc.GetOrganization(c, &scm.GetOrgOptions{Name: course.Name})
	if err != nil {
		logger.Debugf("SyncTasks: Failed to Fetch Course %s due to ERROR : %s", course.Name, err)
		return err
	}
	repos, err := sc.GetRepositories(c, org)
	if err != nil {
		logger.Errorf("SyncTasks: Failed to Fetch Repositories from Course %s due to ERROR: %s", course.Name, err)
		return err
	}

	for _, repo := range repos {
		// checking if it's a course info repository
		if !strings.HasSuffix(repo.Path, "-labs") {
			logger.Debugf("SyncTasks: Skippig these repositories in task creation : %s", repo.Path)
			// not adding issues on course info repository
			continue
		}
		// check if issues already exist
		gitIssues, err := sc.GetRepoIssues(c, &scm.IssueOptions{
			Organization: course.Name,
			Repository:   repo.Path,
		})
		if err != nil {
			logger.Errorf("SyncTasks: Not able to fetch Issues from repository %s, Course %s ", repo.Path, course.Name)
			continue
		}
		logger.Debugf("SyncTasks: Checking Task Creation on Repository: %s", repo.Path)
		for _, assignment := range assignments {
			logger.Debugf("SyncTasks: assignment Elements: %s", assignment)
			logger.Debugf("SyncTasks: assignment TASKS: %s", assignment.Tasks)
			for _, task := range assignment.Tasks {
				// Checking if issue already exist
				logger.Debugf("SyncTasks: assignment.Tasks: %s", assignment.Tasks)
				gitIssue := isExists(gitIssues, task)
				logger.Debugf("SyncTasks: gitIssue: %s", gitIssue)
				logger.Debugf("SyncTasks: task: %s", task)
				if gitIssue != nil {
					// issue already exist
					logger.Debugf("SyncTasks: updating Task on Repository: %s", repo.Path)
					_, err := UpdateIssue(c, sc, course, repo, task, gitIssue)
					if err != nil {
						logger.Errorf("SyncTasks: failed to update task %s on repo %s for course %s : %s", task.Title, repo.Path, course.Name, err)
					}
				} else {
					logger.Debugf("SyncTasks: Creating Task on Repository: %s", repo.Path)
					// issue does not exist, creating new issue on current repository
					_, err := CreateIssue(c, sc, course, repo, task)
					if err != nil {
						logger.Errorf("SyncTasks: failed to create new task %s on repo %s for course %s : %s", task.Title, repo.Path, course.Name, err)
					}
				}
			}
		}
	}
	return nil
}
