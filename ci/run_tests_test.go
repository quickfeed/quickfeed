package ci

import (
	"context"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/kit/score"
	"github.com/autograde/quickfeed/log"
	"github.com/autograde/quickfeed/scm"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"google.golang.org/protobuf/testing/protocmp"
)

// To run this test, please see instructions in the developer guide (dev.md).

// This test uses a test course for experimenting with go.sh behavior.
// The test below will run locally on the test machine, not on the QuickFeed machine.

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

	randomString := qtest.RandomString(t)

	repo := pb.RepoURL{ProviderURL: "github.com", Organization: qfTestOrg}
	info := &AssignmentInfo{
		AssignmentName:     "lab1",
		Script:             "go.sh",
		CreatorAccessToken: accessToken,
		GetURL:             repo.StudentRepoURL(userName),
		TestURL:            repo.TestsRepoURL(),
		RandomSecret:       randomString,
	}
	runData := &RunData{
		Course: &pb.Course{Code: "DAT320"},
		Assignment: &pb.Assignment{
			Name:             info.AssignmentName,
			ContainerTimeout: 1,
		},
		Repo:     &pb.Repository{},
		JobOwner: "muggles",
	}

	runner, err := NewDockerCI(log.Zap(true))
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	ed, err := runData.runTests(runner, info)
	if err != nil {
		t.Fatal(err)
	}
	// We don't actually test anything here since we don't know how many assignments are in QF_TEST_ORG
	t.Logf("\n%s\nExecTime: %v\nSecret: %v\n", ed.out, ed.execTime, info.RandomSecret)
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
		CourseID:         course.ID,
		Name:             "lab1",
		ScriptFile:       "go.sh",
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
	runData := &RunData{
		Course:     course,
		Assignment: assignment,
		Repo: &pb.Repository{
			UserID: 1,
		},
		JobOwner: "test",
	}

	runData.recordResults(zap.NewNop().Sugar(), db, results)
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
	runData.recordResults(zap.NewNop().Sugar(), db, results)

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
	runData.recordResults(zap.NewNop().Sugar(), db, results)
	updatedEnrollment, err := db.GetEnrollmentByCourseAndUser(course.ID, admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updatedEnrollment.RemainingSlipDays(course) != slipDaysBeforeUpdate {
		t.Errorf("Incorrect number of slip days: expected %d, got %d", slipDaysBeforeUpdate, updatedEnrollment.RemainingSlipDays(course))
	}
}
