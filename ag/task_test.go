package ag_test

import (
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

// TestHasChanged tests if HasChanged returns the correct value
func TestHasChanged(t *testing.T) {
	originalTask := &pb.Task{
		ID:              1,
		AssignmentID:    1,
		AssignmentOrder: 1,
		Title:           "This is the original task",
		Body:            "Description description",
		Name:            "lab1/1",
	}

	// Comparing the original task with itself.
	wantResult := false
	gotResult := originalTask.HasChanged(originalTask)

	if diff := cmp.Diff(wantResult, gotResult, protocmp.Transform()); diff != "" {
		t.Errorf("HasChanged mismatch (-wantResult, +gotResult):\n%s", diff)
	}
	// -------------------------------------------------------------------------- //

	// Checking for body change.
	updatedTask := &pb.Task{
		ID:              1,
		AssignmentID:    1,
		AssignmentOrder: 1,
		Title:           "This is the original task",
		Body:            "Different description",
		Name:            "lab1/1",
	}
	wantResult = true
	gotResult = originalTask.HasChanged(updatedTask)

	if diff := cmp.Diff(wantResult, gotResult, protocmp.Transform()); diff != "" {
		t.Errorf("HasChanged mismatch (-wantResult, +gotResult):\n%s", diff)
	}
	// -------------------------------------------------------------------------- //

	// Checking for title change.
	updatedTask.Body = "Description description"
	updatedTask.Title = "A new title"
	wantResult = true
	gotResult = originalTask.HasChanged(updatedTask)

	if diff := cmp.Diff(wantResult, gotResult, protocmp.Transform()); diff != "" {
		t.Errorf("HasChanged mismatch (-wantResult, +gotResult):\n%s", diff)
	}
	// -------------------------------------------------------------------------- //

	// Checking for title and body change.
	updatedTask.Body = "Different description"
	wantResult = true
	gotResult = originalTask.HasChanged(updatedTask)

	if diff := cmp.Diff(wantResult, gotResult, protocmp.Transform()); diff != "" {
		t.Errorf("HasChanged mismatch (-wantResult, +gotResult):\n%s", diff)
	}
	// -------------------------------------------------------------------------- //
}
