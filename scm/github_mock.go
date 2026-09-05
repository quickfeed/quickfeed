package scm

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
)

// routeLabel is the attribute holding the mocked GitHub API route.
const routeLabel = "route"

// MockedGithubSCM implements the SCM interface.
type MockedGithubSCM struct {
	*GithubSCM
	*mockOptions
	repoID int64
}

// SimulateCommit records a commit pushed to the given repository, advancing it one
// commit ahead of the upstream assignments repository it was forked from.
// Tests that only configure the mock at construction can use WithCommitsAhead.
func (s *MockedGithubSCM) SimulateCommit(owner, repo string) error {
	if s.findOrgRepo(owner, repo) == nil {
		return fmt.Errorf("cannot simulate commit: repository %s/%s not found", owner, repo)
	}
	s.aheadBy[repoKey(owner, repo)]++
	return nil
}

// NewMockedGithubSCMClient returns a mocked Github client implementing the SCM interface.
// The logger is captured by the mocked request handlers for the client's lifetime;
// it is therefore taken directly rather than derived from a per-call context.
// This is intentionally breaking the cyclomatic complexity rule (GO-R1005) to keep the
// initialization of all the mock handlers in one place. It is not production code; it is
// only used for testing.
func NewMockedGithubSCMClient(logger *slog.Logger, opts ...MockOption) *MockedGithubSCM { // skipcq: GO-R1005
	mockOpts := newMockOptions()
	for _, o := range opts {
		o(mockOpts)
	}
	s := &MockedGithubSCM{
		mockOptions: mockOpts,
	}

	// To assist with debugging; this may be useful
	// logger.Debug(s.DumpState())

	getOrganizationsByIDHandler := WithRequestMatchHandler(
		getOrganizationsByID,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := mustParse[int64](r.PathValue("id"))
			logger.Debug("mock SCM request", routeLabel, replaceArgs(getOrganizationsByID, id))

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
			logger.Debug("mock SCM request", routeLabel, replaceArgs(getOrgsByOrg, org))

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
			logger.Debug("mock SCM request", routeLabel, replaceArgs(patchOrgsByOrg, org), label.Organization, newOrg.GetLogin())

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
			logger.Debug("mock SCM request", routeLabel, replaceArgs(getOrgsReposByOrg, org))

			found := s.matchOrgFunc(org, func(o github.Organization) {
				foundRepos := make([]github.Repository, 0)
				for i := range s.repos {
					repo := &s.repos[i]
					if repo.GetOrganization().GetLogin() == o.GetLogin() {
						foundRepos = append(foundRepos, *repo)
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
			logger.Debug("mock SCM request", routeLabel, replaceArgs(postOrgsReposByOrg, org), label.Repository, repo.GetName())

			found := s.matchOrgFunc(org, func(o github.Organization) {
				s.repoID++
				repo.ID = &s.repoID
				repo.Owner = &github.User{Login: new(org)}
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
	// repos/%v/%v/forks
	postReposForksByOwnerByRepoHandler := WithRequestMatchHandler(
		postReposForksByOwnerByRepo,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srcOwner := r.PathValue("owner")
			srcRepo := r.PathValue("repo")
			logger.Debug("mock SCM request", routeLabel, replaceArgs(postReposForksByOwnerByRepo, srcOwner, srcRepo))
			opts := mustRead[github.RepositoryCreateForkOptions](r.Body)
			dstOrg := opts.Organization

			found := s.matchOrgFunc(dstOrg, func(o github.Organization) {
				s.repoID++
				fork := github.Repository{
					ID:           &s.repoID,
					Organization: &o,
					Name:         new(opts.Name),
					Owner:        &github.User{Login: new(dstOrg)},
					Fork:         new(true),
					// Record the upstream the fork was created from, mirroring how student
					// and group repositories are forks of the assignments repository.
					Parent: &github.Repository{
						Name:  new(srcRepo),
						Owner: &github.User{Login: new(srcOwner)},
					},
				}
				s.repos = append(s.repos, fork)
				if s.groups[dstOrg] == nil {
					s.groups[dstOrg] = make(map[string][]github.User)
				}
				s.groups[dstOrg][fork.GetName()] = make([]github.User, 0)
				mustWrite(w, fork)
			})
			if !found {
				w.WriteHeader(http.StatusNotFound) // repo not found
			}
		}),
	)
	getOrgsMembershipsByOrgByUsernameHandler := WithRequestMatchHandler(
		getOrgsMembershipsByOrgByUsername,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			org := r.PathValue("org")
			username := r.PathValue("username")
			logger.Debug("mock SCM request", routeLabel, replaceArgs(getOrgsMembershipsByOrgByUsername, org, username))

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
			logger.Debug("mock SCM request", routeLabel, replaceArgs(putOrgsMembershipsByOrgByUsername, org, username), "membership", membership)

			found := s.matchOrgFunc(org, func(o github.Organization) {
				// Check if user already exists
				for i, m := range s.members {
					if m.GetOrganization().GetLogin() == o.GetLogin() && m.GetUser().GetLogin() == username {
						s.members[i].Role = membership.Role
						mustWrite(w, s.members[i])
						return
					}
				}
				// If user not found, and role is admin -> return error
				if strings.EqualFold(membership.GetRole(), "admin") {
					fmt.Println("NOT FOUND", username, "AS ADMIN", membership)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				fmt.Println("ADDING AS MEMBER", username, membership)
				// User not found - create new membership (simulates sending invitation)
				userID := s.getUserID(username)
				newMembership := github.Membership{
					Organization: &github.Organization{Login: new(org)},
					User:         &github.User{ID: new(userID), Login: new(username)},
					Role:         membership.Role,
					State:        new("pending"), // Invitation pending until accepted
				}
				s.members = append(s.members, newMembership)
				mustWrite(w, newMembership)
			})
			if !found {
				w.WriteHeader(http.StatusNotFound) // org not found
			}
		}),
	)
	deleteOrgsMembershipsByOrgByUsernameHandler := WithRequestMatchHandler(
		deleteOrgsMembershipsByOrgByUsername,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			org := r.PathValue("org")
			username := r.PathValue("username")
			logger.Debug("mock SCM request", routeLabel, replaceArgs(deleteOrgsMembershipsByOrgByUsername, org, username))

			found := s.matchOrgFunc(org, func(o github.Organization) {
				for i, m := range s.members {
					if m.GetOrganization().GetLogin() == o.GetLogin() && m.GetUser().GetLogin() == username {
						s.members = slices.Delete(s.members, i, i+1)
						w.WriteHeader(http.StatusNoContent)
						return
					}
				}
				w.WriteHeader(http.StatusNotFound) // no membership record
			})
			if !found {
				w.WriteHeader(http.StatusNotFound) // org not found
			}
		}),
	)
	// Handler for user accepting their own org invitation (PATCH /user/memberships/orgs/{org})
	patchUserMembershipsOrgsByOrgHandler := WithRequestMatchHandler(
		patchUserMembershipsOrgsByOrg,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			org := r.PathValue("org")
			membership := mustRead[github.Membership](r.Body)
			logger.Debug("mock SCM request", routeLabel, replaceArgs(patchUserMembershipsOrgsByOrg, org), "membership", membership)

			// Simulate GitHub's asynchronous invitation job.
			if s.invitationNotReadyFor > 0 {
				s.invitationNotReadyFor--
				w.WriteHeader(http.StatusAccepted)
				return
			}

			// Find the pending membership and activate it
			found := s.matchOrgFunc(org, func(o github.Organization) {
				for i, m := range s.members {
					if m.GetOrganization().GetLogin() == o.GetLogin() {
						// Set state to active (user accepted invitation)
						s.members[i].State = new("active")
						mustWrite(w, s.members[i])
						return
					}
				}
				w.WriteHeader(http.StatusNotFound) // membership not found
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
			logger.Debug("mock SCM request", routeLabel, replaceArgs(getReposByOwnerByRepo, owner, repo))

			for i := range s.repos {
				re := s.repos[i]
				if re.GetOrganization().GetLogin() == owner && re.GetName() == repo {
					re.Owner = &github.User{Login: new(owner)}
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
			logger.Debug("mock SCM request", routeLabel, replaceArgs(deleteReposByOwnerByRepo, owner, repo))

			for i := range s.repos {
				re := s.repos[i]
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
			logger.Debug("mock SCM request", routeLabel, replaceArgs(getRepositoriesByID, id))

			for i := range s.repos {
				repo := &s.repos[i]
				if repo.GetID() == id {
					repo.Owner = &github.User{Login: new(repo.GetOrganization().GetLogin())}
					mustWrite(w, repo)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound) // repo not found
		}),
	)
	getReposCommitsByOwnerByRepoByRefHandler := WithRequestMatchHandler(
		getReposCommitsByOwnerByRepoByRef,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repoName := r.PathValue("repo")
			ref := r.PathValue("ref")
			logger.Debug("mock SCM request", routeLabel, replaceArgs(getReposCommitsByOwnerByRepoByRef, owner, repoName, ref))

			repo := s.findOrgRepo(owner, repoName)
			if repo == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			mustWrite(w, &github.RepositoryCommit{SHA: new(mockRepoHeadSHA(repo))})
		}),
	)
	getReposCompareByOwnerByRepoByBaseByHeadHandler := WithRequestMatchHandler(
		getReposCompareByOwnerByRepoByBaseByHead,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			basehead := r.PathValue("basehead")
			logger.Debug("mock SCM request", routeLabel, replaceArgs(getReposCompareByOwnerByRepoByBaseByHead, owner, repo, basehead))

			if !s.hasOrgRepo(owner, repo) {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			comparison := &github.CommitsComparison{
				AheadBy:      new(0), // Default: no commits ahead
				BehindBy:     new(0),
				TotalCommits: new(0),
				Status:       new("identical"),
			}

			parts := strings.Split(basehead, "...")
			if len(parts) == 2 {
				headRepo := s.findRepoByHeadSHA(parts[1])
				if headRepo != nil {
					if ahead := s.aheadBy[repoKey(owner, headRepo.GetName())]; ahead > 0 {
						comparison.AheadBy = new(ahead)
						comparison.TotalCommits = new(ahead)
						comparison.Status = new("ahead")
					}
				}
			}

			mustWrite(w, comparison)
		}),
	)
	getReposCollaboratorsByOwnerByRepoHandler := WithRequestMatchHandler(
		getReposCollaboratorsByOwnerByRepo,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			logger.Debug("mock SCM request", routeLabel, replaceArgs(getReposCollaboratorsByOwnerByRepo, owner, repo))

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
			logger.Debug("mock SCM request", routeLabel, replaceArgs(putReposCollaboratorsByOwnerByRepoByUsername, owner, repo, username), "options", repoCollaboratorOptions)

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

			userID := s.getUserID(username)
			permissions := map[string]bool{repoCollaboratorOptions.Permission: true}
			ghUser := github.User{ID: new(userID), Login: new(username), Permissions: permissions}
			// this simulates that the user accepts the invitation (mocking the invite response is not supported yet)
			s.groups[owner][repo] = append(collaborators, ghUser)
			s.members = append(s.members, github.Membership{
				Organization: &github.Organization{Login: new(owner)},
				User:         &github.User{ID: new(userID), Login: new(username)},
				Role:         new(repoCollaboratorOptions.Permission),
			})
			invite := github.CollaboratorInvitation{
				Repo: &github.Repository{
					Owner:       &github.User{Login: new(owner)},
					Name:        new(repo),
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
			logger.Debug("mock SCM request", routeLabel, replaceArgs(deleteReposCollaboratorsByOwnerByRepoByUsername, owner, repo, username))

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
	postAppManifestsByCodeConversionsHandler := WithRequestMatchHandler(
		postAppManifestsByCodeConversions,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			code := r.PathValue("code")
			logger.Debug("mock SCM request", routeLabel, replaceArgs(postAppManifestsByCodeConversions, code))
			config, ok := s.appConfigs[code]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusCreated)
			mustWrite(w, config)
		}),
	)
	getUserByIDHandler := WithRequestMatchHandler(
		getUserByID,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := mustParse[int64](r.PathValue("user_id"))
			logger.Debug("mock SCM request", routeLabel, replaceArgs(getUserByID, userID))

			for _, member := range s.members {
				if member.GetUser().GetID() == userID {
					mustWrite(w, member.GetUser())
					return
				}
			}
			// user not found
			w.WriteHeader(http.StatusNotFound)
		}),
	)
	postReposMergeUpstreamByOwnerByRepoHandler := WithRequestMatchHandler(
		postReposMergeUpstreamByOwnerByRepo,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			owner := r.PathValue("owner")
			repo := r.PathValue("repo")
			logger.Debug("mock SCM request", routeLabel, replaceArgs(postReposMergeUpstreamByOwnerByRepo, owner, repo))
			// Always return success for merge-upstream
			result := github.RepoMergeUpstreamResult{
				Message:    new("Successfully fetched and fast-forwarded from upstream"),
				MergeType:  new("fast-forward"),
				BaseBranch: new("main"),
			}
			w.WriteHeader(http.StatusOK)
			mustWrite(w, result)
		}),
	)

	httpClient := NewMockedHTTPClient(
		getOrganizationsByIDHandler,
		getOrgsByOrgHandler,
		patchOrgsByOrgHandler,
		getOrgsReposByOrgHandler,
		postOrgsReposByOrgHandler,
		postReposForksByOwnerByRepoHandler,
		getOrgsMembershipsByOrgByUsernameHandler,
		putOrgsMembershipsByOrgByUsernameHandler,
		patchUserMembershipsOrgsByOrgHandler,
		deleteOrgsMembershipsByOrgByUsernameHandler,
		getReposByOwnerByRepoHandler,
		deleteReposByOwnerByRepoHandler,
		getRepositoriesByIDHandler,
		getReposCommitsByOwnerByRepoByRefHandler,
		getReposCompareByOwnerByRepoByBaseByHeadHandler,
		getReposCollaboratorsByOwnerByRepoHandler,
		putReposCollaboratorsByOwnerByRepoByUsernameHandler,
		deleteReposCollaboratorsByOwnerByRepoByUsernameHandler,
		postReposMergeUpstreamByOwnerByRepoHandler,
		postAppManifestsByCodeConversionsHandler,
		getUserByIDHandler,
	)
	s.GithubSCM = &GithubSCM{
		client:       github.NewClient(httpClient),
		tokenManager: &staticTokenManager{token: "mock-token"},
		providerURL:  "file://" + env.RepositoryPath(),
		createUserClientFn: func(string) *github.Client {
			return github.NewClient(httpClient)
		},
	}
	return s
}
