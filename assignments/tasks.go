package assignments

import (
	"bufio"
	"context"
	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/scm"
	"go.uber.org/zap"
	"os"
	"strings"
)

//func tasks_parser(contents []byte) ([]*pb.Task, error) {
//	//TODO Add parsing of task .md File here
//	var t []*pb.Task
//	t = append(t, &pb.Task{
//		ID:    rand.Uint64(),
//		Title: "title",
//		Body:  string(contents),
//	})
//	return t, nil
//}

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
		//checking if it's a course info repository
		if strings.HasSuffix(repo.Path, "-info") {
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

		for _, assignment := range assignments {
			for _, task := range assignment.Tasks {
				// Checking if issue already exist
				gitIssue, taskIssue := isExists(gitIssues, task)
				if gitIssue != nil && taskIssue != nil {
					// issue already exist
					_, err := UpdateIssue(c, sc, course, repo, task, gitIssue, taskIssue)
					if err != nil {
						logger.Errorf("SyncTasks: failed to update task %s on repo %s for course %s : %s", task.Title, repo.Path, course.Name, err)
					}
				} else {
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
