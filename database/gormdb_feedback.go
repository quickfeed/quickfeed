package database

import (
	"fmt"

	"github.com/quickfeed/quickfeed/qf"
)

// CreateAssignmentFeedback creates a new assignment feedback.
func (db *GormDB) CreateAssignmentFeedback(feedback *qf.AssignmentFeedback) error {
	if err := db.conn.Create(feedback).Error; err != nil {
		return fmt.Errorf("failed to create assignment feedback: %w", err)
	}
	return nil
}

// GetAssignmentFeedback returns assignment feedback matching the given query.
// If userID is specified, returns feedback from that user.
// Otherwise, returns the first feedback found for the assignment.
func (db *GormDB) GetAssignmentFeedback(query *qf.AssignmentFeedbackRequest) (*qf.AssignmentFeedback, error) {
	var feedback qf.AssignmentFeedback
	dbQuery := db.conn.Where("assignment_id = ?", query.GetAssignmentID())

	// If userID is specified, filter by user
	if query.GetUserID() > 0 {
		dbQuery = dbQuery.Where("user_id = ?", query.GetUserID())
	}

	if err := dbQuery.First(&feedback).Error; err != nil {
		return nil, fmt.Errorf("failed to get assignment feedback: %w", err)
	}
	return &feedback, nil
}
