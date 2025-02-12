package hooks

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v62/github"
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
	wh := NewGitHubWebHook(qtest.Logger(t), db, &scm.Manager{}, &ci.Local{}, "secret", stream.NewStreamServices(), nil)
	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, course)

	assignments := []*qf.Assignment{
		{
			CourseID: course.GetID(),
			Order:    1,
			Name:     "lab1",
		},
		{
			CourseID: course.GetID(),
			Order:    2,
			Name:     "lab2",
		},
		{
			CourseID: course.GetID(),
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
			return got[i].GetOrder() < got[j].GetOrder()
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
	wh := NewGitHubWebHook(qtest.Logger(t), db, &scm.Manager{}, &ci.Local{}, "secret", stream.NewStreamServices(), nil)
	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, course)

	tests := []struct {
		name string
		repo *qf.Repository
	}{
		{
			"user repo",
			&qf.Repository{
				UserID:   admin.GetID(),
				RepoType: qf.Repository_USER,
			},
		},
		{
			"group repo",
			&qf.Repository{
				UserID:   admin.GetID(),
				RepoType: qf.Repository_GROUP,
			},
		},
	}

	for _, tt := range tests {
		date := timestamppb.Now()
		wh.updateLastActivityDate(course, tt.repo, admin.GetLogin())
		enrol, err := db.GetEnrollmentByCourseAndUser(course.GetID(), admin.GetID())
		if err != nil {
			t.Fatal(err)
		}
		if !inOneSecondRange(date.GetSeconds(), enrol.GetLastActivityDate().GetSeconds()) {
			t.Errorf("last activity date mismatch: %d, expected %d", enrol.GetLastActivityDate().Seconds, date.GetSeconds())
		}
		// Remove updated date.
		if err := db.UpdateEnrollment(&qf.Enrollment{
			ID:               enrol.GetID(),
			UserID:           admin.GetID(),
			CourseID:         course.GetID(),
			LastActivityDate: nil,
		}); err != nil {
			t.Fatal(err)
		}
	}
}

// inOneSecondRange returns true if a and b are within one second of each other.
func inOneSecondRange(a, b int64) bool {
	diff := a - b
	return diff < 2 && diff > -2
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

func TestIgnorePush(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	wh := NewGitHubWebHook(qtest.Logger(t), db, &scm.Manager{}, &ci.Local{}, "secret", stream.NewStreamServices(), nil)

	repo := qf.RepoURL{ProviderURL: "github.com", Organization: "dat520-2024"}
	usrRepo := &qf.Repository{RepoType: qf.Repository_USER, HTMLURL: repo.StudentRepoURL("user")}
	grpRepo := &qf.Repository{RepoType: qf.Repository_GROUP, HTMLURL: repo.GroupRepoURL("group")}
	pushEventRepo := &github.PushEventRepository{DefaultBranch: github.String("main")}
	pushMain := &github.PushEvent{Ref: github.String("refs/heads/main"), Repo: pushEventRepo}
	pushFeat := &github.PushEvent{Ref: github.String("refs/heads/feat-branch"), Repo: pushEventRepo}
	pullFeat := &qf.PullRequest{ScmRepositoryID: 1, TaskID: 1, IssueID: 1, UserID: 1, Number: 1, SourceBranch: "feat-branch"}

	const ignore bool = true
	tests := []struct {
		name        string
		repo        *qf.Repository
		pushEvent   *github.PushEvent
		pullRequest *qf.PullRequest
		want        bool // true = ignore, false = process
	}{
		{name: "DefaultBranch/UsrRepo", repo: usrRepo, pushEvent: pushMain, want: !ignore},
		{name: "DefaultBranch/GrpRepo", repo: grpRepo, pushEvent: pushMain, want: !ignore},
		{name: "DefaultBranch/UsrRepo/WithPullRequest", repo: usrRepo, pushEvent: pushMain, pullRequest: pullFeat, want: !ignore},
		{name: "DefaultBranch/GrpRepo/WithPullRequest", repo: grpRepo, pushEvent: pushMain, pullRequest: pullFeat, want: !ignore},
		{name: "FeatureBranch/UsrRepo", repo: usrRepo, pushEvent: pushFeat, want: ignore},
		{name: "FeatureBranch/GrpRepo", repo: grpRepo, pushEvent: pushFeat, want: ignore},
		{name: "FeatureBranch/UsrRepo/WithPullRequest", repo: usrRepo, pushEvent: pushFeat, pullRequest: pullFeat, want: ignore},
		{name: "FeatureBranch/GrpRepo/WithPullRequest", repo: grpRepo, pushEvent: pushFeat, pullRequest: pullFeat, want: !ignore},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pullRequest != nil {
				tt.pullRequest.ID = uint64(i)
				if err := db.CreatePullRequest(tt.pullRequest); err != nil {
					t.Fatal(err)
				}
				// Clean up between subtests to avoid pull requests being processed in other subtests.
				t.Cleanup(func() {
					if err := db.UpdatePullRequest(&qf.PullRequest{ID: tt.pullRequest.GetID()}); err != nil {
						t.Fatal(err)
					}
				})
			}
			if got := wh.ignorePush(tt.pushEvent, tt.repo); got != tt.want {
				t.Errorf("ignorePush(%s, %s) = %t, want %t", branchName(tt.pushEvent.GetRef()), tt.repo.Name(), got, tt.want)
			}
		})
	}
}
