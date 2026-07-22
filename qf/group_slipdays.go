package qf

// UpdateSlipDays updates the number of slip days for the given assignment/submission.
// This method is for group submissions.
func (m *Group) UpdateSlipDays(assignment *Assignment, submission *Submission) error {
	return updateSlipDays(m, assignment, submission, submission.IsAllApproved())
}

func (m *Group) addUsedSlipDays(assignmentID uint64, usedDays uint32) {
	m.UsedSlipDays = append(m.GetUsedSlipDays(), &UsedSlipDays{
		AssignmentID: assignmentID,
		GroupID:      m.GetID(),
		UsedDays:     usedDays,
	})
}

func (m *Group) setSlipDaysRemaining(remaining uint32) {
	m.SlipDaysRemaining = remaining
}

// RemainingSlipDays returns the remaining number of slip days for this
// group/course. Note that if the returned amount is negative,
// the group has used up all slip days.
func (m *Group) RemainingSlipDays(c *Course) int32 {
	return remainingSlipDays(m, c)
}

// SetSlipDays updates SlipDaysRemaining field of a group.
func (m *Group) SetSlipDays(c *Course) {
	setSlipDays(m, c)
}

// compile-time assertion for interface compliance.
var _ slipDayHolder = (*Group)(nil)
