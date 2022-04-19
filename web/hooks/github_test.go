package hooks

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/assignments"
	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/internal/qtest"
	logq "github.com/autograde/quickfeed/log"
	"github.com/autograde/quickfeed/scm"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v35/github"
	"go.uber.org/zap"
)

const (
	secret = "the-secret-quickfeed-test"
)

// To run this test, please see instructions in the developer guide (dev.md).

// On macOS, get ngrok using `brew install ngrok`.
// See steps to follow [here](https://groob.io/tutorial/go-github-webhook/).

// To run this test, use the following (replace the forwarding URL with your own):
//
// QF_WEBHOOK_SERVER=https://53c51fa9.ngrok.io go test -v -run TestGitHubWebHook
//
// This will create a new webhook with URL `https://53c51fa9.ngrok.io/webhook`
// for the $QF_TEST_ORG/tests repository for handling push events.
//
// This test will then block waiting for a push event from GitHub; meaning that you
// will manually have to create a push event to the 'tests' repository.
//
// TODO(meling) add code to create a push event to the tests repository.

type foundIssue struct {
	IssueNumber uint64
	Name        string
}

func TestGitHubWebHook(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	serverURL := scm.GetWebHookServer(t)

	logger := logq.Zap(true).Sugar()
	defer func() { _ = logger.Sync() }()

	s, err := scm.NewSCMClient(logger, "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	opt := &scm.CreateHookOptions{
		URL:        serverURL + "/webhook",
		Secret:     secret,
		Repository: &scm.Repository{Owner: qfTestOrg, Path: "tests"},
	}
	err = s.CreateHook(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}

	hooks, err := s.ListHooks(ctx, opt.Repository, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, hook := range hooks {
		t.Logf("hook: %v", hook)
	}

	var db database.Database
	var runner ci.Runner
	webhook := NewGitHubWebHook(logger, db, runner, secret)

	log.Println("starting webhook server")
	http.HandleFunc("/webhook", webhook.Handle)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func TestExtractChanges(t *testing.T) {
	modifiedFiles := []string{
		"go.mod",
		"go.sum",
		"exercise.go",
		"README.md",
		"lab2/fib.go",
		"lab3/detector/fd.go",
		"paxos/proposer.go",
		"/hallo",
		"",
	}
	want := map[string]bool{
		"lab2":  true,
		"lab3":  true,
		"paxos": true,
	}
	got := make(map[string]bool)
	extractChanges(modifiedFiles, got)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("content mismatch (-want +got):\n%s", diff)
	}
}

func TestGitHubPRWebHook(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)
	serverURL := scm.GetWebHookServer(t)

	logger := logq.Zap(true).Sugar()
	defer func() { _ = logger.Sync() }()

	s, err := scm.NewSCMClient(logger, "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	opt := &scm.CreateHookOptions{
		URL:          serverURL + "/webhook",
		Secret:       secret,
		Organization: qfTestOrg,
	}
	err = s.CreateHook(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}

	course := &pb.Course{
		Name:             qfTestOrg,
		OrganizationPath: qfTestOrg,
	}

	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	if err := qtest.PopulateDatabaseWithInitialData(t, db, s, course); err != nil {
		t.Fatal(err)
	}
	if err := populateDatabaseWithTasks(t, ctx, logger, db, s, course); err != nil {
		t.Fatal(err)
	}
	pb.SetAccessToken(course.GetID(), accessToken)
	var runner ci.Runner
	webhook := NewGitHubWebHook(logger, db, runner, secret)

	log.Println("starting webhook server")
	http.HandleFunc("/webhook", webhook.HandlePR)
	log.Fatal(http.ListenAndServe(":4567", nil))
}

func (wh GitHubWebHook) HandlePR(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(wh.secret))
	if err != nil {
		wh.logger.Errorf("Error in request body: %v", err)
		return
	}
	defer r.Body.Close()

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		wh.logger.Errorf("Could not parse github webhook: %v", err)
		return
	}

	switch e := event.(type) {
	case *github.PushEvent:
		repos, err := wh.db.GetRepositories(&pb.Repository{RepositoryID: uint64(e.GetRepo().GetID())})
		if err != nil {
			wh.logger.Errorf("Failed to get repository by remote ID %d from database: %v", e.GetRepo().GetID(), err)
			return
		}
		if len(repos) != 1 {
			wh.logger.Debugf("Ignoring pull request opened event for unknown repository: %s", e.GetRepo().GetFullName())
			return
		}
		repo := repos[0]
		course, err := wh.db.GetCourseByOrganizationID(repo.OrganizationID)
		if err != nil {
			wh.logger.Errorf("Failed to get course from database: %v", err)
			return
		}
		course.Provider = "github"
		// Printing db before
		// repos, err = wh.db.GetRepositoriesWithIssues(&pb.Repository{OrganizationID: course.GetOrganizationID()})
		// for _, repo := range repos {
		// 	fmt.Printf("\nRepository: %s", repo.Name())
		// 	for _, issue := range repo.Issues {
		// 		fmt.Printf("\nIssue ID: %d, issue TaskID: %d", issue.GetID(), issue.GetTaskID())
		// 	}
		// }
		assignments.UpdateFromTestsRepo(wh.logger, wh.db, course)
		// repos, err = wh.db.GetRepositoriesWithIssues(&pb.Repository{OrganizationID: course.GetOrganizationID()})
		// for _, repo := range repos {
		// 	fmt.Printf("\nRepository: %s", repo.Name())
		// 	for _, issue := range repo.Issues {
		// 		fmt.Printf("\nIssue ID: %d, issue TaskID: %d", issue.GetID(), issue.GetTaskID())
		// 	}
		// }
	case *github.PullRequestEvent:
		// wh.logger.Debug(log.IndentJson(e))
		wh.handlePullRequest(e)
	case *github.PullRequestReviewEvent:
		// wh.logger.Debug(log.IndentJson(e))
		wh.handlePullRequestReview(e)
	default:
		wh.logger.Debugf("Ignored event type %s", github.WebHookType(r))
	}
}

func populateDatabaseWithTasks(t *testing.T, ctx context.Context, logger *zap.SugaredLogger, db database.Database, sc scm.SCM, course *pb.Course) error {
	t.Helper()

	org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{Name: course.Name})
	if err != nil {
		return err
	}

	// Find and create assignments
	foundAssignments, _, err := assignments.FetchAssignments(ctx, logger, sc, course)
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

	createdTasks, _, err := db.SynchronizeAssignmentTasks(course, tasks)
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
