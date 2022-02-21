package assignments

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/scm"
	"go.uber.org/zap"
)

// newTask returns a task from markdown contents and associates it with the given assignment.
// The provided markdown contents must contain a title specified on the first line,
// starting with the "# " character sequence, followed by two new line characters.
func newTask(contents []byte, assignment *pb.Assignment, name string) (*pb.Task, error) {
	if !bytes.HasPrefix(contents, []byte("# ")) {
		return nil, fmt.Errorf("task for assignment %s does not start with a # title marker", assignment.Name)
	}
	bodyIndex := bytes.Index(contents, []byte("\n\n"))
	if bodyIndex == -1 {
		return nil, fmt.Errorf("failed to find task body in %s", assignment.Name)
	}
	return &pb.Task{
		AssignmentID: uint64(assignment.Order),
		Title:        string(contents[2:bodyIndex]),
		Body:         string(contents[bodyIndex+2:]),
		Name:         name,
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

// Creates an issue on specified repository. Also creates and returns a dbIssue
func CreateIssue(c context.Context, sc scm.SCM, course *pb.Course, db database.Database, scmRepo *scm.Repository, dbRepo *pb.Repository, task *pb.Task) (*scm.Issue, *pb.Issue, error) {
	newIssue := &scm.CreateIssueOptions{
		Organization: course.Name,
		Repository:   scmRepo.Path,
		Title:        task.Title,
		Body:         task.Body,
	}
	scmIssue, err := sc.CreateIssue(c, newIssue)
	if err != nil {
		return nil, nil, err
	}

	dbIssue := &pb.Issue{
		RepositoryID:       dbRepo.ID,
		GithubRepositoryID: scmRepo.ID,
		Name:               task.Name,
		Title:              task.Title,
		Body:               task.Body,
	}

	// TODO(meling) maybe these need to be recorded in the database; in which case, maybe this should be done outside this function?
	// task.GitIssueID = issue.ID
	// task.IssueNumber = uint32(issue.IssueNumber)
	// task.Status = issue.Status
	return scmIssue, dbIssue, nil
}

// SyncTasks will create Issues in all the git repositories within an Organization
// It will exclude only repository with suffix -info
// It will also update all the existing issues within all the repositories
func SyncTasks(c context.Context, logger *zap.SugaredLogger, sc scm.SCM, course *pb.Course, assignments []*pb.Assignment) error {
	logger.Debugf("SyncTasks: Syncing tasks for all the assignments for Course: %s", course.Name)
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
			logger.Debugf("SyncTasks: Skipping repository: %s", repo.Path)
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
		logger.Debugf("SyncTasks: Beggining task creation on repository: %s", repo.Path)
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
					// issue already exists
					logger.Debugf("SyncTasks: updating Task on Repository: %s", repo.Path)
					_, err := UpdateIssue(c, sc, course, repo, task, gitIssue)
					if err != nil {
						logger.Errorf("SyncTasks: failed to update task %s on repo %s for course %s : %s", task.Title, repo.Path, course.Name, err)
					}
				} else {
					logger.Debugf("SyncTasks: Creating Task on Repository: %s", repo.Path)
					// issue does not exist, creating new issue on current repository
					// _, err := CreateIssue(c, sc, course, repo, task)
					if err != nil {
						logger.Errorf("SyncTasks: failed to create new task %s on repo %s for course %s : %s", task.Title, repo.Path, course.Name, err)
					}
				}
			}
		}
	}
	return nil
}

// Oje - Imagined issue/PR management flow goes as follows:
// 1. Org is created and info/tests/assignments repos are created, as well as "-labs" repos as students enroll
// 2. Teacher updates "tests" repo with assignment. The assignment contains "task-*.md" files.
// 3. As this update happens, a hook will be sent to QF server stating that someone pushed to the tests repo.
// 4. A function/method is triggered on said hook with the purpose of updating all "-labs" repos.
// 5. The function/method goes through all "-labs" repos, creating a new issue for each "task-*.md" file.
// 		- A number of things need to be accounted for here. What should happen if a the user already has an issue corresponding to a given task?
//		  How does one check that an existing task already has an issue associated with it on the users repo.
//		- What happens if the teacher pushes an assignment with "task-*.md" files, creating isssues on all "-labs" repos, but then a user
//		  that had not yet enrolled to the course does so? The user would then be stuck without an issue on said tasks.
// 6. When a user wants their code reviewed, they create a PR, which must in turn be associated with the issue/task they want reviewed.
//		- When a user has closed an issue/task because it was completed. We must make sure that this process does not create a new
//		  issue.

// Oje - Todo list:
// - Should create a test that creates issues on repos, and then checks if these can be associated with existing tasks
// - Make a function that converts from scm-repo to db-repo. Then CreateIssue can take both as argument

// Following is Oje code (placement might be temporary):

func HandleTasks(c context.Context, logger *zap.SugaredLogger, db database.Database, s scm.SCM, course *pb.Course, assignments []*pb.Assignment) error {
	if len(assignments) == 0 {
		return nil
	}
	org, err := s.GetOrganization(c, &scm.GetOrgOptions{Name: course.Name})
	if err != nil {
		return err
	}

	repos, err := s.GetRepositories(c, org)
	if err != nil {
		return err
	}

	for _, scmRepo := range repos {
		// Could maybe get DBRepo first, and then do this test
		if !strings.HasSuffix(scmRepo.Path, "-labs") {
			continue
		}

		dbRepo, err := GetDbRepository(logger, db, scmRepo)
		if err != nil {
			return err
		}

		issues := []*pb.Issue{}
		for _, assignment := range assignments {
			for _, task := range assignment.Tasks {
				_, issue, err := CreateIssue(c, s, course, db, scmRepo, dbRepo, task)
				if err != nil {
					return err
				}
				issues = append(issues, issue)
			}
		}
		UpdateRepositoryIssues(logger, db, dbRepo, issues)
	}

	return nil
}

func UpdateRepositoryIssues(logger *zap.SugaredLogger, db database.Database, repo *pb.Repository, issues []*pb.Issue) error {
	err := db.UpdateRepositoryIssues(repo, issues)
	if err != nil {
		return err
	}
	return nil
}

// Gets dbRepo based on scmRepo
func GetDbRepository(logger *zap.SugaredLogger, db database.Database, scmRepo *scm.Repository) (*pb.Repository, error) {
	repositories, err := db.GetRepositories(&pb.Repository{
		RepositoryID: scmRepo.ID,
	})
	if err != nil {
		return nil, err
	}
	if len(repositories) > 1 {
		// Should only get one repository. Should return a fitting error
		return nil, nil
	}

	return repositories[0], nil
}
