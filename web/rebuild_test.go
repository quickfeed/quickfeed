package web_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
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
	repoPath := t.TempDir()
	t.Setenv("QUICKFEED_REPOSITORY_PATH", repoPath)

	src := filepath.Join(env.TestdataPath(), qtest.MockOrg)
	dst := filepath.Join(repoPath, qtest.MockOrg)
	qtest.PrepareGitRepo(t, src, dst, qf.StudentRepoName("user"))
	qtest.PrepareGitRepo(t, src, dst, qf.TestsRepo)
	qtest.PrepareGitRepo(t, src, dst, qf.AssignmentsRepo)

	mgr := scm.MockManager(t, scm.WithMockOrgs())
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qtest.Logger(t).Desugar()
	q := web.NewQuickFeedService(logger, db, mgr, web.BaseHookOptions{}, &ci.Local{})
	teacher := qtest.CreateFakeUser(t, db)
	qtest.UpdateUser(t, db, &qf.User{ID: teacher.GetID(), IsAdmin: true})

	course := qtest.MockCourses[0]

	qtest.CreateCourse(t, db, teacher, course)
	student1 := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, student1, course)

	student2 := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, student2, course)

	qtest.CreateRepository(t, db, &qf.Repository{
		ScmOrganizationID: 1,
		ScmRepositoryID:   1,
		UserID:            student1.GetID(),
		RepoType:          qf.Repository_USER,
		HTMLURL:           qf.RepoURL{ProviderURL: "github.com", Organization: course.GetScmOrganizationName()}.StudentRepoURL("user"),
	})
	qtest.CreateRepository(t, db, &qf.Repository{
		ScmOrganizationID: 1,
		ScmRepositoryID:   2,
		UserID:            student2.GetID(),
		RepoType:          qf.Repository_USER,
	})

	ctx := context.Background()
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
	qtest.CreateSubmission(t, db, &qf.Submission{
		AssignmentID: 1,
		UserID:       student1.GetID(),
	})
	qtest.CreateSubmission(t, db, &qf.Submission{
		AssignmentID: 1,
		UserID:       student2.GetID(),
	})

	initialSubmissions := qtest.GetSubmissions(t, db, &qf.Submission{AssignmentID: assignment.GetID()})
	if len(initialSubmissions) != 2 {
		t.Fatalf("Expected 2 submissions, got %d", len(initialSubmissions))
	}
	errFailedRebuildSubmission := connect.NewError(connect.CodeInvalidArgument, errors.New("failed to rebuild submission"))
	tests := []struct {
		name    string
		request *connect.Request[qf.RebuildRequest]
		wantErr error
	}{
		{
			name: "Rebuild non-existing submission",
			request: &connect.Request[qf.RebuildRequest]{Msg: &qf.RebuildRequest{
				AssignmentID: assignment.GetID(),
				SubmissionID: 123,
			}},
			wantErr: errFailedRebuildSubmission,
		},
		{
			name: "Wrong assignment ID",
			request: &connect.Request[qf.RebuildRequest]{Msg: &qf.RebuildRequest{
				AssignmentID: 1337,
				SubmissionID: 1,
			}},
			wantErr: errFailedRebuildSubmission,
		},
		{
			name: "Rebuild all submissions with invalid assignment ID",
			request: &connect.Request[qf.RebuildRequest]{Msg: &qf.RebuildRequest{
				AssignmentID: 111,
			}},
			wantErr: connect.NewError(connect.CodeInvalidArgument, errors.New("failed to rebuild submissions")),
		},
		{
			name: "Rebuild existing submission",
			request: &connect.Request[qf.RebuildRequest]{Msg: &qf.RebuildRequest{
				AssignmentID: assignment.GetID(),
				SubmissionID: 1,
			}},
		},
		{
			name: "Rebuild all submissions",
			request: &connect.Request[qf.RebuildRequest]{Msg: &qf.RebuildRequest{
				AssignmentID: assignment.GetID(),
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := q.RebuildSubmissions(ctx, test.request)
			qtest.CheckError(t, err, test.wantErr)
		})
	}
	rebuiltSubmissions := qtest.GetSubmissions(t, db, &qf.Submission{AssignmentID: assignment.GetID()})
	if len(initialSubmissions) != len(rebuiltSubmissions) {
		t.Errorf("Incorrect number of submissions after rebuild: expected %d, got %d", len(initialSubmissions), len(rebuiltSubmissions))
	}
}
