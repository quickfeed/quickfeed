package ag

// cache of access tokens for courses; they are cached here when fetching from database
var accessTokens = make(map[uint64]string)

// SetAccessToken for the given course.
func SetAccessToken(courseID uint64, accessToken string) {
	accessTokens[courseID] = accessToken
}

// GetAccessToken returns the access token for the course.
func (course *Course) GetAccessToken() string {
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
