package ag

// IsValidGroup checks required fields of a group request
func (grp Group) IsValidGroup() bool {
	return grp.Name != "" &&
		grp.CourseID > 0 &&
		len(grp.Users) > 0
}

// IsValidCourse checks required fields of a course request
func (c Course) IsValidCourse() bool {
	return c.Name != "" &&
		c.Code != "" &&
		(c.Provider == "github" || c.Provider == "gitlab" || c.Provider == "fake") &&
		c.DirectoryID != 0 &&
		c.Year != 0 &&
		c.Tag != ""
}

// IsValidUser chacks required fields of a user request
func (u User) IsValidUser() bool {
	return u.ID > 0
}

// IsValidEnrollment checks required fields of an enrollment request
func (req ActionRequest) IsValidEnrollment() bool {
	return req.Status <= Enrollment_Teacher &&
		req.UserID != 0 &&
		req.CourseID != 0
}

// IsValidRequest checks whether RecordRequest fields are valid
func (req RecordRequest) IsValidRequest() bool {
	return req.ID > 0
}

// IsValidRepoRequest checks required fields of a repository request
func (req RepositoryRequest) IsValidRepoRequest() bool {
	return req.CourseID > 0 &&
		req.Type <= Repository_CourseInfo
}

func (req ActionRequest) IsValidRequest() bool {
	return (req.UserID > 0 || req.GroupID > 0) &&
		req.CourseID > 0
}
