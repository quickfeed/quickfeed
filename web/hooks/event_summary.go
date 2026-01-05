package hooks

import "github.com/google/go-github/v62/github"

// PushEventSummary contains the essential fields from a GitHub push event for logging
type PushEventSummary struct {
	Type          string        `json:"type"`
	Ref           string        `json:"ref"`
	RepoID        int64         `json:"repo_id"`
	RepoName      string        `json:"repo_name"`
	DefaultBranch string        `json:"default_branch"`
	CommitID      string        `json:"commit_id"`
	Sender        string        `json:"sender"`
	CommitCount   int           `json:"commit_count"`
	Commits       []CommitFiles `json:"commits,omitempty"`
}

// CommitFiles contains the file changes for a single commit
type CommitFiles struct {
	Modified []string `json:"modified,omitempty"`
	Added    []string `json:"added,omitempty"`
	Removed  []string `json:"removed,omitempty"`
}

// summarizePushEvent extracts only the fields needed for logging and processing from a push event.
// This reduces log size by removing unnecessary metadata (URLs, timestamps, author details, etc.)
// while preserving all information used by QuickFeed's webhook handlers.
func summarizePushEvent(payload *github.PushEvent) *PushEventSummary {
	commits := make([]CommitFiles, 0, len(payload.GetCommits()))
	for _, commit := range payload.GetCommits() {
		commits = append(commits, CommitFiles{
			Modified: commit.Modified,
			Added:    commit.Added,
			Removed:  commit.Removed,
		})
	}

	return &PushEventSummary{
		Type:          "push",
		Ref:           payload.GetRef(),
		RepoID:        payload.GetRepo().GetID(),
		RepoName:      payload.GetRepo().GetFullName(),
		DefaultBranch: payload.GetRepo().GetDefaultBranch(),
		CommitID:      payload.GetHeadCommit().GetID(),
		Sender:        payload.GetSender().GetLogin(),
		CommitCount:   len(payload.GetCommits()),
		Commits:       commits,
	}
}

// PullRequestEventSummary contains the essential fields from a GitHub pull request event for logging
type PullRequestEventSummary struct {
	Type      string `json:"type"`
	Action    string `json:"action"`
	RepoID    int64  `json:"repo_id"`
	RepoName  string `json:"repo_name"`
	Number    int    `json:"pr_number"`
	Title     string `json:"pr_title,omitempty"`
	Body      string `json:"pr_body,omitempty"`
	Merged    bool   `json:"merged,omitempty"`
	SourceRef string `json:"source_ref,omitempty"`
	Sender    string `json:"sender"`
}

// PullRequestReviewEventSummary contains the essential fields from a GitHub pull request review event for logging
type PullRequestReviewEventSummary struct {
	Type        string `json:"type"`
	Action      string `json:"action"`
	RepoID      int64  `json:"repo_id"`
	RepoName    string `json:"repo_name"`
	OrgID       int64  `json:"org_id"`
	Number      int    `json:"pr_number"`
	Title       string `json:"pr_title"`
	ReviewState string `json:"review_state"`
	Sender      string `json:"sender"`
}

// summarizePullRequestEvent extracts only the fields needed for logging and processing from a pull request event.
func summarizePullRequestEvent(payload *github.PullRequestEvent) *PullRequestEventSummary {
	summary := &PullRequestEventSummary{
		Type:     "pull_request",
		Action:   payload.GetAction(),
		RepoID:   payload.GetRepo().GetID(),
		RepoName: payload.GetRepo().GetFullName(),
		Number:   payload.GetNumber(),
		Sender:   payload.GetSender().GetLogin(),
	}

	// Include fields specific to certain actions
	switch payload.GetAction() {
	case "opened":
		summary.Title = payload.GetPullRequest().GetTitle()
		summary.Body = payload.GetPullRequest().GetBody()
		summary.SourceRef = payload.GetPullRequest().GetHead().GetRef()
	case "closed":
		summary.Title = payload.GetPullRequest().GetTitle()
		summary.Merged = payload.GetPullRequest().GetMerged()
	}

	return summary
}

// summarizePullRequestReviewEvent extracts only the fields needed for logging and processing from a pull request review event.
func summarizePullRequestReviewEvent(payload *github.PullRequestReviewEvent) *PullRequestReviewEventSummary {
	return &PullRequestReviewEventSummary{
		Type:        "pull_request_review",
		Action:      payload.GetAction(),
		RepoID:      payload.GetRepo().GetID(),
		RepoName:    payload.GetRepo().GetFullName(),
		OrgID:       payload.GetOrganization().GetID(),
		Number:      payload.GetPullRequest().GetNumber(),
		Title:       payload.GetPullRequest().GetTitle(),
		ReviewState: payload.GetReview().GetState(),
		Sender:      payload.GetSender().GetLogin(),
	}
}

// summarizeEvent summarizes any GitHub event for logging purposes.
func summarizeEvent(event any) any {
	switch e := event.(type) {
	case *github.PushEvent:
		return summarizePushEvent(e)
	case *github.PullRequestEvent:
		return summarizePullRequestEvent(e)
	case *github.PullRequestReviewEvent:
		return summarizePullRequestReviewEvent(e)
	default:
		return event
	}
}
