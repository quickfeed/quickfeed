package hooks

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"testing/synctest"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
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
	synctest.Test(t, func(t *testing.T) {
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
		wh.syncStudentRepos(t.Context(), scmClient, course, "master")

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
	})
}

func TestSyncStudentReposWithErrors(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
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
		wh.syncStudentRepos(t.Context(), scmClient, course, "master")

		if len(scmClient.syncCalls) != 2 {
			t.Errorf("expected 2 sync calls, got %d", len(scmClient.syncCalls))
		}
	})
}

// TestSyncStudentReposNoDuplicateScope guards syncStudentRepos against repeating
// an attribute that the push event's scope already carries. handlePush scopes the
// context to the assignments repository before calling syncStudentRepos, and slog
// keeps duplicate keys, so a repeated attribute shows up twice in one record.
func TestSyncStudentReposNoDuplicateScope(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	wh := NewGitHubWebHook(logger, db, nil, &ci.Local{}, "secret", stream.NewStreamServices(), nil)
	admin := qtest.CreateFakeUser(t, db)
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, admin, course)

	repo := &qf.Repository{
		ScmOrganizationID: course.GetScmOrganizationID(),
		ScmRepositoryID:   1,
		UserID:            admin.ID,
		HTMLURL:           "https://github.com/org/user1-labs",
		RepoType:          qf.Repository_USER,
	}
	if err := db.CreateRepository(repo); err != nil {
		t.Fatal(err)
	}

	// The scope handlePush establishes for a push to the assignments repository.
	ctx := qlog.With(qlog.NewContext(t.Context(), logger),
		label.Repository, qf.AssignmentsRepo,
		label.RepositoryType, qf.Repository_ASSIGNMENTS.String(),
	)
	// A failing sync exercises the per-repository record, which is the one that
	// names a repository other than the one the enclosing scope carries.
	wh.syncStudentRepos(ctx, &mockSCM{errs: []error{errors.New("sync failed")}}, course, "master")

	scopedKeys := []string{label.Repository, label.RepositoryType, label.CourseCode, label.CourseID}
	for record := range strings.SplitSeq(strings.TrimSpace(output.String()), "\n") {
		if record == "" {
			continue
		}
		for _, key := range scopedKeys {
			if count := strings.Count(record, `"`+key+`":`); count > 1 {
				t.Errorf("log record repeats %q %d times, want at most once: %s", key, count, record)
			}
		}
	}
}
