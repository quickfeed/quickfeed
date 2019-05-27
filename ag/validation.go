package ag

func (grp Group) IsValidGroup() bool {
	return grp.Name != "" &&
		grp.Course_ID > 0 &&
		len(grp.Users) > 0
}

func (c Course) IsValidCourse() bool {
	return c.Name != "" &&
		c.Code != "" &&
		(c.Provider == "github" || c.Provider == "gitlab" || c.Provider == "fake") &&
		c.Directory_ID != 0 &&
		c.Year != 0 &&
		c.Tag != ""
}

func (req ActionRequest) IsValidEnrollment() bool {
	return req.Status <= Enrollment_Teacher &&
		req.User_ID != 0 &&
		req.Course_ID != 0
}
