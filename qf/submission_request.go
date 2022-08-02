package qf

// IncludeAll returns true if all assignment types should be returned.
func (req *SubmissionsForCourseRequest) IncludeAll() bool {
	return req.GetType() == SubmissionsForCourseRequest_ALL
}

// Include returns true if the given assignment should be included.
func (req *SubmissionsForCourseRequest) Include(assignment *Assignment) bool {
	switch req.GetType() {
	case SubmissionsForCourseRequest_ALL:
		return true
	case SubmissionsForCourseRequest_GROUP:
		return assignment.IsGroupLab
	case SubmissionsForCourseRequest_INDIVIDUAL:
		return !assignment.IsGroupLab
	default:
		return false
	}
}
