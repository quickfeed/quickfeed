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
		Name:  course.JobName(),
		Image: course.DockerImage(),
		BuildContext: map[string]string{
			ci.Dockerfile: course.GetDockerfile(),
		},
		Commands: []string{`echo -n "Hello from Dockerfile"`},
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
	ctx, cancel := runData.Assignment.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	scmClient, _ := scm.GetTestSCM(t)
	ctx = qlog.NewContext(ctx, qtest.Logger(t))
	results, err := runData.RunTests(ctx, scmClient, runner)
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
	ctx = qlog.NewContext(ctx, qtest.Logger(t))
	results, err := runData.RunTests(ctx, scmClient, runner)
	if err != nil {
		t.Fatal(err)
	}
	if gotStatus := results.GetBuildInfo().GetStatus(); gotStatus != score.RunStatus_TIMEOUT {
		t.Errorf("RunTests(2s timeout) status = %s, want TIMEOUT", gotStatus)
	}
	const wantOut = `The test run timed out.`
	gotOut := results.GetBuildInfo().GetBuildLog()
	if !strings.HasPrefix(gotOut, wantOut) {
		t.Errorf("RunTests(2s timeout) = '%s', want '%s'", gotOut, wantOut)
	}
	if !strings.Contains(gotOut, "Container timeout.") {
		t.Errorf("RunTests(2s timeout) = '%s', want it to contain 'Container timeout.'", gotOut)
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
	qtest.Diff(t, "submission score mismatch", testScores, submission.GetScores(), protocmp.Transform(), protocmp.IgnoreFields(&score.Score{}, "Secret"))
	qtest.Diff(t, "build info mismatch", buildInfo, submission.GetBuildInfo(), protocmp.Transform())

	// When updating submission after deadline: build info (submission and build dates) and slip days must be updated
	newSubmissionDate := qtest.Timestamp(t, "2022-11-12T13:00:00")
	updatedSubmission := recordResults(t, runData, db, results, newSubmissionDate, false)
	enrollment := qtest.GetEnrollment(t, db, course.GetID(), admin.GetID())
	if enrollment.RemainingSlipDays(course) == int32(course.GetSlipDays()) || len(enrollment.GetUsedSlipDays()) < 1 {
		t.Error("Student must have reduced slip days")
	}
	qtest.Diff(t, "build info mismatch", results.GetBuildInfo(), updatedSubmission.GetBuildInfo(), protocmp.Transform())

	// When rebuilding after deadline: delivery date and slip days must stay unchanged, build date must be updated
	wantSubmissionDate := newSubmissionDate
	newDate := qtest.Timestamp(t, "2022-11-13T15:00:00")
	slipDaysBeforeUpdate := enrollment.RemainingSlipDays(course)
	rebuiltSubmission := recordResults(t, runData, db, results, newDate, true)

	qtest.Diff(t, "build date mismatch", newDate, rebuiltSubmission.GetBuildInfo().GetBuildDate(), protocmp.Transform())
	qtest.Diff(t, "submission date mismatch", wantSubmissionDate, rebuiltSubmission.GetBuildInfo().GetSubmissionDate(), protocmp.Transform())

	updatedEnrollment := qtest.GetEnrollment(t, db, course.GetID(), admin.GetID())
	qtest.Diff(t, "slip days mismatch", slipDaysBeforeUpdate, updatedEnrollment.RemainingSlipDays(course))

	// A compilation failure is a failed run with trustworthy zero scores.
	zeroScores := createScores()
	zeroScores[0].Score = 0
	compilationResults := &score.Results{
		BuildInfo: &score.BuildInfo{Status: score.RunStatus_BUILD_FAILURE},
		Scores:    zeroScores,
	}
	compilationDate := qtest.Timestamp(t, "2022-11-14T13:00:00")
	compilationSubmission := recordResults(t, runData, db, compilationResults, compilationDate, false)
	if !compilationSubmission.Failed() {
		t.Error("compilation failure must retain its failed run status")
	}
	if got := compilationSubmission.GetScore(); got != 0 {
		t.Errorf("compilation failure score = %d, want 0", got)
	}
	qtest.Diff(t, "compilation failure scores mismatch", zeroScores, compilationSubmission.GetScores(),
		protocmp.Transform(), protocmp.IgnoreFields(&score.Score{}, "Secret"))
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
	qtest.CreateAssignment(t, db, assignment)

	initialSubmission := &qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       admin.GetID(),
		Score:        80,
		Grades:       []*qf.Grade{{UserID: admin.GetID(), Status: qf.Submission_APPROVED}},
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

	submission := recordResults(t, runData, db, nil, nil, false)

	// make sure all fields were saved correctly in the database
	query := &qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       admin.GetID(),
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
	group := qtest.CreateFakeGroup(t, db, course, 3)

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
			GroupID:  group.GetID(),
		},
		JobOwner: "test",
		CommitID: "deadbeef",
	}

	streamService := stream.NewStreamServices()
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
	wg.Go(func() {
		_ = stream.Run()
	})
}

func TestRecordResultsGroupSlipDays(t *testing.T) {
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
	group := qtest.CreateFakeGroup(t, db, course, 3)

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
	qtest.CreateAssignment(t, db, assignment)
	buildInfo := createBuildInfo(t)
	testScores := createScores()
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

	// Check that submission is recorded correctly
	submission := recordResults(t, runData, db, results, nil, false)
	if submission.IsAllApproved() {
		t.Error("Submission must not be auto approved")
	}
	qtest.Diff(t, "submission score mismatch", testScores, submission.GetScores(), protocmp.Transform(), protocmp.IgnoreFields(&score.Score{}, "Secret"))
	qtest.Diff(t, "build info mismatch", buildInfo, submission.GetBuildInfo(), protocmp.Transform())

	// Verify group slip days not used yet (submission before deadline)
	fetchedGroup, err := db.GetGroup(group.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if fetchedGroup.RemainingSlipDays(course) != int32(course.GetSlipDays()) {
		t.Errorf("Group must have unchanged slip days before deadline, got %d, want %d",
			fetchedGroup.RemainingSlipDays(course), course.GetSlipDays())
	}

	// When updating submission after deadline: build info (submission and build dates) and group slip days must be updated
	newSubmissionDate := qtest.Timestamp(t, "2022-11-12T13:00:00")
	updatedSubmission := recordResults(t, runData, db, results, newSubmissionDate, false)

	// Verify group slip days were reduced
	updatedGroup, err := db.GetGroup(group.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if updatedGroup.RemainingSlipDays(course) == int32(course.GetSlipDays()) || len(updatedGroup.GetUsedSlipDays()) < 1 {
		t.Error("Group must have reduced slip days after late submission")
	}
	qtest.Diff(t, "build info mismatch", results.GetBuildInfo(), updatedSubmission.GetBuildInfo(), protocmp.Transform())

	// Verify individual student enrollments were NOT affected
	for _, user := range group.GetUsers() {
		enrollment, err := db.GetEnrollmentByCourseAndUser(course.GetID(), user.GetID())
		if err != nil {
			t.Fatal(err)
		}
		if enrollment.RemainingSlipDays(course) != int32(course.GetSlipDays()) {
			t.Errorf("Individual student enrollment should not have slip days reduced for group submission, got %d, want %d",
				enrollment.RemainingSlipDays(course), course.GetSlipDays())
		}
	}

	// When rebuilding after deadline: delivery date and slip days must stay unchanged, build date must be updated
	wantSubmissionDate := newSubmissionDate
	newDate := qtest.Timestamp(t, "2022-11-13T15:00:00")
	slipDaysBeforeUpdate := updatedGroup.RemainingSlipDays(course)
	rebuiltSubmission := recordResults(t, runData, db, results, newDate, true)

	qtest.Diff(t, "build date mismatch", newDate, rebuiltSubmission.GetBuildInfo().GetBuildDate(), protocmp.Transform())
	qtest.Diff(t, "submission date mismatch", wantSubmissionDate, rebuiltSubmission.GetBuildInfo().GetSubmissionDate(), protocmp.Transform())

	rebuiltGroup, err := db.GetGroup(group.GetID())
	if err != nil {
		t.Fatal(err)
	}
	qtest.Diff(t, "slip days mismatch", slipDaysBeforeUpdate, rebuiltGroup.RemainingSlipDays(course))
}

// TestRecordResultsGroupSubmitsNonGroupLab pins the design decision that a
// group push to an assignment with IsGroupLab=false is a no-op for slip days:
// the submission isn't graded, so neither the group's pool nor any individual
// enrollment counter is touched.
func TestRecordResultsGroupSubmitsNonGroupLab(t *testing.T) {
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
	group := qtest.CreateFakeGroup(t, db, course, 2)

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
			GroupID:  group.GetID(),
		},
		JobOwner: "test",
		CommitID: "deadbeef",
	}

	// Submit well past the deadline so any active charging path would
	// have produced a deduction.
	lateDate := qtest.Timestamp(t, "2022-11-15T13:00:00")
	_ = recordResults(t, runData, db, results, lateDate, false)

	// Group pool must not be touched: this isn't a group lab.
	fetchedGroup, err := db.GetGroup(group.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if fetchedGroup.RemainingSlipDays(course) != int32(course.GetSlipDays()) {
		t.Errorf("group slip days should be unchanged, got %d, want %d",
			fetchedGroup.RemainingSlipDays(course), course.GetSlipDays())
	}
	if len(fetchedGroup.GetUsedSlipDays()) != 0 {
		t.Errorf("group should have no UsedSlipDays rows, got %d", len(fetchedGroup.GetUsedSlipDays()))
	}

	// Sweep every enrollment in the course, not just group members:
	// any accidental charge against an arbitrary enrollment must fail the test.
	enrollments, err := db.GetEnrollmentsByCourse(course.GetID(), qf.Enrollment_STUDENT, qf.Enrollment_TEACHER)
	if err != nil {
		t.Fatal(err)
	}
	if len(enrollments) == 0 {
		t.Fatal("expected at least one enrollment for the course")
	}
	for _, enrollment := range enrollments {
		if enrollment.RemainingSlipDays(course) != int32(course.GetSlipDays()) {
			t.Errorf("enrollment %d (user %d) slip days should be unchanged, got %d, want %d",
				enrollment.GetID(), enrollment.GetUserID(),
				enrollment.RemainingSlipDays(course), course.GetSlipDays())
		}
		if len(enrollment.GetUsedSlipDays()) != 0 {
			t.Errorf("enrollment %d (user %d) should have no UsedSlipDays rows, got %d",
				enrollment.GetID(), enrollment.GetUserID(), len(enrollment.GetUsedSlipDays()))
		}
	}
}

func recordResults(t *testing.T, runData *ci.RunData, db database.Database, results *score.Results, date *timestamppb.Timestamp, rebuild bool) *qf.Submission {
	if date != nil {
		results.BuildInfo.BuildDate = date
		results.BuildInfo.SubmissionDate = date
	}
	runData.Rebuild = rebuild
	ctx := qlog.NewContext(context.Background(), qtest.Logger(t))
	submission, err := runData.RecordResults(ctx, db, results)
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

// TestRecordResultsFailedRun checks that a failed run does not overwrite the
// previous submission's result: the score, scores, and grades are kept, while
// the failed attempt's date participates in slip-day accounting (issue #1593).
func TestRecordResultsFailedRun(t *testing.T) {
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
	qtest.CreateAssignment(t, db, assignment)
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

	// Record a good submission first.
	goodResults := &score.Results{
		BuildInfo: createBuildInfo(t),
		Scores:    createScores(),
	}
	goodSubmission := recordResults(t, runData, db, goodResults, nil, false)

	// Record a failed run after the deadline with a new commit.
	runData.CommitID = "cafebabe"
	failedResults := &score.Results{
		BuildInfo: &score.BuildInfo{
			SubmissionDate: qtest.Timestamp(t, "2022-11-12T13:00:00"),
			BuildDate:      qtest.Timestamp(t, "2022-11-12T13:00:00"),
			BuildLog:       "The test run produced no test results.",
			ExecTime:       1000,
			Status:         score.RunStatus_NO_SCORES,
		},
	}
	failedSubmission := recordResults(t, runData, db, failedResults, nil, false)

	if failedSubmission.GetScore() != goodSubmission.GetScore() {
		t.Errorf("failed run score = %d, want previous score %d", failedSubmission.GetScore(), goodSubmission.GetScore())
	}
	qtest.Diff(t, "failed run must keep previous scores", goodSubmission.GetScores(), failedSubmission.GetScores(),
		protocmp.Transform(), protocmp.IgnoreFields(&score.Score{}, "Secret"))
	qtest.Diff(t, "failed run must keep previous grades", goodSubmission.GetGrades(), failedSubmission.GetGrades(), protocmp.Transform())
	if failedSubmission.GetCommitHash() != "cafebabe" {
		t.Errorf("failed run commit hash = %s, want cafebabe", failedSubmission.GetCommitHash())
	}
	buildInfo := failedSubmission.GetBuildInfo()
	if buildInfo.GetStatus() != score.RunStatus_NO_SCORES {
		t.Errorf("failed run status = %s, want NO_SCORES", buildInfo.GetStatus())
	}
	qtest.Diff(t, "failed run must keep its submission date",
		failedResults.GetBuildInfo().GetSubmissionDate(), buildInfo.GetSubmissionDate(), protocmp.Transform())
	enrollment := qtest.GetEnrollment(t, db, course.GetID(), admin.GetID())
	if got, want := enrollment.RemainingSlipDays(course), int32(course.GetSlipDays()-1); got != want {
		t.Errorf("slip days after failed run = %d, want %d", got, want)
	}

	// A successful rebuild keeps the failed attempt's submission date and does
	// not account for the same push a second time.
	rebuildDate := qtest.Timestamp(t, "2022-11-13T13:00:00")
	rebuiltSubmission := recordResults(t, runData, db, &score.Results{
		BuildInfo: createBuildInfo(t),
		Scores:    createScores(),
	}, rebuildDate, true)
	qtest.Diff(t, "rebuild must keep failed attempt submission date",
		buildInfo.GetSubmissionDate(), rebuiltSubmission.GetBuildInfo().GetSubmissionDate(), protocmp.Transform())
	enrollment = qtest.GetEnrollment(t, db, course.GetID(), admin.GetID())
	if got, want := enrollment.RemainingSlipDays(course), int32(course.GetSlipDays()-1); got != want {
		t.Errorf("slip days after rebuild = %d, want %d", got, want)
	}

	// The submission fetched from the database must match what was recorded.
	gotSubmission, err := db.GetSubmission(&qf.Submission{AssignmentID: assignment.GetID(), UserID: 1})
	if err != nil {
		t.Fatal(err)
	}
	qtest.Diff(t, "stored scores mismatch", goodSubmission.GetScores(), gotSubmission.GetScores(),
		protocmp.Transform(), protocmp.IgnoreFields(&score.Score{}, "ID", "SubmissionID", "Secret"))
	if gotSubmission.Failed() {
		t.Errorf("stored rebuilt submission status = %s, want SUCCESS", gotSubmission.GetBuildInfo().GetStatus())
	}
}

// TestRecordResultsFailedFirstRun checks that a failed run without any previous
// submission records a zero-score submission carrying the failure status.
func TestRecordResultsFailedFirstRun(t *testing.T) {
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
		CourseID:   course.GetID(),
		Name:       "lab1",
		Deadline:   qtest.Timestamp(t, "2022-11-11T13:00:00"),
		Order:      1,
		ScoreLimit: 70,
	}
	qtest.CreateAssignment(t, db, assignment)
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
	failedResults := &score.Results{
		BuildInfo: &score.BuildInfo{
			SubmissionDate: qtest.Timestamp(t, "2022-11-12T13:00:00"),
			BuildDate:      qtest.Timestamp(t, "2022-11-12T13:00:00"),
			BuildLog:       "The test environment failed before your code could be tested.",
			ExecTime:       1000,
			Status:         score.RunStatus_BUILD_FAILURE,
		},
	}
	submission := recordResults(t, runData, db, failedResults, nil, false)
	if submission.GetScore() != 0 {
		t.Errorf("failed first run score = %d, want 0", submission.GetScore())
	}
	if submission.GetBuildInfo().GetStatus() != score.RunStatus_BUILD_FAILURE {
		t.Errorf("failed first run status = %s, want BUILD_FAILURE", submission.GetBuildInfo().GetStatus())
	}
	if len(submission.GetGrades()) == 0 {
		t.Error("failed first run must initialize grades")
	}
	enrollment := qtest.GetEnrollment(t, db, course.GetID(), admin.GetID())
	if got, want := enrollment.RemainingSlipDays(course), int32(course.GetSlipDays()-1); got != want {
		t.Errorf("slip days after failed first run = %d, want %d", got, want)
	}
}
