package database

import (
	"sort"

	"github.com/quickfeed/quickfeed/qf"
)

// GetCourseSubmissions returns all individual lab submissions by students enrolled in the specified course.
func (db *GormDB) GetCourseSubmissions(request *qf.SubmissionRequest) (*qf.CourseSubmissions, error) {
	var assignmentIDs []uint64
	a := db.conn.Model(&qf.Assignment{}).Where(&qf.Assignment{CourseID: request.GetCourseID()})
	switch request.GetType() {
	case qf.SubmissionRequest_USER:
		// Must use string-based query since GORM does not support boolean false in type-based Where clauses
		a.Where("is_group_lab = ?", false)
	case qf.SubmissionRequest_GROUP:
		a.Where(&qf.Assignment{IsGroupLab: true})
	default: // all
	}
	// the 'order' field of qf.Assignment must be in 'quotes' since otherwise it will be interpreted as SQL
	if err := a.Order("'order'").Pluck("id", &assignmentIDs).Error; err != nil {
		return nil, err
	}
	var submissions []*qf.Submission
	m := db.conn.Model(&qf.Submission{}).Preload("Grades").
		Preload("Reviews").
		Preload("Reviews.GradingBenchmarks").
		Preload("Reviews.GradingBenchmarks.Criteria").
		Preload("Scores")
	if err := m.Where("assignment_id IN ?", assignmentIDs).
		Find(&submissions).Error; err != nil {
		return nil, err
	}

	// fetch course record with all assignments and active enrollments
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

// makeGroupResults returns a map of group ID to Submissions
// for all course groups and all group assignments.
func makeGroupResults(course *qf.Course, submissions []*qf.Submission) map[uint64]*qf.Submissions {
	submissionsMap := make(map[uint64]*qf.Submissions)
	skipGroup := map[uint64]bool{0: true} // skip group ID 0 (no group)
	om := newOrderMap(course.GetAssignments())
	for _, enrollment := range course.GetEnrollments() {
		if skipGroup[enrollment.GetGroupID()] {
			continue // include group enrollment only once
		}
		skipGroup[enrollment.GetGroupID()] = true
		// Note: we (intentionally) use enrollment.GroupID as the key to the map here.
		// This is primarily a convenience for the frontend, which can then use
		// the group ID as the key to the map.
		submissionsMap[enrollment.GetGroupID()] = &qf.Submissions{
			Submissions: choose(submissions, om, func(submission *qf.Submission) bool {
				// include group submissions for this enrollment
				return submission.ByGroup(enrollment.GetGroupID())
			}),
		}
	}
	return submissionsMap
}

// makeUserResults returns a map of enrollment ID to Submissions
// for all course enrollments (students) and all individual assignments.
func makeUserResults(course *qf.Course, submission []*qf.Submission) map[uint64]*qf.Submissions {
	submissionsMap := make(map[uint64]*qf.Submissions)
	om := newOrderMap(course.GetAssignments())
	for _, enrollment := range course.GetEnrollments() {
		submissionsMap[enrollment.GetID()] = &qf.Submissions{
			Submissions: choose(submission, om, func(submission *qf.Submission) bool {
				// include individual submissions for this enrollment
				return submission.ByUser(enrollment.GetUserID())
			}),
		}
	}
	return submissionsMap
}

// makeAllResults returns a map of enrollment ID to Submissions
// for all course enrollments (students and groups) and all individual and group assignments.
func makeAllResults(course *qf.Course, submissions []*qf.Submission) map[uint64]*qf.Submissions {
	submissionsMap := make(map[uint64]*qf.Submissions)
	om := newOrderMap(course.GetAssignments())
	for _, enrollment := range course.GetEnrollments() {
		submissionsMap[enrollment.GetID()] = &qf.Submissions{
			Submissions: choose(submissions, om, func(submission *qf.Submission) bool {
				// include individual and group submissions for this enrollment
				return submission.ByUser(enrollment.GetUserID()) || submission.ByGroup(enrollment.GetGroupID())
			}),
		}
	}
	return submissionsMap
}

func choose(submissions []*qf.Submission, order *orderMap, include func(*qf.Submission) bool) []*qf.Submission {
	var subs []*qf.Submission
	for _, submission := range submissions {
		if include(submission) {
			subs = append(subs, submission)
		}
	}
	// sort submissions by assignment order
	sort.Slice(subs, func(i, j int) bool {
		return order.less(subs[i].GetAssignmentID(), subs[j].GetAssignmentID())
	})
	return subs
}

type orderMap map[uint64]uint32

// newOrderMap creates a new orderMap from a list of assignments.
// The ID of each assignment is mapped to its order.
// Useful for sorting submissions by assignment order
// as the order is not stored in the submission themselves.
func newOrderMap(assignments []*qf.Assignment) *orderMap {
	om := make(orderMap)
	for _, assignment := range assignments {
		om[assignment.GetID()] = assignment.GetOrder()
	}
	return &om
}

func (om orderMap) less(i, j uint64) bool {
	return om[i] < om[j]
}
