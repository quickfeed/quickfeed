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
	"go.uber.org/zap"
)

// Should review this function
func InitializeDbEnvironment(t *testing.T, c context.Context, course *pb.Course, s scm.SCM) (database.Database, func(), error) {
	db, cleanup := qtest.TestDB(t)

	org, err := s.GetOrganization(c, &scm.GetOrgOptions{Name: course.Name})
	if err != nil {
		return nil, nil, err
	}

	repos, err := s.GetRepositories(c, org)
	if err != nil {
		return nil, nil, err
	}

	n := 1
	for _, repo := range repos {
		if !strings.HasSuffix(repo.Path, "-labs") {
			continue
		}

		user := qtest.CreateFakeUser(t, db, uint64(n))
		err = db.CreateRepository(&pb.Repository{
			RepositoryID:   repo.ID,
			OrganizationID: 1,
			UserID:         user.ID,
			HTMLURL:        repo.WebURL,
			RepoType:       pb.Repository_USER,
		})
		if err != nil {
			return nil, nil, err
		}
		n++
	}
	return db, cleanup, nil
}

// TestGetTasks retrieves all tasks of a given course.
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
			t.Logf("\n\nTask AssignmentID: %d\nTask GitIssueID: %d\nTask IssueNumber: %d\nTask Name: %s\nTask Title: %s\nTask Body: %s\n\n", task.AssignmentID, task.GitIssueID, task.IssueNumber, task.Name, task.Title, task.Body)
		}
	}
}

// This test should get all issues on "-labs" repos
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
		if !strings.HasSuffix(repo.Path, "-labs") {
			continue
		}
		fmt.Printf("\n\nIssues on repo %s:\n", repo.Path)
		issues, err := s.GetRepoIssues(ctx, &scm.IssueOptions{
			Organization: course.Name,
			Repository:   repo.Path,
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, issue := range issues {
			fmt.Printf("Issue ID: %d\nIssue IssueNumber: %d\nIssue title: %s\nIssue body: %s\n\n", issue.ID, issue.IssueNumber, issue.Title, issue.Body)
		}
	}

	if err != nil {
		t.Fatal(err)
	}
}

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

	err = HandleTasks(ctx, logger, db, s, course, assignments)
	if err != nil {
		t.Fatal(err)
	}
}

// TestAssociateTasksAndIssues creates issues based on tasks in assignments, then checks association between them.
func TestAssociateTasksAndIssues(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	ctx := context.Background()
	logger := zap.NewNop().Sugar()

	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	course := &pb.Course{
		Name:             qfTestOrg,
		OrganizationPath: qfTestOrg,
	}

	// org, err := s.GetOrganization(ctx, &scm.GetOrgOptions{Name: course.Name})
	// if err != nil {
	// 	t.Fatal(err)
	// }

	db, callback, err := InitializeDbEnvironment(t, ctx, course, s)
	defer callback()
	if err != nil {
		t.Fatal(err)
	}

	assignments, _, err := fetchAssignments(ctx, logger, s, course)
	if len(assignments) == 0 {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	// This test shouldn't actually use HandleTasks, and we are only doing so since it is not yet fully implemented
	err = HandleTasks(ctx, logger, db, s, course, assignments)
	if err != nil {
		t.Fatal(err)
	}

	// TO BE CONTINUED

	// repos, err := s.GetRepositories(c, org)
	// if err != nil {
	// 	return err
	// }

	// db.GetRepositories(&pb.Repository{OrganizationID: org.ID})

}
