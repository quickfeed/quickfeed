package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConvertAssignmentFile(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	
	// Create a test YAML file
	yamlPath := filepath.Join(tempDir, "assignment.yml")
	yamlContent := `order: 1
deadline: "2025-08-26T23:59:00"
scorelimit: 90
autoapprove: false
isgrouplab: false`
	
	err := os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test YAML file: %v", err)
	}
	
	// Test the conversion
	info, err := processAssignmentFile(yamlPath)
	if err != nil {
		t.Fatalf("processAssignmentFile() error = %v", err)
	}
	
	// Verify the target path
	expectedTargetPath := filepath.Join(tempDir, "assignment.json")
	if info.TargetPath != expectedTargetPath {
		t.Errorf("TargetPath = %s, want %s", info.TargetPath, expectedTargetPath)
	}
	
	// Verify the JSON content is valid and contains expected fields
	expectedJSON := `{
  "order": 1,
  "deadline": "2025-08-26T23:59:00",
  "isgrouplab": false,
  "autoapprove": false,
  "scorelimit": 90,
  "reviewers": 0,
  "containertimeout": 0
}`
	
	if string(info.JSONContent) != expectedJSON {
		t.Errorf("JSON content mismatch.\nGot:\n%s\nWant:\n%s", string(info.JSONContent), expectedJSON)
	}
}

func TestConvertAssignmentFileWithAllFields(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	
	// Create a test YAML file with all fields
	yamlPath := filepath.Join(tempDir, "assignment.yaml")
	yamlContent := `order: 2
deadline: "2025-09-02T23:59:00"
scorelimit: 85
autoapprove: true
isgrouplab: true
reviewers: 3
containertimeout: 45`
	
	err := os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test YAML file: %v", err)
	}
	
	// Test the conversion
	info, err := processAssignmentFile(yamlPath)
	if err != nil {
		t.Fatalf("processAssignmentFile() error = %v", err)
	}
	
	// Verify the target path
	expectedTargetPath := filepath.Join(tempDir, "assignment.json")
	if info.TargetPath != expectedTargetPath {
		t.Errorf("TargetPath = %s, want %s", info.TargetPath, expectedTargetPath)
	}
	
	// Verify the JSON content contains all fields
	expectedJSON := `{
  "order": 2,
  "deadline": "2025-09-02T23:59:00",
  "isgrouplab": true,
  "autoapprove": true,
  "scorelimit": 85,
  "reviewers": 3,
  "containertimeout": 45
}`
	
	if string(info.JSONContent) != expectedJSON {
		t.Errorf("JSON content mismatch.\nGot:\n%s\nWant:\n%s", string(info.JSONContent), expectedJSON)
	}
}

func TestConvertAssignmentFileInvalidYAML(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	
	// Create an invalid YAML file
	yamlPath := filepath.Join(tempDir, "assignment.yml")
	yamlContent := `order: invalid
deadline: "2025-08-26T23:59:00"
autoapprove: not_a_bool`
	
	err := os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test YAML file: %v", err)
	}
	
	// Test the conversion should fail
	_, err = processAssignmentFile(yamlPath)
	if err == nil {
		t.Error("processAssignmentFile() should have failed with invalid YAML")
	}
}