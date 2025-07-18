package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConvertAssignments(t *testing.T) {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "test-convert-assignments")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test directories
	lab1Dir := filepath.Join(tempDir, "lab1")
	lab2Dir := filepath.Join(tempDir, "lab2")
	if err := os.MkdirAll(lab1Dir, 0755); err != nil {
		t.Fatalf("Failed to create lab1 directory: %v", err)
	}
	if err := os.MkdirAll(lab2Dir, 0755); err != nil {
		t.Fatalf("Failed to create lab2 directory: %v", err)
	}

	// Create YAML files
	yamlContent1 := `order: 1
name: "lab1"
deadline: "2025-01-15T23:59:00"
scorelimit: 80
autoapprove: false
isgrouplab: false`

	yamlContent2 := `order: 2
name: "lab2"
deadline: "2025-02-15T23:59:00"
scorelimit: 90
autoapprove: true
isgrouplab: true`

	if err := os.WriteFile(filepath.Join(lab1Dir, "assignment.yml"), []byte(yamlContent1), 0644); err != nil {
		t.Fatalf("Failed to create assignment.yml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(lab2Dir, "assignment.yaml"), []byte(yamlContent2), 0644); err != nil {
		t.Fatalf("Failed to create assignment.yaml: %v", err)
	}

	// Run conversion
	convertAssignments([]string{tempDir})

	// Check that JSON files were created
	jsonPath1 := filepath.Join(lab1Dir, "assignment.json")
	jsonPath2 := filepath.Join(lab2Dir, "assignment.json")

	if _, err := os.Stat(jsonPath1); os.IsNotExist(err) {
		t.Errorf("Expected JSON file not created: %s", jsonPath1)
	}
	if _, err := os.Stat(jsonPath2); os.IsNotExist(err) {
		t.Errorf("Expected JSON file not created: %s", jsonPath2)
	}

	// Verify JSON content
	jsonContent1, err := os.ReadFile(jsonPath1)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}
	if len(jsonContent1) == 0 {
		t.Error("JSON file is empty")
	}
}

func TestConvertAssignmentFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "test-convert-single")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a YAML file
	yamlPath := filepath.Join(tempDir, "assignment.yml")
	yamlContent := `order: 1
name: "test-lab"
deadline: "2025-01-15T23:59:00"
scorelimit: 85
autoapprove: false
isgrouplab: false`

	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create YAML file: %v", err)
	}

	// Convert the file
	if err := convertAssignmentFile(yamlPath); err != nil {
		t.Fatalf("Failed to convert assignment file: %v", err)
	}

	// Check that JSON file was created
	jsonPath := filepath.Join(tempDir, "assignment.json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Error("Expected JSON file not created")
	}
}

func TestConvertAssignmentsNoArguments(t *testing.T) {
	// Capture output (this is a simple test, in practice you'd want to capture stdout)
	convertAssignments([]string{})
	// This test mainly ensures no panic occurs
}

func TestConvertAssignmentsInvalidDirectory(t *testing.T) {
	// Test with non-existent directory
	convertAssignments([]string{"/non/existent/directory"})
	// This test mainly ensures no panic occurs
}

func TestEndToEndJSONParsing(t *testing.T) {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "test-end-to-end")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test directories
	lab1Dir := filepath.Join(tempDir, "lab1")
	lab2Dir := filepath.Join(tempDir, "lab2")
	lab3Dir := filepath.Join(tempDir, "lab3")
	
	for _, dir := range []string{lab1Dir, lab2Dir, lab3Dir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create a YAML file
	yamlContent := `order: 1
name: "lab1"
title: "YAML Assignment"
deadline: "2025-01-15T23:59:00"
scorelimit: 80
autoapprove: false
isgrouplab: false`
	
	// Create a JSON file
	jsonContent := `{
  "order": 2,
  "name": "lab2",
  "title": "JSON Assignment",
  "deadline": "2025-02-15T23:59:00",
  "scorelimit": 85,
  "autoapprove": true,
  "isgrouplab": true
}`

	// Create a directory with both YAML and JSON (JSON should take precedence)
	mixedYAML := `order: 3
name: "lab3"
title: "Should be overridden"
deadline: "2025-03-15T23:59:00"
scorelimit: 70
autoapprove: false
isgrouplab: false`

	mixedJSON := `{
  "order": 3,
  "name": "lab3",
  "title": "JSON Takes Precedence",
  "deadline": "2025-03-15T23:59:00",
  "scorelimit": 95,
  "autoapprove": true,
  "isgrouplab": false
}`

	// Write files
	if err := os.WriteFile(filepath.Join(lab1Dir, "assignment.yml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(lab2Dir, "assignment.json"), []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to write JSON file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(lab3Dir, "assignment.yml"), []byte(mixedYAML), 0644); err != nil {
		t.Fatalf("Failed to write mixed YAML file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(lab3Dir, "assignment.json"), []byte(mixedJSON), 0644); err != nil {
		t.Fatalf("Failed to write mixed JSON file: %v", err)
	}

	// Test parsing YAML file
	assignment1, err := parseAssignment(filepath.Join(lab1Dir, "assignment.yml"))
	if err != nil {
		t.Fatalf("Failed to parse YAML assignment: %v", err)
	}
	if assignment1.Title != "YAML Assignment" {
		t.Errorf("Expected title 'YAML Assignment', got %s", assignment1.Title)
	}
	if assignment1.ScoreLimit != 80 {
		t.Errorf("Expected score limit 80, got %d", assignment1.ScoreLimit)
	}

	// Test parsing JSON file
	assignment2, err := parseAssignment(filepath.Join(lab2Dir, "assignment.json"))
	if err != nil {
		t.Fatalf("Failed to parse JSON assignment: %v", err)
	}
	if assignment2.Title != "JSON Assignment" {
		t.Errorf("Expected title 'JSON Assignment', got %s", assignment2.Title)
	}
	if assignment2.ScoreLimit != 85 {
		t.Errorf("Expected score limit 85, got %d", assignment2.ScoreLimit)
	}

	// Test JSON precedence
	assignment3, err := parseAssignment(filepath.Join(lab3Dir, "assignment.json"))
	if err != nil {
		t.Fatalf("Failed to parse JSON assignment: %v", err)
	}
	if assignment3.Title != "JSON Takes Precedence" {
		t.Errorf("Expected title 'JSON Takes Precedence', got %s", assignment3.Title)
	}
	if assignment3.ScoreLimit != 95 {
		t.Errorf("Expected score limit 95, got %d", assignment3.ScoreLimit)
	}

	t.Log("All end-to-end tests passed!")
}