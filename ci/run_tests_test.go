package ci_test

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/internal/rand"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/stream"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// To run this test, please see instructions in the developer guide (dev.md).

// This test uses a test course for experimenting with run.sh behavior.
// The tests below will run locally on the test machine, not on the QuickFeed machine.

func loadDockerfile(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("testdata/Dockerfile")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func testRunData(t *testing.T, runner ci.Runner) *ci.RunData {
	dockerfile := loadDockerfile(t)
	qfTestOrg := scm.GetTestOrganization(t)
	// Only used to fetch the user's GitHub login (user name)
	_, userName := scm.GetTestSCM(t)

	repo := qf.RepoURL{ProviderURL: "github.com", Organization: qfTestOrg}
	course := &qf.Course{
		ID:                  1,
		Code:                "QF101",
		ScmOrganizationName: qfTestOrg,
	}
	course.UpdateDockerfile(dockerfile)

	// Emulate running UpdateFromTestsRepo to ensure the docker image is built before running tests.
	t.Logf("Building %s's Dockerfile:\n%v", course.GetCode(), course.GetDockerfile())
	out, err := runner.Run(context.Background(), &ci.Job{
		Name:       course.JobName(),
		Image:      course.DockerImage(),
		Dockerfile: course.GetDockerfile(),
		Commands:   []string{`echo -n "Hello from Dockerfile"`},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(out)

	return &ci.RunData{
		Course: course,
		Assignment: &qf.Assignment{
			Name:             "lab1",
			ContainerTimeout: 1, // minutes
		},
		Repo: &qf.Repository{
			HTMLURL:  repo.StudentRepoURL(userName),
			RepoType: qf.Repository_USER,
		},
		JobOwner: "muggles",
		CommitID: rand.String()[:7],
	}
}

func TestRunTests(t *testing.T) {
	runner, closeFn := dockerClient(t)
	defer closeFn()

	runData := testRunData(t, runner)
	ctx, cancel := runData.Assignment.WithTimeout(2 * time.Minute)
	defer cancel()

	scmClient, _ := scm.GetTestSCM(t)
	results, err := runData.RunTests(ctx, qtest.Logger(t), scmClient, runner)
	if err != nil {
		t.Fatal(err)
	}
	// We don't actually test anything here since we don't know how many assignments are in QF_TEST_ORG
	t.Logf("%+v", results.BuildInfo.BuildLog)
	results.BuildInfo.BuildLog = "removed"
	t.Logf("%+v\n", qlog.IndentJson(results))
}

func TestRunTestsTimeout(t *testing.T) {
	runner, closeFn := dockerClient(t)
	defer closeFn()

	runData := testRunData(t, runner)
	// Note that this timeout value is susceptible to variation
	ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancel()

	scmClient, _ := scm.GetTestSCM(t)
	results, err := runData.RunTests(ctx, qtest.Logger(t), scmClient, runner)
	if err != nil {
		t.Fatal(err)
	}
	const wantOut = `Container timeout. Please check for infinite loops or other slowness.`
	if results.BuildInfo != nil && !strings.HasPrefix(results.BuildInfo.BuildLog, wantOut) {
		t.Errorf("RunTests(1s timeout) = '%s', got '%s'", wantOut, results.BuildInfo.BuildLog)
	}
}

func TestRecordResults(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &qf.Course{
		Name:              "Test",
		Code:              "DAT320",
		ScmOrganizationID: 1,
		SlipDays:          5,
	}
	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, course)

	assignment := &qf.Assignment{
		CourseID:         course.ID,
		Name:             "lab1",
		Deadline:         qtest.Timestamp(t, "2022-11-11T13:00:00"),
		AutoApprove:      true,
		ScoreLimit:       70,
		Order:            1,
		IsGroupLab:       false,
		ContainerTimeout: 1,
	}
	qtest.CreateAssignment(t, db, assignment)
	buildInfo := createBuildInfo(t)
	testScores := createScores()
	// Must create a new submission with correct scores and build info, not approved
	results := &score.Results{
		BuildInfo: buildInfo,
		Scores:    testScores,
	}
	runData := &ci.RunData{
		Course:     course,
		Assignment: assignment,
		Repo: &qf.Repository{
			RepoType: qf.Repository_USER,
			UserID:   1,
		},
		JobOwner: "test",
		CommitID: "deadbeef",
	}

	// Check that submission is recorded correctly
	submission := recordResults(t, runData, db, results, nil, false)
	if submission.IsApproved(runData.Repo.GetUserID()) {
		t.Error("Submission must not be auto approved")
	}
	qtest.Diff(t, "submission score mismatch", testScores, submission.Scores, protocmp.Transform(), protocmp.IgnoreFields(&score.Score{}, "Secret"))
	qtest.Diff(t, "build info mismatch", buildInfo, submission.BuildInfo, protocmp.Transform())

	// When updating submission after deadline: build info (submission and build dates) and slip days must be updated
	newSubmissionDate := qtest.Timestamp(t, "2022-11-12T13:00:00")
	updatedSubmission := recordResults(t, runData, db, results, newSubmissionDate, false)
	enrollment := qtest.GetEnrollment(t, db, course.ID, admin.ID)
	if enrollment.RemainingSlipDays(course) == int32(course.SlipDays) || len(enrollment.UsedSlipDays) < 1 {
		t.Error("Student must have reduced slip days")
	}
	qtest.Diff(t, "build info mismatch", results.BuildInfo, updatedSubmission.BuildInfo, protocmp.Transform())

	// When rebuilding after deadline: delivery date and slip days must stay unchanged, build date must be updated
	wantSubmissionDate := newSubmissionDate
	newDate := qtest.Timestamp(t, "2022-11-13T15:00:00")
	slipDaysBeforeUpdate := enrollment.RemainingSlipDays(course)
	rebuiltSubmission := recordResults(t, runData, db, results, newDate, true)

	qtest.Diff(t, "build date mismatch", newDate, rebuiltSubmission.BuildInfo.BuildDate, protocmp.Transform())
	qtest.Diff(t, "submission date mismatch", wantSubmissionDate, rebuiltSubmission.BuildInfo.SubmissionDate, protocmp.Transform())

	updatedEnrollment := qtest.GetEnrollment(t, db, course.ID, admin.ID)
	qtest.Diff(t, "slip days mismatch", slipDaysBeforeUpdate, updatedEnrollment.RemainingSlipDays(course))
}

func TestRecordResultsForManualReview(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &qf.Course{
		Name:              "Test",
		ScmOrganizationID: 1,
		SlipDays:          5,
	}
	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, course)

	assignment := &qf.Assignment{
		Order:      1,
		CourseID:   course.ID,
		Name:       "assignment-1",
		Deadline:   qtest.Timestamp(t, "2022-11-11T13:00:00"),
		IsGroupLab: false,
		Reviewers:  1,
	}
	qtest.CreateAssignment(t, db, assignment)

	initialSubmission := &qf.Submission{
		AssignmentID: assignment.ID,
		UserID:       admin.ID,
		Score:        80,
		Grades:       []*qf.Grade{{UserID: admin.ID, Status: qf.Submission_APPROVED}},
		Released:     true,
	}
	if err := db.CreateSubmission(initialSubmission); err != nil {
		t.Fatal(err)
	}

	runData := &ci.RunData{
		Course:     course,
		Assignment: assignment,
		Repo: &qf.Repository{
			RepoType: qf.Repository_USER,
			UserID:   admin.ID,
		},
		JobOwner: "test",
	}

	submission := recordResults(t, runData, db, nil, nil, false)

	// make sure all fields were saved correctly in the database
	query := &qf.Submission{
		AssignmentID: assignment.ID,
		UserID:       admin.ID,
	}
	updatedSubmission, err := db.GetSubmission(query)
	if err != nil {
		t.Fatal(err)
	}

	qtest.Diff(t, "Incorrect submission fields in the database", updatedSubmission, submission, protocmp.Transform())
	// submission must stay approved, released, with score = 80
	qtest.Diff(t, "Incorrect submission after update", initialSubmission, updatedSubmission, protocmp.Transform(), protocmp.IgnoreFields(&qf.Submission{}, "BuildInfo", "Scores"))
}

func TestStreamRecordResults(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &qf.Course{
		Name:              "Test",
		Code:              "DAT320",
		ScmOrganizationID: 1,
		SlipDays:          5,
	}
	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, course)
	group := qtest.CreateGroup(t, db, course, 3)

	assignment := &qf.Assignment{
		CourseID:         course.ID,
		Name:             "lab1",
		Deadline:         qtest.Timestamp(t, "2022-11-11T13:00:00"),
		AutoApprove:      true,
		ScoreLimit:       70,
		Order:            1,
		IsGroupLab:       true,
		ContainerTimeout: 1,
	}
	qtest.CreateAssignment(t, db, assignment)

	results := &score.Results{
		BuildInfo: createBuildInfo(t),
		Scores:    createScores(),
	}

	runData := &ci.RunData{
		Course:     course,
		Assignment: assignment,
		Repo: &qf.Repository{
			RepoType: qf.Repository_GROUP,
			GroupID:  group.ID,
		},
		JobOwner: "test",
		CommitID: "deadbeef",
	}

	streamService := stream.NewStreamServices()
	var streams []*qtest.MockStream[qf.Submission]
	for _, user := range group.Users {
		stream := qtest.NewMockStream[qf.Submission](t)
		streamService.Submission.Add(stream, user.ID)
		streams = append(streams, stream)
	}

	// Add a stream for the admin user
	adminStream := qtest.NewMockStream[qf.Submission](t)
	streamService.Submission.Add(adminStream, admin.ID)

	var wg sync.WaitGroup
	for i := range streams {
		runStream(streams[i], &wg)
	}
	runStream(adminStream, &wg)

	owners, err := runData.GetOwners(db)
	if err != nil {
		t.Fatal(err)
	}

	// Check that submission is recorded correctly
	submission := recordResults(t, runData, db, results, nil, false)
	streamService.Submission.SendTo(submission, owners...)

	if submission.IsAllApproved() {
		t.Error("Submission must not be auto approved")
	}
	updatedSubmission := recordResults(t, runData, db, results, qtest.Timestamp(t, "2022-11-12T13:00:00"), false)
	streamService.Submission.SendTo(updatedSubmission, owners...)

	rebuiltSubmission := recordResults(t, runData, db, results, qtest.Timestamp(t, "2022-11-13T13:00:00"), true)
	streamService.Submission.SendTo(rebuiltSubmission, owners...)

	for i := range streams {
		streams[i].Close()
	}
	adminStream.Close()

	// Wait for all streams to be closed
	wg.Wait()

	// Admin user should have received 0 submissions
	if len(adminStream.Messages) != 0 {
		t.Errorf("Admin user should not have received any submissions, got %d", len(adminStream.Messages))
	}

	// We should have received three submissions for each stream
	numSubmissions := 0
	submissions := []*qf.Submission{submission, updatedSubmission, rebuiltSubmission}
	for _, stream := range streams {
		numSubmissions += len(stream.Messages)

		// Check that the messages are correct
		for i, submission := range submissions {
			qtest.Diff(t, "Incorrect submission", stream.Messages[i], submission, protocmp.Transform())
		}
	}
	if numSubmissions != 9 {
		t.Errorf("Expected 9 messages, got %d", numSubmissions)
	}
}

func runStream(stream *qtest.MockStream[qf.Submission], wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = stream.Run()
	}()
}

func recordResults(t *testing.T, runData *ci.RunData, db database.Database, results *score.Results, date *timestamppb.Timestamp, rebuild bool) *qf.Submission {
	if date != nil {
		results.BuildInfo.BuildDate = date
		results.BuildInfo.SubmissionDate = date
	}
	runData.Rebuild = rebuild
	submission, err := runData.RecordResults(qtest.Logger(t), db, results)
	if err != nil {
		t.Fatal(err)
	}
	return submission
}

func createBuildInfo(t *testing.T) *score.BuildInfo {
	return &score.BuildInfo{
		SubmissionDate: qtest.Timestamp(t, "2022-11-10T13:00:00"),
		BuildDate:      qtest.Timestamp(t, "2022-11-10T13:00:00"),
		BuildLog:       "Testing",
		ExecTime:       33333,
	}
}

func createScores() []*score.Score {
	return []*score.Score{
		{
			Secret:   "secret",
			TestName: "Test",
			Score:    10,
			MaxScore: 15,
			Weight:   1,
		},
	}
}
