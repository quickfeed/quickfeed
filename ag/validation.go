package ag

func (grp Group) IsValidGroup() bool {
	return grp.Name != "" &&
		grp.CourseId > 0 &&
		len(grp.Users) > 0
}

func (c Course) IsValidCourse() bool {
	return c.Name != "" &&
		c.Code != "" &&
		(c.Provider == "github" || c.Provider == "gitlab" || c.Provider == "fake") &&
		c.DirectoryId != 0 &&
		c.Year != 0 &&
		c.Tag != ""
}

func (req ActionRequest) IsValidEnrollment() bool {
	return req.Status <= Enrollment_TEACHER &&
		req.UserId != 0 &&
		req.CourseId != 0
}
