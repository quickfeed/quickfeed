package ag

import (
	"fmt"
	"time"
)

// gracePeriod is the grace period for submissions after the deadline.
// The grace period should be hours in the range 0-23.
// Note grace period applies to all enrollments (courses).
const gracePeriod time.Duration = time.Duration(2 * time.Hour)

// UpdateSlipDays updates the number of slipdays for the given assignment/submission.
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
	if submission.Score < assignment.ScoreLimit && submission.Status != Submission_APPROVED && sinceDeadline > 0 {
		// deadline exceeded; calculate used slipdays for this assignment
		slpDays, slpHours := uint32(sinceDeadline/days), sinceDeadline%days
		// slpHours is hours after deadline, excluding subsequent full-day slipdays after deadline
		if slpHours > gracePeriod {
			slpDays++
		}
		m.updateSlipDays(assignment.GetID(), slpDays)
	}
	return nil
}

// updateSlipDays updates the number of slipdays for the given assignment.
func (m *Enrollment) updateSlipDays(assignmentID uint64, slipDays uint32) {
	for _, val := range m.GetUsedSlipDays() {
		if val.AssignmentID == assignmentID {
			val.UsedSlipDays = slipDays
			return
		}
	}
	// not found; add new entry to the slice
	m.UsedSlipDays = append(m.UsedSlipDays, &UsedSlipDays{
		AssignmentID: assignmentID,
		EnrollmentID: m.ID,
		UsedSlipDays: slipDays,
	})
}

// totalSlipDays returns the total number of slipdays used for this enrollment.
func (m *Enrollment) totalSlipDays() uint32 {
	var total uint32
	for _, val := range m.GetUsedSlipDays() {
		total += val.GetUsedSlipDays()
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
	if m.RemainingSlipDays(c) < 0 {
		m.SlipDaysRemaining = 0
	} else {
		m.SlipDaysRemaining = uint32(m.RemainingSlipDays(c))
	}
}

func (m *Enrollment) IsTeacher() bool {
	return m.GetStatus() == Enrollment_TEACHER
}

func (m *Enrollment) IsStudent() bool {
	return m.GetStatus() == Enrollment_STUDENT
}
