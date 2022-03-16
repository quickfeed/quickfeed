package assignments

import (
	"context"
	"fmt"
	"strings"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/scm"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"google.golang.org/protobuf/testing/protocmp"
)

// When running tests that have anything to do with tasks/issues, it is important that issues have their title corresponding to the name of an associated task.
// For example, if you have an issue that is supposed to be connected to the task "task-1.md" in "lab1", the title of this issue needs to be "lab1/task-1.md".
// Otherwise when creating the database there will be no clear way to know which issue is supposed to be associated with which task.
// This disclaimer applies only when running tests.

// InitializeDbEnvironment initializes a db, based on org.
func InitializeDbEnvironment(t *testing.T, c context.Context, course *pb.Course, s scm.SCM) (database.Database, func(), error) {
	db, cleanup := qtest.TestDB(t)

	org, err := s.GetOrganization(c, &scm.GetOrgOptions{Name: course.Name})
	if err != nil {
		return db, cleanup, err
	}

	// Create course
	admin := qtest.CreateFakeUser(t, db, uint64(1))
	qtest.CreateCourse(t, db, admin, course)

	// Create assignments
	foundAssignments, _, err := fetchAssignments(c, zap.NewNop().Sugar(), s, course)
	if err != nil {
		return db, cleanup, err
	}

	err = db.UpdateAssignments(foundAssignments)
	if err != nil {
		return db, cleanup, err
	}

	// Get repositories
	repos, err := s.GetRepositories(c, org)
	if err != nil {
		return db, cleanup, err
	}

	// Create repositories with issues
	n := 2
	for _, repo := range repos {
		var user *pb.User
		// Hacky solution, but did not quickly find already supplied function for doing this.
		// Does not handle if repo is group repo.
		// Might not even be necessary to handle repos differently in these tests.
		dbRepo := &pb.Repository{}
		switch repo.Path {
		case "course-info":
			dbRepo = &pb.Repository{
				RepositoryID:   repo.ID,
				OrganizationID: org.GetID(),
				HTMLURL:        repo.WebURL,
				RepoType:       pb.Repository_COURSEINFO,
			}
		case "assignments":
			dbRepo = &pb.Repository{
				RepositoryID:   repo.ID,
				OrganizationID: org.GetID(),
				HTMLURL:        repo.WebURL,
				RepoType:       pb.Repository_ASSIGNMENTS,
			}
		case "tests":
			dbRepo = &pb.Repository{
				RepositoryID:   repo.ID,
				OrganizationID: org.GetID(),
				HTMLURL:        repo.WebURL,
				RepoType:       pb.Repository_TESTS,
			}
		default:
			user = qtest.CreateFakeUser(t, db, uint64(n))
			dbRepo = &pb.Repository{
				RepositoryID:   repo.ID,
				OrganizationID: org.GetID(),
				UserID:         user.ID,
				HTMLURL:        repo.WebURL,
				RepoType:       pb.Repository_USER,
			}
		}

		issues := []*pb.Issue{}

		err = db.CreateRepository(dbRepo)
		if err != nil {
			return db, cleanup, err
		}

		existingScmIssues, err := s.GetRepoIssues(c, &scm.IssueOptions{
			Organization: course.Name,
			Repository:   repo.Path,
		})
		if err != nil {
			return db, cleanup, err
		}

		for _, scmIssue := range existingScmIssues {
			dbIssue := &pb.Issue{
				RepositoryID: dbRepo.ID,
				IssueNumber:  uint64(scmIssue.IssueNumber),
				Name:         scmIssue.Title,
				Title:        scmIssue.Title,
				Body:         scmIssue.Body,
			}
			issues = append(issues, dbIssue)
		}
		db.CreateIssues(issues)
		n++
	}

	return db, cleanup, nil
}

// TestInitializeDbEnvironment tests if db is correctly initialized based on preexisting repositories
func TestInitializeDbEnvironment(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	course := &pb.Course{
		Name:             qfTestOrg,
		OrganizationPath: qfTestOrg,
	}

	ctx := context.Background()

	db, callback, err := InitializeDbEnvironment(t, ctx, course, s)
	defer callback()
	if err != nil {
		t.Fatal(err)
	}

	org, err := s.GetOrganization(ctx, &scm.GetOrgOptions{Name: course.Name})
	if err != nil {
		t.Fatal(err)
	}

	repos, err := db.GetRepositoriesWithIssues(&pb.Repository{
		OrganizationID: org.GetID(),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("\n\nRepositories and issues:\n")
	for _, repo := range repos {
		t.Logf("\nRepository ID: %d\nRepository HTMLURL: %s\nRepository UserID: %d", repo.ID, repo.HTMLURL, repo.UserID)
		for _, issue := range repo.Issues {
			t.Logf("\nIssue ID: %d\nIssue RepositoryID: %d\nIssue IssueNumber: %d\nIssue Name: %s\nIssue Title: %s\nIssue Body: %s", issue.ID, issue.RepositoryID, issue.IssueNumber, issue.Name, issue.Title, issue.Body)
		}
	}

	assignments, err := db.GetAssignmentsByCourse(course.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("\n\nAssignments:\n")
	for _, assignment := range assignments {
		t.Logf("\nAssignment ID: %d\nAssignment Name: %s", assignment.ID, assignment.Name)
	}
}

// TestGetTasks retrieves all tasks of a given course via API call.
func TestGetTasks(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)

	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	course := &pb.Course{
		Name:             qfTestOrg,
		OrganizationPath: qfTestOrg,
	}

	assignments, _, err := fetchAssignments(context.Background(), zap.NewNop().Sugar(), s, course)
	if err != nil {
		t.Fatal(err)
	}

	for _, assignment := range assignments {
		for _, task := range assignment.Tasks {
			t.Logf("\nTask AssignmentID: %d\nTask Name: %s\nTask Title: %s\nTask Body: %s\n\n", task.AssignmentID, task.Name, task.Title, task.Body)
		}
	}
}

// TestGetIssuesOnOrg should get all issues on "-labs" repos via API call. Does not get closed issues
func TestGetIssuesOnOrg(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)

	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	course := &pb.Course{
		Name:             qfTestOrg,
		OrganizationPath: qfTestOrg,
	}

	ctx := context.Background()

	org, err := s.GetOrganization(ctx, &scm.GetOrgOptions{Name: course.Name})
	if err != nil {
		t.Fatal(err)
	}

	repos, err := s.GetRepositories(ctx, org)
	if err != nil {
		t.Fatal(err)
	}

	for _, repo := range repos {
		// Should change this test, though there is no good alternative atm.
		if !strings.HasSuffix(repo.Path, "-labs") {
			continue
		}
		t.Logf("\n\nIssues on repo %s:\n", repo.Path)
		issues, err := s.GetRepoIssues(ctx, &scm.IssueOptions{
			Organization: course.Name,
			Repository:   repo.Path,
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, issue := range issues {
			t.Logf("Issue ID: %d\nIssue IssueNumber: %d\nIssue title: %s\nIssue body: %s\n\n", issue.ID, issue.IssueNumber, issue.Title, issue.Body)
		}
	}

	if err != nil {
		t.Fatal(err)
	}
}

// TestHandleTasks runs HandleTasks() on specified org. Results vary depending on which tasks/issues existed prior to running.
func TestHandleTasks(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)

	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	course := &pb.Course{
		Name:             qfTestOrg,
		OrganizationPath: qfTestOrg,
	}

	ctx := context.Background()
	logger := zap.NewNop().Sugar()

	db, callback, err := InitializeDbEnvironment(t, ctx, course, s)
	defer callback()
	if err != nil {
		t.Fatal(err)
	}

	org, err := s.GetOrganization(ctx, &scm.GetOrgOptions{Name: course.Name})
	if err != nil {
		t.Fatal(err)
	}

	// Prints db contents before HandleTasks. This code is also used elsewhere and should be turned into a function if it's going to stick around
	repos, err := db.GetRepositoriesWithIssues(&pb.Repository{
		OrganizationID: org.GetID(),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, repo := range repos {
		t.Logf("\nRepository ID: %d\nRepository HTMLURL: %s\nRepository UserID: %d", repo.ID, repo.HTMLURL, repo.UserID)
		for _, issue := range repo.Issues {
			t.Logf("\nIssue ID: %d\nIssue RepositoryID: %d\nIssue IssueNumber: %d\nIssue Name: %s\nIssue Title: %s\nIssue Body: %s", issue.ID, issue.RepositoryID, issue.IssueNumber, issue.Name, issue.Title, issue.Body)
		}
	}

	assignments, _, err := fetchAssignments(ctx, logger, s, course)
	if err != nil {
		t.Fatal(err)
	}

	err = HandleTasks(ctx, db, s, course, GetTasksFromAssignments(ctx, assignments))
	if err != nil {
		t.Fatal(err)
	}

	// Db contents after
	t.Logf("\n\n\nDB AFTER\n\n\n")
	repos, err = db.GetRepositoriesWithIssues(&pb.Repository{
		OrganizationID: org.GetID(),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, repo := range repos {
		t.Logf("\nRepository ID: %d\nRepository HTMLURL: %s\nRepository UserID: %d", repo.ID, repo.HTMLURL, repo.UserID)
		for _, issue := range repo.Issues {
			t.Logf("\nIssue ID: %d\nIssue RepositoryID: %d\nIssue IssueNumber: %d\nIssue Name: %s\nIssue Title: %s\nIssue Body: %s", issue.ID, issue.RepositoryID, issue.IssueNumber, issue.Name, issue.Title, issue.Body)
		}
	}
}

// TestSynchronizeTasks tests whether tasks are correctly updated in the database
func TestSynchronizeTasks(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	ctx := context.Background()
	admin := qtest.CreateFakeUser(t, db, uint64(1))
	qtest.CreateCourse(t, db, admin, &pb.Course{})

	assignments := []*pb.Assignment{
		{
			CourseID: 1,
			Name:     "Lab1",
			Order:    1,
		},
		{
			CourseID: 1,
			Name:     "Lab2",
			Order:    2,
		},
	}

	for _, assignment := range assignments {
		err := db.CreateAssignment(assignment)
		if err != nil {
			t.Fatal(err)
		}
	}

	tasks := []*pb.Task{
		{
			AssignmentID:    1,
			AssignmentOrder: 1,
			Title:           "Lab1, task1",
			Body:            "Description of task1 in lab1",
			Name:            "Lab1/task1.md",
		},
		{
			AssignmentID:    1,
			AssignmentOrder: 1,
			Title:           "Lab1, task2",
			Body:            "Description of task2 in lab1",
			Name:            "Lab1/task2.md",
		},
		{
			AssignmentID:    2,
			AssignmentOrder: 2,
			Title:           "Lab2, task1",
			Body:            "Description of task1 in lab2",
			Name:            "Lab2/task1.md",
		},
		{
			AssignmentID:    2,
			AssignmentOrder: 2,
			Title:           "Lab2, task2",
			Body:            "Description of task2 in lab2",
			Name:            "Lab2/task2.md",
		},
	}

	err := db.CreateTasks(tasks)
	if err != nil {
		t.Fatal(err)
	}

	// Nothing should happen from this synchronization
	for _, assignment := range assignments {
		err := SynchronizeTasks(ctx, db, assignment, tasks)
		if err != nil {
			t.Fatal(err)
		}
	}
	wantTasks := tasks
	gotTasks, err := db.GetTasks(&pb.Task{})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantTasks, gotTasks, protocmp.Transform()); diff != "" {
		t.Errorf("Synchronization mismatch (-wantTasks, +gotTasks):\n%s", diff)
	}
	// -------------------------------------------------------------------------- //

	// Testing adding one new task, and updating another
	tasks = append(tasks, &pb.Task{
		AssignmentID:    2,
		AssignmentOrder: 2,
		Title:           "Lab2, task3",
		Body:            "Description of task3 in lab2",
		Name:            "Lab2/task3.md",
	})
	tasks[0].Body = "New body for lab1 task1"
	wantTasks = tasks

	for _, assignment := range assignments {
		err := SynchronizeTasks(ctx, db, assignment, tasks)
		if err != nil {
			t.Fatal(err)
		}
	}
	gotTasks, err = db.GetTasks(&pb.Task{})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantTasks, gotTasks, protocmp.Transform()); diff != "" {
		t.Errorf("Synchronization mismatch (-wantTasks, +gotTasks):\n%s", diff)
	}
	// -------------------------------------------------------------------------- //

	// Testing adding new task to db, that is not represented by tasks supplied to SynchronizeTasks
	db.CreateTasks([]*pb.Task{
		{
			AssignmentID:    1,
			AssignmentOrder: 1,
			Title:           "Title title",
			Body:            "This task should not exists in db",
			Name:            "Fake name",
		},
	})
	wantTasks = tasks

	for _, assignment := range assignments {
		err := SynchronizeTasks(ctx, db, assignment, tasks)
		if err != nil {
			t.Fatal(err)
		}
	}

	gotTasks, err = db.GetTasks(&pb.Task{})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantTasks, gotTasks, protocmp.Transform()); diff != "" {
		t.Errorf("Synchronization mismatch (-wantTasks, +gotTasks):\n%s", diff)
	}
	// -------------------------------------------------------------------------- //
}

// TestSynchronizeIssues tests whether issues are correctly updated in the database
func TestSynchronizeIssues(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	ctx := context.Background()
	admin := qtest.CreateFakeUser(t, db, uint64(1))
	qtest.CreateCourse(t, db, admin, &pb.Course{})

	user1 := qtest.CreateFakeUser(t, db, uint64(2))
	user2 := qtest.CreateFakeUser(t, db, uint64(3))
	repositories := []*pb.Repository{
		{
			RepositoryID:   11,
			OrganizationID: 1,
			UserID:         user1.GetID(),
			RepoType:       pb.Repository_USER,
		},
		{
			RepositoryID:   12,
			OrganizationID: 1,
			UserID:         user2.GetID(),
			RepoType:       pb.Repository_USER,
		},
	}

	for _, repository := range repositories {
		err := db.CreateRepository(repository)
		if err != nil {
			t.Fatal(err)
		}
	}

	tasks := []*pb.Task{
		{
			Title: "Lab1 task1",
			Body:  "Description of lab1 in task1",
			Name:  "Lab1/task1.md",
		},
		{
			Title: "Lab2 task1",
			Body:  "Description of lab2 in task1",
			Name:  "Lab2/task1.md",
		},
	}

	issues := []*pb.Issue{
		{
			RepositoryID: 1,
			IssueNumber:  1,
			Name:         "Lab1/task1.md",
			Title:        "Lab1 task1",
			Body:         "Description of lab1 in task1",
		},
		{
			RepositoryID: 1,
			IssueNumber:  1,
			Name:         "Lab2/task1.md",
			Title:        "Lab2 task1",
			Body:         "Description of lab2 in task1",
		},
		{
			RepositoryID: 2,
			IssueNumber:  1,
			Name:         "Lab1/task1.md",
			Title:        "Lab1 task1",
			Body:         "Description of lab1 in task1",
		},
		{
			RepositoryID: 2,
			IssueNumber:  1,
			Name:         "Lab2/task1.md",
			Title:        "Lab2 task1",
			Body:         "Description of lab2 in task1",
		},
	}
	err := db.CreateIssues(issues)
	if err != nil {
		t.Fatal(err)
	}

	// Nothing should happen from this synchronization
	wantIssues := issues
	repositories, err = db.GetRepositoriesWithIssues(&pb.Repository{})
	for _, repo := range repositories {
		err := FakeSynchronizeIssues(ctx, db, repo, tasks)
		if err != nil {
			t.Fatal(err)
		}
	}
	gotIssues, err := db.GetIssues(&pb.Issue{})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantIssues, gotIssues, protocmp.Transform()); diff != "" {
		t.Errorf("Synchronization mismatch (-wantIssues, +gotIssues):\n%s", diff)
	}
	// -------------------------------------------------------------------------- //

	// Testing adding another task, and updating an existing one
	tasks = append(tasks, &pb.Task{
		Title: "Lab2 task2",
		Body:  "Description of lab2 in task2",
		Name:  "Lab2/task2.md",
	})
	tasks[0].Title = "New title"
	newIssues := []*pb.Issue{
		{
			ID:           5,
			RepositoryID: 1,
			IssueNumber:  1,
			Name:         "Lab2/task2.md",
			Title:        "Lab2 task2",
			Body:         "Description of lab2 in task2",
		},
		{
			ID:           6,
			RepositoryID: 2,
			IssueNumber:  1,
			Name:         "Lab2/task2.md",
			Title:        "Lab2 task2",
			Body:         "Description of lab2 in task2",
		},
	}
	issues = append(issues, newIssues...)
	issues[0].Title = "New title"
	issues[2].Title = "New title"
	wantIssues = issues

	repositories, err = db.GetRepositoriesWithIssues(&pb.Repository{})
	for _, repo := range repositories {
		err := FakeSynchronizeIssues(ctx, db, repo, tasks)
		if err != nil {
			t.Fatal(err)
		}
	}
	gotIssues, err = db.GetIssues(&pb.Issue{})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantIssues, gotIssues, protocmp.Transform()); diff != "" {
		t.Errorf("Synchronization mismatch (-wantIssues, +gotIssues):\n%s", diff)
	}
	// -------------------------------------------------------------------------- //

	// Need to check for issue not represented with task (maybe)
}

// Oje - Temporary test for figuring out how to handle storing tasks in the database
func TestUpdateAssignments(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	// ctx := context.Background()
	admin := qtest.CreateFakeUser(t, db, uint64(1))
	qtest.CreateCourse(t, db, admin, &pb.Course{})

	foundAssignments1 := []*pb.Assignment{
		{
			CourseID: 1,
			Name:     "Lab1",
			Order:    1,
		},
		{
			CourseID: 1,
			Name:     "Lab2",
			Order:    2,
		},
	}

	for _, assignment := range foundAssignments1 {
		err := db.CreateAssignment(assignment)
		if err != nil {
			t.Fatal(err)
		}
	}

	tasks := []*pb.Task{
		{
			AssignmentID:    1,
			AssignmentOrder: 1,
			Title:           "Lab1, task1",
			Body:            "Description of task1 in lab1",
			Name:            "Lab1/task1.md",
		},
		{
			AssignmentID:    1,
			AssignmentOrder: 1,
			Title:           "Lab1, task2",
			Body:            "Description of task2 in lab1",
			Name:            "Lab1/task2.md",
		},
		{
			AssignmentID:    2,
			AssignmentOrder: 2,
			Title:           "Lab2, task1",
			Body:            "Description of task1 in lab2",
			Name:            "Lab2/task1.md",
		},
		{
			AssignmentID:    2,
			AssignmentOrder: 2,
			Title:           "Lab2, task2",
			Body:            "Description of task2 in lab2",
			Name:            "Lab2/task2.md",
		},
	}
	foundAssignments1[0].Tasks = tasks[:2]
	foundAssignments1[1].Tasks = tasks[2:]
	err := db.UpdateAssignments(foundAssignments1)
	if err != nil {
		t.Fatal(err)
	}
	dbAssignments, err := db.GetAssignmentsWithTasks(&pb.Assignment{})
	if err != nil {
		t.Fatal(err)
	}

	for _, assignment := range dbAssignments {
		fmt.Printf("\n%v", assignment.Name)
		for _, task := range assignment.Tasks {
			fmt.Printf("\n%v", task)
		}
	}

	// Try adding one assignment
	fmt.Printf("\n\nAdding one new assignment\n\n")
	foundAssignments2 := []*pb.Assignment{}
	for _, assignment := range foundAssignments1 {
		temp := *assignment
		temp.Tasks = []*pb.Task{}
		for _, task := range assignment.Tasks {
			temp2 := *task
			temp.Tasks = append(temp.Tasks, &temp2)
		}
		foundAssignments2 = append(foundAssignments2, &temp)
	}
	for _, assignment := range foundAssignments2 {
		if assignment.Name == "Lab1" {
			assignment.Tasks = append(assignment.Tasks, &pb.Task{
				AssignmentID:    1,
				AssignmentOrder: 1,
				Title:           "New task",
				Body:            "Description description",
				Name:            "NewTask",
			})
		}
	}
	err = db.UpdateAssignments(foundAssignments2)
	if err != nil {
		t.Fatal(err)
	}
	dbAssignments, err = db.GetAssignmentsWithTasks(&pb.Assignment{})
	if err != nil {
		t.Fatal(err)
	}
	for _, assignment := range dbAssignments {
		fmt.Printf("\n%v", assignment.Name)
		for _, task := range assignment.Tasks {
			fmt.Printf("\n%v", task)
		}
	}
	// Seems to work fine
	// ------------------------------------------------------------------//

	// Try adding one assignment, but without previous tasks in slice
	fmt.Printf("\n\nAdding one new assignment, but solo\n\n")
	foundAssignments3 := []*pb.Assignment{}
	for _, assignment := range foundAssignments1 {
		temp := *assignment
		temp.Tasks = []*pb.Task{}
		for _, task := range assignment.Tasks {
			temp2 := *task
			temp.Tasks = append(temp.Tasks, &temp2)
		}
		foundAssignments3 = append(foundAssignments2, &temp)
	}
	for _, assignment := range foundAssignments2 {
		if assignment.Name == "Lab1" {
			assignment.Tasks = []*pb.Task{
				{
					AssignmentID:    1,
					AssignmentOrder: 1,
					Title:           "Only one task now in 'tests'",
					Body:            "Description description",
					Name:            "Solo task",
				},
			}
		}
	}
	err = db.UpdateAssignments(foundAssignments3)
	if err != nil {
		t.Fatal(err)
	}
	dbAssignments, err = db.GetAssignmentsWithTasks(&pb.Assignment{})
	if err != nil {
		t.Fatal(err)
	}
	for _, assignment := range dbAssignments {
		fmt.Printf("\n%v", assignment.Name)
		for _, task := range assignment.Tasks {
			fmt.Printf("\n%v", task)
		}
	}
	// In this situation the solo task is simply appended to the existing list of tasks for that assignment
	// With the current implementation this would cause problems, since tasks in db is supposed to always be whatever is found
	// ------------------------------------------------------------------//
}
