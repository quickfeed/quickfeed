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

// taskName returns the task name as a combination of assignmentName/filename
// excluding the task- prefix and the .md suffix.
func taskName(assignmentName, basePath string) string {
	taskName := basePath[len("task-"):]
	taskName = taskName[:len(taskName)-len(".md")]
	return filepath.Join(assignmentName, taskName)
}

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

// getTasksFromAssignments returns a map, mapping each assignment-order to a map of tasks.
func getTasksFromAssignments(assignments []*pb.Assignment) map[uint32]map[string]*pb.Task {
	taskMap := make(map[uint32]map[string]*pb.Task)
	for _, assignment := range assignments {
		temp := make(map[string]*pb.Task)
		for _, task := range assignment.Tasks {
			temp[task.Name] = task
		}
		taskMap[assignment.Order] = temp
	}
	return taskMap
}

// TODO(Espeland): handleTasks no longer handles late enrolling students, as it only creates, updates and deletes based on how tasks differ from last time checked.
// A different function will have to run when students enroll, creating an issue per task found in the database.
// handleTasks would currently only work in such a way if there are no tasks in tests-repo when a student enrolls. Then this function would catch all created new tasks, and then create an issue from them.
func handleTasks(c context.Context, db database.Database, s scm.SCM, course *pb.Course, assignments []*pb.Assignment) error {
	var createdIssues []*pb.Issue
	var err error
	foundTasks := getTasksFromAssignments(assignments)
	createdTasks, updatedTasks, deletedTasks, err := db.SynchronizeAssignmentTasks(course, foundTasks)
	if err != nil {
		return err
	}

	repos, err := db.GetRepositoriesWithIssues(&pb.Repository{
		OrganizationID: course.GetOrganizationID(),
	})
	if err != nil {
		return err
	}

	// Todo(Espeland): In general, how do we handle if something goes wrong during one of these processes?

	// Deleting issues from database that no longer has an associated task.
	deletedIssues, err := db.DeleteIssuesOfAssociatedTasks(deletedTasks)
	if err != nil {
		return err
	}

	// Creates, updates and deletes issues on all group repositories, based on how tasks differ from last push.
	for _, repo := range repos {
		if !repo.IsGroupRepo() {
			continue
		}
		createdIssues, err = createIssues(c, s, course, repo, createdTasks)
		err = updateIssues(c, s, course, repo, updatedTasks)
		err = deleteIssues(c, s, course, repo, deletedIssues)
	}

	// Creating issues in database, based on issues created on scm.
	err = db.CreateIssues(createdIssues)
	return err
}

// createIssues creates issues on scm based on repository, course and tasks. Returns created issues.
func createIssues(c context.Context, s scm.SCM, course *pb.Course, repo *pb.Repository, tasks []*pb.Task) ([]*pb.Issue, error) {
	createdIssues := []*pb.Issue{}
	for _, task := range tasks {
		issueOptions := &scm.CreateIssueOptions{
			Organization: course.Name,
			Repository:   repo.Name(),
			Title:        task.Title,
			Body:         task.Body,
		}
		scmIssue, err := s.CreateIssue(c, issueOptions)
		if err != nil {
			// TODO(Espeland): Should we return here if there was an error creating the issue? We certainly shouldn't create a db-entry for the issue if there was an error.
			continue
		}
		createdIssues = append(createdIssues, &pb.Issue{
			RepositoryID: repo.ID,
			TaskID:       task.ID,
			IssueNumber:  uint64(scmIssue.IssueNumber),
		})
	}
	return createdIssues, nil
}

// createIssues updates issues based on repository, course and tasks.
func updateIssues(c context.Context, s scm.SCM, course *pb.Course, repo *pb.Repository, tasks []*pb.Task) (err error) {
	taskMap := make(map[uint64]*pb.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	for _, issue := range repo.Issues {
		task, ok := taskMap[issue.TaskID]
		if !ok {
			// Issue does not need to be updated
			continue
		}
		issueOptions := &scm.CreateIssueOptions{
			Organization: course.Name,
			Repository:   repo.Name(),
			Title:        task.Title,
			Body:         task.Body,
		}
		// TODO(Espeland): How do we handle an error while updating a single repository issue?
		_, err = s.EditRepoIssue(c, int(issue.IssueNumber), issueOptions)
	}
	return err
}

func deleteIssues(c context.Context, s scm.SCM, course *pb.Course, repo *pb.Repository, issues []*pb.Issue) error {
	// TODO(Espeland): How do we handle a deleted task? Go-github does not have a way of deleting issues, only closing them.
	return nil
}
