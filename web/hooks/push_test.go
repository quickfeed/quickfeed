package hooks

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

func TestLastActivityDate(t *testing.T) {
	course := qtest.MockCourses[0]
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	wh := NewGitHubWebHook(qtest.Logger(t), db, &scm.Manager{}, &ci.Local{}, "secret")
	admin := qtest.CreateAdminUser(t, db, "fake")
	qtest.CreateCourse(t, db, admin, course)

	date := time.Now().Format("02 Jan")

	tests := []struct {
		name string
		repo *qf.Repository
	}{
		{
			"user repo",
			&qf.Repository{
				UserID:   admin.ID,
				RepoType: qf.Repository_USER,
			},
		},
		{
			"group repo",
			&qf.Repository{
				UserID:   admin.ID,
				RepoType: qf.Repository_GROUP,
			},
		},
	}

	for _, tt := range tests {
		wh.updateLastActivityDate(course, tt.repo, admin.Login)
		enrol, err := db.GetEnrollmentByCourseAndUser(course.ID, admin.ID)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(enrol.LastActivityDate, date); diff != "" {
			t.Errorf("expected last activity date: %s, got %s", date, enrol.LastActivityDate)
		}
		// Remove updated date.
		if err := db.UpdateEnrollment(&qf.Enrollment{
			UserID:           admin.ID,
			CourseID:         course.ID,
			LastActivityDate: "none",
		}); err != nil {
			t.Error(err)
		}
	}
}

func TestBranchName(t *testing.T) {
	tests := []struct {
		ref        string
		wantBranch string
	}{
		{
			"refs/heads/main",
			"main",
		},
		{
			"refs/heads/master",
			"master",
		},
		{
			"/refs/main",
			"main",
		},
	}

	for _, tt := range tests {
		gotBranch := branchName(tt.ref)
		if gotBranch != tt.wantBranch {
			t.Errorf("expected branch name %s, got %s", tt.wantBranch, gotBranch)
		}
	}
}

func TestDefaultBranch(t *testing.T) {
	tests := []struct {
		ref         string
		repoDefault string
		want        bool
	}{
		{
			"refs/heads/main",
			"main",
			true,
		},
		{
			"refs/heads/master",
			"master",
			true,
		},
		{
			"refs/heads/main",
			"master",
			false,
		},
		{
			"refs/heads/master",
			"main",
			false,
		},
	}

	for _, tt := range tests {
		payload := &github.PushEvent{
			Ref: &tt.ref,
			Repo: &github.PushEventRepository{
				DefaultBranch: &tt.repoDefault,
			},
		}
		if isDefaultBranch(payload) != tt.want {
			t.Errorf("default branch: '%s', ref branch: '%s', expected to match: '%v'",
				tt.repoDefault, tt.ref, tt.want)
		}
	}
}
