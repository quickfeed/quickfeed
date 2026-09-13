package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/stream"
)

func TestCanaryTargets(t *testing.T) {
	tests := []struct {
		name      string
		commits   []*github.HeadCommit
		wantNames map[string]bool
		wantAll   bool
	}{
		{
			name:      "NoCommits",
			wantNames: map[string]bool{},
		},
		{
			name:      "AssignmentFolders",
			commits:   []*github.HeadCommit{{Modified: []string{"lab1/lab1.go", "lab2/lab2.go"}}},
			wantNames: map[string]bool{"lab1": true, "lab2": true},
		},
		{
			name:      "NestedPaths",
			commits:   []*github.HeadCommit{{Modified: []string{"lab3/detector/fd.go", "lab3/detector/testdata/in.txt"}}},
			wantNames: map[string]bool{"lab3": true},
		},
		{
			name:      "RootLevelFile",
			commits:   []*github.HeadCommit{{Modified: []string{"go.mod"}}},
			wantNames: map[string]bool{},
			wantAll:   true,
		},
		{
			// A root-level file is outside every assignment folder, whether or
			// not its name begins with a dot; only top-level folders whose name
			// begins with a dot are treated as repository metadata.
			name:      "RootLevelDotFile",
			commits:   []*github.HeadCommit{{Modified: []string{".quickfeedignore"}}},
			wantNames: map[string]bool{},
			wantAll:   true,
		},
		{
			name:      "RootLevelFileWithAssignmentFolder",
			commits:   []*github.HeadCommit{{Modified: []string{"lab1/lab1.go", "README.md"}}},
			wantNames: map[string]bool{"lab1": true},
			wantAll:   true,
		},
		{
			name:      "ScriptsFolder",
			commits:   []*github.HeadCommit{{Modified: []string{"scripts/run.sh"}}},
			wantNames: map[string]bool{},
			wantAll:   true,
		},
		{
			name:      "DotFolder",
			commits:   []*github.HeadCommit{{Modified: []string{".github/workflows/test.yml"}}},
			wantNames: map[string]bool{},
		},
		{
			name:      "EmptyAndLeadingSlashPaths",
			commits:   []*github.HeadCommit{{Modified: []string{"", "/leading", "/"}}},
			wantNames: map[string]bool{},
		},
		{
			// A folder that is not an assignment is returned as a name; only
			// runCanary knows the course's assignments and widens the run.
			name:      "SharedCodeFolder",
			commits:   []*github.HeadCommit{{Modified: []string{"internal/shared.go"}}},
			wantNames: map[string]bool{"internal": true},
		},
		{
			name: "AcrossChangeKinds",
			commits: []*github.HeadCommit{{
				Added:    []string{"lab1/lab1.go"},
				Modified: []string{"lab2/lab2.go"},
				Removed:  []string{"lab3/lab3.go"},
			}},
			wantNames: map[string]bool{"lab1": true, "lab2": true, "lab3": true},
		},
		{
			name: "AcrossCommits",
			commits: []*github.HeadCommit{
				{Added: []string{"lab1/lab1.go"}},
				{Modified: []string{"lab2/lab2.go", ".github/dependabot.yml"}},
				{Removed: []string{"lab1/old.go", "lab4/lab4.go"}},
			},
			wantNames: map[string]bool{"lab1": true, "lab2": true, "lab4": true},
		},
		{
			name: "ScriptsFolderInLaterCommit",
			commits: []*github.HeadCommit{
				{Modified: []string{"lab1/lab1.go"}},
				{Modified: []string{"scripts/Dockerfile"}},
			},
			wantNames: map[string]bool{"lab1": true},
			wantAll:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNames, gotAll := canaryTargets(&github.PushEvent{Commits: tt.commits})
			if diff := cmp.Diff(tt.wantNames, gotNames); diff != "" {
				t.Errorf("canaryTargets() names mismatch (-want +got):\n%s", diff)
			}
			if gotAll != tt.wantAll {
				t.Errorf("canaryTargets() all = %t, want %t", gotAll, tt.wantAll)
			}
		})
	}
}

// canaryTestOrg is the course organization used by the tests below. Its
// repositories are created below a temporary directory; no network, SCM client,
// or Docker daemon is involved, since the canary runs the skeleton code from
// the local clone of the assignments repository with the Local runner.
const canaryTestOrg = "qf1594-2025"

// canaryTestName is the single test the run script below reports a score for.
const canaryTestName = "TestCanary"

// canaryRunScriptFmt is a course run script awarding the skeleton code the
// given number of points out of ten. The Local runner replaces the process
// environment with the job's environment, so the script must only use bash
// builtins; see ci.Local.Run.
const canaryRunScriptFmt = `#image/dummy

echo "{\"Secret\":\"$QUICKFEED_SESSION_SECRET\",\"TestName\":\"` + canaryTestName + `\",\"Score\":%d,\"MaxScore\":10,\"Weight\":1}"
`

// canaryNoOutputScript is a course run script that produces no output and no
// error, which is classified as NO_SCORES; see ci.classifyRun.
const canaryNoOutputScript = `#image/dummy

:
`

// canaryRunMessages are the messages reporting that the canary actually ran.
var canaryRunMessages = []string{
	"test environment canary failed",
	"test environment canary passed",
	"test environment canary could not run",
	"test environment canary finished",
}

// canaryRecord is the subset of a JSON log record that the assertions below
// inspect; the enclosing scope's attributes are ignored.
type canaryRecord struct {
	Level       string `json:"level"`
	Message     string `json:"msg"`
	Assignment  string `json:"assignment"`
	Score       uint32 `json:"score"`
	Problem     string `json:"problem"`
	RunStatus   string `json:"run_status"`
	CloneHead   string `json:"clone_head"`
	Assignments int    `json:"assignments"`
	Problems    int    `json:"problems"`
	Skipped     int    `json:"skipped"`
}

// canaryRecords parses the JSON log records written to output.
func canaryRecords(t *testing.T, output string) []canaryRecord {
	t.Helper()
	var records []canaryRecord
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		var record canaryRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("failed to parse log record %q: %v", line, err)
		}
		records = append(records, record)
	}
	return records
}

// recordsWith returns the parsed records whose message is msg.
func recordsWith(records []canaryRecord, msg string) []canaryRecord {
	var matching []canaryRecord
	for _, record := range records {
		if record.Message == msg {
			matching = append(matching, record)
		}
	}
	return matching
}

// assertNoCanaryRun fails the test if output reports any canary run.
func assertNoCanaryRun(t *testing.T, output string) {
	t.Helper()
	records := canaryRecords(t, output)
	for _, msg := range canaryRunMessages {
		if n := len(recordsWith(records, msg)); n != 0 {
			t.Errorf("runCanary() logged %d %q records, want 0:\n%s", n, msg, output)
		}
	}
}

// setupCanary returns a webhook using the given runner, a course whose run
// script is the given one, the logging context that handlePush establishes
// before calling runCanary, and the buffer receiving the JSON log records.
func setupCanary(t *testing.T, runner ci.Runner, runScript string) (*GitHubWebHook, *qf.Course, context.Context, *bytes.Buffer) {
	t.Helper()
	db, cleanup := qtest.TestDB(t)
	t.Cleanup(cleanup)

	output := new(bytes.Buffer)
	logger := slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	wh := NewGitHubWebHook(logger, db, nil, runner, "secret", stream.NewStreamServices(), nil)
	course := setupCanaryCourse(t, db, runScript)
	ctx, _ := qlog.WithCourse(qlog.NewContext(t.Context(), logger), course)
	return wh, course, ctx, output
}

// setupCanaryCourse creates the course's tests and assignments repositories
// below a fresh repository path, along with the course and its two assignments:
// lab1 is graded by the tests, lab2 is graded manually.
func setupCanaryCourse(t *testing.T, db database.Database, runScript string) *qf.Course {
	t.Helper()
	root := t.TempDir()
	t.Setenv("QUICKFEED_REPOSITORY_PATH", root)
	for _, repo := range []string{qf.TestsRepo, qf.AssignmentsRepo} {
		for _, lab := range []string{"lab1", "lab2"} {
			mkdir(t, filepath.Join(root, canaryTestOrg, repo, lab))
		}
	}
	scripts := filepath.Join(root, canaryTestOrg, qf.TestsRepo, scriptsFolder)
	mkdir(t, scripts)
	if err := os.WriteFile(filepath.Join(scripts, "run.sh"), []byte(runScript), 0o600); err != nil {
		t.Fatal(err)
	}

	course := &qf.Course{
		Name:                "Test Environment Canary",
		Code:                "QF1594",
		Year:                2025,
		Tag:                 "Fall",
		ScmOrganizationID:   1594,
		ScmOrganizationName: canaryTestOrg,
	}
	qtest.CreateCourse(t, db, qtest.CreateFakeUser(t, db), course)
	assignments := []*qf.Assignment{
		{
			CourseID:      course.GetID(),
			Order:         1,
			Name:          "lab1",
			ExpectedTests: []*qf.TestInfo{{TestName: canaryTestName, MaxScore: 10, Weight: 1}},
		},
		{
			CourseID:  course.GetID(),
			Order:     2,
			Name:      "lab2",
			Reviewers: 1, // graded manually; the canary must skip it
		},
	}
	for _, assignment := range assignments {
		if err := db.CreateAssignment(assignment); err != nil {
			t.Fatal(err)
		}
	}
	return course
}

func mkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
}

// initCanaryClone turns dir into a git repository holding a single commit,
// standing in for the go-git clone that the server keeps of a course
// repository, and returns the hash of that commit.
func initCanaryClone(t *testing.T, dir string) string {
	t.Helper()
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# canary\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := worktree.Add("README.md"); err != nil {
		t.Fatal(err)
	}
	// The author is set explicitly, so that the commit does not depend on the
	// git configuration of the machine running the test.
	hash, err := worktree.Commit("add readme", &git.CommitOptions{
		Author: &object.Signature{Name: "QuickFeed", Email: "quickfeed@example.com", When: time.Now()},
	})
	if err != nil {
		t.Fatal(err)
	}
	return hash.String()
}

// canaryPush returns a push event modifying the given files.
func canaryPush(files ...string) *github.PushEvent {
	return &github.PushEvent{
		HeadCommit: &github.HeadCommit{ID: new("9e1b3a7")},
		Commits:    []*github.HeadCommit{{Modified: files}},
	}
}

// TestRunCanary runs the canary end-to-end with the Local runner: the skeleton
// code is taken from the local clone of the assignments repository and scored by
// the course's run script, so the scores below are those of the skeleton code.
func TestRunCanary(t *testing.T) {
	tests := []struct {
		name       string
		runScript  string // the course's run script
		wantLevel  string // level of the expected per-assignment record
		wantMsg    string // message of the expected per-assignment record
		wantScore  uint32 // score the record must report
		wantStatus string // run status the record must report, if any
		wantProbs  int    // problems the summary record must report
	}{
		{
			name:       "SkeletonPassesEveryTest",
			runScript:  fmt.Sprintf(canaryRunScriptFmt, 10),
			wantLevel:  slog.LevelWarn.String(),
			wantMsg:    "test environment canary failed",
			wantScore:  100,
			wantStatus: score.RunStatus_SUCCESS.String(),
			wantProbs:  1,
		},
		{
			name:      "SkeletonScoresZero",
			runScript: fmt.Sprintf(canaryRunScriptFmt, 0),
			wantLevel: slog.LevelInfo.String(),
			wantMsg:   "test environment canary passed",
			wantScore: 0,
		},
		{
			// A run script that prints nothing produces no test results at all,
			// which is a broken test environment rather than a healthy zero.
			name:       "RunProducesNoScores",
			runScript:  canaryNoOutputScript,
			wantLevel:  slog.LevelWarn.String(),
			wantMsg:    "test environment canary failed",
			wantScore:  0,
			wantStatus: score.RunStatus_NO_SCORES.String(),
			wantProbs:  1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wh, course, ctx, output := setupCanary(t, &ci.Local{}, tt.runScript)

			// A push touching lab1 checks lab1, and nothing else. The SCM
			// client is nil: the skeleton run needs no clone.
			wh.runCanary(ctx, nil, course, qf.TestsRepo, canaryPush("lab1/lab1.go"))
			records := canaryRecords(t, output.String())

			got := recordsWith(records, tt.wantMsg)
			if len(got) != 1 {
				t.Fatalf("runCanary() logged %d %q records, want 1:\n%s", len(got), tt.wantMsg, output.String())
			}
			if got[0].Level != tt.wantLevel {
				t.Errorf("runCanary() logged %q at level %s, want %s", tt.wantMsg, got[0].Level, tt.wantLevel)
			}
			if got[0].Assignment != "lab1" {
				t.Errorf("runCanary() logged %q for assignment %q, want %q", tt.wantMsg, got[0].Assignment, "lab1")
			}
			if got[0].Score != tt.wantScore {
				t.Errorf("runCanary() logged score %d for lab1, want %d", got[0].Score, tt.wantScore)
			}
			if got[0].RunStatus != tt.wantStatus {
				t.Errorf("runCanary() logged run status %q for lab1, want %q", got[0].RunStatus, tt.wantStatus)
			}
			if (got[0].Problem != "") != (tt.wantProbs > 0) {
				t.Errorf("runCanary() logged problem %q, want problem: %t", got[0].Problem, tt.wantProbs > 0)
			}

			summary := recordsWith(records, "test environment canary finished")
			if len(summary) != 1 {
				t.Fatalf("runCanary() logged %d summary records, want 1:\n%s", len(summary), output.String())
			}
			if summary[0].Assignments != 1 || summary[0].Problems != tt.wantProbs || summary[0].Skipped != 0 {
				t.Errorf("runCanary() summary = %d assignments, %d problems, %d skipped; want 1, %d, 0",
					summary[0].Assignments, summary[0].Problems, summary[0].Skipped, tt.wantProbs)
			}
			// The manually graded lab2 has no tests to check.
			for _, record := range records {
				if record.Assignment == "lab2" {
					t.Errorf("runCanary() logged %q for the manually graded lab2", record.Message)
				}
			}

			// A push touching only the manually graded lab2 runs nothing.
			output.Reset()
			wh.runCanary(ctx, nil, course, qf.TestsRepo, canaryPush("lab2/lab2.go"))
			assertNoCanaryRun(t, output.String())
		})
	}
}

// TestRunCanaryWidensToUnknownFolder checks that a push to a top-level folder
// that is not an assignment checks every assignment. Such a folder holds code
// shared by the assignments, and nothing in the push says which of them use it.
func TestRunCanaryWidensToUnknownFolder(t *testing.T) {
	const widenMsg = "changed folder is not an assignment; checking every assignment"
	wh, course, ctx, output := setupCanary(t, &ci.Local{}, fmt.Sprintf(canaryRunScriptFmt, 0))

	// A shared-code folder, such as one listed in .quickfeedignore.
	wh.runCanary(ctx, nil, course, qf.TestsRepo, canaryPush("internal/shared.go"))
	records := canaryRecords(t, output.String())

	widened := recordsWith(records, widenMsg)
	if len(widened) != 1 {
		t.Fatalf("runCanary() logged %d %q records, want 1:\n%s", len(widened), widenMsg, output.String())
	}
	if widened[0].Assignment != "internal" {
		t.Errorf("runCanary() logged %q for folder %q, want %q", widenMsg, widened[0].Assignment, "internal")
	}
	// lab1 is the course's only assignment graded by the tests.
	if n := len(recordsWith(records, "test environment canary passed")); n != 1 {
		t.Errorf("runCanary() logged %d passed records for lab1, want 1:\n%s", n, output.String())
	}
}

// contextRunner is a ci.Runner recording whether it ran, and the error of the
// run's context when it did.
type contextRunner struct {
	ran    bool
	ctxErr error
}

func (r *contextRunner) Run(ctx context.Context, _ *ci.Job) (string, error) {
	r.ran, r.ctxErr = true, ctx.Err()
	return "", nil
}

// TestRunCanaryIgnoresWebhookDeadline checks that the handler's webhook
// deadline does not reach the canary's runs. A push to the scripts folder or
// to the repository root selects every auto-graded assignment, and the runs are
// sequential, so a later run may start after the deadline has passed; it must
// then still run, bounded only by its own container timeout.
func TestRunCanaryIgnoresWebhookDeadline(t *testing.T) {
	runner := &contextRunner{}
	wh, course, ctx, _ := setupCanary(t, runner, fmt.Sprintf(canaryRunScriptFmt, 0))
	ctx, cancel := context.WithCancel(ctx)
	cancel() // the webhook deadline passed before the run started

	wh.runCanary(ctx, nil, course, qf.TestsRepo, canaryPush("lab1/lab1.go"))
	if !runner.ran {
		t.Fatal("runCanary() did not run the canary for lab1")
	}
	if runner.ctxErr != nil {
		t.Errorf("runCanary() ran the canary on a done context: %v", runner.ctxErr)
	}
}

// conflictRunner is a ci.Runner reporting that the job's container name is
// taken, which is how a run that is already in progress is reported.
type conflictRunner struct{}

func (conflictRunner) Run(context.Context, *ci.Job) (string, error) {
	return "", ci.ErrConflict
}

// TestRunCanarySkipsDuplicateRun checks that a canary whose run is already in
// progress is counted as skipped, and not as a healthy test environment.
func TestRunCanarySkipsDuplicateRun(t *testing.T) {
	wh, course, ctx, output := setupCanary(t, conflictRunner{}, fmt.Sprintf(canaryRunScriptFmt, 0))

	wh.runCanary(ctx, nil, course, qf.TestsRepo, canaryPush("lab1/lab1.go"))
	records := canaryRecords(t, output.String())

	summary := recordsWith(records, "test environment canary finished")
	if len(summary) != 1 {
		t.Fatalf("runCanary() logged %d summary records, want 1:\n%s", len(summary), output.String())
	}
	if summary[0].Assignments != 1 || summary[0].Skipped != 1 || summary[0].Problems != 0 {
		t.Errorf("runCanary() summary = %d assignments, %d problems, %d skipped; want 1, 0, 1",
			summary[0].Assignments, summary[0].Problems, summary[0].Skipped)
	}
	for _, msg := range []string{"test environment canary failed", "test environment canary passed"} {
		if n := len(recordsWith(records, msg)); n != 0 {
			t.Errorf("runCanary() logged %d %q records for a duplicate run, want 0:\n%s", n, msg, output.String())
		}
	}
}

// TestRunCanarySkipsStaleClone checks that the canary runs only when the local
// clone of the pushed repository is at the pushed commit. The assignments
// update reports success even when refreshing a clone failed, so the clone may
// still hold the skeleton code of an earlier push.
func TestRunCanarySkipsStaleClone(t *testing.T) {
	const skipMsg = "skipping test environment canary: local clone is not at the pushed commit"
	wh, course, ctx, output := setupCanary(t, &ci.Local{}, fmt.Sprintf(canaryRunScriptFmt, 0))
	head := initCanaryClone(t, filepath.Join(course.CloneDir(), qf.TestsRepo))

	// The pushed commit is not the one the clone is at.
	wh.runCanary(ctx, nil, course, qf.TestsRepo, canaryPush("lab1/lab1.go"))
	skips := recordsWith(canaryRecords(t, output.String()), skipMsg)
	if len(skips) != 1 {
		t.Fatalf("runCanary() logged %d %q records, want 1:\n%s", len(skips), skipMsg, output.String())
	}
	if skips[0].Level != slog.LevelWarn.String() {
		t.Errorf("runCanary() logged %q at level %s, want %s", skipMsg, skips[0].Level, slog.LevelWarn.String())
	}
	if skips[0].CloneHead != head {
		t.Errorf("runCanary() logged clone head %q, want %q", skips[0].CloneHead, head)
	}
	assertNoCanaryRun(t, output.String())

	// The clone is at the pushed commit, so the canary runs.
	output.Reset()
	push := canaryPush("lab1/lab1.go")
	push.HeadCommit.ID = new(head)
	wh.runCanary(ctx, nil, course, qf.TestsRepo, push)
	records := canaryRecords(t, output.String())
	if n := len(recordsWith(records, skipMsg)); n != 0 {
		t.Errorf("runCanary() logged %d %q records for the pushed commit, want 0:\n%s", n, skipMsg, output.String())
	}
	if n := len(recordsWith(records, "test environment canary passed")); n != 1 {
		t.Errorf("runCanary() logged %d passed records for lab1, want 1:\n%s", n, output.String())
	}
}
