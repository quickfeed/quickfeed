package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qtest"
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

// canaryTestOrg is the course organization used by TestRunCanary. Its
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

// canaryRecord is the subset of a JSON log record that the assertions below
// inspect; the enclosing scope's attributes are ignored.
type canaryRecord struct {
	Level       string `json:"level"`
	Message     string `json:"msg"`
	Assignment  string `json:"assignment"`
	Score       uint32 `json:"score"`
	Problem     string `json:"problem"`
	Assignments int    `json:"assignments"`
	Problems    int    `json:"problems"`
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
	// The messages that report an actual canary run for an assignment.
	runMessages := []string{
		"test environment canary failed",
		"test environment canary passed",
		"test environment canary could not run",
		"test environment canary finished",
	}
	tests := []struct {
		name      string
		points    int    // points awarded to the skeleton code, out of ten
		wantLevel string // level of the expected per-assignment record
		wantMsg   string // message of the expected per-assignment record
		wantScore uint32 // score the record must report
		wantProbs int    // problems the summary record must report
	}{
		{
			name:      "SkeletonPassesEveryTest",
			points:    10,
			wantLevel: slog.LevelWarn.String(),
			wantMsg:   "test environment canary failed",
			wantScore: 100,
			wantProbs: 1,
		},
		{
			name:      "SkeletonScoresZero",
			points:    0,
			wantLevel: slog.LevelInfo.String(),
			wantMsg:   "test environment canary passed",
			wantScore: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup := qtest.TestDB(t)
			defer cleanup()

			var output bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
			wh := NewGitHubWebHook(logger, db, nil, &ci.Local{}, "secret", stream.NewStreamServices(), nil)
			course := setupCanaryCourse(t, db, fmt.Sprintf(canaryRunScriptFmt, tt.points))
			// The scope handlePush establishes before calling runCanary.
			ctx, _ := qlog.WithCourse(qlog.NewContext(t.Context(), logger), course)

			// A push touching lab1 checks lab1, and nothing else. The SCM
			// client is nil: the skeleton run needs no clone.
			wh.runCanary(ctx, nil, course, canaryPush("lab1/lab1.go"))
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
			if (got[0].Problem != "") != (tt.wantProbs > 0) {
				t.Errorf("runCanary() logged problem %q, want problem: %t", got[0].Problem, tt.wantProbs > 0)
			}

			summary := recordsWith(records, "test environment canary finished")
			if len(summary) != 1 {
				t.Fatalf("runCanary() logged %d summary records, want 1:\n%s", len(summary), output.String())
			}
			if summary[0].Assignments != 1 || summary[0].Problems != tt.wantProbs {
				t.Errorf("runCanary() summary = %d assignments, %d problems; want 1 assignment, %d problems",
					summary[0].Assignments, summary[0].Problems, tt.wantProbs)
			}
			// The manually graded lab2 has no tests to check.
			for _, record := range records {
				if record.Assignment == "lab2" {
					t.Errorf("runCanary() logged %q for the manually graded lab2", record.Message)
				}
			}

			// A push touching only the manually graded lab2 runs nothing.
			output.Reset()
			wh.runCanary(ctx, nil, course, canaryPush("lab2/lab2.go"))
			for _, msg := range runMessages {
				if n := len(recordsWith(canaryRecords(t, output.String()), msg)); n != 0 {
					t.Errorf("runCanary() logged %d %q records for a lab2-only push, want 0:\n%s", n, msg, output.String())
				}
			}
		})
	}
}
