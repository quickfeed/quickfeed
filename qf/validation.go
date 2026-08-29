package qf

// IsValid on void message always returns true.
func (*Void) IsValid() bool {
	return true
}

// IsValid ensures that CourseID is set, and that the group has a name and at least one user.
func (grp *Group) IsValid() bool {
	return grp.GetCourseID() > 0 && grp.GetName() != "" && len(grp.GetUsers()) > 0
}

// IsValid ensures that all required fields of a course are set.
func (c *Course) IsValid() bool {
	return c.GetName() != "" &&
		c.GetCode() != "" &&
		c.GetScmOrganizationID() != 0 &&
		c.GetYear() != 0 &&
		c.GetTag() != ""
}

// IsValid ensures that UserID is set.
func (u *User) IsValid() bool {
	return u.GetID() > 0
}

// IsValid ensures that CourseID and UserID are set, and that the enrollment status is valid.
func (req *Enrollment) IsValid() bool {
	status := req.GetStatus()
	return req.GetCourseID() > 0 &&
		req.GetUserID() > 0 &&
		status >= Enrollment_NONE &&
		status <= Enrollment_TEACHER
}

// IsValid ensures that CourseID is set.
func (req *CourseRequest) IsValid() bool {
	return req.GetCourseID() > 0
}

// IsValid ensures that CourseID is set and either UserID or GroupID is set, but not both.
func (req *RepositoryRequest) IsValid() bool {
	uid, gid := req.GetUserID(), req.GetGroupID()
	return req.GetCourseID() > 0 && ((uid > 0) != (gid > 0))
}

// IsValid ensures that CourseID is set and one of SubmissionID, UserID, GroupID, or Type is set.
func (req *SubmissionRequest) IsValid() bool {
	if req.GetCourseID() == 0 {
		return false // invalid: CourseID must be set
	}
	switch req.GetFetchMode().(type) {
	case nil:
		return false
	case *SubmissionRequest_SubmissionID:
		return req.GetSubmissionID() > 0
	case *SubmissionRequest_UserID:
		return req.GetUserID() > 0
	case *SubmissionRequest_GroupID:
		return req.GetGroupID() > 0
	default: // *SubmissionRequest_Type, requires only CourseID
		return true
	}
}

// IsValid ensures that SubmissionID is set.
// If UserID is 0, the grade is for a group submission.
func (req *Grade) IsValid() bool {
	return req.GetSubmissionID() > 0
}

// IsValid ensures that CourseID is set and either UserID or GroupID is set, but not both.
func (req *GroupRequest) IsValid() bool {
	uid, gid := req.GetUserID(), req.GetGroupID()
	return req.GetCourseID() > 0 && ((uid > 0) != (gid > 0))
}

// IsValid ensures that either CourseID or UserID is set, but not both.
func (req *EnrollmentRequest) IsValid() bool {
	switch req.GetFetchMode().(type) {
	case *EnrollmentRequest_CourseID:
		return req.GetCourseID() > 0
	case *EnrollmentRequest_UserID:
		return req.GetUserID() > 0
	}
	return false
}

// IsValid ensures that both CourseID and AssignmentID are set.
func (req *RebuildRequest) IsValid() bool {
	aid, cid := req.GetAssignmentID(), req.GetCourseID()
	return aid > 0 && cid > 0
}

// IsValid ensures that CourseID is set, and that From is not after To when both are given.
// The handler clamps the interval and limit to their server-enforced maximums, rather than
// rejecting a request for exceeding them.
func (req *CourseLogRequest) IsValid() bool {
	if req.GetCourseID() == 0 {
		return false
	}
	from, to := req.GetFrom(), req.GetTo()
	if from == nil || to == nil {
		return true
	}
	return !from.AsTime().After(to.AsTime())
}

// IsValid ensures that an SCM organization name is set.
func (org *Organization) IsValid() bool {
	// only check the name; the ID is only used in the response
	return org.GetScmOrganizationName() != ""
}

// IsValid ensures that both ReviewerID and SubmissionID are set.
func (r *Review) IsValid() bool {
	return r.GetReviewerID() > 0 && r.GetSubmissionID() > 0
}

// IsValid ensures that CourseID is set and the review is valid.
func (r *ReviewRequest) IsValid() bool {
	return r.GetCourseID() > 0 && r.GetReview().IsValid()
}

// IsValid ensures that CourseID and exactly one target are set.
func (r *NotesRequest) IsValid() bool {
	return r.GetCourseID() > 0 && exactlyOneTarget(r.GetSubmissionID(), r.GetGroupID(), r.GetEnrollmentID())
}

// IsValid ensures that CourseID is set.
// Note is the request type for create, update, and delete, and the three differ
// on what else they require: create needs a body and a target but no ID, update
// needs an ID and a body, and delete needs only an ID. CourseID is all they have
// in common, so the handlers check the rest.
func (n *Note) IsValid() bool {
	return n.GetCourseID() > 0
}

// HasSingleTarget reports whether the note references exactly one submission,
// group, or enrollment.
func (n *Note) HasSingleTarget() bool {
	return exactlyOneTarget(n.GetSubmissionID(), n.GetGroupID(), n.GetEnrollmentID())
}

// exactlyOneTarget returns true if exactly one of the given target IDs is set.
func exactlyOneTarget(ids ...uint64) bool {
	targets := 0
	for _, id := range ids {
		if id > 0 {
			targets++
		}
	}
	return targets == 1
}

// IsValid ensures that the grading benchmark has an AssignmentID and a heading.
func (bm *GradingBenchmark) IsValid() bool {
	return bm.GetAssignmentID() > 0 && bm.GetHeading() != ""
}

// IsValid ensures that the grading criterion has a BenchmarkID and a description.
func (c *GradingCriterion) IsValid() bool {
	return c.GetBenchmarkID() > 0 && c.GetDescription() != ""
}

// IsValid ensures that assignment feedback has CourseID and AssignmentID set and has content.
func (f *AssignmentFeedback) IsValid() bool {
	return f.GetCourseID() > 0 &&
		f.GetAssignmentID() > 0 &&
		f.GetLikedContent() != "" &&
		f.GetImprovementSuggestions() != "" &&
		f.GetTimeSpent() > 0
}

// IsValid ensures that all enrollments are valid and belong to the same course.
func (m *Enrollments) IsValid() bool {
	if len(m.GetEnrollments()) == 0 {
		return false
	}
	for _, e := range m.GetEnrollments() {
		if !e.IsValid() {
			return false
		}
	}
	return m.HasCourseID()
}
