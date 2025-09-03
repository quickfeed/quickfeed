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
	tests := []struct {
		name  string
		input isAssignmentFeedbackRequest_Mode
		valid bool
	}{
		{
			name:  "Invalid request",
			input: nil,
			valid: false,
		},
		{
			name:  "Valid request with course ID",
			input: &AssignmentFeedbackRequest_UserID{UserID: 1},
			valid: true,
		},
		{
			name:  "Valid request with user ID",
			input: &AssignmentFeedbackRequest_AssignmentID{AssignmentID: 1},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &AssignmentFeedbackRequest{
				Mode: tt.input,
			}
			if req.IsValid() != tt.valid {
				t.Errorf("IsValid() = %v, want %v", req.IsValid(), tt.valid)
			}
		})
	}
}
