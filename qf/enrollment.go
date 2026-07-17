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

// UpdateTotalApproved updates the total approved assignments for the current enrollment
// based on the provided submissions.
func (m *Enrollment) UpdateTotalApproved(submissions []*Submission) {
	var totalApproved uint64
	duplicateAssignments := make(map[uint64]struct{})
	for _, s := range submissions {
		// Ignore duplicate approved assignments
		if _, ok := duplicateAssignments[s.GetAssignmentID()]; ok {
			continue
		}
		if s.IsApproved(m.GetUserID()) {
			duplicateAssignments[s.GetAssignmentID()] = struct{}{}
			totalApproved++
		}
	}
	m.TotalApproved = totalApproved
}
