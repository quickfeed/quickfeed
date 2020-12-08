package ag

func (s *Submission) IsApproved() bool {
	return s.GetStatus() == Submission_APPROVED
}
