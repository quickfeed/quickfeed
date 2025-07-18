package assignments

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestParseJSONAssignment(t *testing.T) {
	jsonContent := `{
		"order": 1,
		"name": "lab1",
		"title": "Getting Started with Go",
		"deadline": "2025-01-15T23:59:00",
		"scorelimit": 80,
		"autoapprove": false,
		"isgrouplab": false,
		"hoursmin": 10,
		"hoursmax": 15
	}`

	assignment, err := newAssignmentFromFile([]byte(jsonContent), "lab1", 1, "assignment.json")
	if err != nil {
		t.Fatalf("Failed to parse JSON assignment: %v", err)
	}

	want := &qf.Assignment{
		Name:       "lab1",
		CourseID:   1,
		Order:      1,
		ScoreLimit: 80,
		Deadline:   qtest.Timestamp(t, "2025-01-15T23:59:00"),
		AutoApprove: false,
		IsGroupLab: false,
	}

	if diff := cmp.Diff(want, assignment, protocmp.Transform()); diff != "" {
		t.Errorf("JSON assignment parsing mismatch (-want +got):\n%s", diff)
	}
}

func TestParseYAMLAssignment(t *testing.T) {
	yamlContent := `order: 1
name: "lab1"
title: "Getting Started with Go"
deadline: "2025-01-15T23:59:00"
scorelimit: 80
autoapprove: false
isgrouplab: false
hoursmin: 10
hoursmax: 15`

	assignment, err := newAssignmentFromFile([]byte(yamlContent), "lab1", 1, "assignment.yml")
	if err != nil {
		t.Fatalf("Failed to parse YAML assignment: %v", err)
	}

	want := &qf.Assignment{
		Name:       "lab1",
		CourseID:   1,
		Order:      1,
		ScoreLimit: 80,
		Deadline:   qtest.Timestamp(t, "2025-01-15T23:59:00"),
		AutoApprove: false,
		IsGroupLab: false,
	}

	if diff := cmp.Diff(want, assignment, protocmp.Transform()); diff != "" {
		t.Errorf("YAML assignment parsing mismatch (-want +got):\n%s", diff)
	}
}

func TestParseUnsupportedFormat(t *testing.T) {
	content := `some content`

	_, err := newAssignmentFromFile([]byte(content), "lab1", 1, "assignment.txt")
	if err == nil {
		t.Error("Expected error for unsupported file format, got nil")
	}
	
	expectedError := "unsupported assignment file format: .txt"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}