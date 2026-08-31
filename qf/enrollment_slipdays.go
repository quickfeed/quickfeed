package qf

// UpdateSlipDays updates the number of slip days for the given assignment/submission.
func (m *Enrollment) UpdateSlipDays(assignment *Assignment, submission *Submission) error {
	return updateSlipDays(m, assignment, submission, submission.IsApproved(m.GetUserID()))
}

func (m *Enrollment) addUsedSlipDays(assignmentID uint64, usedDays uint32) {
	m.UsedSlipDays = append(m.GetUsedSlipDays(), &UsedSlipDays{
		AssignmentID: assignmentID,
		EnrollmentID: m.GetID(),
		UsedDays:     usedDays,
	})
}

func (m *Enrollment) setSlipDaysRemaining(remaining uint32) {
	m.SlipDaysRemaining = remaining
}

// RemainingSlipDays returns the remaining number of slip days for this
// user/course enrollment. Note that if the returned amount is negative,
// the user has used up all slip days.
func (m *Enrollment) RemainingSlipDays(c *Course) int32 {
	return remainingSlipDays(m, c)
}

// SetSlipDays updates SlipDaysRemaining field of an enrollment.
func (m *Enrollment) SetSlipDays(c *Course) {
	setSlipDays(m, c)
}

// compile-time assertion for interface compliance.
var _ slipDayHolder = (*Enrollment)(nil)
