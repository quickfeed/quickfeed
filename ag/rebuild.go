package ag

// SetSubmissionID sets the submission ID to be rebuilt.
// Setting the submission ID will clear the CourseID field.
func (r *RebuildRequest) SetSubmissionID(id uint64) {
	r.RebuildType = &RebuildRequest_SubmissionID{SubmissionID: id}
}

// SetCourseID sets the course ID to be rebuilt.
// Setting the course ID will clear the SubmissionID field.
func (r *RebuildRequest) SetCourseID(id uint64) {
	r.RebuildType = &RebuildRequest_CourseID{CourseID: id}
}
