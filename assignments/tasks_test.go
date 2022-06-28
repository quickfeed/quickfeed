package assignments

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/log"
	"github.com/autograde/quickfeed/scm"
)

func assignmentsWithTasks(courseID uint64) []*pb.Assignment {
	return []*pb.Assignment{
		{
			CourseID:    courseID,
			Name:        "lab1",
			ScriptFile:  "go.sh",
			Deadline:    "12.01.2022",
			AutoApprove: false,
			Order:       1,
			IsGroupLab:  false,
			Tasks: []*pb.Task{
				{Title: "lab1/fib, 1", Name: "lab1/fib", AssignmentOrder: 1, Body: "Implement fibonacci"},
				{Title: "lab1/luc, 1", Name: "lab1/luc", AssignmentOrder: 1, Body: "Implement lucas numbers"},
			},
		},
		{
			CourseID:    courseID,
			Name:        "lab2",
			ScriptFile:  "go.sh",
			Deadline:    "12.12.2021",
			AutoApprove: false,
			Order:       2,
			IsGroupLab:  false,
			Tasks: []*pb.Task{
				{Title: "lab2/add, 2", Name: "lab2/add", AssignmentOrder: 2, Body: "Implement addition"},
				{Title: "lab2/sub, 2", Name: "lab2/sub", AssignmentOrder: 2, Body: "Implement subtraction"},
				{Title: "lab2/mul, 2", Name: "lab2/mul", AssignmentOrder: 2, Body: "Implement multiplication"},
			},
		},
	}
}

type foundIssue struct {
	IssueNumber uint64
	Name        string
}

// The test environment creates tasks based on the issues found on student repositories in the organization.
// This is so that we can emulate the result of a previous push to the tests repository.
// For these tasks to be created correctly the title of each issue must use the following format:
//   "<task name>, <assignment order>".
//
// For example, if an issue is related to "task-hello_world.md" in "lab1",
// then the title of the corresponding issue should be:
//   "lab1/hello_world, 1" (assuming lab1 has assignment order 1).
//
// It is also recommended that issues are created on all student repositories, and that they are the same.

// populateDatabaseWithTasks based on the given course's organization.
func populateDatabaseWithTasks(t *testing.T, db database.Database, sc scm.SCM, course *pb.Course) error {
	t.Helper()

	// Assignments that will be updated
	assignmentsWithTasks := assignmentsWithTasks(course.ID)
	if err := db.UpdateAssignments(assignmentsWithTasks); err != nil {
		return err
	}

	ctx := context.Background()
	org := &pb.Organization{
		Path: course.OrganizationPath,
	}
	repos, err := sc.GetRepositories(ctx, org)
	if err != nil {
		return err
	}

	foundIssues := make(map[uint64]map[string]*foundIssue)
	tasks := make(map[uint32]map[string]*pb.Task)

	// Finds issues, and creates tasks based on them
	for _, repo := range repos {
		existingScmIssues, err := sc.GetIssues(ctx, &scm.RepositoryOptions{
			Owner: course.Name,
			Path:  repo.Path,
		})
		if err != nil {
			return err
		}

		if len(existingScmIssues) == 0 {
			continue
		}
		foundIssues[repo.ID] = make(map[string]*foundIssue)
		for _, scmIssue := range existingScmIssues {
			fmt.Printf("issue: %v\n", scmIssue)
			splitTitle := strings.Split(scmIssue.Title, ", ")
			name := splitTitle[0]
			temp, err := strconv.Atoi(splitTitle[len(splitTitle)-1])
			if err != nil {
				continue
			}
			assignmentOrder := uint32(temp)
			foundIssues[repo.ID][name] = &foundIssue{IssueNumber: uint64(scmIssue.IssueNumber), Name: name}

			if _, ok := tasks[assignmentOrder]; !ok {
				tasks[assignmentOrder] = make(map[string]*pb.Task)
			}
			tasks[assignmentOrder][name] = &pb.Task{Title: scmIssue.Title, Body: scmIssue.Body, Name: name, AssignmentOrder: assignmentOrder}
		}
	}

	createdTasks, _, err := db.SynchronizeAssignmentTasks(course, tasks)
	if err != nil {
		return err
	}
	for _, t := range createdTasks {
		fmt.Printf("t: %v\n", t)
	}

	dbRepos, err := db.GetRepositoriesWithIssues(&pb.Repository{
		OrganizationID: course.GetOrganizationID(),
	})
	if err != nil {
		return err
	}

	issuesToCreate := []*pb.Issue{}
	for _, repo := range dbRepos {
		if !repo.IsGroupRepo() {
			continue
		}
		for _, task := range createdTasks {
			foundIssue, ok := foundIssues[repo.RepositoryID][task.Name]
			if !ok {
				continue
			}
			issuesToCreate = append(issuesToCreate, &pb.Issue{RepositoryID: repo.ID, TaskID: task.ID, IssueNumber: foundIssue.IssueNumber})
		}
	}
	for _, t := range issuesToCreate {
		fmt.Printf("i: %v\n", t)
	}

	return db.CreateIssues(issuesToCreate)
}

// TestHandleTasks runs handleTasks() on the specified organization.
// Results vary depending on what tasks/issues existed prior to running.
func TestHandleTasks(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	qfTestUser := scm.GetTestUser(t)
	accessToken := scm.GetAccessToken(t)

	logger := log.Zap(false).Sugar()
	sc, err := scm.NewSCMClient(logger, "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	course := &pb.Course{
		Name:             qfTestOrg,
		OrganizationPath: qfTestOrg,
	}

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	if err = qtest.PopulateDatabaseWithInitialData(t, db, sc, course); err != nil {
		t.Fatal(err)
	}
	if err = populateDatabaseWithTasks(t, db, sc, course); err != nil {
		t.Fatal(err)
	}

	assignments := assignmentsWithTasks(course.ID)
	ctx := context.Background()
	if err = handleTasks(ctx, db, sc, course, assignments); err != nil {
		t.Fatal(err)
	}
	// TODO(meling) Check that we get the expected assignments back from github...
	for _, a := range assignments {
		for _, t := range a.GetTasks() {
			fmt.Printf("B: %v\n", t)
		}
	}

	opt := &scm.RepositoryOptions{
		Owner: qfTestOrg,
		Path:  pb.StudentRepoName(qfTestUser),
	}
	issues, err := sc.GetIssues(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}
	for _, issue := range issues {
		t.Logf("issue: %v", issue)
	}
}
