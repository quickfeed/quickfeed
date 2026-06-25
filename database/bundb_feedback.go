package database

import (
	"context"
	"fmt"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateAssignmentFeedback creates a new assignment feedback and a corresponding receipt.
func (db *BunDB) CreateAssignmentFeedback(feedback *qf.AssignmentFeedback, userID uint64) error {
	ctx := context.Background()
	return db.conn.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		feedback.CreatedAt = timestamppb.Now()
		if _, err := tx.NewInsert().Model(feedback).Exec(ctx); err != nil {
			return fmt.Errorf("failed to create assignment feedback: %w", err)
		}
		receipt := &qf.FeedbackReceipt{
			AssignmentID: feedback.GetAssignmentID(),
			UserID:       userID,
		}
		if _, err := tx.NewInsert().Model(receipt).Exec(ctx); err != nil {
			return fmt.Errorf("failed to create feedback receipt: %w", err)
		}
		return nil
	})
}

// GetAssignmentFeedback returns a list of assignment feedbacks for the given course.
func (db *BunDB) GetAssignmentFeedback(query *qf.CourseRequest) (*qf.AssignmentFeedbacks, error) {
	ctx := context.Background()
	var feedbacks []*qf.AssignmentFeedback
	if err := db.conn.NewSelect().Model(&feedbacks).
		Where("course_id = ?", query.GetCourseID()).
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to get assignment feedback: %w", err)
	}
	return &qf.AssignmentFeedbacks{Feedbacks: feedbacks}, nil
}
