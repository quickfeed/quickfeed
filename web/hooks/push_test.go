package hooks

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/stream"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestExtractAssignments(t *testing.T) {
	course := qtest.MockCourses[0]
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	wh := NewGitHubWebHook(qtest.Logger(t), db, &scm.Manager{}, &ci.Local{}, "secret", stream.NewStreamServices())
	admin := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateCourse(t, db, admin, course)

	assignments := []*qf.Assignment{
		{
			CourseID: course.ID,
			Order:    1,
			Name:     "lab1",
		},
		{
			CourseID: course.ID,
			Order:    2,
			Name:     "lab2",
		},
		{
			CourseID: course.ID,
			Order:    3,
			Name:     "lab3",
		},
	}
	for _, assignment := range assignments {
		if err := db.CreateAssignment(assignment); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name            string
		modified        []string
		added           []string
		removed         []string
		wantAssignments []*qf.Assignment
	}{
		{
			"modified lab1, lab3",
			[]string{"lab1/lab123.go", "lab3/temp.go"},
			[]string{},
			[]string{},
			[]*qf.Assignment{assignments[0], assignments[2]},
		},
		{
			"added lab2",
			[]string{},
			[]string{"lab2/assignment.go"},
			[]string{},
			[]*qf.Assignment{assignments[1]},
		},
		{
			"removed lab1, modified lab1",
			[]string{"lab1/name.go"},
			[]string{},
			[]string{"lab1/lab1.go"},
			[]*qf.Assignment{assignments[0]},
		},
		{
			"modified lab1, added lab2, removed lab3",
			[]string{"lab1/test.go"},
			[]string{"lab2/assignment.go"},
			[]string{"lab3/lab1.go"},
			assignments,
		},
	}

	for _, tt := range tests {
		got := wh.extractAssignments(&github.PushEvent{
			Commits: []*github.HeadCommit{
				{
					Modified: tt.modified,
					Added:    tt.added,
					Removed:  tt.removed,
				},
			},
		}, course)
		sort.Slice(got, func(i, j int) bool {
			return got[i].Order < got[j].Order
		})
		if diff := cmp.Diff(tt.wantAssignments, got, protocmp.Transform()); diff != "" {
			t.Errorf("%s: mismatch (-want, +got):\n%s", tt.name, diff)
		}
	}
}

func TestLastActivityDate(t *testing.T) {
	course := qtest.MockCourses[0]
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	wh := NewGitHubWebHook(qtest.Logger(t), db, &scm.Manager{}, &ci.Local{}, "secret", stream.NewStreamServices())
	admin := qtest.CreateFakeUser(t, db, 1)
	qtest.CreateCourse(t, db, admin, course)

	date := timestamppb.Now()
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
		if diff := cmp.Diff(date.Seconds, enrol.LastActivityDate.Seconds); diff != "" {
			t.Errorf("last activity date mismatch: (-want, +got):\n%s", diff)
		}
		// Remove updated date.
		if err := db.UpdateEnrollment(&qf.Enrollment{
			UserID:           admin.ID,
			CourseID:         course.ID,
			LastActivityDate: nil,
		}); err != nil {
			t.Fatal(err)
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
