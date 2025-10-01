package database

import (
	"fmt"

	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// CreateAssignmentFeedback creates a new assignment feedback.
func (db *GormDB) CreateAssignmentFeedback(feedback *qf.AssignmentFeedback, userID uint64) error {
	err := db.conn.Transaction(func(tx *gorm.DB) error {
		// Set the creation timestamp
		feedback.CreatedAt = timestamppb.Now()
		if err := tx.Create(feedback).Error; err != nil {
			return fmt.Errorf("failed to create assignment feedback: %w", err)
		}
		// Create a receipt for the feedback
		receipt := &qf.FeedbackReceipt{
			AssignmentID: feedback.GetAssignmentID(),
			UserID:       userID,
		}
		if err := tx.Create(receipt).Error; err != nil {
			return fmt.Errorf("failed to create feedback receipt: %w", err)
		}
		return nil
	})
	return err
}

// GetAssignmentFeedback returns a list of assignment feedback matching the given course.
func (db *GormDB) GetAssignmentFeedback(query *qf.CourseRequest) (*qf.AssignmentFeedbacks, error) {
	var feedbacks []*qf.AssignmentFeedback
	if err := db.conn.Model(&qf.AssignmentFeedback{}).Where(&qf.AssignmentFeedback{
		CourseID: query.GetCourseID(),
	}).Find(&feedbacks).Error; err != nil {
		return nil, fmt.Errorf("failed to get assignment feedback: %w", err)
	}
	return &qf.AssignmentFeedbacks{Feedbacks: feedbacks}, nil
}
