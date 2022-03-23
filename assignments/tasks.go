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

// TODO(Espeland): Ordering of tasks (See teacher.md)

// newTask returns a task from markdown contents and associates it with the given assignment.
// The provided markdown contents must contain a title specified on the first line,
// starting with the "# " character sequence, followed by two new line characters.
func newTask(contents []byte, assignmentOrder uint32, name string) (*pb.Task, error) {
	if !bytes.HasPrefix(contents, []byte("# ")) {
		return nil, fmt.Errorf("task with name: %s, does not start with a # title marker", name)
	}
	bodyIndex := bytes.Index(contents, []byte("\n\n"))
	if bodyIndex == -1 {
		return nil, fmt.Errorf("failed to find task body in task: %s", name)
	}

	task := &pb.Task{
		AssignmentOrder: assignmentOrder,
		Title:           string(contents[2:bodyIndex]),
		Body:            string(contents[bodyIndex+2:]),
		Name:            name,
	}

	return task, nil
}

// Updates an issue on specified repository
func updateScmIssue(c context.Context, sc scm.SCM, course *pb.Course, repo *pb.Repository, issue *pb.Issue) (*scm.Issue, error) {
	issueOptions := &scm.CreateIssueOptions{
		Organization: course.Name,
		Repository:   repo.Name(),
		Title:        issue.Title,
		Body:         issue.Body,
	}
	return sc.EditRepoIssue(c, int(issue.IssueNumber), issueOptions)
}

// Creates an issue on specified repository.
func createScmIssue(c context.Context, sc scm.SCM, course *pb.Course, repo *pb.Repository, task *pb.Task) (*scm.Issue, error) {
	issueOptions := &scm.CreateIssueOptions{
		Organization: course.Name,
		Repository:   repo.Name(),
		Title:        task.Title,
		Body:         task.Body,
	}
	return sc.CreateIssue(c, issueOptions)
}

// This is more of a converter function, and cannot currently return an error. Should probably be renamed or something.
func createIssue(c context.Context, repo *pb.Repository, task *pb.Task, scmIssue *scm.Issue) (*pb.Issue, error) {
	issue := &pb.Issue{
		RepositoryID: repo.ID,
		IssueNumber:  uint64(scmIssue.IssueNumber),
		Name:         task.Name,
		Title:        task.Title,
		Body:         task.Body,
	}
	return issue, nil
}

// getTasksFromAssignments returns a list of all tasks of given assignments.
func getTasksFromAssignments(assignments []*pb.Assignment) []*pb.Task {
	tasks := []*pb.Task{}
	for _, assignment := range assignments {
		tasks = append(tasks, assignment.Tasks...)
	}
	return tasks
}

// taskName returns the task name as a combination of assignmentName/filename
// excluding the task- prefix and the .md suffix.
func taskName(assignmentName, basePath string) string {
	taskName := basePath[len("task-"):]
	taskName = taskName[:len(taskName)-len(".md")]
	return filepath.Join(assignmentName, taskName)
}

func handleTasks(c context.Context, db database.Database, s scm.SCM, course *pb.Course, assignments []*pb.Assignment) error {
	tasks := getTasksFromAssignments(assignments)
	repos, err := db.GetRepositoriesWithIssues(&pb.Repository{
		OrganizationID: course.GetOrganizationID(),
	})
	if err != nil {
		return err
	}

	existingAssignments, err := db.GetAssignmentsByCourse(course.GetID(), false)
	if err != nil {
		return err
	}

	for _, assignment := range existingAssignments {
		if err = synchronizeTasks(c, db, assignment, tasks); err != nil {
			return err
		}
	}

	// Synchronize issues on all student repositories.
	for _, repo := range repos {
		if !repo.IsStudentRepo() {
			continue
		}
		err = synchronizeIssues(c, db, course, s, repo, tasks)
	}

	// TODO(Espeland): What should happen if there is an error while syncing issues?
	if err != nil {
		return err
	}

	return nil
}

// synchronizeTasks synchronizes tasks in the database with the ones of given assignment.
func synchronizeTasks(c context.Context, db database.Database, assignment *pb.Assignment, tasks []*pb.Task) error {
	tasksToBeCreated := []*pb.Task{}
	tasksToBeUpdated := []*pb.Task{}
	taskMap := make(map[string]*pb.Task)
	for _, task := range tasks {
		if task.AssignmentOrder == assignment.Order {
			taskMap[task.Name] = task
		}
	}

	existingTasks, err := db.GetTasks(&pb.Task{AssignmentID: assignment.GetID()})
	if err != nil {
		return err
	}

	for _, existingTask := range existingTasks {
		task, ok := taskMap[existingTask.Name]
		if !ok {
			// There exists a task in db, that is not represented by a task found in scm.
			db.DeleteTask(existingTask)
			continue
		}
		if existingTask.HasChanged(task) {
			// Task has been changed
			existingTask.Title = task.Title
			existingTask.Body = task.Body
			tasksToBeUpdated = append(tasksToBeUpdated, existingTask)
		}
		delete(taskMap, existingTask.Name)
	}

	// Only tasks that there is no existing record of remains
	for _, task := range taskMap {
		tasksToBeCreated = append(tasksToBeCreated, task)
	}

	err = db.CreateTasks(tasksToBeCreated)
	if err != nil {
		return err
	}
	err = db.UpdateTasks(tasksToBeUpdated)
	if err != nil {
		return err
	}

	return nil
}

// synchronizeIssues synchronizes database and scm with issues based on tasks found
func synchronizeIssues(c context.Context, db database.Database, course *pb.Course, s scm.SCM, repo *pb.Repository, tasks []*pb.Task) error {
	issuesToBeCreated := []*pb.Issue{}
	issuesToBeUpdated := []*pb.Issue{}
	tasksMap := make(map[string]*pb.Task)
	for _, task := range tasks {
		tasksMap[task.Name] = task
	}

	// Loops through existing issues on repo.
	for _, issue := range repo.Issues {
		task, ok := tasksMap[issue.Name]
		if !ok {
			// What should happen if task does not exist for issue?
			continue
		}
		// Might be necessary to check if the issue is closed here, otherwise we might end up editing already closed issues.
		// If this becomes a necessity, then the issue data-record must have a closed (true/false) field, or a status field.
		if !(task.Title == issue.Title && task.Body == issue.Body) {
			// Issue needs to be updated here
			issue.Title = task.Title
			issue.Body = task.Body
			issuesToBeUpdated = append(issuesToBeUpdated, issue)
		}
		delete(tasksMap, issue.Name)
	}

	// Only tasks that do not have an issue associated with them remain.
	for _, task := range tasksMap {
		// Creates the actual issue on a scm
		scmIssue, err := createScmIssue(c, s, course, repo, task)
		if err != nil {
			return err
		}
		// Creates issue to be saved in db
		issue, err := createIssue(c, repo, task, scmIssue)
		if err != nil {
			return err
		}
		issuesToBeCreated = append(issuesToBeCreated, issue)
	}

	// Updates issues on scm.
	for _, issue := range issuesToBeUpdated {
		_, err := updateScmIssue(c, s, course, repo, issue)
		if err != nil {
			return err
		}
	}

	err := db.CreateIssues(issuesToBeCreated)
	if err != nil {
		return err
	}
	err = db.UpdateIssues(issuesToBeUpdated)
	if err != nil {
		return err
	}

	return nil
}

// Only used for testing
func fakeSynchronizeIssues(c context.Context, db database.Database, repo *pb.Repository, tasks []*pb.Task) error {
	issuesToBeCreated := []*pb.Issue{}
	issuesToBeUpdated := []*pb.Issue{}
	tasksMap := make(map[string]*pb.Task)
	for _, task := range tasks {
		tasksMap[task.Name] = task
	}

	// Loops through existing issues on repo.
	for _, issue := range repo.Issues {
		task, ok := tasksMap[issue.Name]
		if !ok {
			// What should happen if task does not exist for issue?
			continue
		}
		if !(task.Title == issue.Title && task.Body == issue.Body) {
			// Issue needs to be updated here
			issue.Title = task.Title
			issue.Body = task.Body
			issuesToBeUpdated = append(issuesToBeUpdated, issue)
		}
		delete(tasksMap, issue.Name)
	}

	// Only tasks that do not have an issue associated with them remain.
	for _, task := range tasksMap {
		// Creates issue to be saved in db
		issue, err := createIssue(c, repo, task, &scm.Issue{
			IssueNumber: 1,
		})
		if err != nil {
			return err
		}
		issuesToBeCreated = append(issuesToBeCreated, issue)
	}

	err := db.CreateIssues(issuesToBeCreated)
	if err != nil {
		return err
	}
	err = db.UpdateIssues(issuesToBeUpdated)
	if err != nil {
		return err
	}

	return nil
}
