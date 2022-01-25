package hooks

import (
	"context"
	"log"
	"net/http"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/internal/qtest"
	logq "github.com/autograde/quickfeed/log"
	"github.com/autograde/quickfeed/scm"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
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

	// TODO(meling) db is nil; will cause handling of push event to panic; will need a database with content for this to work fully.
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

func TestRecordResultsForManualReview(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	logger := logq.Zap(true).Sugar()
	defer func() { _ = logger.Sync() }()
	var runner ci.Runner
	testhook := NewGitHubWebHook(logger, db, runner, secret)

	course := &pb.Course{
		Name:           "Test",
		OrganizationID: 1,
		SlipDays:       5,
	}
	admin := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateCourse(t, db, admin, course)

	assignment := &pb.Assignment{
		Order:      1,
		CourseID:   course.ID,
		Name:       "assignment-1",
		Deadline:   "2022-11-11T13:00:00",
		IsGroupLab: false,
		Reviewers:  1,
	}
	if err := db.CreateAssignment(assignment); err != nil {
		t.Fatal(err)
	}

	initialSubmission := &pb.Submission{
		AssignmentID: assignment.ID,
		UserID:       admin.ID,
		Score:        80,
		Status:       pb.Submission_APPROVED,
		Released:     true,
	}
	if err := db.CreateSubmission(initialSubmission); err != nil {
		t.Fatal(err)
	}

	runData := &ci.RunData{
		Course:     course,
		Assignment: assignment,
		Repo: &pb.Repository{
			UserID: 1,
		},
		JobOwner: "test",
	}

	testhook.recordSubmissionWithoutTests(runData)
	query := &pb.Submission{
		AssignmentID: assignment.ID,
		UserID:       admin.ID,
	}
	updatedSubmission, err := db.GetSubmission(query)
	if err != nil {
		t.Fatal(err)
	}
	// submission must stay approved, released, with score = 80
	if diff := cmp.Diff(initialSubmission, updatedSubmission, protocmp.Transform(), protocmp.IgnoreFields(&pb.Submission{}, "BuildInfo", "Scores")); diff != "" {
		t.Errorf("Incorrect submission after update. Want: %+v, got %+v", initialSubmission, updatedSubmission)
	}
}
