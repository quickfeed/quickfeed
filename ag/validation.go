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
	return req.GetStatus() <= Enrollment_Teacher &&
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
		req.GetType() <= Repository_CourseInfo
}

// IsValidRequest checks required fields of an update group request
func (req ActionRequest) IsValidRequest() bool {
	return (req.GetUserID() > 0 || req.GetGroupID() > 0) &&
		req.GetCourseID() > 0
}
