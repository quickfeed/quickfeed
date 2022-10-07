package qf

// IsValid on void message always returns true.
func (*Void) IsValid() bool {
	return true
}

// IsValid checks required fields of a group request
func (grp *Group) IsValid() bool {
	return grp.GetName() != "" && grp.GetCourseID() > 0
}

// IsValid checks required fields of a course request
func (c *Course) IsValid() bool {
	return c.GetName() != "" &&
		c.GetCode() != "" &&
		(c.GetProvider() == "github" || c.GetProvider() == "fake") &&
		c.GetOrganizationID() != 0 &&
		c.GetYear() != 0 &&
		c.GetTag() != ""
}

// IsValid checks required fields of a user request
func (u *User) IsValid() bool {
	return u.GetID() > 0
}

// IsValid ensures that user ID is set
func (u *UserRequest) IsValid() bool {
	return u.GetUserID() > 0
}

// IsValid checks required fields of an enrollment request.
func (req *Enrollment) IsValid() bool {
	return req.GetStatus() <= Enrollment_TEACHER &&
		req.GetUserID() > 0 && req.GetCourseID() > 0
}

// IsValid ensures that course ID is set
func (req *CourseRequest) IsValid() bool {
	return req.GetCourseID() > 0
}

// IsValid ensures that user ID is set
func (req *EnrollmentStatusRequest) IsValid() bool {
	return req.GetUserID() > 0
}

// IsValid checks whether OrgRequest fields are valid
func (req *OrgRequest) IsValid() bool {
	return req.GetOrgName() != ""
}

// IsValid checks that all requested repo types are valid types and course ID field is set
func (req *URLRequest) IsValid() bool {
	if req.GetCourseID() < 1 {
		return false
	}
	for _, r := range req.GetRepoTypes() {
		if r <= Repository_NONE {
			return false
		}
	}
	return true
}

// IsValid checks that the request has positive course ID
// and either user ID or group ID is set
func (req *RepositoryRequest) IsValid() bool {
	uid, gid := req.GetUserID(), req.GetGroupID()
	return req.GetCourseID() > 0 &&
		(uid == 0 && gid > 0) ||
		(uid > 0 && gid == 0)
}

// IsValid checks required fields of an action request.
// It must have a positive course ID and
// a positive user ID or group ID but not both.
func (req *SubmissionRequest) IsValid() bool {
	uid, gid := req.GetUserID(), req.GetGroupID()
	return req.GetCourseID() > 0 &&
		(uid == 0 && gid > 0) ||
		(uid > 0 && gid == 0)
}

// IsValid ensures that both submission and course IDs are set
func (req *UpdateSubmissionRequest) IsValid() bool {
	return req.GetCourseID() > 0 && req.GetSubmissionID() > 0
}

// IsValid ensures that group ID is provided
func (req *GetGroupRequest) IsValid() bool {
	return req.GetGroupID() > 0
}

// IsValid ensures that course ID and group or user IDs are set
func (req *GroupRequest) IsValid() bool {
	uid, gid := req.GetUserID(), req.GetGroupID()
	return (uid > 0 || gid > 0) && req.GetCourseID() > 0
}

// IsValid checks that course ID is positive.
func (req *EnrollmentRequest) IsValid() bool {
	return req.GetCourseID() > 0
}

// IsValid ensures that course ID is provided.
func (req *SubmissionsForCourseRequest) IsValid() bool {
	return req.GetCourseID() != 0
}

// IsValid ensures that both course and assignment IDs are set.
func (req *RebuildRequest) IsValid() bool {
	aid, cid := req.GetAssignmentID(), req.GetCourseID()
	return aid > 0 && cid > 0
}

// IsValid checks that either ID or path field is set
func (org *Organization) IsValid() bool {
	id, path := org.GetID(), org.GetName()
	return id > 0 || path != ""
}

// IsValid ensures that course ID and submission ID are present.
func (req *SubmissionReviewersRequest) IsValid() bool {
	return req.CourseID > 0 && req.SubmissionID > 0
}

// IsValid ensures that a review always has a reviewer and a submission IDs.
func (r *Review) IsValid() bool {
	return r.ReviewerID > 0 && r.SubmissionID > 0
}

// IsValid ensures that course ID is provided and the review is valid.
func (r *ReviewRequest) IsValid() bool {
	return r.CourseID > 0 && r.Review.IsValid()
}

// IsValid ensures that a grading benchmark always belongs to an assignment
// and is not empty.
func (bm *GradingBenchmark) IsValid() bool {
	return bm.AssignmentID > 0 && bm.Heading != ""
}

// IsValid ensures that a criterion always belongs to a grading benchmark
// and is not empty.
func (c *GradingCriterion) IsValid() bool {
	return c.BenchmarkID > 0 && c.Description != ""
}

// IsValid ensures that course code, year, and student login are set
func (r *CourseUserRequest) IsValid() bool {
	return r.CourseCode != "" && r.UserLogin != "" && r.CourseYear > 2019
}

func (m *Enrollments) IsValid() bool {
	if len(m.Enrollments) == 0 {
		return false
	}
	for _, e := range m.Enrollments {
		if !e.IsValid() {
			return false
		}
	}
	return m.HasCourseID()
}
