package scm

import (
	"net/http"
	"slices"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"
)

var jsonFolderContent = `[
  {
    "name": "Dockerfile",
    "path": "scripts/Dockerfile",
    "sha": "873c7550c0fc40b07cf173382bc93028f8f87c06",
    "size": 316,
    "type": "file"
  },
  {
    "name": "run.sh",
    "path": "scripts/run.sh",
    "sha": "fa3515649d92a369bb4c212760bf54b5d4d00d4e",
    "size": 1381,
    "type": "file"
  }
]`

var (
	meling  = github.User{Login: github.String("meling")}
	leslie  = github.User{Login: github.String("leslie")}
	lamport = github.User{Login: github.String("lamport")}
	jostein = github.User{Login: github.String("jostein")}
	foo     = github.User{Login: github.String("foo")} // organization (user/owner)
	bar     = github.User{Login: github.String("bar")} // organization (user/owner)
)

// MockedGithubSCM implements the SCM interface.
type MockedGithubSCM struct {
	*GithubSCM
	*mockOptions
	issueID   int64
	commentID int64
}

// NewMockedGithubSCMClient returns a mocked Github client implementing the SCM interface.
func NewMockedGithubSCMClient(logger *zap.SugaredLogger, opts ...MockOption) *MockedGithubSCM {
	mockOpts := newMockOptions()
	for _, o := range opts {
		o(mockOpts)
	}
	s := &MockedGithubSCM{
		mockOptions: mockOpts,
	}

	// initial empty issues map: owner -> repo -> issues
	s.issues = make(map[string]map[string][]github.Issue)
	for _, repo := range s.repos {
		if s.issues[*repo.Organization.Login] == nil {
			s.issues[*repo.Organization.Login] = make(map[string][]github.Issue)
		}
		s.issues[*repo.Organization.Login][*repo.Name] = make([]github.Issue, 0)
	}
	// initial empty comments map: owner -> repo -> issue ID -> comments
	s.comments = make(map[string]map[string]map[int64][]github.IssueComment)
	for org, repo := range s.issues {
		s.comments[org] = make(map[string]map[int64][]github.IssueComment)
		for re, issues := range repo {
			s.comments[org][re] = make(map[int64][]github.IssueComment)
			for _, issue := range issues {
				s.comments[org][re][issue.GetID()] = []github.IssueComment{}
			}
		}
	}

	getByIDHandler := WithRequestMatchHandler(
		getByID,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := mustParse[int64](r.PathValue("id"))
			for _, org := range s.orgs {
				if org.GetID() == id {
					mustWrite(w, org)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		}),
	)
	getOrgsByOrgHandler := WithRequestMatchHandler(
		getOrgsByOrg,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			org := r.PathValue("org")
			found := s.matchOrgFunc(org, func(o github.Organization) {
				mustWrite(w, o)
			})
			if !found {
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)
	getOrgsReposByOrgHandler := WithRequestMatchHandler(
		getOrgsReposByOrg,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			org := r.PathValue("org")
			found := s.matchOrgFunc(org, func(o github.Organization) {
				foundRepos := make([]github.Repository, 0)
				for _, repo := range s.repos {
					if repo.GetOrganization().GetLogin() == o.GetLogin() {
						foundRepos = append(foundRepos, repo)
					}
				}
				mustWrite(w, foundRepos)
			})
			if !found {
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)
	getOrgsMembershipsByOrgByUsernameHandler := WithRequestMatchHandler(
		getOrgsMembershipsByOrgByUsername,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			org := r.PathValue("org")
			username := r.PathValue("username")
			found := s.matchOrgFunc(org, func(o github.Organization) {
				for _, m := range s.members {
					if m.GetOrganization().GetLogin() == o.GetLogin() && m.GetUser().GetLogin() == username {
						mustWrite(w, m)
						return
					}
				}
				w.WriteHeader(http.StatusNotFound)
			})
			if !found {
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)
	getReposContentsByOwnerByRepoByPathHandler := WithRequestMatchHandler(
		getReposContentsByOwnerByRepoByPath,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// we only care about the owner and repo; we ignore the path component
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			if !s.hasOrgRepo(owner, repo) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			mustWrite(w, jsonFolderContent)
		}),
	)
	getReposCollaboratorsByOwnerByRepoHandler := WithRequestMatchHandler(
		getReposCollaboratorsByOwnerByRepo,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")

			collaborators := s.groups[owner][repo]
			if collaborators == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			mustWrite(w, collaborators)
		}),
	)
	putReposCollaboratorsByOwnerByRepoByUsernameHandler := WithRequestMatchHandler(
		putReposCollaboratorsByOwnerByRepoByUsername,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			username := r.PathValue("username")

			collaborators := s.groups[owner][repo]
			if collaborators == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if slices.ContainsFunc(collaborators, func(u github.User) bool { return u.GetLogin() == username }) {
				// already exists; no need to add again
				w.WriteHeader(http.StatusNoContent)
				return
			}

			ghUser := github.User{Login: github.String(username)}
			s.groups[owner][repo] = append(collaborators, ghUser)
			invite := github.CollaboratorInvitation{
				Repo:    &github.Repository{Owner: &github.User{Login: github.String(owner)}, Name: github.String(repo)},
				Invitee: &ghUser,
			}
			w.WriteHeader(http.StatusCreated)
			mustWrite(w, invite)
		}),
	)
	deleteReposCollaboratorsByOwnerByRepoByUsernameHandler := WithRequestMatchHandler(
		deleteReposCollaboratorsByOwnerByRepoByUsername,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			username := r.PathValue("username")

			collaborators := s.groups[owner][repo]
			if collaborators == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			collaborators = slices.DeleteFunc(collaborators, func(u github.User) bool {
				return u.GetLogin() == username
			})
			s.groups[owner][repo] = collaborators
			w.WriteHeader(http.StatusNoContent)
		}),
	)
	postIssueByOwnerByRepoHandler := WithRequestMatchHandler(
		postIssueByOwnerByRepo,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			issue := mustRead[github.Issue](r.Body)

			if !s.hasOrgRepo(owner, repo) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if issue.ID != nil || issue.Number != nil || issue.Repository != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if s.issues[owner] == nil {
				s.issues[owner] = make(map[string][]github.Issue)
			}
			if s.issues[owner][repo] == nil {
				s.issues[owner][repo] = make([]github.Issue, 0)
			}
			nextIssueNumber := 1
			for _, ghIssue := range s.issues[owner][repo] {
				if *ghIssue.Number >= nextIssueNumber {
					nextIssueNumber = *ghIssue.Number + 1
				}
			}
			s.issueID++
			issue.ID = github.Int64(s.issueID)
			issue.Number = github.Int(nextIssueNumber)
			issue.Repository = &github.Repository{
				Owner: &github.User{Login: github.String(owner)},
				Name:  github.String(repo),
			}
			s.issues[owner][repo] = append(s.issues[owner][repo], issue)
			w.WriteHeader(http.StatusCreated)
			mustWrite(w, issue)
		}),
	)
	patchIssueByOwnerByRepoByIssueNumberHandler := WithRequestMatchHandler(
		patchIssueByOwnerByRepoByIssueNumber,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			issueNumber := mustParse[int](r.PathValue("issue_number"))
			issue := mustRead[github.Issue](r.Body)

			for i, ghIssue := range s.issues[owner][repo] {
				if *ghIssue.Number == issueNumber {
					issue.ID = ghIssue.ID
					issue.Number = &issueNumber
					issue.Repository = &github.Repository{
						Owner: &github.User{Login: github.String(owner)},
						Name:  github.String(repo),
					}
					s.issues[owner][repo][i] = issue
					w.WriteHeader(http.StatusOK)
					mustWrite(w, issue)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		}),
	)
	getIssueByOwnerByRepoByIssueNumberHandler := WithRequestMatchHandler(
		getIssueByOwnerByRepoByIssueNumber,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			issueNumber := mustParse[int](r.PathValue("issue_number"))

			for _, issue := range s.issues[owner][repo] {
				if *issue.Number == issueNumber {
					mustWrite(w, issue)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		}),
	)
	getIssuesByOwnerByRepoHandler := WithRequestMatchHandler(
		getIssuesByOwnerByRepo,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")

			issues := s.issues[owner][repo]
			if issues == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			mustWrite(w, issues)
		}),
	)
	postIssueCommentByOwnerByRepoByIssueNumberHandler := WithRequestMatchHandler(
		postIssueCommentByOwnerByRepoByIssueNumber,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			issueNumber := mustParse[int](r.PathValue("issue_number"))
			comment := mustRead[github.IssueComment](r.Body)

			for _, ghIssue := range s.issues[owner][repo] {
				if *ghIssue.Number == issueNumber {
					s.commentID++
					comment.ID = github.Int64(s.commentID)
					s.comments[owner][repo][*ghIssue.ID] = append(s.comments[owner][repo][*ghIssue.ID], comment)
					w.WriteHeader(http.StatusCreated)
					mustWrite(w, comment)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		}),
	)
	patchIssueCommentByOwnerByRepoByCommentIDHandler := WithRequestMatchHandler(
		patchIssueCommentByOwnerByRepoByCommentID,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			commentID := mustParse[int64](r.PathValue("comment_id"))
			comment := mustRead[github.IssueComment](r.Body)

			for _, ghIssue := range s.issues[owner][repo] {
				for i, ghComment := range s.comments[owner][repo][*ghIssue.ID] {
					if *ghComment.ID == commentID {
						comment.ID = ghComment.ID
						s.comments[owner][repo][*ghIssue.ID][i] = comment
						w.WriteHeader(http.StatusOK)
						mustWrite(w, comment)
						return
					}
				}
			}
			w.WriteHeader(http.StatusNotFound)
		}),
	)
	postPullReviewersByOwnerByRepoByPullNumberHandler := WithRequestMatchHandler(
		postPullReviewersByOwnerByRepoByPullNumber,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			pullNumber := mustParse[int](r.PathValue("pull_number"))
			reviewers := mustRead[github.ReviewersRequest](r.Body)

			if _, exists := s.reviewers[owner][repo][pullNumber]; !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			s.reviewers[owner][repo][pullNumber] = reviewers
			users := make([]*github.User, 0, len(reviewers.Reviewers))
			for _, reviewer := range reviewers.Reviewers {
				users = append(users, &github.User{Login: github.String(reviewer)})
			}
			pr := github.PullRequest{
				Number:             github.Int(pullNumber),
				RequestedReviewers: users,
			}
			w.WriteHeader(http.StatusCreated)
			mustWrite(w, pr)
		}),
	)
	// Mock query handler for fetching the issue ID based on issue number
	queryHandler := func(w http.ResponseWriter, vars map[string]any) {
		owner := vars["repositoryOwner"].(string)
		repo := vars["repositoryName"].(string)
		issueNumber := int(vars["issueNumber"].(float64))

		var id int64
		for _, issue := range s.issues[owner][repo] {
			if issue.GetNumber() == issueNumber {
				id = issue.GetID()
				break
			}
		}
		if id == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		respBody := map[string]any{
			"data": map[string]any{
				"repository": map[string]any{
					"issue": map[string]any{
						"id": id,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mustWrite(w, respBody)
	}
	// Mock mutation handler for deleting an issue based on issue ID
	mutationHandler := func(w http.ResponseWriter, vars map[string]any) {
		id := int64(vars["issueId"].(float64))

		var foundRepo string
		for owner := range s.issues {
			for repo := range s.issues[owner] {
				for _, issue := range s.issues[owner][repo] {
					if issue.GetID() == id {
						foundRepo = repo
						issues := s.issues[owner][repo]
						issues = slices.DeleteFunc(issues, func(i github.Issue) bool { return i.GetID() == id })
						s.issues[owner][repo] = issues
						break
					}
				}
			}
		}
		if foundRepo == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		respBody := map[string]any{
			"data": map[string]any{
				"deleteIssue": map[string]any{
					"repository": map[string]any{
						"name": foundRepo,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mustWrite(w, respBody)
	}
	graphQLHandler := WithRequestMatchHandler("/graphql", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		req := mustRead[request](r.Body)

		if strings.HasPrefix(req.Query, "mutation") {
			mutationHandler(w, req.Variables["input"].(map[string]any))
		} else {
			queryHandler(w, req.Variables)
		}
	}))

	httpClient := NewMockedHTTPClient(
		getByIDHandler,
		getOrgsByOrgHandler,
		getOrgsReposByOrgHandler,
		getOrgsMembershipsByOrgByUsernameHandler,
		getReposContentsByOwnerByRepoByPathHandler,
		getReposCollaboratorsByOwnerByRepoHandler,
		putReposCollaboratorsByOwnerByRepoByUsernameHandler,
		deleteReposCollaboratorsByOwnerByRepoByUsernameHandler,
		postIssueByOwnerByRepoHandler,
		patchIssueByOwnerByRepoByIssueNumberHandler,
		getIssueByOwnerByRepoByIssueNumberHandler,
		getIssuesByOwnerByRepoHandler,
		postIssueCommentByOwnerByRepoByIssueNumberHandler,
		patchIssueCommentByOwnerByRepoByCommentIDHandler,
		postPullReviewersByOwnerByRepoByPullNumberHandler,
		graphQLHandler,
	)
	s.GithubSCM = &GithubSCM{
		logger:      logger,
		client:      github.NewClient(httpClient),
		clientV4:    githubv4.NewClient(httpClient),
		providerURL: "github.com",
	}
	return s
}
