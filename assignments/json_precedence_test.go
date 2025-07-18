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

func TestJSONPrecedenceOverYAML(t *testing.T) {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "test-json-precedence")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test directories
	lab1Dir := filepath.Join(tempDir, "lab1")
	lab2Dir := filepath.Join(tempDir, "lab2")
	
	for _, dir := range []string{lab1Dir, lab2Dir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Lab1: Only JSON file
	jsonContent := `{
  "order": 1,
  "deadline": "2025-01-15T23:59:00",
  "scorelimit": 85,
  "autoapprove": true,
  "isgrouplab": false
}`

	// Lab2: Both YAML and JSON files (JSON should take precedence)
	yamlContent := `order: 2
deadline: "2025-02-15T23:59:00"
scorelimit: 70
autoapprove: false
isgrouplab: true`

	jsonContent2 := `{
  "order": 2,
  "deadline": "2025-02-15T23:59:00",
  "scorelimit": 90,
  "autoapprove": true,
  "isgrouplab": false
}`

	// Write files
	if err := os.WriteFile(filepath.Join(lab1Dir, "assignment.json"), []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to write JSON file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(lab2Dir, "assignment.yml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(lab2Dir, "assignment.json"), []byte(jsonContent2), 0644); err != nil {
		t.Fatalf("Failed to write JSON file: %v", err)
	}

	// Test the repository content parsing
	assignments, _, err := readTestsRepositoryContent(tempDir, 1)
	if err != nil {
		t.Fatalf("Failed to read repository content: %v", err)
	}

	if len(assignments) != 2 {
		t.Fatalf("Expected 2 assignments, got %d", len(assignments))
	}

	// Find assignments by name
	var lab1Assignment, lab2Assignment *qf.Assignment
	for _, assignment := range assignments {
		if assignment.Name == "lab1" {
			lab1Assignment = assignment
		} else if assignment.Name == "lab2" {
			lab2Assignment = assignment
		}
	}

	if lab1Assignment == nil {
		t.Fatal("Lab1 assignment not found")
	}
	if lab2Assignment == nil {
		t.Fatal("Lab2 assignment not found")
	}

	// Test lab1 (JSON only)
	expectedLab1 := &qf.Assignment{
		Name:       "lab1",
		CourseID:   1,
		Order:      1,
		ScoreLimit: 85,
		Deadline:   qtest.Timestamp(t, "2025-01-15T23:59:00"),
		AutoApprove: true,
		IsGroupLab: false,
	}

	if diff := cmp.Diff(expectedLab1, lab1Assignment, protocmp.Transform()); diff != "" {
		t.Errorf("Lab1 assignment mismatch (-want +got):\n%s", diff)
	}

	// Test lab2 (JSON should take precedence over YAML)
	expectedLab2 := &qf.Assignment{
		Name:       "lab2",
		CourseID:   1,
		Order:      2,
		ScoreLimit: 90, // JSON value, not YAML value (70)
		Deadline:   qtest.Timestamp(t, "2025-02-15T23:59:00"),
		AutoApprove: true,  // JSON value, not YAML value (false)
		IsGroupLab: false,  // JSON value, not YAML value (true)
	}

	if diff := cmp.Diff(expectedLab2, lab2Assignment, protocmp.Transform()); diff != "" {
		t.Errorf("Lab2 assignment mismatch (-want +got):\n%s", diff)
	}

	// Specifically test that JSON values took precedence
	if lab2Assignment.ScoreLimit != 90 {
		t.Errorf("Expected JSON score limit 90, got %d (YAML would be 70)", lab2Assignment.ScoreLimit)
	}
	if !lab2Assignment.AutoApprove {
		t.Errorf("Expected JSON autoapprove true, got %v (YAML would be false)", lab2Assignment.AutoApprove)
	}
	if lab2Assignment.IsGroupLab {
		t.Errorf("Expected JSON isgrouplab false, got %v (YAML would be true)", lab2Assignment.IsGroupLab)
	}
}