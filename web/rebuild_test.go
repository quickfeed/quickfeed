package web_test

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
)

func TestSimulatedRebuildWorkPoolWithErrCount(t *testing.T) {
	t.Skip("Disabled: mainly used for testing the work pool logic used in rebuildSubmissions")
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
	_, mgr := scm.MockSCMManager(t)
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qlog.Logger(t).Desugar()
	q := web.NewQuickFeedService(logger, db, mgr, web.BaseHookOptions{}, &ci.Local{})
	teacher := qtest.CreateFakeUser(t, db, 1)
	err := db.UpdateUser(&qf.User{ID: teacher.ID, IsAdmin: true})
	if err != nil {
		t.Fatal(err)
	}
	course := qf.Course{
		Name:             "QuickFeed Test Course",
		Code:             "qf101",
		Provider:         "fake",
		OrganizationID:   1,
		OrganizationName: qtest.MockOrg,
	}
	if err := db.CreateCourse(teacher.ID, &course); err != nil {
		t.Fatal(err)
	}
	student1 := qtest.CreateFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: student1.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   student1.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	student2 := qtest.CreateFakeUser(t, db, 4)
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: student2.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   student2.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	repo := qf.RepoURL{ProviderURL: "github.com", Organization: course.OrganizationName}
	repo1 := qf.Repository{
		OrganizationID: 1,
		RepositoryID:   1,
		UserID:         student1.ID,
		RepoType:       qf.Repository_USER,
		HTMLURL:        repo.StudentRepoURL("user"),
	}
	if err := db.CreateRepository(&repo1); err != nil {
		t.Fatal(err)
	}
	repo2 := qf.Repository{
		OrganizationID: 1,
		RepositoryID:   2,
		UserID:         student2.ID,
		RepoType:       qf.Repository_USER,
	}
	if err := db.CreateRepository(&repo2); err != nil {
		t.Fatal(err)
	}

	ctx := auth.WithUserContext(context.Background(), teacher)
	assignment := &qf.Assignment{
		CourseID: course.ID,
		Name:     "lab1",
		RunScriptContent: `#image/quickfeed:go
printf "AssignmentName: lab1\n"
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

	if err = db.CreateSubmission(&qf.Submission{
		AssignmentID: 1,
		UserID:       student1.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err = db.CreateSubmission(&qf.Submission{
		AssignmentID: 1,
		UserID:       student2.ID,
	}); err != nil {
		t.Fatal(err)
	}

	// try to rebuild non-existing submission
	rebuildRequest := connect.Request[qf.RebuildRequest]{Msg: &qf.RebuildRequest{
		AssignmentID: assignment.ID,
		SubmissionID: 123,
	}}
	if _, err := q.RebuildSubmissions(ctx, &rebuildRequest); err == nil {
		t.Errorf("Expected error: record not found")
	}

	os.Setenv("QUICKFEED_REPOSITORY_PATH", filepath.Join(env.Root(), "testdata", "courses"))
	// rebuild existing submission
	rebuildRequest.Msg.SubmissionID = 1
	if _, err := q.RebuildSubmissions(ctx, &rebuildRequest); err != nil {
		t.Fatalf("Failed to rebuild submission: %s", err)
	}
	submissions, err := db.GetSubmissions(&qf.Submission{AssignmentID: assignment.ID})
	if err != nil {
		t.Fatalf("Failed to get created submissions: %s", err)
	}

	// make sure wrong assignment ID returns error
	request := &connect.Request[qf.RebuildRequest]{Msg: &qf.RebuildRequest{}}

	request.Msg.SubmissionID = course.ID
	request.Msg.AssignmentID = 1337
	if _, err = q.RebuildSubmissions(ctx, request); err == nil {
		t.Fatal("Expected error: record not found")
	}

	request.Msg.AssignmentID = assignment.ID
	if _, err = q.RebuildSubmissions(ctx, request); err != nil {
		t.Fatalf("Failed to rebuild submissions: %s", err)
	}
	rebuiltSubmissions, err := db.GetSubmissions(&qf.Submission{AssignmentID: assignment.ID})
	if err != nil {
		t.Fatalf("Failed to get created submissions: %s", err)
	}
	if len(submissions) != len(rebuiltSubmissions) {
		t.Errorf("Incorrect number of submissions after rebuild: expected %d, got %d", len(submissions), len(rebuiltSubmissions))
	}
}
