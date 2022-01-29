package assignments

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/scm"
	"go.uber.org/zap"
)

func readTaskFiles(path string) (*pb.Task, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	task := &pb.Task{}

	sc := bufio.NewScanner(f)
	title_flag := true
	for sc.Scan() {
		line := sc.Text() // GET the line string
		if strings.HasPrefix(line, "#") && title_flag {
			task.Title = line
			title_flag = false
		} else {
			task.Body = task.Body + "\n" + line
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return task, nil
}

func isExists(gitIssues []*scm.Issue, task *pb.Task) (gitIssue *scm.Issue, taskIssue *pb.Issue) {
	for _, taskIssue = range task.Issues {
		for _, gitIssue = range gitIssues {
			if taskIssue.GitIssueID == gitIssue.ID {
				return gitIssue, taskIssue
			}
		}
	}
	return nil, nil
}

func findTasksFiles(dir string) ([]*pb.Task, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, err
	}
	var taskContents []*pb.Task
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Walk unable to read path; stop walking the tree
			return err
		}
		if !info.IsDir() {
			filename := filepath.Base(path)
			if strings.HasSuffix(filename, taskFile) {
				task, err := readTaskFiles(path)
				if err != nil {
					return err
				}
				taskContents = append(taskContents, task)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return taskContents, nil
}

func UpdateIssue(c context.Context, sc scm.SCM, course *pb.Course, repo *scm.Repository, task *pb.Task, gitIssue *scm.Issue, taskIssue *pb.Issue) (issue *scm.Issue, err error) {
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
	Issue, err := sc.EditRepoIssue(c, updateIssue, newIssue)
	return Issue, err
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
	taskIssue := &pb.Issue{
		GitIssueID:  issue.ID,
		TaskID:      task.ID,
		IssueNumber: uint32(issue.IssueNumber),
		Status:      issue.Status,
	}
	task.Issues = append(task.Issues, taskIssue)
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
				gitIssue, taskIssue := isExists(gitIssues, task)
				logger.Debugf("SyncTasks: gitIssue: %s", gitIssue)
				logger.Debugf("SyncTasks: taskIssue: %s", taskIssue)
				if gitIssue != nil && taskIssue != nil {
					// issue already exist
					logger.Debugf("SyncTasks: updating Task on Repository: %s", repo.Path)
					_, err := UpdateIssue(c, sc, course, repo, task, gitIssue, taskIssue)
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
