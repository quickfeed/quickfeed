package qf

import "fmt"

// UpdateSlipDays updates the number of slip days for the given assignment/submission.
// This method is for group submissions.
func (m *Group) UpdateSlipDays(assignment *Assignment, submission *Submission) error {
	if m.GetCourseID() != assignment.GetCourseID() {
		return fmt.Errorf("invariant violation (group.GetCourseID() != assignment.GetCourseID()) (%d != %d)", m.GetCourseID(), assignment.GetCourseID())
	}
	if assignment.GetID() != submission.GetAssignmentID() {
		return fmt.Errorf("invariant violation (assignment.GetID() != submission.GetAssignmentID()) (%d != %d)", assignment.GetID(), submission.GetAssignmentID())
	}
	sinceDeadline := assignment.SinceDeadline(submission.GetBuildInfo().GetSubmissionDate().AsTime())

	// if score is less than limit and it's not yet approved, update slip days if deadline has passed
	if submission.GetScore() < assignment.GetScoreLimit() && !submission.IsAllApproved() && sinceDeadline > 0 {
		// deadline exceeded; calculate used slip days for this assignment
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
func (m *Group) internalUpdateSlipDays(assignmentID uint64, slipDays uint32) {
	for _, val := range m.GetUsedSlipDays() {
		if val.GetAssignmentID() == assignmentID {
			val.UsedDays = slipDays
			return
		}
	}
	// not found; add new entry to the slice
	m.UsedSlipDays = append(m.GetUsedSlipDays(), &UsedSlipDays{
		AssignmentID: assignmentID,
		GroupID:      m.GetID(),
		UsedDays:     slipDays,
	})
}

// totalSlipDays returns the total number of slip days used for this group.
func (m *Group) totalSlipDays() uint32 {
	var total uint32
	for _, val := range m.GetUsedSlipDays() {
		total += val.GetUsedDays()
	}
	return total
}

// RemainingSlipDays returns the remaining number of slip days for this
// group/course. Note that if the returned amount is negative,
// the group has used up all slip days.
func (m *Group) RemainingSlipDays(c *Course) int32 {
	if m.GetCourseID() != c.GetID() {
		return 0
	}
	return int32(c.GetSlipDays() - m.totalSlipDays())
}

// SetSlipDays updates SlipDaysRemaining field of a group.
func (m *Group) SetSlipDays(c *Course) {
	remaining := m.RemainingSlipDays(c)
	if remaining < 0 {
		m.SlipDaysRemaining = 0
	} else {
		m.SlipDaysRemaining = uint32(remaining)
	}
}
