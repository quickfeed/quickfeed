package qf

// FetchID returns course ID
func (r *Course) FetchID(_ string) uint64 {
	return r.GetID()
}

// FetchID returns user ID
func (u *User) FetchID(_ string) uint64 {
	return u.GetID()
}

// FetchID returns course ID
func (r *Enrollments) FetchID(_ string) uint64 {
	return r.GetCourseID()
}

// FetchID returns course ID
func (r *Group) FetchID(_ string) uint64 {
	return r.GetCourseID()
}

// FetchID returns user or course ID
func (r *GroupRequest) FetchID(role string) uint64 {
	switch role {
	case "user":
		return r.GetUserID()
	case "course":
		return r.GetCourseID()
	}
	return 0
}

// FetchID returns group ID
func (r *GetGroupRequest) FetchID(_ string) uint64 {
	return r.GetGroupID()
}

// FetchID returns course ID
func (r *CourseRequest) FetchID(_ string) uint64 {
	return r.GetCourseID()
}

// FetchID returns course ID
func (r *EnrollmentRequest) FetchID(_ string) uint64 {
	return r.GetCourseID()
}

// FetchID returns user ID
func (r *EnrollmentStatusRequest) FetchID(_ string) uint64 {
	return r.GetUserID()
}

// FetchID returns user, group, or course ID
func (r *SubmissionRequest) FetchID(role string) uint64 {
	switch role {
	case "user":
		return r.GetUserID()
	case "group":
		return r.GetGroupID()
	case "course":
		return r.GetCourseID()
	}
	return 0
}

// FetchID returns course ID
func (r *SubmissionsForCourseRequest) FetchID(_ string) uint64 {
	return r.GetCourseID()
}

// FetchID returns course ID
func (r *UpdateSubmissionRequest) FetchID(_ string) uint64 {
	return r.GetCourseID()
}

// FetchID returns course ID
func (r *RebuildRequest) FetchID(_ string) uint64 {
	return r.GetCourseID()
}

// FetchID returns course ID
func (r *RepositoryRequest) FetchID(_ string) uint64 {
	return r.GetCourseID()
}

// FetchID returns course ID
func (r *ReviewRequest) FetchID(_ string) uint64 {
	return r.GetCourseID()
}

// FetchID returns course ID
func (r *SubmissionReviewersRequest) FetchID(_ string) uint64 {
	return r.GetCourseID()
}
