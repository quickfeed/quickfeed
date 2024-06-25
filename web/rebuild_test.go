package web_test

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/fileop"
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

// prepareGitRepo creates copies src/repo folder to dst and initializes
// dst/repo as a git repository and adds a single file lab1/lab1.go.
func prepareGitRepo(src, dst, repo string) error {
	if err := fileop.CopyDir(filepath.Join(src, repo), dst); err != nil {
		return err
	}
	gitRepo := filepath.Join(dst, repo)
	r, err := git.PlainInit(gitRepo, false)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	_, err = w.Add("lab1")
	if err != nil {
		return err
	}
	_, err = w.Commit("added lab1", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@itest.run",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func TestRebuildSubmissions(t *testing.T) {
	repoPath := t.TempDir()
	t.Setenv("QUICKFEED_REPOSITORY_PATH", repoPath)

	src := filepath.Join(env.TestdataPath(), qtest.MockOrg)
	dst := filepath.Join(repoPath, qtest.MockOrg)
	err := prepareGitRepo(src, dst, qf.StudentRepoName("user"))
	if err != nil {
		t.Fatal(err)
	}
	err = prepareGitRepo(src, dst, qf.TestsRepo)
	if err != nil {
		t.Fatal(err)
	}
	err = prepareGitRepo(src, dst, qf.AssignmentsRepo)
	if err != nil {
		t.Fatal(err)
	}

	_, mgr := scm.MockSCMManager(t, scm.WithMockOrgs())
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qtest.Logger(t).Desugar()
	q := web.NewQuickFeedService(logger, db, mgr, web.BaseHookOptions{}, &ci.Local{})
	teacher := qtest.CreateFakeUser(t, db)
	err = db.UpdateUser(&qf.User{ID: teacher.ID, IsAdmin: true})
	if err != nil {
		t.Fatal(err)
	}
	course := qf.Course{
		Name:                "QuickFeed Test Course",
		Code:                "qf101",
		ScmOrganizationID:   1,
		ScmOrganizationName: qtest.MockOrg,
	}
	if err := db.CreateCourse(teacher.ID, &course); err != nil {
		t.Fatal(err)
	}
	student1 := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, student1, &course)

	student2 := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, student2, &course)

	repo := qf.RepoURL{ProviderURL: "github.com", Organization: course.ScmOrganizationName}
	repo1 := qf.Repository{
		ScmOrganizationID: 1,
		ScmRepositoryID:   1,
		UserID:            student1.ID,
		RepoType:          qf.Repository_USER,
		HTMLURL:           repo.StudentRepoURL("user"),
	}
	if err := db.CreateRepository(&repo1); err != nil {
		t.Fatal(err)
	}
	repo2 := qf.Repository{
		ScmOrganizationID: 1,
		ScmRepositoryID:   2,
		UserID:            student2.ID,
		RepoType:          qf.Repository_USER,
	}
	if err := db.CreateRepository(&repo2); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
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
