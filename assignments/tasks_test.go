package assignments

import (
	"context"
	"strings"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/scm"
	"go.uber.org/zap"
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

	// Get created tasks
	tasks := []*pb.Task{}
	for _, assignment := range foundAssignments {
		dbTasks, err := db.GetTasks(&pb.Task{AssignmentID: uint64(assignment.GetOrder())})
		if err != nil {
			return db, cleanup, err
		}
		tasks = append(tasks, dbTasks...)
	}
	taskMap := make(map[string]*pb.Task)
	for _, task := range tasks {
		taskMap[task.Name] = task
	}

	// Create repositories
	repos, err := s.GetRepositories(c, org)
	if err != nil {
		return db, cleanup, err
	}

	// Create issues for repositories
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
			correspondingTask, ok := taskMap[scmIssue.Title]
			dbIssue := &pb.Issue{}
			if !ok {
				dbIssue = &pb.Issue{
					RepositoryID: dbRepo.ID,
					IssueNumber:  uint64(scmIssue.IssueNumber),
					Name:         scmIssue.Title,
					Title:        scmIssue.Title,
					Body:         scmIssue.Body,
				}
			} else {
				dbIssue = &pb.Issue{
					RepositoryID: dbRepo.ID,
					TaskID:       correspondingTask.ID,
					IssueNumber:  uint64(scmIssue.IssueNumber),
					Name:         scmIssue.Title,
					Title:        scmIssue.Title,
					Body:         scmIssue.Body,
				}
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

	repos, err := GetRepositoriesByOrgID(db, org.ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("\n\nRepositories and issues:\n")
	for _, repo := range repos {
		t.Logf("\nRepository ID: %d\nRepository HTMLURL: %s\nRepository UserID: %d", repo.ID, repo.HTMLURL, repo.UserID)
		for _, issue := range repo.Issues {
			t.Logf("\nIssue ID: %d\nIssue RepositoryID: %d\nIssue TaskID: %d\nIssue IssueNumber: %d\nIssue Name: %s\nIssue Title: %s\nIssue Body: %s", issue.ID, issue.RepositoryID, issue.TaskID, issue.IssueNumber, issue.Name, issue.Title, issue.Body)
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
	repos, err := GetRepositoriesByOrgID(db, org.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, repo := range repos {
		t.Logf("\nRepository ID: %d\nRepository HTMLURL: %s\nRepository UserID: %d", repo.ID, repo.HTMLURL, repo.UserID)
		for _, issue := range repo.Issues {
			t.Logf("\nIssue ID: %d\nIssue RepositoryID: %d\nIssue TaskID: %d\nIssue IssueNumber: %d\nIssue Name: %s\nIssue Title: %s\nIssue Body: %s", issue.ID, issue.RepositoryID, issue.TaskID, issue.IssueNumber, issue.Name, issue.Title, issue.Body)
		}
	}

	assignments, _, err := fetchAssignments(ctx, logger, s, course)
	if err != nil {
		t.Fatal(err)
	}

	err = HandleTasks(ctx, db, s, course, assignments)
	if err != nil {
		t.Fatal(err)
	}

	// Db contents after
	t.Logf("\n\n\nDB AFTER\n\n\n")
	repos, err = GetRepositoriesByOrgID(db, org.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, repo := range repos {
		t.Logf("\nRepository ID: %d\nRepository HTMLURL: %s\nRepository UserID: %d", repo.ID, repo.HTMLURL, repo.UserID)
		for _, issue := range repo.Issues {
			t.Logf("\nIssue ID: %d\nIssue RepositoryID: %d\nIssue TaskID: %d\nIssue IssueNumber: %d\nIssue Name: %s\nIssue Title: %s\nIssue Body: %s", issue.ID, issue.RepositoryID, issue.TaskID, issue.IssueNumber, issue.Name, issue.Title, issue.Body)
		}
	}
}
