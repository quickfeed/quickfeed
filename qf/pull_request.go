package qf

func (pr *PullRequest) SetApproved() {
	pr.Stage = PullRequest_APPROVED
}

func (pr *PullRequest) SetReview() {
	pr.Stage = PullRequest_REVIEW
}

func (pr *PullRequest) SetDraft() {
	pr.Stage = PullRequest_DRAFT
}

// IsApproved returns true if a pull request is at the approved stage.
func (pr *PullRequest) IsApproved() bool {
	return pr.GetStage() == PullRequest_APPROVED
}

// HasReviewers returns true if a pull request is in the approved or review stage.
// This implies that it should have had reviewers assigned to it.
func (pr *PullRequest) HasReviewers() bool {
	return pr.GetStage() == PullRequest_APPROVED || pr.GetStage() == PullRequest_REVIEW
}

// HasFeedbackComment returns true if the pull request has a comment associated with it.
func (pr *PullRequest) HasFeedbackComment() bool {
	return pr.GetScmCommentID() > 0
}

// Checks if a pull request is valid for creation.
func (pr *PullRequest) Valid() bool {
	return pr.GetScmRepositoryID() > 0 && pr.GetTaskID() > 0 &&
		pr.GetIssueID() > 0 && pr.GetSourceBranch() != "" && pr.GetNumber() > 0 &&
		pr.GetUserID() > 0
}
