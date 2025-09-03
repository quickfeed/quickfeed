package database

import (
	"fmt"

	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateAssignmentFeedback creates a new assignment feedback.
func (db *GormDB) CreateAssignmentFeedback(feedback *qf.AssignmentFeedback) error {
	// Set the creation timestamp
	feedback.CreatedAt = timestamppb.Now()
	if err := db.conn.Create(feedback).Error; err != nil {
		return fmt.Errorf("failed to create assignment feedback: %w", err)
	}
	return nil
}

// GetAssignmentFeedback returns a list of assignment feedback matching the given query.
// If userID is specified, returns feedback from that user only.
// Otherwise, returns all feedback for the assignment.
func (db *GormDB) GetAssignmentFeedback(query *qf.AssignmentFeedbackRequest) (*qf.AssignmentFeedbacks, error) {
	var feedbacks []*qf.AssignmentFeedback
	dbQuery := db.conn.Model(&qf.AssignmentFeedback{})

	if query.GetAssignmentID() > 0 || query.GetUserID() > 0 {
		switch query.GetMode().(type) {
		case *qf.AssignmentFeedbackRequest_AssignmentID:
			dbQuery = dbQuery.Where("assignment_id = ?", query.GetAssignmentID())
		case *qf.AssignmentFeedbackRequest_UserID:
			dbQuery = dbQuery.Where("user_id = ?", query.GetUserID())
		}
	}

	dbQuery = dbQuery.Where("course_id = ?", query.GetCourseID())

	if err := dbQuery.Find(&feedbacks).Error; err != nil {
		return nil, fmt.Errorf("failed to get assignment feedback: %w", err)
	}
	return &qf.AssignmentFeedbacks{Feedbacks: feedbacks}, nil
}
