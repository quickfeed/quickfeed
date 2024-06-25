package scm

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/internal/env"
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

// MockedGithubSCM implements the SCM interface.
type MockedGithubSCM struct {
	*GithubSCM
	*mockOptions
	repoID      int64
	issueID     int64
	issueNumber map[string]int // owner/repo -> issue number
	commentID   int64
}

// nextIssueNumber returns the next issue number for the given owner and repo.
func (s *MockedGithubSCM) nextIssueNumber(owner, repo string) *int {
	key := fmt.Sprintf("%s/%s", owner, repo)
	if s.issueNumber == nil {
		s.issueNumber = make(map[string]int)
	}
	s.issueNumber[key]++
	return github.Int(s.issueNumber[key])
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

	if s.issues == nil {
		// initial empty issues map: owner -> repo -> issues
		s.issues = make(map[string]map[string][]github.Issue)
	}
	for _, repo := range s.repos {
		org := repo.GetOrganization().GetLogin()
		if s.issues[org] == nil {
			s.issues[org] = make(map[string][]github.Issue)
			s.issues[org][repo.GetName()] = make([]github.Issue, 0)
		}
	}
	if s.comments == nil {
		// initial empty comments map: owner -> repo -> issue ID -> comments
		s.comments = make(map[string]map[string]map[int64][]github.IssueComment)
	}
	for org, repos := range s.issues {
		if s.comments[org] == nil {
			s.comments[org] = make(map[string]map[int64][]github.IssueComment)
		}
		for repo := range repos {
			if s.comments[org][repo] == nil {
				s.comments[org][repo] = make(map[int64][]github.IssueComment)
			}
		}
	}

	getOrganizationsByIDHandler := WithRequestMatchHandler(
		getOrganizationsByID,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := mustParse[int64](r.PathValue("id"))
			for _, org := range s.orgs {
				if org.GetID() == id {
					mustWrite(w, org)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound) // org not found
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
				w.WriteHeader(http.StatusNotFound) // org not found
			}
		}),
	)
	patchOrgsByOrgHandler := WithRequestMatchHandler(
		patchOrgsByOrg,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			org := r.PathValue("org")
			newOrg := mustRead[github.Organization](r.Body)

			found := s.matchOrgFunc(org, func(o github.Organization) {
				o.Login = newOrg.Login
				o.DefaultRepoPermission = newOrg.DefaultRepoPermission
				o.MembersCanCreateRepos = newOrg.MembersCanCreateRepos
				w.WriteHeader(http.StatusOK)
				mustWrite(w, o)
			})
			if !found {
				w.WriteHeader(http.StatusNotFound) // org not found
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
				w.WriteHeader(http.StatusNotFound) // org not found
			}
		}),
	)
	postOrgsReposByOrgHandler := WithRequestMatchHandler(
		postOrgsReposByOrg,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			org := r.PathValue("org")
			repo := mustRead[github.Repository](r.Body)

			found := s.matchOrgFunc(org, func(o github.Organization) {
				s.repoID++
				repo.ID = &s.repoID
				repo.Owner = &github.User{Login: github.String(org)}
				repo.Organization = &o
				s.repos = append(s.repos, repo)
				if s.groups[org] == nil {
					s.groups[org] = make(map[string][]github.User)
				}
				s.groups[org][repo.GetName()] = make([]github.User, 0)
				mustWrite(w, repo)
			})
			if !found {
				w.WriteHeader(http.StatusNotFound) // org not found
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
				w.WriteHeader(http.StatusNotFound) // member not found
			})
			if !found {
				w.WriteHeader(http.StatusNotFound) // org not found
			}
		}),
	)
	putOrgsMembershipsByOrgByUsernameHandler := WithRequestMatchHandler(
		putOrgsMembershipsByOrgByUsername,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			org := r.PathValue("org")
			username := r.PathValue("username")
			membership := mustRead[github.Membership](r.Body)

			found := s.matchOrgFunc(org, func(o github.Organization) {
				for i, m := range s.members {
					if m.GetOrganization().GetLogin() == o.GetLogin() && m.GetUser().GetLogin() == username {
						s.members[i].Role = membership.Role
						mustWrite(w, membership)
						return
					}
				}
				w.WriteHeader(http.StatusNotFound) // member not found
			})
			if !found {
				w.WriteHeader(http.StatusNotFound) // org not found
			}
		}),
	)
	deleteOrgsMembersByOrgByUsernameHandler := WithRequestMatchHandler(
		deleteOrgsMembersByOrgByUsername,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			org := r.PathValue("org")
			username := r.PathValue("username")

			found := s.matchOrgFunc(org, func(o github.Organization) {
				for i, m := range s.members {
					if m.GetOrganization().GetLogin() == o.GetLogin() && m.GetUser().GetLogin() == username {
						s.members = slices.Delete(s.members, i, i+1)
						w.WriteHeader(http.StatusNoContent)
						return
					}
				}
				w.WriteHeader(http.StatusNotFound) // member not found
			})
			if !found {
				w.WriteHeader(http.StatusNotFound) // org not found
			}
		}),
	)
	getReposByOwnerByRepoHandler := WithRequestMatchHandler(
		getReposByOwnerByRepo,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			for _, re := range s.repos {
				if re.GetOrganization().GetLogin() == owner && re.GetName() == repo {
					re.Owner = &github.User{Login: github.String(owner)}
					mustWrite(w, re)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound) // repo not found
		}),
	)
	deleteReposByOwnerByRepoHandler := WithRequestMatchHandler(
		deleteReposByOwnerByRepo,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			for i, re := range s.repos {
				if re.GetOrganization().GetLogin() == owner && re.GetName() == repo {
					s.repos = slices.Delete(s.repos, i, i+1)
					delete(s.groups[owner], repo)
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound) // repo not found
		}),
	)
	getRepositoriesByIDHandler := WithRequestMatchHandler(
		getRepositoriesByID,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := mustParse[int64](r.PathValue("repository_id"))
			for _, repo := range s.repos {
				if repo.GetID() == id {
					repo.Owner = &github.User{Login: github.String(repo.GetOrganization().GetLogin())}
					mustWrite(w, repo)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound) // repo not found
		}),
	)
	getReposContentsByOwnerByRepoByPathHandler := WithRequestMatchHandler(
		getReposContentsByOwnerByRepoByPath,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// we only care about the owner and repo; we ignore the path component
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			if !s.hasOrgRepo(owner, repo) {
				w.WriteHeader(http.StatusNotFound) // org and repo not found
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
				w.WriteHeader(http.StatusNotFound) // org and repo not found
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
			repoCollaboratorOptions := mustRead[github.RepositoryAddCollaboratorOptions](r.Body)

			collaborators := s.groups[owner][repo]
			if collaborators == nil {
				if !slices.ContainsFunc(s.repos, func(r github.Repository) bool {
					return r.GetOrganization().GetLogin() == owner && r.GetName() == repo
				}) {
					w.WriteHeader(http.StatusNotFound) // org and repo not found
					return
				}
				collaborators = make([]github.User, 0)
				if s.groups[owner] == nil {
					s.groups[owner] = make(map[string][]github.User)
				}
				s.groups[owner][repo] = collaborators
			}
			if slices.ContainsFunc(collaborators, func(u github.User) bool { return u.GetLogin() == username }) {
				// already exists; no need to add again
				w.WriteHeader(http.StatusNoContent)
				return
			}

			permissions := map[string]bool{repoCollaboratorOptions.Permission: true}
			ghUser := github.User{Login: github.String(username), Permissions: permissions}
			// this simulates that the user accepts the invitation (mocking the invite response is not supported yet)
			s.groups[owner][repo] = append(collaborators, ghUser)
			s.members = append(s.members, github.Membership{
				Organization: &github.Organization{Login: github.String(owner)},
				User:         &github.User{Login: github.String(username)},
				Role:         github.String(repoCollaboratorOptions.Permission),
			})
			invite := github.CollaboratorInvitation{
				Repo: &github.Repository{
					Owner:       &github.User{Login: github.String(owner)},
					Name:        github.String(repo),
					Permissions: permissions,
				},
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
				w.WriteHeader(http.StatusNotFound) // org and repo not found
				return
			}

			collaborators = slices.DeleteFunc(collaborators, func(u github.User) bool {
				return u.GetLogin() == username
			})
			s.groups[owner][repo] = collaborators
			w.WriteHeader(http.StatusNoContent)
		}),
	)
	postReposIssuesByOwnerByRepoHandler := WithRequestMatchHandler(
		postReposIssuesByOwnerByRepo,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			issueReq := mustRead[github.IssueRequest](r.Body)

			if !s.hasOrgRepo(owner, repo) {
				w.WriteHeader(http.StatusNotFound) // org and repo not found
				return
			}
			if issueReq.Title == nil || issueReq.Body == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if s.issues[owner] == nil {
				s.issues[owner] = make(map[string][]github.Issue)
			}
			if s.issues[owner][repo] == nil {
				s.issues[owner][repo] = make([]github.Issue, 0)
			}

			s.issueID++
			issue := github.Issue{
				ID:       github.Int64(s.issueID),
				Number:   s.nextIssueNumber(owner, repo),
				Title:    issueReq.Title,
				Body:     issueReq.Body,
				Assignee: &github.User{Name: issueReq.Assignee},
				Repository: &github.Repository{
					Owner: &github.User{Login: github.String(owner)},
					Name:  github.String(repo),
				},
			}
			s.issues[owner][repo] = append(s.issues[owner][repo], issue)
			w.WriteHeader(http.StatusCreated)
			mustWrite(w, issue)
		}),
	)
	patchReposIssuesByOwnerByRepoByIssueNumberHandler := WithRequestMatchHandler(
		patchReposIssuesByOwnerByRepoByIssueNumber,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			issueNumber := mustParse[int](r.PathValue("issue_number"))
			issueReq := mustRead[github.IssueRequest](r.Body)

			for i, ghIssue := range s.issues[owner][repo] {
				if *ghIssue.Number == issueNumber {
					ghIssue.Title = issueReq.Title
					ghIssue.Body = issueReq.Body
					ghIssue.Assignee = &github.User{Name: issueReq.Assignee}
					ghIssue.State = issueReq.State
					s.issues[owner][repo][i] = ghIssue
					w.WriteHeader(http.StatusOK)
					mustWrite(w, ghIssue)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound) // issue not found
		}),
	)
	getReposIssuesByOwnerByRepoByIssueNumberHandler := WithRequestMatchHandler(
		getReposIssuesByOwnerByRepoByIssueNumber,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			issueNumber := mustParse[int](r.PathValue("issue_number"))

			for _, issue := range s.issues[owner][repo] {
				if *issue.Number == issueNumber {
					w.WriteHeader(http.StatusOK)
					mustWrite(w, issue)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound) // issue not found
		}),
	)
	getReposIssuesByOwnerByRepoHandler := WithRequestMatchHandler(
		getReposIssuesByOwnerByRepo,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")

			issues := s.issues[owner][repo]
			if issues == nil {
				w.WriteHeader(http.StatusNotFound) // org and repo not found
				return
			}
			w.WriteHeader(http.StatusOK)
			mustWrite(w, issues)
		}),
	)
	postReposIssuesCommentsByOwnerByRepoByIssueNumberHandler := WithRequestMatchHandler(
		postReposIssuesCommentsByOwnerByRepoByIssueNumber,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			issueNumber := mustParse[int](r.PathValue("issue_number"))
			comment := mustRead[github.IssueComment](r.Body)

			for _, ghIssue := range s.issues[owner][repo] {
				if *ghIssue.Number == issueNumber {
					s.commentID++
					comment.ID = github.Int64(s.commentID)
					if s.comments[owner][repo] == nil {
						s.comments[owner][repo] = make(map[int64][]github.IssueComment)
					}
					s.comments[owner][repo][*ghIssue.ID] = append(s.comments[owner][repo][*ghIssue.ID], comment)
					w.WriteHeader(http.StatusCreated)
					mustWrite(w, comment)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound) // issue not found
		}),
	)
	patchReposIssuesCommentsByOwnerByRepoByCommentIDHandler := WithRequestMatchHandler(
		patchReposIssuesCommentsByOwnerByRepoByCommentID,
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
			w.WriteHeader(http.StatusNotFound) // comment not found
		}),
	)
	postReposPullsRequestedReviewersByOwnerByRepoByPullNumberHandler := WithRequestMatchHandler(
		postReposPullsRequestedReviewersByOwnerByRepoByPullNumber,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			pullNumber := mustParse[int](r.PathValue("pull_number"))
			reviewers := mustRead[github.ReviewersRequest](r.Body)

			if _, exists := s.reviewers[owner][repo][pullNumber]; !exists {
				w.WriteHeader(http.StatusNotFound) // pull request not found
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
			w.WriteHeader(http.StatusNotFound) // issue not found
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
			w.WriteHeader(http.StatusNotFound) // issue not found
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
		getOrganizationsByIDHandler,
		getOrgsByOrgHandler,
		patchOrgsByOrgHandler,
		getOrgsReposByOrgHandler,
		postOrgsReposByOrgHandler,
		getOrgsMembershipsByOrgByUsernameHandler,
		putOrgsMembershipsByOrgByUsernameHandler,
		deleteOrgsMembersByOrgByUsernameHandler,
		getReposByOwnerByRepoHandler,
		deleteReposByOwnerByRepoHandler,
		getRepositoriesByIDHandler,
		getReposContentsByOwnerByRepoByPathHandler,
		getReposCollaboratorsByOwnerByRepoHandler,
		putReposCollaboratorsByOwnerByRepoByUsernameHandler,
		deleteReposCollaboratorsByOwnerByRepoByUsernameHandler,
		postReposIssuesByOwnerByRepoHandler,
		patchReposIssuesByOwnerByRepoByIssueNumberHandler,
		getReposIssuesByOwnerByRepoByIssueNumberHandler,
		getReposIssuesByOwnerByRepoHandler,
		postReposIssuesCommentsByOwnerByRepoByIssueNumberHandler,
		patchReposIssuesCommentsByOwnerByRepoByCommentIDHandler,
		postReposPullsRequestedReviewersByOwnerByRepoByPullNumberHandler,
		graphQLHandler,
	)
	s.GithubSCM = &GithubSCM{
		logger:      logger,
		client:      github.NewClient(httpClient),
		clientV4:    githubv4.NewClient(httpClient),
		providerURL: "file://" + env.RepositoryPath(),
	}
	return s
}
