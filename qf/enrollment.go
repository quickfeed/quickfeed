package qf

import (
	"fmt"
	"time"
)

// gracePeriod is the grace period for submissions after the deadline.
// The grace period should be hours in the range 0-23.
// Note grace period applies to all enrollments (courses).
const gracePeriod time.Duration = time.Duration(2 * time.Hour)

// UpdateSlipDays updates the number of slip days for the given assignment/submission.
func (m *Enrollment) UpdateSlipDays(start time.Time, assignment *Assignment, submission *Submission) error {
	if m.GetCourseID() != assignment.GetCourseID() {
		return fmt.Errorf("invariant violation (enrollment.CourseID != assignment.CourseID) (%d != %d)", m.CourseID, assignment.CourseID)
	}
	if assignment.GetID() != submission.GetAssignmentID() {
		return fmt.Errorf("invariant violation (assignment.ID != submission.AssignmentID) (%d != %d)", assignment.ID, submission.AssignmentID)
	}
	sinceDeadline, err := assignment.SinceDeadline(start)
	if err != nil {
		return err
	}
	// if score is less than limit and it's not yet approved, update slip days if deadline has passed
	if submission.Score < assignment.ScoreLimit && !submission.IsApproved(m.GetUserID()) && sinceDeadline > 0 {
		// deadline exceeded; calculate used slipdays for this assignment
		slpDays, slpHours := uint32(sinceDeadline/days), sinceDeadline%days
		// slpHours is hours after deadline, excluding subsequent full-day slip days after deadline
		if slpHours > gracePeriod {
			slpDays++
		}
		m.internalUpdateSlipDays(assignment.GetID(), slpDays)
	}
	return nil
}

// internalUpdateSlipDays updates the number of slip days for the given assignment.
func (m *Enrollment) internalUpdateSlipDays(assignmentID uint64, slipDays uint32) {
	for _, val := range m.GetUsedSlipDays() {
		if val.AssignmentID == assignmentID {
			val.UsedDays = slipDays
			return
		}
	}
	// not found; add new entry to the slice
	m.UsedSlipDays = append(m.UsedSlipDays, &UsedSlipDays{
		AssignmentID: assignmentID,
		EnrollmentID: m.ID,
		UsedDays:     slipDays,
	})
}

// totalSlipDays returns the total number of slip days used for this enrollment.
func (m *Enrollment) totalSlipDays() uint32 {
	var total uint32
	for _, val := range m.GetUsedSlipDays() {
		total += val.GetUsedDays()
	}
	return total
}

// RemainingSlipDays returns the remaining number of slip days for this
// user/course enrollment. Note that if the returned amount is negative,
// the user has used up all slip days.
func (m *Enrollment) RemainingSlipDays(c *Course) int32 {
	if m.GetCourseID() != c.GetID() {
		return 0
	}
	return int32(c.GetSlipDays() - m.totalSlipDays())
}

// SetSlipDays updates SlipDaysRemaining field of an enrollment.
func (m *Enrollment) SetSlipDays(c *Course) {
	remaining := m.RemainingSlipDays(c)
	if remaining < 0 {
		m.SlipDaysRemaining = 0
	} else {
		m.SlipDaysRemaining = uint32(remaining)
	}
}

func (m *Enrollment) IsNone() bool {
	return m.GetStatus() == Enrollment_NONE
}

func (m *Enrollment) IsPending() bool {
	return m.GetStatus() == Enrollment_PENDING
}

func (m *Enrollment) IsStudent() bool {
	return m.GetStatus() == Enrollment_STUDENT
}

func (m *Enrollment) IsTeacher() bool {
	return m.GetStatus() == Enrollment_TEACHER
}

// IsAdmin returns true if the enrolled user is an admin.
func (m *Enrollment) IsAdmin() bool {
	return m.GetUser().GetIsAdmin()
}

// Name returns the name of the enrolled user.
func (m *Enrollment) Name() string {
	return m.GetUser().GetName()
}

// GetCourseID returns the course ID for a slice of enrollments
func (m *Enrollments) GetCourseID() uint64 {
	enrollments := m.GetEnrollments()
	if len(enrollments) == 0 {
		return 0
	}
	return enrollments[0].GetCourseID()
}

// HasCourseID checks all enrollments have the same Course ID
func (m *Enrollments) HasCourseID() bool {
	enrollments := m.GetEnrollments()
	if len(enrollments) == 0 {
		return false
	}
	courseID := enrollments[0].GetCourseID()
	for _, e := range enrollments {
		if e.GetCourseID() != courseID {
			return false
		}
	}
	return true
}

// UserIDs returns the user IDs in these enrollments.
func (m *Enrollments) UserIDs() []uint64 {
	userIDs := make([]uint64, 0)
	for _, enrollment := range m.GetEnrollments() {
		userIDs = append(userIDs, enrollment.GetUserID())
	}
	return userIDs
}
