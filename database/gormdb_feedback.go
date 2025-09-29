package database

import (
	"fmt"

	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateAssignmentFeedback creates a new assignment feedback.
func (db *GormDB) CreateAssignmentFeedback(feedback *qf.AssignmentFeedback, userID uint64) error {
	// Set the creation timestamp
	feedback.CreatedAt = timestamppb.Now()
	if err := db.conn.Create(feedback).Error; err != nil {
		return fmt.Errorf("failed to create assignment feedback: %w", err)
	}
	// Create a receipt for the feedback
	receipt := &qf.FeedbackReceipt{
		AssignmentID: feedback.GetAssignmentID(),
		UserID:       userID,
	}
	if err := db.conn.Create(receipt).Error; err != nil {
		return fmt.Errorf("failed to create feedback receipt: %w", err)
	}
	return nil
}

// GetAssignmentFeedback returns a list of assignment feedback matching the given query.
// If userID is specified, returns feedback from that user only.
// Otherwise, returns all feedback for the assignment.
func (db *GormDB) GetAssignmentFeedback(query *qf.CourseRequest) (*qf.AssignmentFeedbacks, error) {
	var feedbacks []*qf.AssignmentFeedback
	dbQuery := db.conn.Model(&qf.AssignmentFeedback{})
	dbQuery = dbQuery.Where("course_id = ?", query.GetCourseID())

	if err := dbQuery.Find(&feedbacks).Error; err != nil {
		return nil, fmt.Errorf("failed to get assignment feedback: %w", err)
	}
	return &qf.AssignmentFeedbacks{Feedbacks: feedbacks}, nil
}
