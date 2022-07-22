package ci_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"google.golang.org/protobuf/testing/protocmp"
)

// To run this test, please see instructions in the developer guide (dev.md).

// This test uses a test course for experimenting with run.sh behavior.
// The tests below will run locally on the test machine, not on the QuickFeed machine.

func loadRunScript(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("testdata/run.sh")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func testRunData(t *testing.T, runScriptContent string) *ci.RunData {
	qfTestOrg := scm.GetTestOrganization(t)
	// Only used to fetch the user's GitHub login (user name)
	s := scm.GetTestSCM(t)
	userName, err := s.GetUserName(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	repo := qf.RepoURL{ProviderURL: "github.com", Organization: qfTestOrg}
	courseID := uint64(1)
	qf.SetAccessToken(courseID, scm.GetAccessToken(t))
	runData := &ci.RunData{
		Course: &qf.Course{
			ID:               courseID,
			Code:             "QF101",
			Provider:         "github",
			OrganizationPath: qfTestOrg,
		},
		Assignment: &qf.Assignment{
			Name:             "lab1",
			RunScriptContent: runScriptContent,
			ContainerTimeout: 1, // minutes
		},
		Repo: &qf.Repository{
			HTMLURL:  repo.StudentRepoURL(userName),
			RepoType: qf.Repository_USER,
		},
		JobOwner: "muggles",
		CommitID: "deadbeef",
	}
	return runData
}

func TestRunTests(t *testing.T) {
	runScriptContent := loadRunScript(t)
	runData := testRunData(t, runScriptContent)

	runner, closeFn := dockerClient(t)
	defer closeFn()
	ctx, cancel := runData.Assignment.WithTimeout(2 * time.Minute)
	defer cancel()
	results, err := runData.RunTests(ctx, qtest.Logger(t), runner)
	if err != nil {
		t.Fatal(err)
	}
	// We don't actually test anything here since we don't know how many assignments are in QF_TEST_ORG
	t.Logf("%+v", results.BuildInfo.BuildLog)
	results.BuildInfo.BuildLog = "removed"
	t.Logf("%+v\n", qlog.IndentJson(results))
}

func TestRunTestsTimeout(t *testing.T) {
	runScriptContent := loadRunScript(t)
	runData := testRunData(t, runScriptContent)

	runner, closeFn := dockerClient(t)
	defer closeFn()
	// Note that this timeout value is susceptible to variation
	ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancel()
	results, err := runData.RunTests(ctx, qtest.Logger(t), runner)
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
		Name:           "Test",
		Code:           "DAT320",
		OrganizationID: 1,
		SlipDays:       5,
	}
	admin := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateCourse(t, db, admin, course)

	assignment := &qf.Assignment{
		CourseID: course.ID,
		Name:     "lab1",
		RunScriptContent: `#image/quickfeed:go
printf "AssignmentName: {{ .AssignmentName }}\n"
printf "RandomSecret: {{ .RandomSecret }}\n"
`,
		Deadline:         "2022-11-11T13:00:00",
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
		BuildDate: "2022-11-10T13:00:00",
		BuildLog:  "Testing",
		ExecTime:  33333,
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
			UserID: 1,
		},
		JobOwner: "test",
		CommitID: "deadbeef",
	}

	// Check that submission is recorded correctly
	submission, err := runData.RecordResults(qtest.Logger(t), db, results)
	if err != nil {
		t.Fatal(err)
	}
	if submission.Status == qf.Submission_APPROVED {
		t.Error("Submission must not be auto approved")
	}
	if diff := cmp.Diff(testScores, submission.Scores, protocmp.Transform(), protocmp.IgnoreFields(&score.Score{}, "Secret")); diff != "" {
		t.Errorf("Incorrect submission scores. Want: %+v, got %+v", testScores, submission.Scores)
	}
	if diff := cmp.Diff(buildInfo.BuildDate, submission.BuildInfo.BuildDate); diff != "" {
		t.Errorf("Incorrect build date. Want: %s, got %s", buildInfo.BuildDate, submission.BuildInfo.BuildDate)
	}

	// When updating submission after deadline: build info and slip days must be updated
	newBuildDate := "2022-11-12T13:00:00"
	results.BuildInfo.BuildDate = newBuildDate
	updatedSubmission, err := runData.RecordResults(qtest.Logger(t), db, results)
	if err != nil {
		t.Fatal(err)
	}
	enrollment, err := db.GetEnrollmentByCourseAndUser(course.ID, admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if enrollment.RemainingSlipDays(course) == int32(course.SlipDays) || len(enrollment.UsedSlipDays) < 1 {
		t.Error("Student must have reduced slip days")
	}
	if updatedSubmission.BuildInfo.BuildDate != newBuildDate {
		t.Errorf("Incorrect build date: want %s, got %s", newBuildDate, updatedSubmission.BuildInfo.BuildDate)
	}

	// When rebuilding after deadline: delivery date and slip days must stay unchanged
	runData.Rebuild = true
	results.BuildInfo.BuildDate = "2022-11-13T13:00:00"
	slipDaysBeforeUpdate := enrollment.RemainingSlipDays(course)
	submission, err = runData.RecordResults(qtest.Logger(t), db, results)
	if err != nil {
		t.Fatal(err)
	}
	if submission.BuildInfo.BuildDate != newBuildDate {
		t.Errorf("Incorrect build date: want %s, got %s", newBuildDate, submission.BuildInfo.BuildDate)
	}
	updatedEnrollment, err := db.GetEnrollmentByCourseAndUser(course.ID, admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updatedEnrollment.RemainingSlipDays(course) != slipDaysBeforeUpdate {
		t.Errorf("Incorrect number of slip days: expected %d, got %d", slipDaysBeforeUpdate, updatedEnrollment.RemainingSlipDays(course))
	}
}

func TestRecordResultsForManualReview(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &qf.Course{
		Name:           "Test",
		OrganizationID: 1,
		SlipDays:       5,
	}
	admin := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateCourse(t, db, admin, course)

	assignment := &qf.Assignment{
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

	initialSubmission := &qf.Submission{
		AssignmentID: assignment.ID,
		UserID:       admin.ID,
		Score:        80,
		Status:       qf.Submission_APPROVED,
		Released:     true,
	}
	if err := db.CreateSubmission(initialSubmission); err != nil {
		t.Fatal(err)
	}

	runData := &ci.RunData{
		Course:     course,
		Assignment: assignment,
		Repo: &qf.Repository{
			UserID: 1,
		},
		JobOwner: "test",
	}

	submission, err := runData.RecordResults(qtest.Logger(t), db, nil)
	if err != nil {
		t.Fatal(err)
	}

	// make sure all fields were saved correctly in the database
	query := &qf.Submission{
		AssignmentID: assignment.ID,
		UserID:       admin.ID,
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
