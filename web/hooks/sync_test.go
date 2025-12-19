package hooks

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/stream"
)

type mockSCM struct {
	scm.SCM
	syncCalls map[string]int
	mu        sync.Mutex
	errs      []error
}

func (m *mockSCM) SyncFork(_ context.Context, opt *scm.SyncForkOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.syncCalls == nil {
		m.syncCalls = make(map[string]int)
	}
	m.syncCalls[opt.Repository]++

	if len(m.errs) > 0 {
		err := m.errs[0]
		m.errs = m.errs[1:]
		return err
	}
	return nil
}

func TestSyncStudentRepos(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	wh := NewGitHubWebHook(qtest.Logger(t), db, nil, &ci.Local{}, "secret", stream.NewStreamServices(), nil)
	admin := qtest.CreateFakeUser(t, db)
	user2 := qtest.CreateFakeUser(t, db)
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, admin, course)

	// Create some repositories
	repos := []*qf.Repository{
		{
			ScmOrganizationID: course.GetScmOrganizationID(),
			ScmRepositoryID:   1,
			UserID:            admin.ID,
			HTMLURL:           "https://github.com/org/user1-labs",
			RepoType:          qf.Repository_USER,
		},
		{
			ScmOrganizationID: course.GetScmOrganizationID(),
			ScmRepositoryID:   2,
			UserID:            user2.ID,
			HTMLURL:           "https://github.com/org/user2-labs",
			RepoType:          qf.Repository_USER,
		},
		{
			ScmOrganizationID: course.GetScmOrganizationID(),
			ScmRepositoryID:   3,
			HTMLURL:           "https://github.com/org/assignments",
			RepoType:          qf.Repository_ASSIGNMENTS,
		},
	}
	for _, repo := range repos {
		if err := db.CreateRepository(repo); err != nil {
			t.Fatal(err)
		}
	}

	scmClient := &mockSCM{}
	wh.syncStudentRepos(context.Background(), scmClient, course, "master")

	if len(scmClient.syncCalls) != 2 {
		t.Errorf("expected 2 sync calls, got %d", len(scmClient.syncCalls))
	}
	if scmClient.syncCalls["user1-labs"] != 1 {
		t.Errorf("expected 1 sync call for user1-labs, got %d", scmClient.syncCalls["user1-labs"])
	}
	if scmClient.syncCalls["user2-labs"] != 1 {
		t.Errorf("expected 1 sync call for user2-labs, got %d", scmClient.syncCalls["user2-labs"])
	}
	if scmClient.syncCalls["assignments"] != 0 {
		t.Errorf("expected 0 sync calls for assignments, got %d", scmClient.syncCalls["assignments"])
	}
}

func TestSyncStudentReposWithErrors(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	wh := NewGitHubWebHook(qtest.Logger(t), db, nil, &ci.Local{}, "secret", stream.NewStreamServices(), nil)
	admin := qtest.CreateFakeUser(t, db)
	user2 := qtest.CreateFakeUser(t, db)
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, admin, course)

	repos := []*qf.Repository{
		{
			ScmOrganizationID: course.GetScmOrganizationID(),
			ScmRepositoryID:   1,
			UserID:            admin.ID,
			HTMLURL:           "https://github.com/org/user1-labs",
			RepoType:          qf.Repository_USER,
		},
		{
			ScmOrganizationID: course.GetScmOrganizationID(),
			ScmRepositoryID:   2,
			UserID:            user2.ID,
			HTMLURL:           "https://github.com/org/user2-labs",
			RepoType:          qf.Repository_USER,
		},
	}
	for _, repo := range repos {
		if err := db.CreateRepository(repo); err != nil {
			t.Fatal(err)
		}
	}

	scmClient := &mockSCM{
		errs: []error{errors.New("sync failed"), nil},
	}
	wh.syncStudentRepos(context.Background(), scmClient, course, "master")

	if len(scmClient.syncCalls) != 2 {
		t.Errorf("expected 2 sync calls, got %d", len(scmClient.syncCalls))
	}
}

func TestSyncForkWithRetry(t *testing.T) {
	wh := GitHubWebHook{
		logger: qtest.Logger(t),
	}

	t.Run("Success", func(t *testing.T) {
		scmClient := &mockSCM{}
		err := wh.syncForkWithRetry(context.Background(), scmClient, "org", "repo", "master")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scmClient.syncCalls["repo"] != 1 {
			t.Errorf("expected 1 sync call, got %d", scmClient.syncCalls["repo"])
		}
	})

	t.Run("RateLimitRetrySuccess", func(t *testing.T) {
		// This test might be slow due to 1s initial delay.
		// We use a short reset time to avoid long wait if possible,
		// but the code adds 1s anyway if Reset.After(now) is true.
		// If we don't set Reset, it uses initialRetryDelay (1s).
		scmClient := &mockSCM{
			errs: []error{
				&github.RateLimitError{
					Rate: github.Rate{
						Reset: github.Timestamp{Time: time.Now().Add(100 * time.Millisecond)},
					},
					Response: &http.Response{},
				},
			},
		}
		start := time.Now()
		err := wh.syncForkWithRetry(context.Background(), scmClient, "org", "repo", "master")
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scmClient.syncCalls["repo"] != 2 {
			t.Errorf("expected 2 sync calls, got %d", scmClient.syncCalls["repo"])
		}
		if duration < time.Second {
			t.Errorf("expected retry delay of at least 1s, got %v", duration)
		}
	})

	t.Run("AbuseRateLimitRetrySuccess", func(t *testing.T) {
		retryAfter := 500 * time.Millisecond
		scmClient := &mockSCM{
			errs: []error{
				&github.AbuseRateLimitError{
					RetryAfter: &retryAfter,
					Response:   &http.Response{},
				},
			},
		}
		start := time.Now()
		err := wh.syncForkWithRetry(context.Background(), scmClient, "org", "repo", "master")
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scmClient.syncCalls["repo"] != 2 {
			t.Errorf("expected 2 sync calls, got %d", scmClient.syncCalls["repo"])
		}
		if duration < retryAfter {
			t.Errorf("expected retry delay of at least %v, got %v", retryAfter, duration)
		}
	})

	t.Run("NonRetryableError", func(t *testing.T) {
		expectedErr := errors.New("permanent error")
		scmClient := &mockSCM{
			errs: []error{expectedErr},
		}
		err := wh.syncForkWithRetry(context.Background(), scmClient, "org", "repo", "master")
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected error %v, got %v", expectedErr, err)
		}
		if scmClient.syncCalls["repo"] != 1 {
			t.Errorf("expected 1 sync call, got %d", scmClient.syncCalls["repo"])
		}
	})

	t.Run("MaxRetriesExceeded", func(t *testing.T) {
		scmClient := &mockSCM{
			errs: []error{
				errors.New("retryable error"), // This won't be retried unless it's a rate limit error
			},
		}
		// Wait, the current implementation only retries on RateLimitError or AbuseRateLimitError.
		// Let's test max retries with rate limit errors.
		scmClient.errs = []error{
			&github.RateLimitError{Response: &http.Response{}},
			&github.RateLimitError{Response: &http.Response{}},
			&github.RateLimitError{Response: &http.Response{}},
			&github.RateLimitError{Response: &http.Response{}},
		}
		// maxSyncRetries is 3, so it should try 4 times total (0, 1, 2, 3)
		err := wh.syncForkWithRetry(context.Background(), scmClient, "org", "repo", "master")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if scmClient.syncCalls["repo"] != 4 {
			t.Errorf("expected 4 sync calls, got %d", scmClient.syncCalls["repo"])
		}
	})
}
