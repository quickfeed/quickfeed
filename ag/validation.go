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

// IsValidEnrollment checks required fields of an enrollment request
func (req ActionRequest) IsValidEnrollment() bool {
	return req.Status <= Enrollment_Teacher &&
		req.UserID != 0 &&
		req.CourseID != 0
}
