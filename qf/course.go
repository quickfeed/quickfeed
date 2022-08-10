package qf

// cache of access tokens for courses; they are cached here when fetching from database
var accessTokens = make(map[uint64]string)

// SetAccessTokenForCourse for the given course.
func SetAccessTokenForCourse(courseID uint64, accessToken string) {
	accessTokens[courseID] = accessToken
}

// GetAccessTokenForCourse returns the access token for the course.
func (course *Course) GetAccessTokenForCourse() string {
	return accessTokens[course.GetID()]
}

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
