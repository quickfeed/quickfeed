package qf

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAssignmentFeedbackValidation(t *testing.T) {
	tests := []struct {
		name      string
		feedback  *AssignmentFeedback
		wantValid bool
	}{
		{
			name: "valid feedback with all fields",
			feedback: &AssignmentFeedback{
				CourseID:               1,
				AssignmentID:           1,
				LikedContent:           "Great assignment with clear instructions",
				ImprovementSuggestions: "Could use more examples",
				TimeSpent:              2,
				CreatedAt:              timestamppb.New(time.Now()),
			},
			wantValid: true,
		},
		{
			name: "invalid feedback missing assignment ID",
			feedback: &AssignmentFeedback{
				CourseID:               1,
				LikedContent:           "Good assignment",
				ImprovementSuggestions: "Needs improvement",
			},
			wantValid: false,
		},
		{
			name: "invalid feedback with no content",
			feedback: &AssignmentFeedback{
				AssignmentID: 1,
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.feedback.IsValid()
			if got != tt.wantValid {
				t.Errorf("AssignmentFeedback.IsValid() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

func TestAssignmentFeedbackRequestValidation(t *testing.T) {
	tests := []struct {
		name      string
		request   *CourseRequest
		wantValid bool
	}{
		{
			name: "valid request with all required fields",
			request: &CourseRequest{
				CourseID: 1,
			},
			wantValid: true,
		},
		{
			name: "invalid request missing course ID",
			request: &CourseRequest{
				CourseID: 0,
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.request.IsValid()
			if got != tt.wantValid {
				t.Errorf("AssignmentFeedbackRequest.IsValid() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}
