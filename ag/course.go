package ag

// cache of access tokens for courses; they are cached here when fetching from database
var accessTokens = make(map[uint64]string)

// SetAccessToken for the given course.
func SetAccessToken(courseID uint64, accessToken string) {
	accessTokens[courseID] = accessToken
}

func (course *Course) GetAccessToken() string {
	return accessTokens[course.GetID()]
}
