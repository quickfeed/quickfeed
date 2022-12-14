package qf

// IncludeAll returns true if all assignment types should be returned.
func (req *SubmissionRequest) IncludeAll() bool {
	return req.GetType() == SubmissionRequest_ALL
}

// Include returns true if the given assignment should be included.
func (req *SubmissionRequest) Include(assignment *Assignment) bool {
	switch req.GetType() {
	case SubmissionRequest_ALL:
		return true
	case SubmissionRequest_GROUP:
		return assignment.IsGroupLab
	case SubmissionRequest_USER:
		return !assignment.IsGroupLab
	default:
		return false
	}
}
