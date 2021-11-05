package web_test

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/scm"
	"github.com/autograde/quickfeed/web"
	"go.uber.org/zap"
)

func TestSimulatedRebuildWorkpoolWithErrCount(t *testing.T) {
	// this tests the workpool logic used in rebuildSubmissions
	for _, maxContainers := range []int{3, 5, 6, 8, 10} {
		for _, errRate := range []int{2, 3, 4, 5, 6} {
			for _, numSubs := range []int{10, 15, 20, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 150, 500} {
				t.Run(fmt.Sprintf("containers=%d/errRate=%d/submissions=%d", maxContainers, errRate, numSubs), func(t *testing.T) {
					submissions := make([]int, numSubs)
					for i := range submissions {
						submissions[i] = i
					}
					sem := make(chan struct{}, maxContainers)
					errCnt := int32(0)
					var wg sync.WaitGroup
					wg.Add(len(submissions))
					for _, submission := range submissions {
						submission := submission
						// the counting semaphore limits concurrency to maxContainers
						go func() {
							sem <- struct{}{} // acquire semaphore
							// here we are rebuilding submission
							if submission%errRate == 0 { // simulate error every errRate submission
								// count the error
								atomic.AddInt32(&errCnt, 1)
							}
							<-sem // release semaphore
							wg.Done()
						}()
					}
					// wait for all submissions to finish rebuilding
					wg.Wait()
					close(sem)

					expectedErrCnt := int32(math.Ceil(float64(len(submissions)) / float64(errRate)))
					if errCnt != expectedErrCnt {
						t.Errorf("errCnt != expectedErrCnt ==> %d != %d == (%d/%d)", errCnt, expectedErrCnt, len(submissions), errRate)
					}
				})
			}
		}
	}
}

func TestRebuildSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	teacher := qtest.CreateFakeUser(t, db, 1)
	err := db.UpdateUser(&pb.User{ID: teacher.ID, IsAdmin: true})
	if err != nil {
		t.Fatal(err)
	}
	var course pb.Course
	course.Provider = "fake"
	course.OrganizationID = 1
	if err := db.CreateCourse(teacher.ID, &course); err != nil {
		t.Fatal(err)
	}
	student1 := qtest.CreateFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: student1.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   student1.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	student2 := qtest.CreateFakeUser(t, db, 4)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: student2.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   student2.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	repo1 := pb.Repository{
		OrganizationID: 1,
		RepositoryID:   1,
		UserID:         student1.ID,
		RepoType:       pb.Repository_USER,
	}
	if err := db.CreateRepository(&repo1); err != nil {
		t.Fatal(err)
	}
	repo2 := pb.Repository{
		OrganizationID: 1,
		RepositoryID:   2,
		UserID:         student2.ID,
		RepoType:       pb.Repository_USER,
	}
	if err := db.CreateRepository(&repo2); err != nil {
		t.Fatal(err)
	}
	fakeProvider, scms := qtest.FakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	ctx := withUserContext(context.Background(), teacher)

	_, err = fakeProvider.CreateOrganization(context.Background(), &scm.OrganizationOptions{Path: "path", Name: "name"})
	if err != nil {
		t.Fatal(err)
	}
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

	if err = db.CreateSubmission(&pb.Submission{
		AssignmentID: 1,
		UserID:       student1.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err = db.CreateSubmission(&pb.Submission{
		AssignmentID: 1,
		UserID:       student2.ID,
	}); err != nil {
		t.Fatal(err)
	}

	// rebuild a single submission
	var rebuildRequest pb.RebuildRequest
	rebuildRequest.AssignmentID = assignment.ID
	rebuildRequest.SubmissionID = 123
	if _, err := ags.RebuildSubmission(ctx, &rebuildRequest); err == nil {
		t.Errorf("Expected error: record not found")
	}
	rebuildRequest.SubmissionID = 1
	if _, err := ags.RebuildSubmission(ctx, &rebuildRequest); err != nil {
		t.Fatalf("Failed to rebuild submission: %s", err)
	}
	submissions, err := db.GetSubmissions(&pb.Submission{AssignmentID: assignment.ID})
	if err != nil {
		t.Fatalf("Failed to get created submissions: %s", err)
	}

	// make sure wrong course ID returns error
	var request pb.AssignmentRequest
	request.CourseID = 15
	if _, err = ags.RebuildSubmissions(ctx, &request); err == nil {
		t.Fatal("Expected error: record not found")
	}

	// make sure wrong assignment ID returns error
	request.CourseID = course.ID
	request.AssignmentID = 1337
	if _, err = ags.RebuildSubmissions(ctx, &request); err == nil {
		t.Fatal("Expected error: record not found")
	}

	request.AssignmentID = assignment.ID
	if _, err = ags.RebuildSubmissions(ctx, &request); err != nil {
		t.Fatalf("Failed to rebuild submissions: %s", err)
	}
	rebuiltSubmissions, err := db.GetSubmissions(&pb.Submission{AssignmentID: assignment.ID})
	if err != nil {
		t.Fatalf("Failed to get created submissions: %s", err)
	}
	if len(submissions) != len(rebuiltSubmissions) {
		t.Errorf("Incorrect number of submissions after rebuild: expected %d, got %d", len(submissions), len(rebuiltSubmissions))
	}

	// check access control
	ctx = withUserContext(ctx, student1)
	if _, err = ags.RebuildSubmissions(ctx, &request); err == nil {
		t.Fatal("Expected error: authentication failed")
	}
}
