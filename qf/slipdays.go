package qf

import (
	"fmt"
	"time"
)

// gracePeriod is the grace period for submissions after the deadline.
// The grace period should be hours in the range 0-23.
// Note grace period applies to all enrollments and groups (courses).
const gracePeriod = 2 * time.Hour

// slipDayHolder is implemented by *Group and *Enrollment, allowing the slip-day
// bookkeeping below to be shared between the two instead of duplicated.
type slipDayHolder interface {
	GetCourseID() uint64
	GetUsedSlipDays() []*UsedSlipDays
	addUsedSlipDays(assignmentID uint64, usedDays uint32)
	setSlipDaysRemaining(uint32)
}

// updateSlipDays updates the number of slip days used for the given assignment/submission.
// approved indicates whether the submission should be considered approved for the purposes
// of halting slip-day accrual; *Enrollment and *Group differ only in how that is determined.
func updateSlipDays(m slipDayHolder, assignment *Assignment, submission *Submission, approved bool) error {
	if m.GetCourseID() != assignment.GetCourseID() {
		return fmt.Errorf("invariant violation (GetCourseID() != assignment.GetCourseID()) (%d != %d)", m.GetCourseID(), assignment.GetCourseID())
	}
	if assignment.GetID() != submission.GetAssignmentID() {
		return fmt.Errorf("invariant violation (assignment.GetID() != submission.GetAssignmentID()) (%d != %d)", assignment.GetID(), submission.GetAssignmentID())
	}
	sinceDeadline := assignment.SinceDeadline(submission.GetBuildInfo().GetSubmissionDate().AsTime())

	// if score is less than limit and it's not yet approved, update slip days if deadline has passed
	if submission.GetScore() < assignment.GetScoreLimit() && !approved && sinceDeadline > 0 {
		updateUsedSlipDays(m, assignment.GetID(), usedSlipDaysSinceDeadline(sinceDeadline))
	}
	return nil
}

// usedSlipDaysSinceDeadline returns the number of slip days used for the time
// elapsed since an assignment deadline.
func usedSlipDaysSinceDeadline(sinceDeadline time.Duration) uint32 {
	slipDays := uint32(sinceDeadline / days)
	if sinceDeadline%days > gracePeriod {
		slipDays++
	}
	return slipDays
}

// updateUsedSlipDays updates the number of slip days used for the given assignment,
// adding a new entry if one doesn't already exist.
func updateUsedSlipDays(m slipDayHolder, assignmentID uint64, slipDays uint32) {
	for _, val := range m.GetUsedSlipDays() {
		if val.GetAssignmentID() == assignmentID {
			val.UsedDays = slipDays
			return
		}
	}
	m.addUsedSlipDays(assignmentID, slipDays)
}

// totalSlipDays returns the total number of slip days used.
func totalSlipDays(m slipDayHolder) uint32 {
	var total uint32
	for _, val := range m.GetUsedSlipDays() {
		total += val.GetUsedDays()
	}
	return total
}

// remainingSlipDays returns the remaining number of slip days for the given course.
// Note that if the returned amount is negative, all slip days have been used up.
func remainingSlipDays(m slipDayHolder, c *Course) int32 {
	if m.GetCourseID() != c.GetID() {
		return 0
	}
	return int32(c.GetSlipDays()) - int32(totalSlipDays(m))
}

// setSlipDays updates the SlipDaysRemaining field.
func setSlipDays(m slipDayHolder, c *Course) {
	remaining := remainingSlipDays(m, c)
	if remaining < 0 {
		m.setSlipDaysRemaining(0)
	} else {
		m.setSlipDaysRemaining(uint32(remaining))
	}
}
