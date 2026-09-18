package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/quickfeed/quickfeed/assignments"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
)

func TestCourseCode(t *testing.T) {
	tests := []struct {
		name string
		org  string
		want string
	}{
		{name: "TrailingYear", org: "dat320-2025", want: "DAT320"},
		{name: "TrailingYearUpperCase", org: "DAT320-2025", want: "DAT320"},
		{name: "QuickFeedTestCourse", org: "qf101-2022", want: "QF101"},
		{name: "NoYear", org: "my-course", want: "MY-COURSE"},
		{name: "NoYearNoDash", org: "dat320", want: "DAT320"},
		{name: "ThreeDigitSuffix", org: "dat320-202", want: "DAT320-202"},
		{name: "FiveDigitSuffix", org: "dat320-20255", want: "DAT320-20255"},
		{name: "YearInTheMiddle", org: "dat320-2025-spring", want: "DAT320-2025-SPRING"},
		{name: "OnlyYear", org: "2025", want: "2025"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := courseCode(tc.org); got != tc.want {
				t.Errorf("courseCode(%q) = %q, want %q", tc.org, got, tc.want)
			}
		})
	}
}

// TestRunUsageErrors checks that every usage error is reported as an error,
// without panicking, so that the tool exits non-zero.
func TestRunUsageErrors(t *testing.T) {
	t.Setenv("QUICKFEED_REPOSITORY_PATH", t.TempDir())
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "NoSubcommand", args: nil, want: "no subcommand given"},
		{name: "UnknownSubcommand", args: []string{"frobnicate"}, want: `unknown subcommand "frobnicate"`},
		{name: "CloneWithoutCourse", args: []string{"clone"}, want: "missing required flag -course"},
		{name: "RunWithoutCourse", args: []string{"run", "-lab", "lab1"}, want: "missing required flag -course"},
		{name: "RunWithoutLab", args: []string{"run", "-course", "dat320-2025"}, want: "missing required flag -lab"},
		{
			name: "RunWithoutSource",
			args: []string{"run", "-course", "dat320-2025", "-lab", "lab1"},
			want: "missing the code to test",
		},
		{
			name: "RunWithTwoSources",
			args: []string{"run", "-course", "dat320-2025", "-lab", "lab1", "-user", "meling", "-group", "meling-group"},
			want: "conflicting sources for the code to test",
		},
		{name: "CheckWithoutCourse", args: []string{"check"}, want: "missing required flag -course"},
		{name: "UnknownFlag", args: []string{"check", "-nosuchflag"}, want: "flag provided but not defined"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := run(tc.args, &stdout, &stderr)
			if err == nil {
				t.Fatalf("run(%q) error = nil, want %q", tc.args, tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("run(%q) error = %q, want it to contain %q", tc.args, err, tc.want)
			}
		})
	}
}

// TestRunHelp checks that asking for help is not an error, so that the tool
// exits zero.
func TestRunHelp(t *testing.T) {
	for _, args := range [][]string{
		{"-help"},
		{"help"},
		{"clone", "-help"},
		{"run", "-help"},
		{"check", "-help"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := run(args, &stdout, &stderr); err != nil {
				t.Fatalf("run(%q) error = %v, want nil", args, err)
			}
			out := stdout.String() + stderr.String()
			if !strings.Contains(out, "qcm") {
				t.Errorf("run(%q) printed %q, want usage text", args, out)
			}
		})
	}
}

// TestCheckMissingTestsRepository checks that a course whose repositories have
// not been cloned is reported, rather than dereferencing a missing directory.
func TestCheckMissingTestsRepository(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("QUICKFEED_REPOSITORY_PATH", dir)
	var stdout, stderr bytes.Buffer
	err := run([]string{"check", "-course", "dat320-2025", "-dir", dir}, &stdout, &stderr)
	if err == nil {
		t.Fatal("run(check) error = nil, want an error for the missing tests repository")
	}
	if !strings.Contains(err.Error(), "reading tests repository") {
		t.Errorf("run(check) error = %q, want it to name the tests repository", err)
	}
}

// assignmentJSON is a minimal assignment.json for an auto-graded assignment.
const assignmentJSON = `{"order": 1, "deadline": "01-12-2025T23:59"}`

// TestCheckUnknownLab checks that a misspelled -lab is reported as an error
// naming the course's assignments. Without the check, every test run would be
// skipped and the command would report success without checking anything.
func TestCheckUnknownLab(t *testing.T) {
	const org = "dat320-2025"
	dir := writeTestsRepo(t, org, map[string]string{
		filepath.Join("lab1", "assignment.json"): assignmentJSON,
		filepath.Join(scriptsDir, runScriptFile): "#image/dat320\necho hello\n",
	})
	var stdout, stderr bytes.Buffer
	err := run([]string{"check", "-course", org, "-dir", dir, "-lab", "lab9"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("run(check -lab lab9) error = nil, want an error for the unknown assignment")
	}
	for _, want := range []string{`assignment "lab9" not found`, "lab1"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("run(check -lab lab9) error = %q, want it to contain %q", err, want)
		}
	}
	if out := stdout.String(); out != "" {
		t.Errorf("run(check -lab lab9) printed %q, want no checks to have run", out)
	}
}

// TestRunEmptyDockerfile checks that a course whose scripts/Dockerfile is
// present but empty is reported, rather than silently treated as a course
// without a Dockerfile, whose image would then never be built.
func TestRunEmptyDockerfile(t *testing.T) {
	const org = "dat320-2025"
	dir := writeTestsRepo(t, org, map[string]string{
		filepath.Join("lab1", "assignment.json"): assignmentJSON,
		filepath.Join(scriptsDir, runScriptFile): "#image/dat320\necho hello\n",
		filepath.Join(scriptsDir, ci.Dockerfile): "\n\n",
	})
	var stdout, stderr bytes.Buffer
	err := run([]string{"run", "-course", org, "-dir", dir, "-lab", "lab1", "-submission", t.TempDir()}, &stdout, &stderr)
	if err == nil {
		t.Fatal("run(run) error = nil, want an error for the empty Dockerfile")
	}
	if want := "scripts/Dockerfile is empty"; !strings.Contains(err.Error(), want) {
		t.Errorf("run(run) error = %q, want it to contain %q", err, want)
	}
}

// TestValidateSetsRepositoryPath pins that -dir is what the ci package resolves
// a course's clone directory from; the flag was previously declared but unused.
func TestValidateSetsRepositoryPath(t *testing.T) {
	t.Setenv("QUICKFEED_REPOSITORY_PATH", "unset")
	dir := t.TempDir()
	c := commonFlags{org: "dat320-2025", dir: dir}
	if err := c.validate(); err != nil {
		t.Fatal(err)
	}
	if got, want := os.Getenv("QUICKFEED_REPOSITORY_PATH"), dir; got != want {
		t.Errorf("validate() set QUICKFEED_REPOSITORY_PATH to %q, want %q", got, want)
	}
	if got, want := c.course().CloneDir(), filepath.Join(dir, "dat320-2025"); got != want {
		t.Errorf("course().CloneDir() = %q, want %q", got, want)
	}
	if got, want := c.testsDir(), filepath.Join(dir, "dat320-2025", qf.TestsRepo); got != want {
		t.Errorf("testsDir() = %q, want %q", got, want)
	}
	if got, want := c.code, "DAT320"; got != want {
		t.Errorf("validate() derived code %q, want %q", got, want)
	}
}

func TestParseRunScripts(t *testing.T) {
	testsDir := t.TempDir()
	writeRepoFile(t, testsDir, filepath.Join(scriptsDir, runScriptFile), "#image/dat320\necho hello\n")
	writeRepoFile(t, testsDir, filepath.Join("lab1", runScriptFile), "#image/golang:1.26\necho lab1\n")
	writeRepoFile(t, testsDir, filepath.Join("lab2", runScriptFile), "#image/dat320\n\n\n")
	parsed := []*qf.Assignment{{Name: "lab1"}, {Name: "lab2"}, {Name: "lab3"}}

	scripts, err := parseRunScripts(testsDir, parsed)
	if err != nil {
		t.Fatal(err)
	}
	if !scripts.hasDefault {
		t.Error("parseRunScripts() hasDefault = false, want true")
	}
	if got, want := scripts.images["lab1/run.sh"], "golang:1.26"; got != want {
		t.Errorf("parseRunScripts() lab1 image = %q, want %q", got, want)
	}
	if got, want := scripts.missing, []string{"lab3"}; len(got) != 1 || got[0] != want[0] {
		t.Errorf("parseRunScripts() missing = %q, want %q", got, want)
	}
	if len(scripts.failures) != 1 || !strings.Contains(scripts.failures[0], "lab2/run.sh") {
		t.Errorf("parseRunScripts() failures = %q, want the lab2 script to fail", scripts.failures)
	}
	if got, want := scripts.usersOf("dat320"), []string{"scripts/run.sh"}; len(got) != 1 || got[0] != want[0] {
		t.Errorf("usersOf(dat320) = %q, want %q", got, want)
	}
}

func TestCheckRunScripts(t *testing.T) {
	tests := []struct {
		name    string
		scripts *runScripts
		want    string
	}{
		{
			name:    "DefaultScriptCoversEveryAssignment",
			scripts: &runScripts{images: map[string]string{"scripts/run.sh": "dat320"}, hasDefault: true, missing: []string{"lab1"}},
			want:    pass,
		},
		{
			name:    "PerAssignmentScripts",
			scripts: &runScripts{images: map[string]string{"lab1/run.sh": "dat320"}},
			want:    pass,
		},
		{
			name:    "NoScriptForAssignment",
			scripts: &runScripts{images: map[string]string{}, missing: []string{"lab2"}},
			want:    fail,
		},
		{
			name:    "UnparsableScript",
			scripts: &runScripts{images: map[string]string{}, hasDefault: true, failures: []string{"scripts/run.sh: empty run script"}},
			want:    fail,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := checkRunScripts(tc.scripts); got.result != tc.want {
				t.Errorf("checkRunScripts() = %+v, want %s", got, tc.want)
			}
		})
	}
}

func TestCheckDockerfile(t *testing.T) {
	course := &qf.Course{Code: "DAT320", ScmOrganizationName: "dat320-2025"}
	// A course without a Dockerfile whose run scripts name prebuilt images.
	ck := newChecker(io.Discard)
	prebuilt := &runScripts{images: map[string]string{"scripts/run.sh": "golang:1.26"}}
	got := checkDockerfile(t.Context(), nil, course, nil, prebuilt, true, ck)
	if got.result != skip {
		t.Errorf("checkDockerfile(no Dockerfile, prebuilt image) = %+v, want %s", got, skip)
	}
	// A course whose run script names the course image, but has no Dockerfile.
	own := &runScripts{images: map[string]string{"scripts/run.sh": "dat320"}}
	got = checkDockerfile(t.Context(), nil, course, nil, own, true, ck)
	if got.result != fail {
		t.Errorf("checkDockerfile(no Dockerfile, course image) = %+v, want %s", got, fail)
	}
	if !strings.Contains(strings.Join(got.details, "\n"), "dat320") {
		t.Errorf("checkDockerfile() details = %q, want it to name the course image", got.details)
	}
	// A Dockerfile that is not built because the user said not to.
	buildContext := map[string]string{ci.Dockerfile: "FROM golang:1.26\n"}
	got = checkDockerfile(t.Context(), nil, course, buildContext, own, false, ck)
	if got.result != skip {
		t.Errorf("checkDockerfile(-build=false) = %+v, want %s", got, skip)
	}
	// A Dockerfile that is present but empty is a mistake, not an absent one.
	empty := map[string]string{ci.Dockerfile: "\n  \n"}
	got = checkDockerfile(t.Context(), nil, course, empty, prebuilt, true, ck)
	if got.result != fail {
		t.Errorf("checkDockerfile(empty Dockerfile) = %+v, want %s", got, fail)
	}
	if !strings.Contains(strings.Join(got.details, "\n"), "scripts/Dockerfile is empty") {
		t.Errorf("checkDockerfile() details = %q, want it to name the empty Dockerfile", got.details)
	}
}

func TestCourseDockerfile(t *testing.T) {
	tests := []struct {
		name         string
		buildContext map[string]string
		want         string
		wantErr      bool
	}{
		{name: "NoDockerfile", buildContext: map[string]string{"go.mod": "module example\n"}},
		{name: "NoBuildContext", buildContext: nil},
		{name: "Dockerfile", buildContext: map[string]string{ci.Dockerfile: "FROM golang:1.26\n"}, want: "FROM golang:1.26\n"},
		{name: "EmptyDockerfile", buildContext: map[string]string{ci.Dockerfile: ""}, wantErr: true},
		{name: "BlankDockerfile", buildContext: map[string]string{ci.Dockerfile: "\n \t\n"}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := courseDockerfile(tc.buildContext)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("courseDockerfile(%q) error = nil, want an error", tc.buildContext)
				}
				return
			}
			if err != nil {
				t.Fatalf("courseDockerfile(%q) error = %v, want nil", tc.buildContext, err)
			}
			if got != tc.want {
				t.Errorf("courseDockerfile(%q) = %q, want %q", tc.buildContext, got, tc.want)
			}
		})
	}
}

// TestCheckSkeletonManuallyGraded checks that a manually graded assignment is
// skipped rather than run: it has no tests to score the handout code with.
func TestCheckSkeletonManuallyGraded(t *testing.T) {
	course := &qf.Course{Code: "DAT320", ScmOrganizationName: "dat320-2025"}
	parsed := []*qf.Assignment{{Name: "lab1", Reviewers: 1}}
	ck := newChecker(io.Discard)
	checkSkeleton(t.Context(), nil, course, parsed, "lab1", 0, ck)
	got := ck.results
	if len(got) != 1 || got[0].result != skip {
		t.Fatalf("checkSkeleton(manually graded) = %+v, want a single %s", got, skip)
	}
	if !strings.Contains(strings.Join(got[0].details, "\n"), "not an auto-graded assignment") {
		t.Errorf("checkSkeleton() details = %q, want it to say why the run was skipped", got[0].details)
	}
}

func TestCheckContent(t *testing.T) {
	dir := t.TempDir()
	c := commonFlags{org: "dat320-2025", dir: dir}
	if err := c.validate(); err != nil {
		t.Fatal(err)
	}
	// Neither repository exists, so the comparison is skipped with a note.
	got := checkContent(&c, nil, nil)
	if got.result != pass {
		t.Errorf("checkContent(no issues) = %+v, want %s", got, pass)
	}
	if !strings.Contains(strings.Join(got.details, "\n"), "skipped the cross-repository comparison") {
		t.Errorf("checkContent() details = %q, want the skipped comparison note", got.details)
	}
	issues := []assignments.RepoIssue{{Assignment: "lab1", File: "lab1/tests.json", Problem: "duplicate test name"}}
	got = checkContent(&c, nil, issues)
	if got.result != fail {
		t.Errorf("checkContent(issues) = %+v, want %s", got, fail)
	}
	if !strings.Contains(strings.Join(got.details, "\n"), "lab1/tests.json: duplicate test name") {
		t.Errorf("checkContent() details = %q, want it to list the issue", got.details)
	}
}

func TestAutoGraded(t *testing.T) {
	parsed := []*qf.Assignment{
		{Name: "lab1"},
		{Name: "lab2", Reviewers: 1},
		{Name: "lab3"},
	}
	if got := names(autoGraded(parsed, "")); len(got) != 2 || got[0] != "lab1" || got[1] != "lab3" {
		t.Errorf("autoGraded(all) = %q, want the auto-graded assignments", got)
	}
	if got := names(autoGraded(parsed, "lab3")); len(got) != 1 || got[0] != "lab3" {
		t.Errorf("autoGraded(lab3) = %q, want [lab3]", got)
	}
	if got := autoGraded(parsed, "lab2"); len(got) != 0 {
		t.Errorf("autoGraded(lab2) = %+v, want nothing for a manually graded assignment", got)
	}
}

func TestSkeletonResult(t *testing.T) {
	healthy := &score.Results{Scores: []*score.Score{
		{TestName: "TestLint", Score: 1, MaxScore: 1, Weight: 1},
		{TestName: "TestWork", Score: 0, MaxScore: 100, Weight: 30},
	}}
	got := skeletonResult("skeleton lab1", healthy, ci.MaxSkeletonScore)
	if got.result != pass {
		t.Errorf("skeletonResult(healthy) = %+v, want %s", got, pass)
	}
	// The test that awards the skeleton its few percent is still listed.
	if !strings.Contains(strings.Join(got.details, "\n"), "TestLint (1/1)") {
		t.Errorf("skeletonResult(healthy) details = %q, want it to list TestLint", got.details)
	}
	leaked := &score.Results{Scores: []*score.Score{
		{TestName: "TestA", Score: 10, MaxScore: 10, Weight: 1},
		{TestName: "TestB", Score: 0, MaxScore: 10, Weight: 1},
	}}
	got = skeletonResult("skeleton lab1", leaked, ci.MaxSkeletonScore)
	if got.result != fail {
		t.Errorf("skeletonResult(50%%) = %+v, want %s", got, fail)
	}
	for _, want := range []string{"scored 50%", "TestA (10/10)"} {
		if !strings.Contains(strings.Join(got.details, "\n"), want) {
			t.Errorf("skeletonResult(50%%) details = %q, want it to contain %q", got.details, want)
		}
	}
	failed := &score.Results{BuildInfo: &score.BuildInfo{Status: score.RunStatus_BUILD_FAILURE}}
	if got := skeletonResult("skeleton lab1", failed, ci.MaxSkeletonScore); got.result != fail || !strings.Contains(strings.Join(got.details, "\n"), "BUILD_FAILURE") {
		t.Errorf("skeletonResult(build failure) = %+v, want %s naming the status", got, fail)
	}
}

func TestSolutionResult(t *testing.T) {
	full := &score.Results{Scores: []*score.Score{{TestName: "TestA", Score: 10, MaxScore: 10, Weight: 1}}}
	if got := solutionResult("solution lab1", full); got.result != pass {
		t.Errorf("solutionResult(100%%) = %+v, want %s", got, pass)
	}
	partial := &score.Results{Scores: []*score.Score{
		{TestName: "TestA", Score: 10, MaxScore: 10, Weight: 1},
		{TestName: "TestB", Score: 4, MaxScore: 10, Weight: 1},
	}}
	got := solutionResult("solution lab1", partial)
	if got.result != fail {
		t.Errorf("solutionResult(70%%) = %+v, want %s", got, fail)
	}
	if !strings.Contains(strings.Join(got.details, "\n"), "TestB (4/10)") {
		t.Errorf("solutionResult() details = %q, want it to list TestB", got.details)
	}
	failed := &score.Results{BuildInfo: &score.BuildInfo{Status: score.RunStatus_TIMEOUT}}
	if got := solutionResult("solution lab1", failed); got.result != fail || !strings.Contains(strings.Join(got.details, "\n"), "TIMEOUT") {
		t.Errorf("solutionResult(timeout) = %+v, want %s naming the status", got, fail)
	}
}

func TestPrintChecks(t *testing.T) {
	var buf bytes.Buffer
	err := printChecks(&buf, []checkResult{
		{name: "run scripts", result: pass, details: []string{"parsed scripts/run.sh"}},
		{name: "dockerfile", result: fail, details: []string{"build failed\nline two\n"}},
		{name: "skeleton lab1", result: fail, details: []string{
			"skeleton code scored 20%, expected at most 5%",
			"tests passing on the skeleton code:",
			"TestA (1/1)",
			"TestB (9/10)",
		}},
	})
	if err == nil {
		t.Fatal("printChecks() error = nil, want an error for the failed checks")
	}
	if !strings.Contains(err.Error(), "2 of 3 checks failed") {
		t.Errorf("printChecks() error = %q, want the number of failed checks", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	// Header, three checks, and four continuation lines.
	if len(lines) != 8 {
		t.Fatalf("printChecks() printed %d lines, want 8:\n%s", len(lines), buf.String())
	}
	if !strings.HasPrefix(lines[0], "CHECK") {
		t.Errorf("printChecks() header = %q, want it to start with CHECK", lines[0])
	}
	// Every detail is on a line of its own, and continuation lines are indented
	// to the DETAILS column, where the check's first detail starts.
	detailsCol := strings.Index(lines[1], "parsed scripts/run.sh")
	wantContinuations := []string{"line two", "tests passing on the skeleton code:", "TestA (1/1)", "TestB (9/10)"}
	for _, want := range wantContinuations {
		found := false
		for _, line := range lines {
			if strings.TrimSpace(line) == want {
				found = true
				if col := strings.Index(line, want); col != detailsCol {
					t.Errorf("continuation %q starts at column %d, want the DETAILS column %d", want, col, detailsCol)
				}
			}
		}
		if !found {
			t.Errorf("printChecks() has no line holding only %q:\n%s", want, buf.String())
		}
	}
	if !strings.Contains(lines[2], "build failed") || strings.Contains(lines[2], "line two") {
		t.Errorf("printChecks() dockerfile row = %q, want only the first detail line", lines[2])
	}
}

func TestPrintResults(t *testing.T) {
	var buf bytes.Buffer
	printResults(&buf, &score.Results{
		BuildInfo: &score.BuildInfo{BuildLog: "compiling"},
		Scores: []*score.Score{
			{TestName: "TestA", Score: 5, MaxScore: 10, Weight: 1},
			{TestName: "TestB", Score: 10, MaxScore: 10, Weight: 1},
		},
	})
	out := buf.String()
	for _, want := range []string{"compiling", "TEST", "TestA", "TestB", "Total: 75%"} {
		if !strings.Contains(out, want) {
			t.Errorf("printResults() = %q, want it to contain %q", out, want)
		}
	}
}

func TestFindAssignment(t *testing.T) {
	parsed := []*qf.Assignment{{Name: "lab1"}, {Name: "lab2"}}
	if got, err := findAssignment(parsed, "lab2"); err != nil || got.GetName() != "lab2" {
		t.Errorf("findAssignment(lab2) = %v, %v, want lab2", got, err)
	}
	_, err := findAssignment(parsed, "lab9")
	if err == nil {
		t.Fatal("findAssignment(lab9) error = nil, want an error")
	}
	if !strings.Contains(err.Error(), "lab1, lab2") {
		t.Errorf("findAssignment(lab9) error = %q, want it to list the available assignments", err)
	}
	if _, err := findAssignment(nil, "lab1"); err == nil || !strings.Contains(err.Error(), "no assignments") {
		t.Errorf("findAssignment(no assignments) error = %v, want it to say so", err)
	}
}

func names(as []*qf.Assignment) []string {
	out := make([]string, len(as))
	for i, a := range as {
		out[i] = a.GetName()
	}
	return out
}

func writeRepoFile(t *testing.T, testsDir, rel, content string) {
	t.Helper()
	path := filepath.Join(testsDir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// writeTestsRepo creates the tests repository of the given course organization,
// holding the given repository-relative files, and returns the root directory
// to pass as -dir. The repository path in the environment is pointed at the
// same directory, as the commands themselves do, so that the ci package
// resolves the course's clone directory to it.
func writeTestsRepo(t *testing.T, org string, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	testsDir := filepath.Join(dir, org, qf.TestsRepo)
	for rel, content := range files {
		writeRepoFile(t, testsDir, rel, content)
	}
	t.Setenv("QUICKFEED_REPOSITORY_PATH", dir)
	return dir
}
