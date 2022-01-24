package ci_test

import (
	"context"
	"os"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/kit/score"
	"github.com/autograde/quickfeed/scm"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"google.golang.org/protobuf/testing/protocmp"
)

// To run this test, please see instructions in the developer guide (dev.md).

// This test uses a test course for experimenting with run.sh behavior.
// The test below will run locally on the test machine, not on the QuickFeed machine.

func loadRunScript(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("testdata/run.sh")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func testRunData(qfTestOrg string, userName, accessToken, scriptTemplate string) *ci.RunData {
	repo := pb.RepoURL{ProviderURL: "github.com", Organization: qfTestOrg}
	courseID := uint64(1)
	pb.SetAccessToken(courseID, accessToken)
	runData := &ci.RunData{
		Course: &pb.Course{
			ID:   courseID,
			Code: "DAT320",
		},
		Assignment: &pb.Assignment{
			Name:             "lab1",
			ScriptFile:       scriptTemplate,
			ContainerTimeout: 1,
		},
		Repo: &pb.Repository{
			HTMLURL:  repo.StudentRepoURL(userName),
			RepoType: pb.Repository_USER,
		},
		JobOwner: "muggles",
	}
	return runData
}

func TestRunTests(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)

	// Only used to fetch the user's GitHub login (user name)
	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}
	userName, err := s.GetUserName(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	scriptTemplate := loadRunScript(t)
	runData := testRunData(qfTestOrg, userName, accessToken, scriptTemplate)

	runner, closeFn := dockerClient(t)
	defer closeFn()
	results, err := runData.RunTests(zap.NewNop().Sugar(), runner)
	if err != nil {
		t.Fatal(err)
	}
	// We don't actually test anything here since we don't know how many assignments are in QF_TEST_ORG
	t.Logf("%+v\n", results)
}

func TestRunTestsTimeout(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)

	// Only used to fetch the user's GitHub login (user name)
	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}
	userName, err := s.GetUserName(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// TODO(meling) fix this so that it actually times out
	scriptTemplate := loadRunScript(t)
	runData := testRunData(qfTestOrg, userName, accessToken, scriptTemplate)

	runner, closeFn := dockerClient(t)
	defer closeFn()
	results, err := runData.RunTests(zap.NewNop().Sugar(), runner)
	if err != nil {
		t.Fatal(err)
	}
	// We don't actually test anything here since we don't know how many assignments are in QF_TEST_ORG
	t.Logf("%+v\n", results)
}

func TestRecordResults(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &pb.Course{
		Name:           "Test",
		OrganizationID: 1,
		SlipDays:       5,
	}
	admin := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateCourse(t, db, admin, course)

	assignment := &pb.Assignment{
		CourseID: course.ID,
		Name:     "lab1",
		ScriptFile: `#image/quickfeed:go
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
		Repo: &pb.Repository{
			UserID: 1,
		},
		JobOwner: "test",
	}

	// TODO Get submission here from record results
	runData.RecordResults(zap.NewNop().Sugar(), db, results)
	submission, err := db.GetSubmission(&pb.Submission{AssignmentID: assignment.ID, UserID: admin.ID})
	if err != nil {
		t.Fatal(err)
	}
	if submission.Status == pb.Submission_APPROVED {
		t.Error("Submission must not be auto approved")
	}
	if diff := cmp.Diff(testScores, submission.Scores, protocmp.Transform(), protocmp.IgnoreFields(&score.Score{}, "Secret")); diff != "" {
		t.Errorf("Incorrect submission scores. Want: %+v, got %+v", testScores, submission.Scores)
	}
	if diff := cmp.Diff(buildInfo.BuildDate, submission.BuildInfo.BuildDate); diff != "" {
		t.Errorf("Incorrect build date. Want: %s, got %s", buildInfo.BuildDate, submission.BuildInfo.BuildDate)
	}

	// Updating submission after deadline: build info and slip days must be updated
	newBuildDate := "2022-11-12T13:00:00"
	results.BuildInfo.BuildDate = newBuildDate
	runData.RecordResults(zap.NewNop().Sugar(), db, results)

	enrollment, err := db.GetEnrollmentByCourseAndUser(course.ID, admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if enrollment.RemainingSlipDays(course) == int32(course.SlipDays) || len(enrollment.UsedSlipDays) < 1 {
		t.Error("Student must have reduced slip days")
	}
	updatedSubmission, err := db.GetSubmission(&pb.Submission{AssignmentID: assignment.ID, UserID: admin.ID})
	if err != nil {
		t.Fatal(err)
	}
	if updatedSubmission.BuildInfo.BuildDate != newBuildDate {
		t.Errorf("Incorrect build date: want %s, got %s", newBuildDate, updatedSubmission.BuildInfo.BuildDate)
	}

	// Rebuilding after deadline: delivery date and slip days must stay unchanged
	runData.Rebuild = true
	results.BuildInfo.BuildDate = "2022-11-13T13:00:00"
	slipDaysBeforeUpdate := enrollment.RemainingSlipDays(course)
	runData.RecordResults(zap.NewNop().Sugar(), db, results)
	updatedEnrollment, err := db.GetEnrollmentByCourseAndUser(course.ID, admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updatedEnrollment.RemainingSlipDays(course) != slipDaysBeforeUpdate {
		t.Errorf("Incorrect number of slip days: expected %d, got %d", slipDaysBeforeUpdate, updatedEnrollment.RemainingSlipDays(course))
	}
}
