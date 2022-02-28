package assignments

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/scm"
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

// Creates an issue on specified repository.
func CreateScmIssue(c context.Context, sc scm.SCM, course *pb.Course, repo *pb.Repository, task *pb.Task) (*scm.Issue, error) {
	newIssue := &scm.CreateIssueOptions{
		Organization: course.Name,
		Repository:   filepath.Base(repo.GetHTMLURL()), // Needs to be of type "tests", not "https://github.com/qf101/tests". This is a very hacky solution. pb.Repository should probably have a field "Name" that is set upon creation.
		Title:        task.Title,
		Body:         task.Body,
	}
	issue, err := sc.CreateIssue(c, newIssue)
	if err != nil {
		return nil, err
	}
	return issue, nil
}

// This is more of a converter function, and cannot currently return an error. Should probably be renamed or something.
func CreateDbIssue(c context.Context, repo *pb.Repository, task *pb.Task) (*pb.Issue, error) {
	issue := &pb.Issue{
		RepositoryID:       repo.ID,
		GithubRepositoryID: repo.RepositoryID,
		Name:               task.Name,
		Title:              task.Title,
		Body:               task.Body,
	}
	return issue, nil
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
// - Currently there is a db-record for tasks, it is however not used, but the struct is used. Since we now have a db-record for issues,
// 	 we should not need the db-record for tasks, however the struct will be necessary. Therefore we need an equivalent

// Following is Oje code (placement might be temporary):

func HandleTasks(c context.Context, db database.Database, s scm.SCM, course *pb.Course, assignments []*pb.Assignment) error {
	if len(assignments) == 0 {
		return nil
	}
	org, err := s.GetOrganization(c, &scm.GetOrgOptions{Name: course.Name})
	if err != nil {
		return err
	}

	repos, err := GetRepositoriesByOrgID(db, org.ID)
	if err != nil {
		return err
	}

	for _, repo := range repos {
		if !repo.IsStudentRepo() {
			continue
		}

		tasks := make(map[string]*pb.Task)
		for _, assignment := range assignments {
			for _, task := range assignment.Tasks {
				tasks[task.Name] = task
			}
		}
		err = HandleTasksForRepo(c, db, course, s, repo, tasks)
		if err != nil {
			return err
		}
	}

	return nil
}

func HandleTasksForRepo(c context.Context, db database.Database, course *pb.Course, s scm.SCM, repo *pb.Repository, tasks map[string]*pb.Task) error {
	newOrAlteredIssues := []*pb.Issue{}
	for _, issue := range repo.Issues {
		task, ok := tasks[issue.Name]
		if !ok {
			// What should happen if task does not exist for issue?
			continue
		}
		if !(task.Title == issue.Title && task.Body == issue.Body) {
			// Issue needs to be updated here
			issue.Title = task.Title
			issue.Body = task.Body
			newOrAlteredIssues = append(newOrAlteredIssues, issue)
			// UpdateIssue(c, s, course, )
			continue
		}
		delete(tasks, issue.Name)
	}

	// Only tasks that have no issue associated with them remain. There must be created an issue for them.
	for _, task := range tasks {
		// Creates the actual issue on a scm
		_, err := CreateScmIssue(c, s, course, repo, task)
		if err != nil {
			return err
		}
		// Creates issue to be saved in db
		issue, err := CreateDbIssue(c, repo, task)
		if err != nil {
			return err
		}
		newOrAlteredIssues = append(newOrAlteredIssues, issue)
	}
	// This creates new record instead of updating existing one. TBC
	UpdateRepositoryIssues(db, repo, newOrAlteredIssues)
	return nil
}

func UpdateRepositoryIssues(db database.Database, repo *pb.Repository, issues []*pb.Issue) error {
	err := db.UpdateRepositoryIssues(repo, issues)
	if err != nil {
		return err
	}
	return nil
}

// Gets dbRepo based on orgID. Should maybe be moved to be a separate method in gormdb_repository.go of some kind
func GetRepositoriesByOrgID(db database.Database, orgID uint64) ([]*pb.Repository, error) {
	repositories, err := db.GetRepositoriesWithIssues(&pb.Repository{
		OrganizationID: orgID,
	})
	if err != nil {
		return nil, err
	}

	return repositories, nil
}
