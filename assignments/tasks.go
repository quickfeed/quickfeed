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

	task := &pb.Task{
		AssignmentID: uint64(assignment.Order),
		Title:        string(contents[2:bodyIndex]),
		Body:         string(contents[bodyIndex+2:]),
		Name:         name,
	}

	return task, nil
}

// Updates an issue on specified repository
func UpdateScmIssue(c context.Context, sc scm.SCM, course *pb.Course, repo *pb.Repository, issue *pb.Issue) (*scm.Issue, error) {
	newIssue := &scm.CreateIssueOptions{
		Organization: course.Name,
		Repository:   filepath.Base(repo.GetHTMLURL()), // Todo
		Title:        issue.Title,
		Body:         issue.Body,
	}
	updateIssue := &scm.IssueOptions{
		Organization: course.Name,
		Repository:   filepath.Base(repo.GetHTMLURL()), // Todo
		IssueNumber:  int(issue.IssueNumber),
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
func CreateIssue(c context.Context, repo *pb.Repository, task *pb.Task, scmIssue *scm.Issue) (*pb.Issue, error) {
	issue := &pb.Issue{
		RepositoryID: repo.ID,
		TaskID:       task.ID,
		IssueNumber:  uint64(scmIssue.IssueNumber),
		Name:         task.Name,
		Title:        task.Title,
		Body:         task.Body,
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
// - When assignments are parsed from API call, the assignmentID field is 0, but order is set to assignmentID in yaml file.
//	 Should check that this doesn't lead to problems when creating/updating assignments in db, and also their related tasks.
// - Need to create a test that tests the synchronization of tasks. Currently in TestHandleTasks(), all tasks are already created in db before running.
// - There are many fmt.Printf()'s scattered across different functions. These need to be removed.
// - SynchronizeTasks now synchs per assignment, instead of just for the entire course. Should review the possibility of changing
// - Repositories in database do not have a "Name"-field.
// - SynchronizeTasks might not be necessary since assignments are updated in UpdateFromTestsRepo. Should generally review UpdateFromTestsRepo.
// - Current implementation is highly reliant on assignments being created an updated correctly in database. Need to review whether or not this is actually happening.
// - Implement DeleteTask() in SynchronizeTasks.

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

	// Loops through all assignments found in "tests"-repo
	tasks := []*pb.Task{}
	for _, assignment := range assignments {
		synchronizedTasks, err := SynchronizeTasks(c, db, assignment)
		if err != nil {
			return err
		}
		tasks = append(tasks, synchronizedTasks...)
	}

	// Loops through all student repos
	for _, repo := range repos {
		if !repo.IsStudentRepo() {
			continue
		}

		// Remember to remove
		fmt.Printf("\n\nHandeling tasks fro repo: %s\n", repo.HTMLURL)

		err = SynchronizeIssues(c, db, course, s, repo, tasks)
		if err != nil {
			return err
		}
	}

	return nil
}

// SynchronizeTasks synchronizes tasks in the database, with the ones of given assignment.
// Returns a slice of tasks as they appear in the database.
func SynchronizeTasks(c context.Context, db database.Database, assignment *pb.Assignment) ([]*pb.Task, error) {
	// Here foundTasks represents tasks that have been found by running readTestsRepositoryContent().
	// While existingTasks represent tasks that are found in the database for this given assignment.

	tasksToBeCreated := []*pb.Task{}
	tasksToBeUpdated := []*pb.Task{}
	tasksToReturn := []*pb.Task{}
	foundTasks := make(map[string]*pb.Task)
	for _, task := range assignment.Tasks {
		foundTasks[task.Name] = task
	}

	existingTasks, err := db.GetTasks(&pb.Task{AssignmentID: uint64(assignment.GetOrder())}) // Check todo
	if err != nil {
		return nil, err
	}

	for _, existingTask := range existingTasks {
		foundTask, ok := foundTasks[existingTask.Name]
		if !ok {
			// There exists a task in db, that is not represented by a task found in scm.
			// This task should be deleted
			// DeleteTask()
			continue
		}
		if !(foundTask.Title == existingTask.Title && foundTask.Body == existingTask.Body) {
			// Task has been changed
			existingTask.Title = foundTask.Title
			existingTask.Body = foundTask.Body
			tasksToBeUpdated = append(tasksToBeUpdated, existingTask)
		}
		tasksToReturn = append(tasksToReturn, existingTask)
		delete(foundTasks, existingTask.Name)
	}

	// Only tasks that there is no existing record of remains
	for _, task := range foundTasks {
		tasksToBeCreated = append(tasksToBeCreated, task)
		tasksToReturn = append(tasksToReturn, task)
	}

	err = db.CreateTasks(tasksToBeCreated)
	if err != nil {
		return nil, err
	}
	err = db.UpdateTasks(tasksToBeUpdated)
	if err != nil {
		return nil, err
	}

	return tasksToReturn, nil
}

// SynchronizeIssues synchronizes database and scm with issues based on tasks found
func SynchronizeIssues(c context.Context, db database.Database, course *pb.Course, s scm.SCM, repo *pb.Repository, tasks []*pb.Task) error {
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

	// Only tasks that do not have an issue with them remain.
	for _, task := range tasksMap {
		// Creates the actual issue on a scm
		scmIssue, err := CreateScmIssue(c, s, course, repo, task)
		if err != nil {
			return err
		}
		// Creates issue to be saved in db
		issue, err := CreateIssue(c, repo, task, scmIssue)
		if err != nil {
			return err
		}
		issuesToBeCreated = append(issuesToBeCreated, issue)
	}

	// Updates issues on scm.
	for _, issue := range issuesToBeUpdated {
		_, err := UpdateScmIssue(c, s, course, repo, issue)
		if err != nil {
			return err
		}
	}

	// Remember to remove
	fmt.Printf("\nIssues to be created:\n")
	for _, issue := range issuesToBeCreated {
		fmt.Printf("%v\n", issue)
	}

	// Remember to remove
	fmt.Printf("\nIssues to be updated:\n")
	for _, issue := range issuesToBeUpdated {
		fmt.Printf("%v\n", issue)
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

// Gets dbRepo based on orgID. Might not be necessary to have this
func GetRepositoriesByOrgID(db database.Database, orgID uint64) ([]*pb.Repository, error) {
	repositories, err := db.GetRepositoriesWithIssues(&pb.Repository{
		OrganizationID: orgID,
	})
	if err != nil {
		return nil, err
	}

	return repositories, nil
}
