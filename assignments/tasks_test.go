package assignments

import (
	"context"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/scm"
	"go.uber.org/zap"
)

type issueInformation struct {
	IssueNumber uint64
	Name        string
}

// When running tests that have anything to do with tasks/issues, it is important that issues have their title corresponding to the name of an associated task.
// For example, if you have an issue that is supposed to be connected to the task "task-hello_world.md" in "lab1", the title of this issue needs to be "lab1/hello_world".
// Otherwise when creating the database there will be no clear way to know which issue is supposed to be associated with which task.
// InitializeDbEnvironment is supposed to create a database environment that would be representative of the state of the organization on the previous push to the tests repository.

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

	// Find and create assignments
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

	foundIssues := make(map[uint64]map[string]*issueInformation)
	tasks := make(map[string]*pb.Task)

	// Create repositories
	n := 2
	for _, repo := range repos {
		var user *pb.User
		// Might not even be necessary to handle repos differently in these tests.
		var dbRepo *pb.Repository
		switch repo.Path {
		case "course-" + pb.InfoRepo:
			dbRepo = &pb.Repository{
				RepositoryID:   repo.ID,
				OrganizationID: org.GetID(),
				HTMLURL:        repo.WebURL,
				RepoType:       pb.Repository_COURSEINFO,
			}
		case pb.AssignmentRepo:
			dbRepo = &pb.Repository{
				RepositoryID:   repo.ID,
				OrganizationID: org.GetID(),
				HTMLURL:        repo.WebURL,
				RepoType:       pb.Repository_ASSIGNMENTS,
			}
		case pb.TestsRepo:
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
				// Since tasks are only to be managed for group-repositories, we assume while testing, that every student-repository is a group-repository
				RepoType: pb.Repository_GROUP,
			}
		}

		err = db.CreateRepository(dbRepo)
		if err != nil {
			return db, cleanup, err
		}

		existingScmIssues, err := s.GetRepoIssues(c, &scm.RepositoryOptions{
			Owner: course.Name,
			Path:  repo.Path,
		})
		if err != nil {
			return db, cleanup, err
		}

		if len(existingScmIssues) == 0 {
			continue
		}
		foundIssues[repo.ID] = make(map[string]*issueInformation)
		for _, scmIssue := range existingScmIssues {
			foundIssues[repo.ID][scmIssue.Title] = &issueInformation{
				IssueNumber: uint64(scmIssue.IssueNumber),
				Name:        scmIssue.Title,
			}
			tasks[scmIssue.Title] = &pb.Task{
				Title: scmIssue.Title,
				Body:  scmIssue.Body,
				Name:  scmIssue.Title,
			}
		}
		n++
	}

	// We remove from foundTasks every task that is not represented by an issue. This way, the database is initialized with tasks based on issues and tasks, and not just the tasks found in the tests repository.
	foundTasks := getTasksFromAssignments(foundAssignments)
	for _, taskMap := range foundTasks {
		for _, task := range taskMap {
			if _, ok := tasks[task.Name]; !ok {
				delete(taskMap, task.Name)
			} else {
				task.Title = tasks[task.Name].Title
				task.Body = tasks[task.Name].Body
			}
		}
	}
	createdTasks, _, _, err := db.SynchronizeAssignmentTasks(course, foundTasks)
	if err != nil {
		return db, cleanup, err
	}

	createdTasksMap := make(map[string]*pb.Task)
	for _, createdTask := range createdTasks {
		createdTasksMap[createdTask.Name] = createdTask
	}

	dbRepos, err := db.GetRepositoriesWithIssues(&pb.Repository{
		OrganizationID: course.GetOrganizationID(),
	})
	if err != nil {
		return db, cleanup, err
	}

	issuesToCreate := []*pb.Issue{}
	for _, repo := range dbRepos {
		if !repo.IsGroupRepo() {
			continue
		}
		for _, createdTask := range createdTasksMap {
			foundIssue := foundIssues[repo.RepositoryID][createdTask.Name]
			issuesToCreate = append(issuesToCreate, &pb.Issue{
				RepositoryID: repo.ID,
				TaskID:       createdTasksMap[foundIssue.Name].ID,
				IssueNumber:  foundIssue.IssueNumber,
			})
		}
	}
	err = db.CreateIssues(issuesToCreate)
	return db, cleanup, err
}

// TestInitializeDbEnvironment tests if db is correctly initialized.
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
			t.Logf("\nIssue ID: %d\nIssue RepositoryID: %d\nIssue TaskID: %d\nIssue IssueNumber: %d", issue.ID, issue.RepositoryID, issue.TaskID, issue.IssueNumber)
		}
	}

	assignments, err := db.GetAssignmentsByCourse(course.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("\n\nAssignments:\n")
	for _, assignment := range assignments {
		t.Logf("\nAssignment ID: %d\nAssignment Name: %s", assignment.ID, assignment.Name)
		tasks, _ := db.GetTasks(&pb.Task{AssignmentID: assignment.GetID()})
		for _, task := range tasks {
			t.Logf("\nTask ID: %d\nTask AssignmentID: %d\nTask Name: %s\nTask Title: %s\nTask Body: %s", task.ID, task.AssignmentID, task.Name, task.Title, task.Body)
		}
	}
}

// TestHandleTasks runs handleTasks() on specified org. Results vary depending on what tasks/issues existed prior to running.
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

	assignments, _, err := fetchAssignments(ctx, logger, s, course)
	if err != nil {
		t.Fatal(err)
	}

	err = handleTasks(ctx, db, s, course, assignments)
	if err != nil {
		t.Fatal(err)
	}
}
