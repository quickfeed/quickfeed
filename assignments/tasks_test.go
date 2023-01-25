package assignments

import (
	"context"
	"testing"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

// TestSynchronizeTasksWithIssues synchronizes tasks with issues on user repositories
// on the specified test organization. The test deletes existing issues on the test
// organization first, before synchronizing the tasks to issues. The test will leave
// behind newly created issues on the user repositories for manual inspection.
func TestSynchronizeTasksWithIssues(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	scmApp := scm.GetAppSCM(t)
	course, assignments, repos := initDatabase(t, db, scmApp)
	ctx := context.Background()

	// Delete all issues on student repositories
	repoFn(repos, func(repo *scm.Repository) {
		if err := scmApp.DeleteIssues(ctx, &scm.RepositoryOptions{
			Owner: course.ScmOrganizationName,
			Path:  repo.Path,
		}); err != nil {
			t.Fatal(err)
		}
		t.Logf("Deleted issues at repo: %+v", repo.Path)
	})

	// Create issues on student repositories for the first assignment's tasks
	first := assignments[:1]
	t.Logf("Synchronizing tasks with issues for assignment %d with %d tasks", first[0].GetOrder(), len(first[0].Tasks))
	if err := synchronizeTasksWithIssues(ctx, db, scmApp, course, first); err != nil {
		t.Fatal(err)
	}

	// Check if the issues were created
	repoFn(repos, func(repo *scm.Repository) {
		scmIssues, err := scmApp.GetIssues(ctx, &scm.RepositoryOptions{
			Owner: course.ScmOrganizationName,
			Path:  repo.Path,
		})
		if err != nil {
			t.Fatal(err)
		}
		issues := make(map[string]*scm.Issue)
		for _, issue := range scmIssues {
			t.Logf("Found issue (%s): %s", repo.Path, issue.Title)
			issues[issue.Title] = issue
		}
		for _, task := range first[0].Tasks {
			if _, ok := issues[task.Title]; !ok {
				t.Errorf("task.Title = %s not found in repo %s", task.Title, repo.Path)
			}
		}
	})

	// Create issues on student repositories for the second assignment's tasks
	second := assignments[1:]
	t.Logf("Synchronizing tasks with issues for assignment %d with %d tasks", second[0].GetOrder(), len(second[0].Tasks))
	if err := synchronizeTasksWithIssues(ctx, db, scmApp, course, second); err != nil {
		t.Fatal(err)
	}

	// Check if the issues were created
	repoFn(repos, func(repo *scm.Repository) {
		scmIssues, err := scmApp.GetIssues(ctx, &scm.RepositoryOptions{
			Owner: course.ScmOrganizationName,
			Path:  repo.Path,
		})
		if err != nil {
			t.Fatal(err)
		}
		issues := make(map[string]*scm.Issue)
		for _, issue := range scmIssues {
			t.Logf("Found issue (%s): %s", repo.Path, issue.Title)
			issues[issue.Title] = issue
		}
		for _, task := range second[0].Tasks {
			if _, ok := issues[task.Title]; !ok {
				t.Errorf("task.Title = %s not found in repo %s", task.Title, repo.Path)
			}
		}
	})
}

func repoFn(repos []*scm.Repository, fn func(repo *scm.Repository)) {
	for _, repo := range repos {
		if qf.RepoType(repo.Path).IsCourseRepo() {
			continue
		}
		fn(repo)
	}
}

// initDatabase creates initial data-records based on the QF_TEST_ORG organization.
// This function is used for testing task and pull request related functionality.
func initDatabase(t *testing.T, db database.Database, sc scm.SCM) (*qf.Course, []*qf.Assignment, []*scm.Repository) {
	t.Helper()

	qfTestOrg := scm.GetTestOrganization(t)
	ctx := context.Background()
	org, err := sc.GetOrganization(ctx, &scm.OrganizationOptions{Name: qfTestOrg})
	if err != nil {
		t.Fatal(err)
	}

	course := &qf.Course{
		Name:                "QuickFeed Test Course",
		ScmOrganizationName: org.GetScmOrganizationName(),
		ScmOrganizationID:   org.GetScmOrganizationID(),
	}
	admin := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateCourse(t, db, admin, course)

	repos, err := sc.GetRepositories(ctx, org)
	if err != nil {
		t.Fatal(err)
	}

	// Add repositories to the database
	for _, scmRepo := range repos {
		repo := &qf.Repository{
			ScmRepositoryID:   scmRepo.ID,
			ScmOrganizationID: org.GetScmOrganizationID(),
			HTMLURL:           scmRepo.HTMLURL,
			RepoType:          qf.RepoType(scmRepo.Path),
		}
		if repo.IsUserRepo() {
			user := qtest.CreateFakeUser(t, db, 0)
			qtest.EnrollStudent(t, db, user, course)
			group := &qf.Group{
				Name:     repo.UserName(),
				CourseID: course.GetID(),
				Users:    []*qf.User{user},
			}
			if err := db.CreateGroup(group); err != nil {
				t.Fatal(err)
			}
			// For testing purposes, assume all student repositories are group repositories
			// since tasks and pull requests are only supported for groups anyway.
			repo.RepoType = qf.Repository_GROUP
			repo.GroupID = group.GetID()
		}
		t.Logf("Creating repo in database: %v", scmRepo.Path)
		if err = db.CreateRepository(repo); err != nil {
			t.Fatal(err)
		}
	}

	assignments := []*qf.Assignment{
		{
			CourseID:    course.GetID(),
			Name:        "lab1",
			Deadline:    qtest.Timestamp(t, "2022-12-01T19:00:00"),
			AutoApprove: false,
			Order:       1,
			IsGroupLab:  false,
			Tasks: []*qf.Task{
				{Title: "Fibonacci", Name: "fib", AssignmentOrder: 1, Body: "Implement fibonacci"},
				{Title: "Lucas Numbers", Name: "luc", AssignmentOrder: 1, Body: "Implement lucas numbers"},
			},
		},
		{
			CourseID:    course.GetID(),
			Name:        "lab2",
			Deadline:    qtest.Timestamp(t, "2022-12-12T19:00:00"),
			AutoApprove: false,
			Order:       2,
			IsGroupLab:  false,
			Tasks: []*qf.Task{
				{Title: "Addition", Name: "add", AssignmentOrder: 2, Body: "Implement addition"},
				{Title: "Subtraction", Name: "sub", AssignmentOrder: 2, Body: "Implement subtraction"},
				{Title: "Multiplication", Name: "mul", AssignmentOrder: 2, Body: "Implement multiplication"},
			},
		},
	}
	for _, assignment := range assignments {
		if err := db.CreateAssignment(assignment); err != nil {
			t.Fatal(err)
		}
	}
	return course, assignments, repos
}
