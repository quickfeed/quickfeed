package assignments

import (
	"context"
	"strconv"
	"strings"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/log"
	"github.com/autograde/quickfeed/scm"
	"go.uber.org/zap"
)

type foundIssue struct {
	IssueNumber uint64
	Name        string
}

// The test environment creates tasks based on the issues found on student repositories in the organization.
// This is so that we can emulate the result of a previous push to the tests repository.
// For these tasks to be created correctly the title of each task must use the following format:
//   "<task name>, <assignment order>".
//
// For example if an issue is related to "task-hello_world.md" in "lab1",
// then the title of the corresponding issue should be:
//   "lab1/hello_world, 1" (assuming lab1 has assignment order 1).
//
// It is also recommended that issues are created on all student repositories, and that they are the same.

// populateDatabaseWithTasks based on the given course's organization.
func populateDatabaseWithTasks(t *testing.T, ctx context.Context, logger *zap.SugaredLogger, db database.Database, sc scm.SCM, course *pb.Course) error {
	t.Helper()

	org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{Name: course.Name})
	if err != nil {
		return err
	}

	// Find and create assignments
	foundAssignments, _, err := fetchAssignments(ctx, logger, sc, course)
	if err != nil {
		return err
	}

	if err = db.UpdateAssignments(foundAssignments); err != nil {
		return err
	}

	repos, err := sc.GetRepositories(ctx, org)
	if err != nil {
		return err
	}

	foundIssues := make(map[uint64]map[string]*foundIssue)
	tasks := make(map[uint32]map[string]*pb.Task)

	// Finds issues, and creates tasks based on them
	for _, repo := range repos {
		existingScmIssues, err := sc.GetRepoIssues(ctx, &scm.RepositoryOptions{
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

	createdTasks, _, _, err := db.SynchronizeAssignmentTasks(course, tasks)
	if err != nil {
		return err
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

	return db.CreateIssues(issuesToCreate)
}

// TestHandleTasks runs handleTasks() on the specified organization.
// Results vary depending on what tasks/issues existed prior to running.
func TestHandleTasks(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)

	logger := log.Zap(false).Sugar()
	scm, err := scm.NewSCMClient(logger, "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	course := &pb.Course{
		Name:             qfTestOrg,
		OrganizationPath: qfTestOrg,
	}

	ctx := context.Background()
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	if err = qtest.PopulateDatabaseWithInitialData(t, ctx, db, scm, course); err != nil {
		t.Fatal(err)
	}
	if err = populateDatabaseWithTasks(t, ctx, logger, db, scm, course); err != nil {
		t.Fatal(err)
	}

	assignments, _, err := fetchAssignments(ctx, logger, scm, course)
	if err != nil {
		t.Fatal(err)
	}

	if err = handleTasks(ctx, db, scm, course, assignments); err != nil {
		t.Fatal(err)
	}
}
