package assignments

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestParseWithInvalidDir(t *testing.T) {
	const dir = "invalid/dir"
	_, _, _, err := readTestsRepositoryContent(dir, 0)
	if err == nil {
		t.Errorf("want no such file or directory error, got nil")
	}
}

const (
	y1 = `order: 1
name: "For loops"
deadline: "27-08-2017 12:00"
autoapprove: false
`
	y2 = `order: 2
name: "Nested loops"
deadline: "27-08-2018 12:00"
autoapprove: false
`
	y3 = `order: 3
name: "Nested loops"
deadline: "27-08-2018 12:00"
autoapprove: false
`
	yOldAssignmentIDField = `assignmentid: 3
name: "Big salary"
deadline: "27-08-2019 12:00"
autoapprove: false
`
	yUnknownFields = `order: 1
subject: "Go Programming for Fun and Profit"
name: "For loops"
deadline: "27-08-2017 12:00"
grading: "Pass/Fail"
expected_effort: "10 hours"
autoapprove: false
`

	script   = `Default script`
	script1  = `Script for Lab1`
	script2  = `Script for updating tests`
	df       = `A dockerfile in training`
	criteria = `
	[
		{
			"heading": "First benchmark",
			"criteria": [
				{
					"description": "Test 1",
					"points": 5
				},
				{
					"description": "Test 2",
					"points": 10
				}
			]
		},
		{
			"heading": "Second benchmark",
			"criteria": [
				{
					"description": "Test 3",
					"points": 5
				}
			]
		}
	]`
)

func writeFile(t *testing.T, testsDir, path, filename, content string) {
	t.Helper()
	dir := filepath.Join(testsDir, path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestParse(t *testing.T) {
	testsDir := t.TempDir()

	for _, c := range []struct {
		path, filename, content string
	}{
		{"lab1", "assignment.yaml", y1},
		{"lab2", "assignment.yaml", y2},
		{"scripts", runScript, script},
		{"scripts", updateTestsScript, script2},
		{"lab1", runScript, script1},
		{"scripts", "Dockerfile", df},
		{"lab2", "criteria.json", criteria},
	} {
		writeFile(t, testsDir, c.path, c.filename, c.content)
	}

	wantCriteria := []*qf.GradingBenchmark{
		{
			Heading: "First benchmark",
			Criteria: []*qf.GradingCriterion{
				{
					Description: "Test 1",
					Points:      5,
				},
				{
					Description: "Test 2",
					Points:      10,
				},
			},
		},
		{
			Heading: "Second benchmark",
			Criteria: []*qf.GradingCriterion{
				{
					Description: "Test 3",
					Points:      5,
				},
			},
		},
	}

	// We expect assignment names to be set based on
	// assignment folder names.
	wantAssignment1 := &qf.Assignment{
		Name:        "lab1",
		Deadline:    qtest.Timestamp(t, "2017-08-27T12:00:00"),
		AutoApprove: false,
		Order:       1,
		ScoreLimit:  80,
	}

	wantAssignment2 := &qf.Assignment{
		Name:              "lab2",
		Deadline:          qtest.Timestamp(t, "2018-08-27T12:00:00"),
		AutoApprove:       false,
		Order:             2,
		ScoreLimit:        80,
		GradingBenchmarks: wantCriteria,
	}

	assignments, dockerfile, updateTestsScript, err := readTestsRepositoryContent(testsDir, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(assignments) != 2 {
		t.Errorf("len(assignments) = %d, want %d", len(assignments), 2)
	}
	if dockerfile != df {
		t.Errorf("Incorrect dockerfile\n Want: %s\n Got: %s\n", df, dockerfile)
	}
	if updateTestsScript != script2 {
		t.Errorf("Incorrect updateTestsScript\n Want: %s\n Got: %s\n", script2, updateTestsScript)
	}
	if diff := cmp.Diff(assignments[0], wantAssignment1, protocmp.Transform()); diff != "" {
		t.Errorf("readTestsRepositoryContent() mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(assignments[1], wantAssignment2, protocmp.Transform()); diff != "" {
		t.Errorf("readTestsRepositoryContent() mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(assignments[1].GradingBenchmarks, wantCriteria, protocmp.Transform()); diff != "" {
		t.Errorf("readTestsRepositoryContent() mismatch when parsing criteria (-want +got):\n%s", diff)
	}
}

func TestParseOldAssignmentIDField(t *testing.T) {
	testsDir := t.TempDir()

	// test for use of assignmentid: 3 instead of order
	for _, c := range []struct {
		path, filename, content string
	}{
		{"lab3", "assignment.yaml", yOldAssignmentIDField},
	} {
		writeFile(t, testsDir, c.path, c.filename, c.content)
	}
	_, _, _, err := readTestsRepositoryContent(testsDir, 0)
	if err == nil {
		t.Fatal("want error: 'assignment order must be greater than 0', got nil")
	}
}

func TestParseOneBadAssignmentAmongCorrectOnes(t *testing.T) {
	testsDir := t.TempDir()

	// lab1 and lab2 are correct
	// lab3 contains an old assignmentid field
	for _, c := range []struct {
		path, filename, content string
	}{
		{"lab1", "assignment.yml", y1},
		{"lab2", "assignment.yml", y2},
		{"lab3", "assignment.yml", yOldAssignmentIDField},
	} {
		writeFile(t, testsDir, c.path, c.filename, c.content)
	}

	// Since lab3 contains an old assignmentid field, this will return an error
	_, _, _, err := readTestsRepositoryContent(testsDir, 0)
	if err == nil {
		t.Fatal("want error: 'assignment order must be greater than 0', got nil")
	}
}

func TestParseUnknownFields(t *testing.T) {
	testsDir := t.TempDir()

	for _, c := range []struct {
		path, filename, content string
	}{
		{"lab1", "assignment.yaml", yUnknownFields},
	} {
		writeFile(t, testsDir, c.path, c.filename, c.content)
	}

	// We expect assignment names to be set based on
	// assignment folder names.
	wantAssignment1 := &qf.Assignment{
		Name:        "lab1",
		Deadline:    qtest.Timestamp(t, "2017-08-27T12:00:00"),
		AutoApprove: false,
		Order:       1,
		ScoreLimit:  80,
	}

	assignments, _, _, err := readTestsRepositoryContent(testsDir, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(assignments) != 1 {
		t.Errorf("len(assignments) = %d, want %d", len(assignments), 1)
	}
	if diff := cmp.Diff(assignments[0], wantAssignment1, protocmp.Transform()); diff != "" {
		t.Errorf("readTestsRepositoryContent() mismatch (-want +got):\n%s", diff)
	}
}

func TestParseAndSaveAssignment(t *testing.T) {
	testsDir := t.TempDir()

	for _, c := range []struct {
		path, filename, content string
	}{
		{"lab1", "assignment.yml", y1},
		{"lab2", "assignment.yml", y2},
	} {
		writeFile(t, testsDir, c.path, c.filename, c.content)
	}
	course := &qf.Course{
		ID:   1,
		Name: "Test course",
	}

	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "admin", Login: "admin"})
	qtest.CreateCourse(t, db, admin, course)

	assignments, _, _, err := readTestsRepositoryContent(testsDir, course.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(assignments) != 2 {
		t.Errorf("len(assignments) = %d, want %d", len(assignments), 1)
	}

	// Save the assignments to the database
	if err := db.UpdateAssignments(assignments); err != nil {
		t.Fatal(err)
	}

	// Read the assignments from the database
	gotAssignments, err := db.GetAssignmentsByCourse(course.ID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(assignments, gotAssignments, protocmp.Transform()); diff != "" {
		t.Errorf("readTestsRepositoryContent() mismatch (-want +got):\n%s", diff)
	}

	// Add a new assignment to the list of assignments we expect to get from the database
	writeFile(t, testsDir, "lab3", "assignment.yml", y3)

	// Parse the new assignment
	newAssignments, _, _, err := readTestsRepositoryContent(testsDir, course.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Should be 3 assignments in the database now
	if len(newAssignments) != 3 {
		t.Errorf("len(assignments) = %d, want %d", len(newAssignments), 3)
	}

	// Save the new assignments to the database
	if err := db.UpdateAssignments(newAssignments); err != nil {
		t.Fatal(err)
	}

	// Read the assignments from the database
	gotNewAssignments, err := db.GetAssignmentsByCourse(course.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Should be 3 assignments now
	if len(gotNewAssignments) != 3 {
		t.Errorf("len(assignments) = %d, want %d", len(gotNewAssignments), 3)
	}

	// Check that the new assignments are the same as the ones we parsed
	if diff := cmp.Diff(newAssignments, gotNewAssignments, protocmp.Transform()); diff != "" {
		t.Errorf("readTestsRepositoryContent() mismatch (-want +got):\n%s", diff)
	}
}

func TestFixDeadline(t *testing.T) {
	deadlineTests := []struct {
		in, want string
	}{
		{"2020-01-23T18:00:20", "2020-01-23T18:00:20"},
		{"2020-01-23 18:00:20", "2020-01-23T18:00:20"},
		{"2020-01-23T18:00", "2020-01-23T18:00:00"},
		{"2020-01-23 18:00", "2020-01-23T18:00:00"},
		{"2020-01-23T1800", "2020-01-23T18:00:00"},
		{"2020-01-23 1800", "2020-01-23T18:00:00"},
		{"2020-01-23T18", "2020-01-23T18:00:00"},
		{"2020-01-23 18", "2020-01-23T18:00:00"},
		{"2020-01-23 6pm", "2020-01-23T18:00:00"},
		{"2020-01-23 6am", "2020-01-23T06:00:00"},
		//
		{"2020-1-23T18:00:20", "2020-01-23T18:00:20"},
		{"2020-1-23 18:00:20", "2020-01-23T18:00:20"},
		{"2020-1-23T18:00", "2020-01-23T18:00:00"},
		{"2020-1-23 18:00", "2020-01-23T18:00:00"},
		{"2020-1-23T1800", "2020-01-23T18:00:00"},
		{"2020-1-23 1800", "2020-01-23T18:00:00"},
		{"2020-1-23T18", "2020-01-23T18:00:00"},
		{"2020-1-23 18", "2020-01-23T18:00:00"},
		{"2020-1-23 6pm", "2020-01-23T18:00:00"},
		{"2020-1-23 6am", "2020-01-23T06:00:00"},
		//
		{"2020-1-1T18:00:20", "2020-01-01T18:00:20"},
		{"2020-1-1 18:00:20", "2020-01-01T18:00:20"},
		{"2020-1-1T18:00", "2020-01-01T18:00:00"},
		{"2020-1-1 18:00", "2020-01-01T18:00:00"},
		{"2020-1-1T1800", "2020-01-01T18:00:00"},
		{"2020-1-1 1800", "2020-01-01T18:00:00"},
		{"2020-1-1T18", "2020-01-01T18:00:00"},
		{"2020-1-1 18", "2020-01-01T18:00:00"},
		{"2020-1-1 6pm", "2020-01-01T18:00:00"},
		{"2020-1-1 6am", "2020-01-01T06:00:00"},
		//
		{"23-01-2020T18:00:20", "2020-01-23T18:00:20"},
		{"23-01-2020 18:00:20", "2020-01-23T18:00:20"},
		{"23-01-2020T18:00", "2020-01-23T18:00:00"},
		{"23-01-2020 18:00", "2020-01-23T18:00:00"},
		{"23-01-2020T1800", "2020-01-23T18:00:00"},
		{"23-01-2020 1800", "2020-01-23T18:00:00"},
		{"23-01-2020T18", "2020-01-23T18:00:00"},
		{"23-01-2020 18", "2020-01-23T18:00:00"},
		{"23-01-2020 6pm", "2020-01-23T18:00:00"},
		{"23-01-2020 6am", "2020-01-23T06:00:00"},
		//
		{"23-1-2020T18:00:20", "2020-01-23T18:00:20"},
		{"23-1-2020 18:00:20", "2020-01-23T18:00:20"},
		{"23-1-2020T18:00", "2020-01-23T18:00:00"},
		{"23-1-2020 18:00", "2020-01-23T18:00:00"},
		{"23-1-2020T1800", "2020-01-23T18:00:00"},
		{"23-1-2020 1800", "2020-01-23T18:00:00"},
		{"23-1-2020T18", "2020-01-23T18:00:00"},
		{"23-1-2020 18", "2020-01-23T18:00:00"},
		{"23-1-2020 6pm", "2020-01-23T18:00:00"},
		{"23-1-2020 6am", "2020-01-23T06:00:00"},
		//
		{"1-1-2020T18:00:20", "2020-01-01T18:00:20"},
		{"1-1-2020 18:00:20", "2020-01-01T18:00:20"},
		{"1-1-2020T18:00", "2020-01-01T18:00:00"},
		{"1-1-2020 18:00", "2020-01-01T18:00:00"},
		{"1-1-2020T1800", "2020-01-01T18:00:00"},
		{"1-1-2020 1800", "2020-01-01T18:00:00"},
		{"1-1-2020T18", "2020-01-01T18:00:00"},
		{"1-1-2020 18", "2020-01-01T18:00:00"},
		{"1-1-2020 6pm", "2020-01-01T18:00:00"},
		{"1-1-2020 6am", "2020-01-01T06:00:00"},
		//
		{"1-12-2020T18:00:20", "2020-12-01T18:00:20"},
		{"1-12-2020 18:00:20", "2020-12-01T18:00:20"},
		{"1-12-2020T18:00", "2020-12-01T18:00:00"},
		{"1-12-2020 18:00", "2020-12-01T18:00:00"},
		{"1-12-2020T1800", "2020-12-01T18:00:00"},
		{"1-12-2020 1800", "2020-12-01T18:00:00"},
		{"1-12-2020T18", "2020-12-01T18:00:00"},
		{"1-12-2020 18", "2020-12-01T18:00:00"},
		{"1-12-2020 6pm", "2020-12-01T18:00:00"},
		{"1-12-2020 6am", "2020-12-01T06:00:00"},
		{"1-12-2020 6:59pm", "2020-12-01T18:59:00"},
		{"1-12-2020 6:59:30pm", "2020-12-01T18:59:30"},
	}
	for _, c := range deadlineTests {
		fixed, err := FixDeadline(c.in)
		if err != nil {
			t.Errorf("FixDeadline(%q) returned error %v", c.in, err)
		}
		// format deadline to match wanted layout
		got := fixed.AsTime().Format(qf.TimeLayout)
		if got != c.want {
			t.Errorf("FixDeadline(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
