package scm

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v62/github"
	"github.com/gosimple/slug"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"
	"google.golang.org/protobuf/testing/protocmp"
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

var (
	orgFoo = github.Organization{ID: github.Int64(123), Login: foo.Login}
	orgBar = github.Organization{ID: github.Int64(456), Login: bar.Login}
)

var (
	fooMelingRepo = github.Repository{Organization: &orgFoo, Owner: &foo, Name: github.String("meling-labs")}
	barMelingRepo = github.Repository{Organization: &orgBar, Owner: &bar, Name: github.String("meling-labs")}
	fooJosieRepo  = github.Repository{Organization: &orgFoo, Owner: &foo, Name: github.String("josie-labs")}
)

// MockedGithubSCM implements the SCM interface.
type MockedGithubSCM struct {
	*GithubSCM
	orgs      []github.Organization
	repos     []github.Repository
	members   []github.Membership
	groups    map[string]map[string][]github.User                   // map: owner -> repo -> collaborators
	issues    map[string]map[string][]github.Issue                  // map: owner -> repo -> issues
	comments  map[string]map[string]map[int64][]github.IssueComment // map: owner -> repo -> issue ID -> comments
	reviewers map[string]map[string]map[int]github.ReviewersRequest // map: owner -> repo -> pull requests ID -> reviewers
	issueID   int64
	commentID int64
}

// NewMockedGithubSCMClient returns a mocked Github client implementing the SCM interface.
func NewMockedGithubSCMClient(logger *zap.SugaredLogger) *MockedGithubSCM {
	s := &MockedGithubSCM{}

	// setup mock data based on qtest.MockCourses (complete course organizations and four repositories)
	mockRepos := []string{"info", "assignments", "tests", qf.StudentRepoName("meling")}
	for _, course := range qtest.MockCourses {
		ghOrg := github.Organization{
			ID:    github.Int64(int64(course.ScmOrganizationID)),
			Login: github.String(slug.Make(course.ScmOrganizationName)),
		}
		s.orgs = append(s.orgs, ghOrg)
		for _, repo := range mockRepos {
			ghRepo := github.Repository{Organization: &ghOrg, Name: github.String(repo)}
			s.repos = append(s.repos, ghRepo)
		}
	}

	// setup mock data based with partial organizations, repositories, memberships, and groups
	s.orgs = append(s.orgs, orgFoo, orgBar)
	// initial memberships: user -> role; two members; one owner, one member
	s.members = []github.Membership{
		{Organization: &orgFoo, User: &meling, Role: github.String(OrgOwner)},
		{Organization: &orgBar, User: &meling, Role: github.String(OrgMember)},
	}
	// initial repositories: for organization foo; bar has no repositories
	repos := []github.Repository{
		{Organization: &orgFoo, Name: github.String("info")},
		{Organization: &orgFoo, Name: github.String("assignments")},
		{Organization: &orgFoo, Name: github.String("tests")},
		{Organization: &orgFoo, Name: github.String("meling-labs")},
		{Organization: &orgFoo, Name: github.String("josie-labs")},
	}
	s.repos = append(s.repos, repos...)

	// initial groups map: owner -> repo -> collaborators (only group repos should have collaborators)
	s.groups = map[string]map[string][]github.User{
		"foo": {
			"info":        {},
			"assignments": {},
			"tests":       {},
			"meling-labs": {},
			"groupX":      {lamport},
		},
		"bar": {
			"groupY": {leslie},
			"groupZ": {},
		},
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
	// initial reviewers map: owner -> repo -> pull requests ID -> reviewers
	s.reviewers = map[string]map[string]map[int]github.ReviewersRequest{
		"foo": {
			"meling-labs": {
				1: {Reviewers: []string{"meling", "leslie"}},
				2: {Reviewers: []string{"lamport", "jostein"}},
			},
			"josie-labs": {
				1: {Reviewers: []string{"meling", "leslie"}},
				2: {Reviewers: []string{"lamport", "jostein"}},
			},
		},
	}

	matchFn := func(orgName string, f func(github.Organization)) bool {
		for _, org := range s.orgs {
			if org.GetLogin() == orgName {
				f(org)
				return true
			}
		}
		return false
	}
	hasOrgRepo := func(orgName, repoName string) bool {
		for _, repo := range s.repos {
			if repo.GetOrganization().GetLogin() == orgName && repo.GetName() == repoName {
				return true
			}
		}
		return false
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
			found := matchFn(org, func(o github.Organization) {
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
			found := matchFn(org, func(o github.Organization) {
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
			found := matchFn(org, func(o github.Organization) {
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
			if !hasOrgRepo(owner, repo) {
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

			if !hasOrgRepo(owner, repo) {
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
			return
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

func TestMockGetOrganization(t *testing.T) {
	orgFoo := &qf.Organization{ScmOrganizationID: 123, ScmOrganizationName: "foo"}
	orgBar := &qf.Organization{ScmOrganizationID: 456, ScmOrganizationName: "bar"}

	tests := []struct {
		name    string
		org     *OrganizationOptions // cannot be nil
		wantOrg *qf.Organization
		wantErr bool
	}{
		{name: "IncompleteRequest", org: &OrganizationOptions{}, wantOrg: nil, wantErr: true},
		{name: "IncompleteRequest", org: &OrganizationOptions{Username: "meling"}, wantOrg: nil, wantErr: true},
		{name: "IncompleteRequest", org: &OrganizationOptions{NewCourse: true}, wantOrg: nil, wantErr: true},
		{name: "IncompleteRequest", org: &OrganizationOptions{NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},

		{name: "CompleteRequest", org: &OrganizationOptions{ID: 123}, wantOrg: orgFoo, wantErr: false},
		{name: "CompleteRequest", org: &OrganizationOptions{ID: 456}, wantOrg: orgBar, wantErr: false},
		{name: "CompleteRequest/Missing", org: &OrganizationOptions{ID: 789}, wantOrg: nil, wantErr: true}, // 789 does not exist

		{name: "CompleteRequest", org: &OrganizationOptions{ID: 123, NewCourse: true}, wantOrg: nil, wantErr: true},
		{name: "CompleteRequest", org: &OrganizationOptions{ID: 456, NewCourse: true}, wantOrg: orgBar, wantErr: false},
		{name: "CompleteRequest/Missing", org: &OrganizationOptions{ID: 789, NewCourse: true}, wantOrg: nil, wantErr: true}, // 789 does not exist

		{name: "CompleteRequest", org: &OrganizationOptions{ID: 123, Username: "meling"}, wantOrg: orgFoo, wantErr: false},     // meling is owner of foo
		{name: "CompleteRequest", org: &OrganizationOptions{ID: 456, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is only member of bar, not owner
		{name: "CompleteRequest/Missing", org: &OrganizationOptions{ID: 789, Username: "meling"}, wantOrg: nil, wantErr: true}, // 789 does not exist

		{name: "CompleteRequest", org: &OrganizationOptions{ID: 123, NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is owner of foo, but foo is not empty (not new course)
		{name: "CompleteRequest", org: &OrganizationOptions{ID: 456, NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is only member of bar, not owner
		{name: "CompleteRequest/Missing", org: &OrganizationOptions{ID: 789, NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true}, // 789 does not exist

		{name: "CompleteRequest", org: &OrganizationOptions{Name: "foo"}, wantOrg: orgFoo, wantErr: false},
		{name: "CompleteRequest", org: &OrganizationOptions{Name: "bar"}, wantOrg: orgBar, wantErr: false},
		{name: "CompleteRequest/Missing", org: &OrganizationOptions{Name: "baz"}, wantOrg: nil, wantErr: true}, // baz does not exist

		{name: "CompleteRequest", org: &OrganizationOptions{Name: "foo", NewCourse: true}, wantOrg: nil, wantErr: true},
		{name: "CompleteRequest", org: &OrganizationOptions{Name: "bar", NewCourse: true}, wantOrg: orgBar, wantErr: false},
		{name: "CompleteRequest/Missing", org: &OrganizationOptions{Name: "baz", NewCourse: true}, wantOrg: nil, wantErr: true}, // baz does not exist

		{name: "CompleteRequest", org: &OrganizationOptions{Name: "foo", Username: "meling"}, wantOrg: orgFoo, wantErr: false},     // meling is owner of foo
		{name: "CompleteRequest", org: &OrganizationOptions{Name: "bar", Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is only member of bar, not owner
		{name: "CompleteRequest/Missing", org: &OrganizationOptions{Name: "baz", Username: "meling"}, wantOrg: nil, wantErr: true}, // baz does not exist

		{name: "CompleteRequest", org: &OrganizationOptions{Name: "foo", NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is owner of foo
		{name: "CompleteRequest", org: &OrganizationOptions{Name: "bar", NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true},         // meling is only member of bar, not owner
		{name: "CompleteRequest/Missing", org: &OrganizationOptions{Name: "baz", NewCourse: true, Username: "meling"}, wantOrg: nil, wantErr: true}, // baz does not exist
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"ID", "Name", "Username", "NewCourse"}, tt.org.ID, tt.org.Name, tt.org.Username, tt.org.NewCourse)
		t.Run(name, func(t *testing.T) {
			gotOrg, gotErr := s.GetOrganization(context.Background(), tt.org)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetOrganization() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.wantOrg, gotOrg, protocmp.Transform()); diff != "" {
				t.Errorf("GetOrganization() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMockGetRepositories(t *testing.T) {
	tests := []struct {
		name    string
		org     *qf.Organization
		want    []*Repository
		wantErr bool
	}{
		{name: "IncompleteRequest/NilOrg", org: nil, want: nil, wantErr: true},
		{name: "IncompleteRequest/NoOrgName", org: &qf.Organization{}, want: nil, wantErr: true},
		{name: "CompleteRequest/NotFound", org: &qf.Organization{ScmOrganizationName: "bar"}, want: []*Repository{}, wantErr: false},
		{name: "CompleteRequest/FiveRepos", org: &qf.Organization{ScmOrganizationName: "foo"}, want: []*Repository{
			{OrgID: 123, Path: "info"},
			{OrgID: 123, Path: "assignments"},
			{OrgID: 123, Path: "tests"},
			{OrgID: 123, Path: "meling-labs"},
			{OrgID: 123, Path: "josie-labs"},
		}},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetRepositories(context.Background(), tt.org)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepositories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("GetRepositories() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMockRepositoryIsEmpty(t *testing.T) {
	tests := []struct {
		name      string
		opt       *RepositoryOptions
		wantEmpty bool
	}{
		{name: "IncompleteRequest", opt: &RepositoryOptions{}, wantEmpty: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Owner: "foo"}, wantEmpty: true},
		{name: "IncompleteRequest", opt: &RepositoryOptions{Path: "info"}, wantEmpty: true},

		{name: "CompleteRequest/Empty", opt: &RepositoryOptions{Owner: "bar", Path: "info"}, wantEmpty: true},
		{name: "CompleteRequest/Empty", opt: &RepositoryOptions{Owner: "bar", Path: "assignments"}, wantEmpty: true},
		{name: "CompleteRequest/Empty", opt: &RepositoryOptions{Owner: "bar", Path: "tests"}, wantEmpty: true},
		{name: "CompleteRequest/Empty", opt: &RepositoryOptions{Owner: "bar", Path: "meling-labs"}, wantEmpty: true},

		{name: "CompleteRequest/NonEmpty", opt: &RepositoryOptions{Owner: "foo", Path: "info"}, wantEmpty: false},
		{name: "CompleteRequest/NonEmpty", opt: &RepositoryOptions{Owner: "foo", Path: "assignments"}, wantEmpty: false},
		{name: "CompleteRequest/NonEmpty", opt: &RepositoryOptions{Owner: "foo", Path: "tests"}, wantEmpty: false},
		{name: "CompleteRequest/NonEmpty", opt: &RepositoryOptions{Owner: "foo", Path: "meling-labs"}, wantEmpty: false},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Owner", "Path"}, tt.opt.Owner, tt.opt.Path)
		t.Run(name, func(t *testing.T) {
			gotIsEmpty := s.RepositoryIsEmpty(context.Background(), tt.opt)
			if gotIsEmpty != tt.wantEmpty {
				t.Errorf("RepositoryIsEmpty() = %v, want %v", gotIsEmpty, tt.wantEmpty)
			}
		})
	}
}

func TestMockUpdateGroupMembers(t *testing.T) {
	tests := []struct {
		name      string
		org       *GroupOptions
		wantUsers []github.User
		wantErr   bool
	}{
		{name: "IncompleteRequest", org: &GroupOptions{}, wantErr: true},
		{name: "IncompleteRequest", org: &GroupOptions{Organization: "foo"}, wantErr: true},
		{name: "IncompleteRequest", org: &GroupOptions{GroupName: "a"}, wantErr: true},
		{name: "IncompleteRequest", org: &GroupOptions{Users: []string{"meling"}}, wantErr: true},
		{name: "IncompleteRequest", org: &GroupOptions{Organization: "foo", Users: []string{"meling"}}, wantErr: true},
		{name: "IncompleteRequest", org: &GroupOptions{GroupName: "a", Users: []string{"meling"}}, wantErr: true},

		{name: "CompleteRequest/NotFound", org: &GroupOptions{Organization: "foo", GroupName: "a"}, wantErr: true},
		{name: "CompleteRequest/NotFound", org: &GroupOptions{Organization: "x", GroupName: "info"}, wantErr: true},
		{name: "CompleteRequest/NotFound", org: &GroupOptions{Organization: "foo", GroupName: "a", Users: []string{"meling"}}, wantErr: true},
		{name: "CompleteRequest/NotFound", org: &GroupOptions{Organization: "x", GroupName: "info", Users: []string{"meling"}}, wantErr: true},

		{name: "CompleteRequest", org: &GroupOptions{Organization: "foo", GroupName: "info", Users: []string{}}, wantErr: false, wantUsers: []github.User{}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "foo", GroupName: "groupX", Users: []string{"meling"}}, wantErr: false, wantUsers: []github.User{meling}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "foo", GroupName: "groupX", Users: []string{"meling", "leslie"}}, wantErr: false, wantUsers: []github.User{meling, leslie}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "foo", GroupName: "groupX", Users: []string{"meling", "leslie", "lamport"}}, wantErr: false, wantUsers: []github.User{meling, leslie, lamport}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "bar", GroupName: "groupY", Users: []string{"leslie", "lamport"}}, wantErr: false, wantUsers: []github.User{leslie, lamport}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "bar", GroupName: "groupY", Users: []string{"leslie"}}, wantErr: false, wantUsers: []github.User{leslie}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "bar", GroupName: "groupY", Users: []string{}}, wantErr: false, wantUsers: []github.User{}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "bar", GroupName: "groupZ", Users: []string{"leslie"}}, wantErr: false, wantUsers: []github.User{leslie}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "bar", GroupName: "groupZ", Users: []string{}}, wantErr: false, wantUsers: []github.User{}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "bar", GroupName: "groupZ", Users: []string{"leslie", "lamport"}}, wantErr: false, wantUsers: []github.User{leslie, lamport}},
		{name: "CompleteRequest", org: &GroupOptions{Organization: "bar", GroupName: "groupZ", Users: []string{"jostein"}}, wantErr: false, wantUsers: []github.User{jostein}},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t))
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"Organization", "GroupName", "Users"}, tt.org.Organization, tt.org.GroupName, tt.org.Users)
		t.Run(name, func(t *testing.T) {
			if err := s.UpdateGroupMembers(context.Background(), tt.org); (err != nil) != tt.wantErr {
				t.Errorf("UpdateGroupMembers() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantUsers == nil {
				return
			}
			// verify the state of the groups after the test
			if diff := cmp.Diff(tt.wantUsers, s.groups[tt.org.Organization][tt.org.GroupName], protocmp.Transform()); diff != "" {
				t.Errorf("UpdateGroupMembers() mismatch (-want +got):\n%s", diff)
			}
		})
	}

	// expected state after calling the sequence of UpdateGroupMembers
	// owner -> repo -> collaborators
	wantGroups := map[string]map[string][]github.User{
		"foo": {
			"info":        {},
			"assignments": {},
			"tests":       {},
			"meling-labs": {},
			"groupX":      {meling, leslie, lamport},
		},
		"bar": {
			"groupY": {},
			"groupZ": {jostein},
		},
	}
	// verify the state of the groups after the sequence of UpdateGroupMembers
	if diff := cmp.Diff(wantGroups, s.groups, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateGroupMembers() mismatch (-want +got):\n%s", diff)
	}
}
