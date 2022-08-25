package qf

// SetSlipDays sets number of remaining slip days for each course enrollment
func (course *Course) SetSlipDays() {
	for _, e := range course.Enrollments {
		e.SetSlipDays(course)
	}

	for _, g := range course.Groups {
		g.SetSlipDays(course)
	}
}

func (course *Course) TeacherEnrollments() []*Enrollment {
	enrolledTeachers := []*Enrollment{}
	for _, enrollment := range course.Enrollments {
		if enrollment.IsTeacher() {
			enrolledTeachers = append(enrolledTeachers, enrollment)
		}
	}
	return enrolledTeachers
}

// Dummy implementation of the interceptor.userIDs interface.
// Marks this message type to be evaluated for token refresh.
func (*Course) UserIDs() []uint64 {
	return []uint64{}
}
