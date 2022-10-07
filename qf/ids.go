package qf

// IDFor returns zero.
func (*Void) IDFor(_ string) uint64 {
	return 0
}

// IDFor returns course ID.
func (r *Course) IDFor(_ string) uint64 {
	return r.GetID()
}

// IDFor returns user ID.
func (r *User) IDFor(_ string) uint64 {
	return r.GetID()
}

// IDFor returns user ID.
func (r *Enrollment) IDFor(_ string) uint64 {
	return r.GetUserID()
}

// IDFor returns course ID.
func (r *Enrollments) IDFor(_ string) uint64 {
	return r.GetCourseID()
}

// IDFor returns course ID.
func (r *Group) IDFor(_ string) uint64 {
	return r.GetCourseID()
}

// IDFor returns user, group, or course ID.
func (r *GroupRequest) IDFor(role string) uint64 {
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

// IDFor returns group ID.
func (r *GetGroupRequest) IDFor(_ string) uint64 {
	return r.GetGroupID()
}

// IDFor returns course ID.
func (r *CourseRequest) IDFor(_ string) uint64 {
	return r.GetCourseID()
}

// IDFor returns course ID.
func (r *EnrollmentRequest) IDFor(_ string) uint64 {
	return r.GetCourseID()
}

// IDFor returns user ID.
func (r *EnrollmentStatusRequest) IDFor(_ string) uint64 {
	return r.GetUserID()
}

// IDFor returns user, group, or course ID.
func (r *SubmissionRequest) IDFor(role string) uint64 {
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

// IDFor returns course ID.
func (r *UpdateSubmissionsRequest) IDFor(_ string) uint64 {
	return r.GetCourseID()
}

// IDFor returns course ID.
func (r *SubmissionsForCourseRequest) IDFor(_ string) uint64 {
	return r.GetCourseID()
}

// IDFor returns course ID.
func (r *UpdateSubmissionRequest) IDFor(_ string) uint64 {
	return r.GetCourseID()
}

// IDFor returns course ID.
func (r *RebuildRequest) IDFor(_ string) uint64 {
	return r.GetCourseID()
}

// IDFor returns course ID.
func (r *RepositoryRequest) IDFor(_ string) uint64 {
	return r.GetCourseID()
}

// IDFor returns course ID.
func (r *ReviewRequest) IDFor(_ string) uint64 {
	return r.GetCourseID()
}

// IDFor returns course ID.
func (r *SubmissionReviewersRequest) IDFor(_ string) uint64 {
	return r.GetCourseID()
}

// IDFor returns course ID.
func (r *URLRequest) IDFor(_ string) uint64 {
	return r.GetCourseID()
}

// IDFor returns course ID.
func (r *GradingBenchmark) IDFor(_ string) uint64 {
	return r.GetCourseID()
}

// IDFor returns course ID.
func (r *GradingCriterion) IDFor(_ string) uint64 {
	return r.GetCourseID()
}

// IDFor returns 0, this request is only used by admins.
func (*OrgRequest) IDFor(_ string) uint64 {
	return 0
}

// IDFor returns 0.
func (*CourseUserRequest) IDFor(_ string) uint64 {
	return 0
}
