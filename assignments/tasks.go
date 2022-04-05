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

// Message added to the body of an issue when closing it, since there is no support for deleting issues.
const deleteMsg = "\n**The task associated with this issue has been deleted by the teaching staff.**\n"

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

	return &pb.Task{
		AssignmentOrder: assignmentOrder,
		Title:           string(contents[2:bodyIndex]),
		Body:            string(contents[bodyIndex+2:]),
		Name:            name,
	}, nil
}

// tasksFromAssignments returns a map, mapping each assignment-order to a map of tasks.
func tasksFromAssignments(assignments []*pb.Assignment) map[uint32]map[string]*pb.Task {
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

// mapTasksByID transforms the given tasks to a map from taskID to task.
func mapTasksByID(tasks []*pb.Task) map[uint64]*pb.Task {
	taskMap := make(map[uint64]*pb.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}
	return taskMap
}

func handleTasks(ctx context.Context, db database.Database, sc scm.SCM, course *pb.Course, assignments []*pb.Assignment) error {
	tasksFromTestsRepo := tasksFromAssignments(assignments)
	createdTasks, updatedTasks, deletedTasks, err := db.SynchronizeAssignmentTasks(course, tasksFromTestsRepo)
	if err != nil {
		return err
	}

	repos, err := db.GetRepositoriesWithIssues(&pb.Repository{
		OrganizationID: course.GetOrganizationID(),
	})
	if err != nil {
		return err
	}

	// Creates, updates and deletes issues on all group repositories, based on how tasks differ from last push.
	createdIssues := []*pb.Issue{}
	for _, repo := range repos {
		if !repo.IsGroupRepo() {
			continue
		}
		repoCreatedIssues, err := createIssues(ctx, sc, course, repo, createdTasks)
		if err != nil {
			return err
		}
		createdIssues = append(createdIssues, repoCreatedIssues...)
		if err = updateIssues(ctx, sc, course, repo, updatedTasks, false); err != nil {
			return err
		}
		if err = updateIssues(ctx, sc, course, repo, deletedTasks, true); err != nil {
			return err
		}
	}

	// Deleting issues from database that no longer has an associated task.
	if err = db.DeleteIssuesOfAssociatedTasks(deletedTasks); err != nil {
		return err
	}
	// Creating issues in database, based on issues created on scm.
	return db.CreateIssues(createdIssues)
}

// createIssues creates issues on scm based on repository, course and tasks. Returns created issues.
func createIssues(ctx context.Context, sc scm.SCM, course *pb.Course, repo *pb.Repository, tasks []*pb.Task) ([]*pb.Issue, error) {
	createdIssues := []*pb.Issue{}
	for _, task := range tasks {
		issueOptions := &scm.CreateIssueOptions{
			Organization: course.GetOrganizationPath(),
			Repository:   repo.Name(),
			Title:        task.Title,
			Body:         task.Body,
		}
		scmIssue, err := sc.CreateIssue(ctx, issueOptions)
		if err != nil {
			return createdIssues, err
		}
		createdIssues = append(createdIssues, &pb.Issue{
			RepositoryID: repo.ID,
			TaskID:       task.ID,
			IssueNumber:  uint64(scmIssue.IssueNumber),
		})
	}
	return createdIssues, nil
}

// updateIssues updates issues based on repository, course and tasks. It handles deleted tasks by closing them and inserting a statement into the body.
func updateIssues(ctx context.Context, sc scm.SCM, course *pb.Course, repo *pb.Repository, tasks []*pb.Task, handleDeletion bool) error {
	taskMap := mapTasksByID(tasks)
	for _, issue := range repo.Issues {
		task, ok := taskMap[issue.TaskID]
		if !ok {
			// Issue does not need to be updated
			continue
		}
		issueOptions := &scm.CreateIssueOptions{
			Organization: course.GetOrganizationPath(),
			Repository:   repo.Name(),
			Title:        task.Title,
		}
		if handleDeletion {
			issueOptions.State = "closed"
			issueOptions.Body = deleteMsg + task.Body
		} else {
			issueOptions.Body = task.Body
		}

		if _, err := sc.EditRepoIssue(ctx, int(issue.IssueNumber), issueOptions); err != nil {
			return err
		}
	}
	return nil
}
