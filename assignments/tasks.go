package assignments

import (
	"bytes"
	"context"
	"fmt"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

// taskName returns the task name based on the filename
// excluding the task- prefix and the .md suffix.
func taskName(filename string) string {
	taskName := filename[len("task-"):]
	return taskName[:len(taskName)-len(".md")]
}

// newTask returns a task from markdown contents and associates it with the given assignment.
// The provided markdown contents must contain a title specified on the first line,
// starting with the "# " character sequence, followed by two new line characters.
func newTask(contents []byte, assignmentOrder uint32, name string) (*qf.Task, error) {
	if !bytes.HasPrefix(contents, []byte("# ")) {
		return nil, fmt.Errorf("task with name: %s, does not start with a # title marker", name)
	}
	bodyIndex := bytes.Index(contents, []byte("\n\n"))
	if bodyIndex == -1 {
		return nil, fmt.Errorf("failed to find task body in task: %s", name)
	}

	return &qf.Task{
		AssignmentOrder: assignmentOrder,
		Title:           string(contents[2:bodyIndex]),
		Body:            string(contents[bodyIndex+2:]),
		Name:            name,
	}, nil
}

// tasksFromAssignments returns a map, mapping each assignment-order to a map of tasks.
func tasksFromAssignments(assignments []*qf.Assignment) map[uint32]map[string]*qf.Task {
	taskMap := make(map[uint32]map[string]*qf.Task)
	for _, assignment := range assignments {
		temp := make(map[string]*qf.Task)
		for _, task := range assignment.Tasks {
			temp[task.Name] = task
		}
		taskMap[assignment.Order] = temp
	}
	return taskMap
}

// mapTasksByID transforms the given tasks to a map from taskID to task.
func mapTasksByID(tasks []*qf.Task) map[uint64]*qf.Task {
	taskMap := make(map[uint64]*qf.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}
	return taskMap
}

// synchronizeTasksWithIssues synchronizes tasks with issues on SCM's group repositories.
func synchronizeTasksWithIssues(ctx context.Context, db database.Database, sc scm.SCM, course *qf.Course, assignments []*qf.Assignment) error {
	tasksFromTestsRepo := tasksFromAssignments(assignments)
	createdTasks, updatedTasks, err := db.SynchronizeAssignmentTasks(course, tasksFromTestsRepo)
	if err != nil {
		return err
	}

	repos, err := db.GetRepositoriesWithIssues(&qf.Repository{
		ScmOrganizationID: course.GetScmOrganizationID(),
	})
	if err != nil {
		return err
	}

	// Creates, updates and deletes issues on all group repositories, based on how tasks differ from last push.
	// The created issues will be created by the QuickFeed user (the App owner).
	var createdIssues []*qf.Issue
	for _, repo := range repos {
		if !repo.IsGroupRepo() {
			continue
		}
		repoCreatedIssues, err := createIssues(ctx, sc, course, repo, createdTasks)
		if err != nil {
			return err
		}
		createdIssues = append(createdIssues, repoCreatedIssues...)
		if err = updateIssues(ctx, sc, course, repo, updatedTasks); err != nil {
			return err
		}
	}
	// Create issues in the database based on issues created on the scm.
	return db.CreateIssues(createdIssues)
}

// createIssues creates issues on scm based on repository, course and tasks. Returns created issues.
func createIssues(ctx context.Context, sc scm.SCM, course *qf.Course, repo *qf.Repository, tasks []*qf.Task) ([]*qf.Issue, error) {
	var createdIssues []*qf.Issue
	for _, task := range tasks {
		issueOptions := &scm.IssueOptions{
			Organization: course.GetScmOrganizationName(),
			Repository:   repo.Name(),
			Title:        task.Title,
			Body:         task.Body,
		}
		scmIssue, err := sc.CreateIssue(ctx, issueOptions)
		if err != nil {
			return nil, err
		}
		createdIssues = append(createdIssues, &qf.Issue{
			RepositoryID:   repo.ID,
			TaskID:         task.ID,
			ScmIssueNumber: uint64(scmIssue.Number),
		})
	}
	return createdIssues, nil
}

// updateIssues updates issues based on repository, course and tasks. It handles deleted tasks by closing them and inserting a statement into the body.
func updateIssues(ctx context.Context, sc scm.SCM, course *qf.Course, repo *qf.Repository, tasks []*qf.Task) error {
	taskMap := mapTasksByID(tasks)
	for _, issue := range repo.Issues {
		task, ok := taskMap[issue.TaskID]
		if !ok {
			// Issue does not need to be updated
			continue
		}
		issueOptions := &scm.IssueOptions{
			Organization: course.GetScmOrganizationName(),
			Repository:   repo.Name(),
			Title:        task.Title,
			Body:         task.Body,
			Number:       int(issue.ScmIssueNumber),
		}
		if task.IsDeleted() {
			issueOptions.State = "closed"
		}

		if _, err := sc.UpdateIssue(ctx, issueOptions); err != nil {
			return err
		}
	}
	return nil
}
