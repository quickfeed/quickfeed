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
	qfTestOrg := scm.GetTestOrganization(t)
	s, _ := scm.GetTestSCM(t)

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &qf.Course{
		Name:             "QuickFeed Test Course",
		OrganizationName: qfTestOrg,
	}
	if err := PopulateDatabaseWithInitialData(t, db, s, course); err != nil {
		t.Fatal(err)
	}

	assignments := qtest.AssignmentsWithTasks(course.ID)
	for _, assignment := range assignments {
		assignment.CourseID = course.GetID()
		if err := db.CreateAssignment(assignment); err != nil {
			t.Error(err)
		}
	}

	ctx := context.Background()
	repos, err := s.GetRepositories(ctx, &qf.Organization{Name: course.OrganizationName})
	if err != nil {
		t.Fatal(err)
	}

	// Delete all issues on student repositories
	repoFn(repos, func(repo *scm.Repository) {
		if err := s.DeleteIssues(ctx, &scm.RepositoryOptions{
			Owner: course.OrganizationName,
			Path:  repo.Path,
		}); err != nil {
			t.Fatal(err)
		}
		t.Logf("Deleted issues at repo: %+v", repo.Path)
	})

	// Create issues on student repositories for the first assignment's tasks
	first := assignments[:1]
	t.Logf("Synchronizing tasks with issues for assignment %d with %d tasks", first[0].GetOrder(), len(first[0].Tasks))
	if err := synchronizeTasksWithIssues(ctx, db, s, course, first); err != nil {
		t.Fatal(err)
	}

	// Check if the issues were created
	repoFn(repos, func(repo *scm.Repository) {
		scmIssues, err := s.GetIssues(ctx, &scm.RepositoryOptions{
			Owner: course.OrganizationName,
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
	if err := synchronizeTasksWithIssues(ctx, db, s, course, second); err != nil {
		t.Fatal(err)
	}

	// Check if the issues were created
	repoFn(repos, func(repo *scm.Repository) {
		scmIssues, err := s.GetIssues(ctx, &scm.RepositoryOptions{
			Owner: course.OrganizationName,
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

// PopulateDatabaseWithInitialData creates initial data-records based on organization
// This function was created with the intent of being used for testing task and pull request related functionality.
func PopulateDatabaseWithInitialData(t *testing.T, db database.Database, sc scm.SCM, course *qf.Course) error {
	t.Helper()

	ctx := context.Background()
	org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{Name: course.OrganizationName})
	if err != nil {
		return err
	}
	course.OrganizationID = org.GetID()
	admin := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateCourse(t, db, admin, course)

	repos, err := sc.GetRepositories(ctx, org)
	if err != nil {
		return err
	}

	// Create repositories
	nxtRemoteID := uint64(2)
	for _, repo := range repos {
		dbRepo := &qf.Repository{
			RepositoryID:   repo.ID,
			OrganizationID: org.GetID(),
			HTMLURL:        repo.HTMLURL,
			RepoType:       qf.RepoType(repo.Path),
		}
		if dbRepo.IsUserRepo() {
			user := qtest.CreateFakeUser(t, db, nxtRemoteID)
			nxtRemoteID++
			qtest.EnrollStudent(t, db, user, course)
			group := &qf.Group{
				Name:     dbRepo.UserName(),
				CourseID: course.GetID(),
				Users:    []*qf.User{user},
			}
			if err := db.CreateGroup(group); err != nil {
				return err
			}
			// For testing purposes, assume all student repositories are group repositories
			// since tasks and pull requests are only supported for groups anyway.
			dbRepo.RepoType = qf.Repository_GROUP
			dbRepo.GroupID = group.GetID()
		}

		t.Logf("create repo: %v", dbRepo)
		if err = db.CreateRepository(dbRepo); err != nil {
			return err
		}
	}
	return nil
}
