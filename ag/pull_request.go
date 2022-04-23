package ag

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
	return pr.Stage == PullRequest_APPROVED
}

// HasReviewers returns true if a pull request is in the approved or review stage.
// This indicates that it should have had reviewers assigned to it.
func (pr *PullRequest) HasReviewers() bool {
	return pr.Stage == PullRequest_APPROVED || pr.Stage == PullRequest_REVIEW
}

// Checks if a pull request is valid for creation.
func (pr *PullRequest) IsValid() bool {
	return pr.ExternalRepositoryID > 0 && pr.TaskID > 0 &&
		pr.IssueID > 0 && pr.SourceBranchName != "" && pr.Number > 0 &&
		pr.UserID > 0
}
