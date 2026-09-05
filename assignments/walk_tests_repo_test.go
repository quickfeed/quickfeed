package assignments

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

const testsFolder = "testdata/tests"

func TestWalkTestsRepository(t *testing.T) {
	// map of expected files in the testdata/tests folder, relative to that folder.
	// Note: run.sh is ignored by walkTestsRepository so they are not included here.
	wantFiles := map[string]struct{}{
		"scripts/Dockerfile":   {},
		"scripts/go.mod":       {},
		"scripts/go.sum":       {},
		"lab1/assignment.json": {},
		"lab1/tests.json":      {},
		"lab2/assignment.json": {},
		"lab2/tests.json":      {},
		"lab3/assignment.json": {},
		"lab4/assignment.json": {},
		"lab4/criteria.json":   {},
		"lab5/assignment.json": {},
		"lab5/criteria.json":   {},
	}
	files, err := walkTestsRepository(testsFolder)
	if err != nil {
		t.Fatal(err)
	}
	for filename := range files {
		if _, ok := wantFiles[filename]; !ok {
			t.Errorf("unexpected file %q in %s", filename, testsFolder)
		}
	}
	for wantFilename := range wantFiles {
		if _, ok := files[wantFilename]; !ok {
			t.Errorf("missing file %q in %s", wantFilename, testsFolder)
		}
	}
	if len(files) != len(wantFiles) {
		t.Errorf("expected %d files, got %d", len(wantFiles), len(files))
	}
}

func TestReadTestsRepositoryContent(t *testing.T) {
	wantDockerfile := "FROM golang:1.25-alpine\nRUN apk update && apk add --no-cache git bash build-base\nWORKDIR /quickfeed\nCOPY go.mod go.sum ./\nRUN go mod download\n"
	wantAssignments := []*qf.Assignment{
		{
			Name:       "lab1",
			CourseID:   1,
			Order:      1,
			ScoreLimit: 80,
			Deadline:   qtest.Timestamp(t, "2019-01-24T14:00:00"),
			ExpectedTests: []*qf.TestInfo{
				{TestName: "TestGitQuestionsAG", MaxScore: 10, Weight: 1},
				{TestName: "TestMissingSemesterQuestionsAG", MaxScore: 9, Weight: 1},
				{TestName: "TestShellQuestionsAG", MaxScore: 20, Weight: 1},
			},
		},
		{
			Name:       "lab2",
			CourseID:   1,
			Order:      2,
			ScoreLimit: 80,
			Deadline:   qtest.Timestamp(t, "2019-01-31T16:00:00"),
			ExpectedTests: []*qf.TestInfo{
				{TestName: "Test0Formatting", MaxScore: 1, Weight: 5},
				{TestName: "Test0Lint", MaxScore: 1, Weight: 5},
				{TestName: "Test0TODOItems", MaxScore: 1, Weight: 5},
				{TestName: "Test0VetCheck", MaxScore: 1, Weight: 5},
				{TestName: "TestGrpc_ProtoGeneration", MaxScore: 2, Weight: 20},
				{TestName: "TestGrpc_RequestSequence", MaxScore: 14, Weight: 50},
				{TestName: "TestGrpc_ServerRaceCondition", MaxScore: 1, Weight: 50},
				{TestName: "TestNetworkQuestions", MaxScore: 5, Weight: 1},
				{TestName: "TestWeb_Counter", MaxScore: 5, Weight: 10},
				{TestName: "TestWeb_FizzBuzz", MaxScore: 18, Weight: 30},
				{TestName: "TestWeb_NonExisting", MaxScore: 6, Weight: 10},
				{TestName: "TestWeb_Redirect", MaxScore: 4, Weight: 20},
				{TestName: "TestWeb_Root", MaxScore: 1, Weight: 10},
				{TestName: "TestWeb_ServerFull", MaxScore: 39, Weight: 20},
			},
		},
		{
			Name:       "lab3",
			CourseID:   1,
			Order:      3,
			ScoreLimit: 80,
			Deadline:   qtest.Timestamp(t, "2019-02-14T23:00:00"),
			IsGroupLab: true,
		},
		{
			Name:       "lab4",
			CourseID:   1,
			Order:      4,
			ScoreLimit: 80,
			Deadline:   qtest.Timestamp(t, "2019-03-15T16:00:00"),
			IsGroupLab: true,
			GradingBenchmarks: []*qf.GradingBenchmark{
				{
					Heading:  "Assignment 1",
					CourseID: 1,
					Criteria: []*qf.GradingCriterion{
						{
							CourseID:    1,
							Description: "Links work",
						},
						{
							CourseID:    1,
							Description: "Images are links, opening in a new tab",
						},
					},
				},
			},
		},
		{
			Name:       "lab5",
			CourseID:   1,
			Order:      5,
			ScoreLimit: 80,
			Deadline:   qtest.Timestamp(t, "2025-07-21T16:00:00"),
			IsGroupLab: true,
			GradingBenchmarks: []*qf.GradingBenchmark{
				{
					CourseID: 1,
					Criteria: []*qf.GradingCriterion{
						{
							CourseID: 1,
						},
					},
				},
			},
		},
	}

	gotAssignments, gotBuildContext, gotIssues, err := readTestsRepositoryContent(testsFolder, 1)
	if err != nil {
		t.Fatal(err)
	}
	if gotBuildContext[ci.Dockerfile] != wantDockerfile {
		t.Errorf("got Dockerfile %q, want %q", gotBuildContext[ci.Dockerfile], wantDockerfile)
	}
	if diff := cmp.Diff(wantAssignments, gotAssignments, protocmp.Transform()); diff != "" {
		t.Errorf("readTestsRepositoryContent() mismatch (-wantAssignments +gotAssignments):\n%s", diff)
	}
	// lab5's criteria.json has an empty heading and an empty description.
	// Labs 3-5 are auto-graded without a tests.json; criteria alone does not
	// make an assignment manually graded.
	wantIssues := []RepoIssue{
		{Assignment: "lab5", File: "lab5/criteria.json", Problem: "benchmark with empty heading"},
		{Assignment: "lab5", File: "lab5/criteria.json", Problem: `criterion with empty description in benchmark ""`},
		{Assignment: "lab3", File: "lab3/tests.json", Problem: "missing or empty tests.json: all submissions will score zero"},
		{Assignment: "lab4", File: "lab4/tests.json", Problem: "missing or empty tests.json: all submissions will score zero"},
		{Assignment: "lab5", File: "lab5/tests.json", Problem: "missing or empty tests.json: all submissions will score zero"},
	}
	if diff := cmp.Diff(wantIssues, gotIssues); diff != "" {
		t.Errorf("readTestsRepositoryContent() issue mismatch (-want +got):\n%s", diff)
	}
}

func TestBuildContextContainsModuleFiles(t *testing.T) {
	_, gotBuildContext, _, err := readTestsRepositoryContent(testsFolder, 1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that build context contains go.mod and go.sum files
	wantFiles := []string{ci.Dockerfile, "go.mod", "go.sum"}
	for _, filename := range wantFiles {
		if _, ok := gotBuildContext[filename]; !ok {
			t.Errorf("build context missing expected file %q", filename)
		}
	}

	// Verify go.mod content
	if !strings.Contains(gotBuildContext["go.mod"], "github.com/quickfeed/build-context") {
		t.Errorf("go.mod does not contain expected module name, got: %s", gotBuildContext["go.mod"])
	}

	// Verify go.sum content
	if !strings.Contains(gotBuildContext["go.sum"], "github.com/relab/container") {
		t.Errorf("go.sum does not contain expected dependency, got: %s", gotBuildContext["go.sum"])
	}

	// Verify Dockerfile contains COPY and go mod download instructions
	dockerfile := gotBuildContext[ci.Dockerfile]
	if !strings.Contains(dockerfile, "COPY go.mod go.sum ./") {
		t.Errorf("Dockerfile does not contain COPY instruction for module files")
	}
	if !strings.Contains(dockerfile, "go mod download") {
		t.Errorf("Dockerfile does not contain go mod download instruction")
	}
}

func TestReadTestsRepositoryContentBadContent(t *testing.T) {
	// Check that readTestsRepositoryContent reports bad content as issues
	// instead of aborting, and excludes the affected assignments.
	tests := []struct {
		name        string
		folder      string
		wantProblem string
	}{
		{name: "InvalidTypes", folder: "testdata/invalid-tests/invalid-types", wantProblem: "unmarshaling"},
		{name: "NegativeInteger", folder: "testdata/invalid-tests/negative-integer", wantProblem: "unmarshaling"},
		{name: "MissingAssignment1", folder: "testdata/invalid-tests/missing-assignment-json1", wantProblem: "missing"},
		{name: "MissingAssignment2", folder: "testdata/invalid-tests/missing-assignment-json2", wantProblem: "missing"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assignments, _, issues, err := readTestsRepositoryContent(tc.folder, 1)
			if err != nil {
				t.Fatal(err)
			}
			if len(assignments) != 0 {
				t.Errorf("len(assignments) = %d, want 0", len(assignments))
			}
			if len(issues) == 0 {
				t.Fatal("expected issues, got none")
			}
			if !slices.ContainsFunc(issues, func(issue RepoIssue) bool {
				return strings.Contains(issue.Problem, tc.wantProblem)
			}) {
				t.Errorf("issues %q do not mention %q", issues, tc.wantProblem)
			}
		})
	}
}

func TestReadTestsRepositoryContentIssues(t *testing.T) {
	testsDir := t.TempDir()

	const badTestsJSON = `[
		{"TestName": "TestA", "MaxScore": 10, "Weight": 1},
		{"TestName": "", "MaxScore": 5, "Weight": 1},
		{"TestName": "TestB", "MaxScore": 0, "Weight": 1},
		{"TestName": "TestC", "MaxScore": 5, "Weight": 0},
		{"TestName": "TestA", "MaxScore": 7, "Weight": 2}
	]`
	for _, c := range []struct {
		path, filename, content string
	}{
		{"lab1", "assignment.json", j1},
		{"lab1", "tests.json", badTestsJSON},
		{"lab2", "tests.json", testJson},              // no assignment.json for lab2
		{"lab3", "assignment.json", j3},               //
		{"lab3", "Dockerfile", "FROM golang"},         // Dockerfile outside scripts
		{"internal/pkg", "tests.json", "not json"},    // nested; must be ignored
		{"lab1/testdata", "criteria.json", "ignored"}, // nested; must be ignored
		{"scripts", "Dockerfile", df},
		{"", "criteria.json", "[]"}, // repository root; must be reported
	} {
		writeFile(t, testsDir, c.path, c.filename, c.content)
	}

	assignments, gotBuildContext, issues, err := readTestsRepositoryContent(testsDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	// lab1 and lab3 must survive; lab2 has no assignment.json
	if len(assignments) != 2 {
		t.Errorf("len(assignments) = %d, want 2", len(assignments))
	}
	// The invalid tests.json entries must be dropped, keeping only TestA
	wantTests := []*qf.TestInfo{{TestName: "TestA", MaxScore: 10, Weight: 1}}
	if diff := cmp.Diff(wantTests, assignments[0].GetExpectedTests(), protocmp.Transform()); diff != "" {
		t.Errorf("ExpectedTests mismatch (-want +got):\n%s", diff)
	}
	if gotBuildContext[ci.Dockerfile] != df {
		t.Errorf("got Dockerfile %q, want %q", gotBuildContext[ci.Dockerfile], df)
	}

	wantIssues := []RepoIssue{
		{File: "criteria.json", Problem: "file must be inside an assignment folder"},
		{Assignment: "lab1", File: "lab1/tests.json", Problem: "test entry with empty test name"},
		{Assignment: "lab1", File: "lab1/tests.json", Problem: `test "TestB" must have max score greater than 0`},
		{Assignment: "lab1", File: "lab1/tests.json", Problem: `test "TestC" must have weight greater than 0`},
		{Assignment: "lab1", File: "lab1/tests.json", Problem: `duplicate test name "TestA"`},
		{Assignment: "lab2", File: "lab2/tests.json", Problem: `missing "lab2/assignment.json"`},
		{Assignment: "lab3", File: "lab3/Dockerfile", Problem: "Dockerfile must be in the scripts folder"},
		{Assignment: "lab3", File: "lab3/tests.json", Problem: "missing or empty tests.json: all submissions will score zero"},
	}
	if diff := cmp.Diff(wantIssues, issues); diff != "" {
		t.Errorf("readTestsRepositoryContent() issue mismatch (-want +got):\n%s", diff)
	}
}

func TestProcessCriteriaFileNullEntries(t *testing.T) {
	const criteria = `[
		null,
		{"heading":"Valid benchmark","criteria":[null,{"description":"Valid criterion"}]}
	]`
	assignment := &qf.Assignment{}
	problems, err := processCriteriaFile([]byte(criteria), assignment, 1)
	if err != nil {
		t.Fatal(err)
	}
	wantProblems := []string{
		"null benchmark",
		`null criterion in benchmark "Valid benchmark"`,
	}
	if diff := cmp.Diff(wantProblems, problems); diff != "" {
		t.Errorf("processCriteriaFile() problem mismatch (-want +got):\n%s", diff)
	}
	wantBenchmarks := []*qf.GradingBenchmark{
		{
			CourseID: 1,
			Heading:  "Valid benchmark",
			Criteria: []*qf.GradingCriterion{
				{CourseID: 1, Description: "Valid criterion"},
			},
		},
	}
	if diff := cmp.Diff(wantBenchmarks, assignment.GetGradingBenchmarks(), protocmp.Transform()); diff != "" {
		t.Errorf("processCriteriaFile() benchmark mismatch (-want +got):\n%s", diff)
	}
}

func TestReadTestsRepositoryContentCollectsIssuesForBrokenAssignment(t *testing.T) {
	testsDir := t.TempDir()
	for _, filename := range []string{assignmentFile, criteriaFile, testsFile} {
		writeFile(t, testsDir, "lab1", filename, "{")
	}

	assignments, _, issues, err := readTestsRepositoryContent(testsDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(assignments) != 0 {
		t.Errorf("len(assignments) = %d, want 0", len(assignments))
	}
	wantFiles := []string{
		"lab1/assignment.json",
		"lab1/criteria.json",
		"lab1/tests.json",
	}
	if len(issues) != len(wantFiles) {
		t.Fatalf("len(issues) = %d, want %d: %v", len(issues), len(wantFiles), issues)
	}
	for i, issue := range issues {
		if issue.File != wantFiles[i] {
			t.Errorf("issues[%d].File = %q, want %q", i, issue.File, wantFiles[i])
		}
		if !strings.Contains(issue.Problem, "unmarshaling") {
			t.Errorf("issues[%d].Problem = %q, want unmarshaling error", i, issue.Problem)
		}
	}
}

func TestMissingTestsIssuesWithCriteria(t *testing.T) {
	criteria := []*qf.GradingBenchmark{{Heading: "Benchmark"}}
	assignments := map[string]*qf.Assignment{
		"auto":   {GradingBenchmarks: criteria},
		"manual": {Reviewers: 1, GradingBenchmarks: criteria},
	}
	want := []RepoIssue{{
		Assignment: "auto",
		File:       "auto/tests.json",
		Problem:    "missing or empty tests.json: all submissions will score zero",
	}}
	if diff := cmp.Diff(want, missingTestsIssues(assignments)); diff != "" {
		t.Errorf("missingTestsIssues() mismatch (-want +got):\n%s", diff)
	}
}
