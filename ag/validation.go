package ag

// IsValidGroup checks required fields of a group request
func (grp Group) IsValidGroup() bool {
	return grp.GetName() != "" &&
		grp.GetCourseID() > 0 &&
		len(grp.GetUsers()) > 0
}

// IsValidCourse checks required fields of a course request
func (c Course) IsValidCourse() bool {
	return c.GetName() != "" &&
		c.GetCode() != "" &&
		(c.GetProvider() == "github" || c.GetProvider() == "gitlab" || c.GetProvider() == "fake") &&
		c.GetDirectoryID() != 0 &&
		c.GetYear() != 0 &&
		c.GetTag() != ""
}

// IsValidUser chacks required fields of a user request
func (u User) IsValidUser() bool {
	return u.GetID() > 0
}

// IsValidEnrollment checks required fields of an enrollment request
func (req ActionRequest) IsValidEnrollment() bool {
	return req.GetStatus() <= Enrollment_TEACHER &&
		req.GetUserID() != 0 &&
		req.GetCourseID() != 0
}

// IsValidRequest checks whether RecordRequest fields are valid
func (req RecordRequest) IsValidRequest() bool {
	return req.GetID() > 0
}

// IsValidRepoRequest checks required fields of a repository request
func (req RepositoryRequest) IsValidRepoRequest() bool {
	return req.GetCourseID() > 0 &&
		req.GetType() <= Repository_COURSEINFO
}

// IsValidRequest checks required fields of an update group request
func (req ActionRequest) IsValidRequest() bool {
	return (req.GetUserID() > 0 || req.GetGroupID() > 0) &&
		req.GetCourseID() > 0
}

// IsValidRequest checks that course ID is a positive number
func (req EnrollmentRequest) IsValidRequest() bool {
	return req.GetCourseID() > 0
}

// IsValidProvider validates provider string coming from front end
func (l Providers) IsValidProvider(provider string) bool {
	isValid := false
	for _, p := range l.GetProviders() {
		if p == provider {
			isValid = true
		}
	}
	return isValid
}
