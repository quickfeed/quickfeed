package assignments

import (
	"context"
	"testing"

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
	s := scm.GetTestSCM(t)

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &qf.Course{
		Name:             "QuickFeed Test Course",
		OrganizationName: qfTestOrg,
		Provider:         "github",
	}
	if err := qtest.PopulateDatabaseWithInitialData(t, db, s, course); err != nil {
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
