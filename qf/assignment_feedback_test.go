package qf

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAssignmentFeedbackValidation(t *testing.T) {
	// Test valid feedback
	validFeedback := &AssignmentFeedback{
		AssignmentID:           1,
		LikedContent:           "Great assignment with clear instructions",
		ImprovementSuggestions: "Could use more examples",
		TimeSpent:              2,
		CreatedAt:              timestamppb.New(time.Now()),
	}

	if !validFeedback.IsValid() {
		t.Error("Valid feedback should pass validation")
	}

	// Test invalid feedback (missing assignment ID)
	invalidFeedback := &AssignmentFeedback{
		LikedContent:           "Good assignment",
		ImprovementSuggestions: "Needs improvement",
	}

	if invalidFeedback.IsValid() {
		t.Error("Feedback without assignment ID should fail validation")
	}

	// Test invalid feedback (no content)
	emptyFeedback := &AssignmentFeedback{
		AssignmentID: 1,
	}

	if emptyFeedback.IsValid() {
		t.Error("Feedback without content should fail validation")
	}
}

func TestAssignmentFeedbackRequestValidation(t *testing.T) {
	// Test valid request
	validRequest := &AssignmentFeedbackRequest{
		CourseID:     1,
		AssignmentID: 1,
		UserID:       123,
	}

	if !validRequest.IsValid() {
		t.Error("Valid request should pass validation")
	}

	// Test invalid request (missing course ID)
	invalidRequest := &AssignmentFeedbackRequest{
		AssignmentID: 1,
	}

	if invalidRequest.IsValid() {
		t.Error("Request without course ID should fail validation")
	}

	// Test invalid request (missing assignment ID)
	invalidRequest2 := &AssignmentFeedbackRequest{
		CourseID: 1,
	}

	if invalidRequest2.IsValid() {
		t.Error("Request without assignment ID should fail validation")
	}
}
