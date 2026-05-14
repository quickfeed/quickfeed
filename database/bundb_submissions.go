package database

import (
	"context"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/uptrace/bun"
)

// GetCourseSubmissions returns the latest course submissions of the requested submission type.
// It reuses the same helper functions as the GORM implementation since they are ORM-agnostic.
func (db *BunDB) GetCourseSubmissions(request *qf.SubmissionRequest) (*qf.CourseSubmissions, error) {
	ctx := context.Background()

	var assignmentIDs []uint64
	q := db.conn.NewSelect().Model((*qf.Assignment)(nil)).
		Column("id").
		Where("course_id = ?", request.GetCourseID())
	switch request.GetType() {
	case qf.SubmissionRequest_USER:
		q = q.Where("is_group_lab = ?", false)
	case qf.SubmissionRequest_GROUP:
		q = q.Where("is_group_lab = ?", true)
	}
	q = q.OrderExpr("\"order\"")
	if err := q.Scan(ctx, &assignmentIDs); err != nil {
		return nil, err
	}

	var submissions []*qf.Submission
	if err := db.conn.NewSelect().
		Model(&submissions).
		Relation("Grades").
		Relation("Reviews").
		Relation("Reviews.GradingBenchmarks").
		Relation("Reviews.GradingBenchmarks.Criteria").
		Relation("Scores").
		Where("assignment_id IN (?)", bun.In(assignmentIDs)).
		Scan(ctx); err != nil {
		return nil, err
	}

	course, err := db.GetCourseByStatus(request.GetCourseID(), qf.Enrollment_TEACHER)
	if err != nil {
		return nil, err
	}

	var submissionsMap map[uint64]*qf.Submissions
	switch request.GetType() {
	case qf.SubmissionRequest_GROUP:
		submissionsMap = makeGroupResults(course, submissions)
	case qf.SubmissionRequest_USER:
		submissionsMap = makeUserResults(course, submissions)
	case qf.SubmissionRequest_ALL:
		submissionsMap = makeAllResults(course, submissions)
	}
	return &qf.CourseSubmissions{Submissions: submissionsMap}, nil
}
