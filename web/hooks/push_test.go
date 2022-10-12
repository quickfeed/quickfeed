package hooks

import (
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestPullRequestPushPayload(t *testing.T) {
	course := qtest.MockCourses[0]
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	wh := NewGitHubWebHook(qtest.Logger(t), db, &scm.Manager{}, &ci.Local{}, "secret")
	admin := qtest.CreateAdminUser(t, db, "fake")
	qtest.CreateCourse(t, db, admin, course)

	if err := db.CreateAssignment(&qf.Assignment{
		CourseID: course.ID,
		Order:    1,
		Name:     "lab1",
	}); err != nil {
		t.Fatal(err)
	}

	task := &qf.Task{
		AssignmentID: 1,
		Name:         "test_task",
	}

	taskMap := make(map[uint32]map[string]*qf.Task)
	tasks := make(map[string]*qf.Task)
	tasks[task.Name] = task
	taskMap[1] = tasks
	if _, _, err := db.SynchronizeAssignmentTasks(course, taskMap); err != nil {
		t.Fatal(err)
	}

	pr := &qf.PullRequest{
		SourceBranch:    "main",
		ScmRepositoryID: 1,
		UserID:          admin.ID,
		TaskID:          1,
		IssueID:         1,
		Number:          1,
	}
	if err := db.CreatePullRequest(pr); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		ref      string
		repoID   int64
		wantTask string
		wantPR   *qf.PullRequest
		wantErr  bool
	}{
		{
			"existing PR",
			"refs/heads/main",
			1,
			task.Name,
			pr,
			false,
		},
		{
			"wrong branch",
			"refs/heads/test",
			1,
			"",
			nil,
			true,
		},
		{
			"wrong repo",
			"refs/heads/main",
			11,
			"",
			nil,
			true,
		},
	}
	for _, tt := range tests {
		gotPR, gotTask, err := wh.handlePullRequestPushPayload(&github.PushEvent{
			Ref: &tt.ref,
			Repo: &github.PushEventRepository{
				ID: &tt.repoID,
			},
		})

		if (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error %v, got = %v, ", tt.name, tt.wantErr, err)
		}
		if diff := cmp.Diff(gotTask, tt.wantTask); diff != "" {
			t.Errorf("%s: expected task name '%s', got '%s'", tt.name, tt.wantTask, gotTask)
		}
		if diff := cmp.Diff(gotPR, tt.wantPR, protocmp.Transform()); diff != "" {
			t.Errorf("%s: mismatch (-want PR, +got PR):\n%s", tt.name, diff)
		}
	}
}

func TestExtractAssignments(t *testing.T) {
	course := qtest.MockCourses[0]
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	wh := NewGitHubWebHook(qtest.Logger(t), db, &scm.Manager{}, &ci.Local{}, "secret")
	admin := qtest.CreateAdminUser(t, db, "fake")
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
