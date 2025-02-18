package ci_test

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/internal/rand"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/stream"
	"google.golang.org/protobuf/testing/protocmp"
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

func setupRunData(t *testing.T, runner ci.Runner) *ci.RunData {
	dockerfile := loadDockerfile(t)
	qfTestOrg := scm.GetTestOrganization(t)
	// Only used to fetch the user's GitHub login (username)
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

	runData := setupRunData(t, runner)
	ctx, cancel := runData.Assignment.WithTimeout(2 * time.Minute)
	defer cancel()

	scmClient, _ := scm.GetTestSCM(t)
	results, err := runData.RunTests(ctx, qtest.Logger(t), scmClient, runner)
	if err != nil {
		t.Fatal(err)
	}
	// We don't actually test anything here since we don't know how many assignments are in QF_TEST_ORG
	t.Logf("%+v", results.GetBuildInfo().GetBuildLog())
	results.BuildInfo.BuildLog = "removed"
	t.Logf("%+v\n", qlog.IndentJson(results))
}

func TestRunTestsTimeout(t *testing.T) {
	runner, closeFn := dockerClient(t)
	defer closeFn()

	runData := setupRunData(t, runner)
	ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancel()

	scmClient, _ := scm.GetTestSCM(t)
	results, err := runData.RunTests(ctx, qtest.Logger(t), scmClient, runner)
	if err != nil {
		t.Fatal(err)
	}
	const wantOut = `Container timeout. Please check for infinite loops or other slowness.`
	if results.BuildInfo != nil && !strings.HasPrefix(results.BuildInfo.GetBuildLog(), wantOut) {
		t.Errorf("RunTests(1s timeout) = '%s', got '%s'", wantOut, results.BuildInfo.GetBuildLog())
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
		CourseID:         course.GetID(),
		Name:             "lab1",
		Deadline:         qtest.Timestamp(t, "2022-11-11T13:00:00"),
		AutoApprove:      true,
		ScoreLimit:       70,
		Order:            1,
		IsGroupLab:       false,
		ContainerTimeout: 1,
	}
	if err := db.CreateAssignment(assignment); err != nil {
		t.Fatal(err)
	}

	buildInfo := &score.BuildInfo{
		SubmissionDate: qtest.Timestamp(t, "2022-11-10T13:00:00"),
		BuildDate:      qtest.Timestamp(t, "2022-11-10T13:00:00"),
		BuildLog:       "Testing",
		ExecTime:       33333,
	}
	testScores := []*score.Score{
		{
			Secret:   "secret",
			TestName: "Test",
			Score:    10,
			MaxScore: 15,
			Weight:   1,
		},
	}
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
	submission, err := runData.RecordResults(qtest.Logger(t), db, results)
	if err != nil {
		t.Fatal(err)
	}
	if submission.IsApproved(runData.Repo.GetUserID()) {
		t.Error("Submission must not be auto approved")
	}
	if diff := cmp.Diff(testScores, submission.GetScores(), protocmp.Transform(), protocmp.IgnoreFields(&score.Score{}, "Secret")); diff != "" {
		t.Errorf("submission score mismatch: (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(buildInfo.GetBuildDate(), submission.GetBuildInfo().GetBuildDate(), protocmp.Transform()); diff != "" {
		t.Errorf("build date mismatch: (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(buildInfo.GetSubmissionDate(), submission.GetBuildInfo().GetSubmissionDate(), protocmp.Transform()); diff != "" {
		t.Errorf("submission date mismatch: (-want +got):\n%s", diff)
	}

	// When updating submission after deadline: build info (submission and build dates) and slip days must be updated
	newSubmissionDate := qtest.Timestamp(t, "2022-11-12T13:00:00")
	results.BuildInfo.BuildDate = newSubmissionDate
	results.BuildInfo.SubmissionDate = newSubmissionDate
	updatedSubmission, err := runData.RecordResults(qtest.Logger(t), db, results)
	if err != nil {
		t.Fatal(err)
	}
	enrollment, err := db.GetEnrollmentByCourseAndUser(course.GetID(), admin.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if enrollment.RemainingSlipDays(course) == int32(course.GetSlipDays()) || len(enrollment.GetUsedSlipDays()) < 1 {
		t.Error("Student must have reduced slip days")
	}
	if diff := cmp.Diff(newSubmissionDate, updatedSubmission.GetBuildInfo().GetBuildDate(), protocmp.Transform()); diff != "" {
		t.Errorf("build date mismatch: (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(newSubmissionDate, updatedSubmission.GetBuildInfo().GetSubmissionDate(), protocmp.Transform()); diff != "" {
		t.Errorf("submission date mismatch: (-want +got):\n%s", diff)
	}

	// When rebuilding after deadline: delivery date and slip days must stay unchanged, build date must be updated
	runData.Rebuild = true
	wantSubmissionDate := newSubmissionDate
	newDate := qtest.Timestamp(t, "2022-11-13T15:00:00")
	results.BuildInfo.BuildDate = newDate
	results.BuildInfo.SubmissionDate = newDate
	slipDaysBeforeUpdate := enrollment.RemainingSlipDays(course)
	rebuiltSubmission, err := runData.RecordResults(qtest.Logger(t), db, results)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(newDate, rebuiltSubmission.GetBuildInfo().GetBuildDate(), protocmp.Transform()); diff != "" {
		t.Errorf("build date mismatch: (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(wantSubmissionDate, rebuiltSubmission.GetBuildInfo().GetSubmissionDate(), protocmp.Transform()); diff != "" {
		t.Errorf("submission date mismatch: (-want +got):\n%s", diff)
	}
	updatedEnrollment, err := db.GetEnrollmentByCourseAndUser(course.GetID(), admin.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(slipDaysBeforeUpdate, updatedEnrollment.RemainingSlipDays(course)); diff != "" {
		t.Errorf("slip days mismatch: (-want +got):\n%s", diff)
	}
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
		CourseID:   course.GetID(),
		Name:       "assignment-1",
		Deadline:   qtest.Timestamp(t, "2022-11-11T13:00:00"),
		IsGroupLab: false,
		Reviewers:  1,
	}
	if err := db.CreateAssignment(assignment); err != nil {
		t.Fatal(err)
	}

	initialSubmission := &qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       admin.GetID(),
		Score:        80,
		Grades:       []*qf.Grade{{UserID: admin.GetID(), Status: qf.Submission_APPROVED}},
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
			UserID:   admin.GetID(),
		},
		JobOwner: "test",
	}

	submission, err := runData.RecordResults(qtest.Logger(t), db, nil)
	if err != nil {
		t.Fatal(err)
	}

	// make sure all fields were saved correctly in the database
	query := &qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       admin.GetID(),
	}
	updatedSubmission, err := db.GetSubmission(query)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(updatedSubmission, submission, protocmp.Transform()); diff != "" {
		t.Errorf("Incorrect submission fields in the database. Want: %+v, got %+v", initialSubmission, updatedSubmission)
	}

	// submission must stay approved, released, with score = 80
	if diff := cmp.Diff(initialSubmission, updatedSubmission, protocmp.Transform(), protocmp.IgnoreFields(&qf.Submission{}, "BuildInfo", "Scores")); diff != "" {
		t.Errorf("Incorrect submission after update. Want: %+v, got %+v", initialSubmission, updatedSubmission)
	}
}

func TestStreamRecordResults(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	streamService := stream.NewStreamServices()

	course := &qf.Course{
		Name:              "Test",
		Code:              "DAT320",
		ScmOrganizationID: 1,
		SlipDays:          5,
	}
	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, course)

	groupMember1 := qtest.CreateFakeUser(t, db)
	groupMember2 := qtest.CreateFakeUser(t, db)
	groupMember3 := qtest.CreateFakeUser(t, db)
	for _, user := range []*qf.User{groupMember1, groupMember2, groupMember3} {
		qtest.EnrollStudent(t, db, user, course)
	}
	group := &qf.Group{
		CourseID: course.GetID(),
		Name:     "group-1",
		Users: []*qf.User{
			groupMember1,
			groupMember2,
			groupMember3,
		},
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}

	assignment := &qf.Assignment{
		CourseID:         course.GetID(),
		Name:             "lab1",
		Deadline:         qtest.Timestamp(t, "2022-11-11T13:00:00"),
		AutoApprove:      true,
		ScoreLimit:       70,
		Order:            1,
		IsGroupLab:       true,
		ContainerTimeout: 1,
	}
	if err := db.CreateAssignment(assignment); err != nil {
		t.Fatal(err)
	}

	buildInfo := &score.BuildInfo{
		BuildDate:      qtest.Timestamp(t, "2022-11-10T13:00:00"),
		SubmissionDate: qtest.Timestamp(t, "2022-11-10T13:00:00"),
		BuildLog:       "Testing",
		ExecTime:       33333,
	}
	testScores := []*score.Score{
		{
			Secret:   "secret",
			TestName: "Test",
			Score:    10,
			MaxScore: 15,
			Weight:   1,
		},
	}

	results := &score.Results{
		BuildInfo: buildInfo,
		Scores:    testScores,
	}
	runData := &ci.RunData{
		Course:     course,
		Assignment: assignment,
		Repo: &qf.Repository{
			RepoType: qf.Repository_GROUP,
			GroupID:  group.GetID(),
		},
		JobOwner: "test",
		CommitID: "deadbeef",
	}

	var streams []*qtest.MockStream[qf.Submission]
	for _, user := range group.GetUsers() {
		mockStream := qtest.NewMockStream[qf.Submission](t)
		streamService.Submission.Add(mockStream, user.GetID())
		streams = append(streams, mockStream)
	}

	// Add a stream for the admin user
	adminStream := qtest.NewMockStream[qf.Submission](t)
	streamService.Submission.Add(adminStream, admin.GetID())

	var wg sync.WaitGroup
	for i := range streams {
		runStream(streams[i], &wg)
	}
	runStream(adminStream, &wg)

	// Check that submission is recorded correctly
	submission, err := runData.RecordResults(qtest.Logger(t), db, results)
	if err != nil {
		t.Fatal(err)
	}

	owners, err := runData.GetOwners(db)
	if err != nil {
		t.Fatal(err)
	}
	streamService.Submission.SendTo(submission, owners...)
	if submission.IsAllApproved() {
		t.Error("Submission must not be auto approved")
	}

	newBuildDate := qtest.Timestamp(t, "2022-11-12T13:00:00")
	results.BuildInfo.BuildDate = newBuildDate
	updatedSubmission, err := runData.RecordResults(qtest.Logger(t), db, results)
	if err != nil {
		t.Fatal(err)
	}
	streamService.Submission.SendTo(updatedSubmission, owners...)

	runData.Rebuild = true
	results.BuildInfo.BuildDate = qtest.Timestamp(t, "2022-11-13T13:00:00")
	rebuiltSubmission, err := runData.RecordResults(qtest.Logger(t), db, results)
	if err != nil {
		t.Fatal(err)
	}
	streamService.Submission.SendTo(rebuiltSubmission, owners...)

	for _, mockStream := range streams {
		mockStream.Close()
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
	for _, mockStream := range streams {
		numSubmissions += len(mockStream.Messages)
	}
	if numSubmissions != 9 {
		t.Errorf("Expected 9 messages, got %d", numSubmissions)
	}

	// Check that the messages are correct
	submissions := []*qf.Submission{submission, updatedSubmission, rebuiltSubmission}
	for _, mockStream := range streams {
		for i, submission := range submissions {
			if diff := cmp.Diff(mockStream.Messages[i], submission, protocmp.Transform()); diff != "" {
				t.Errorf("Incorrect submission. Want: %+v, got %+v", submission, mockStream.Messages[i])
			}
		}
	}
}

func runStream(stream *qtest.MockStream[qf.Submission], wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = stream.Run()
	}()
}
