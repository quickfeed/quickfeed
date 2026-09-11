package main

import (
	"bytes"
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
	writeScript(t, testsDir, filepath.Join(scriptsDir, runScriptFile), "#image/dat320\necho hello\n")
	writeScript(t, testsDir, filepath.Join("lab1", runScriptFile), "#image/golang:1.26\necho lab1\n")
	writeScript(t, testsDir, filepath.Join("lab2", runScriptFile), "#image/dat320\n\n\n")
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
	prebuilt := &runScripts{images: map[string]string{"scripts/run.sh": "golang:1.26"}}
	got := checkDockerfile(t.Context(), nil, course, nil, prebuilt, true)
	if got.result != skip {
		t.Errorf("checkDockerfile(no Dockerfile, prebuilt image) = %+v, want %s", got, skip)
	}
	// A course whose run script names the course image, but has no Dockerfile.
	own := &runScripts{images: map[string]string{"scripts/run.sh": "dat320"}}
	got = checkDockerfile(t.Context(), nil, course, nil, own, true)
	if got.result != fail {
		t.Errorf("checkDockerfile(no Dockerfile, course image) = %+v, want %s", got, fail)
	}
	if !strings.Contains(got.details, "dat320") {
		t.Errorf("checkDockerfile() details = %q, want it to name the course image", got.details)
	}
	// A Dockerfile that is not built because the user said not to.
	buildContext := map[string]string{ci.Dockerfile: "FROM golang:1.26\n"}
	got = checkDockerfile(t.Context(), nil, course, buildContext, own, false)
	if got.result != skip {
		t.Errorf("checkDockerfile(-build=false) = %+v, want %s", got, skip)
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
	if !strings.Contains(got.details, "skipped the cross-repository comparison") {
		t.Errorf("checkContent() details = %q, want the skipped comparison note", got.details)
	}
	issues := []assignments.RepoIssue{{Assignment: "lab1", File: "lab1/tests.json", Problem: "duplicate test name"}}
	got = checkContent(&c, nil, issues)
	if got.result != fail {
		t.Errorf("checkContent(issues) = %+v, want %s", got, fail)
	}
	if !strings.Contains(got.details, "lab1/tests.json: duplicate test name") {
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
	if !strings.Contains(got.details, "TestB (4/10)") {
		t.Errorf("solutionResult() details = %q, want it to list TestB", got.details)
	}
	failed := &score.Results{BuildInfo: &score.BuildInfo{Status: score.RunStatus_TIMEOUT}}
	if got := solutionResult("solution lab1", failed); got.result != fail || !strings.Contains(got.details, "TIMEOUT") {
		t.Errorf("solutionResult(timeout) = %+v, want %s naming the status", got, fail)
	}
}

func TestPrintChecks(t *testing.T) {
	var buf bytes.Buffer
	err := printChecks(&buf, []checkResult{
		{name: "run scripts", result: pass, details: "parsed scripts/run.sh"},
		{name: "dockerfile", result: fail, details: "build failed\nline two\n"},
	})
	if err == nil {
		t.Fatal("printChecks() error = nil, want an error for the failed check")
	}
	if !strings.Contains(err.Error(), "1 of 2 checks failed") {
		t.Errorf("printChecks() error = %q, want the number of failed checks", err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "CHECK") {
		t.Errorf("printChecks() = %q, want a table header", out)
	}
	if !strings.Contains(out, "build failed | line two") {
		t.Errorf("printChecks() = %q, want the multi-line details folded onto one line", out)
	}
	if strings.Count(out, "\n") != 3 {
		t.Errorf("printChecks() printed %d lines, want 3", strings.Count(out, "\n"))
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

func writeScript(t *testing.T, testsDir, rel, content string) {
	t.Helper()
	path := filepath.Join(testsDir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
