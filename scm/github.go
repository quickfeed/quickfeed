package scm

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"

	"go.uber.org/zap"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// GithubSCM implements the SCM interface.
type GithubSCM struct {
	logger      *zap.SugaredLogger
	client      *github.Client
	clientV4    *githubv4.Client
	config      *Config
	token       string
	providerURL string
	tokenURL    string
}

// NewGithubSCMClient returns a new Github client implementing the SCM interface.
func NewGithubSCMClient(logger *zap.SugaredLogger, token string) *GithubSCM {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	return &GithubSCM{
		logger:      logger,
		client:      github.NewClient(httpClient),
		clientV4:    githubv4.NewClient(httpClient),
		token:       token,
		providerURL: "https://github.com",
	}
}

// GetOrganization returns the organization specified by the options; if ID is provided,
// the ID is used to fetch the organization, otherwise the name is used.
// If NewCourse is true, the organization is checked for existing course repositories.
// If Username is provided, the organization is only returned if the user is an owner.
// The organization must allow third-party access for this to work.
func (s *GithubSCM) GetOrganization(ctx context.Context, opt *OrganizationOptions) (*qf.Organization, error) {
	const op Op = "GetOrganization"
	if !opt.valid() {
		return nil, E(op, M("failed to get organization"), fmt.Errorf("missing fields: %+v", *opt))
	}
	var githubOrg *github.Organization
	var err error
	if opt.ID > 0 {
		githubOrg, _, err = s.client.Organizations.GetByID(ctx, int64(opt.ID))
		if err != nil {
			return nil, E(op, M("failed to get organization by ID: %d", opt.ID), err)
		}
	} else {
		githubOrg, _, err = s.client.Organizations.Get(ctx, opt.Name)
		if err != nil {
			return nil, E(op, M("failed to get organization %s", opt.Name), err)
		}
	}

	orgName := githubOrg.GetLogin()

	// If getting organization for the purpose of creating a new course,
	// ensure that the organization does not already contain any course repositories.
	if opt.NewCourse {
		repos, err := s.GetRepositories(ctx, orgName)
		if err != nil {
			// this code path can only happen if there is an issue with accessing GitHub since
			// we already checked that the organization exists; returning the underlying error.
			return nil, err
		}
		if isDirty(repos) {
			return nil, E(op, M("%s: course repositories %s: %w", orgName, repoNames, ErrAlreadyExists))
		}
	}

	// If user name is provided, return the organization only if the user is one of its owners.
	// This is used together with NewCourse to ensure that the user has access to create a new course.
	if opt.Username != "" {
		m := M("%s: permission denied", orgName)
		membership, _, err := s.client.Organizations.GetOrgMembership(ctx, opt.Username, orgName)
		if err != nil {
			return nil, E(op, m, fmt.Errorf("failed to get membership: %w", err))
		}
		// membership role must be "admin"
		if membership.GetRole() != OrgOwner {
			return nil, E(op, m, fmt.Errorf("%s: %w", opt.Username, ErrNotOwner))
		}
	}

	return &qf.Organization{ScmOrganizationID: uint64(githubOrg.GetID()), ScmOrganizationName: orgName}, nil
}

// GetRepositories implements the SCM interface.
func (s *GithubSCM) GetRepositories(ctx context.Context, org string) ([]*Repository, error) {
	const op Op = "GetRepositories"
	if org == "" {
		return nil, E(op, "organization name must be provided")
	}
	repos, _, err := s.client.Repositories.ListByOrg(ctx, org, &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return nil, E(op, M("failed to get repositories for %s", org), err)
	}
	repositories := make([]*Repository, len(repos))
	for i, repo := range repos {
		repositories[i] = toRepository(repo)
	}
	return repositories, nil
}

// RepositoryIsEmpty implements the SCM interface
func (s *GithubSCM) RepositoryIsEmpty(ctx context.Context, opt *RepositoryOptions) bool {
	repo, err := s.getRepository(ctx, opt)
	if err != nil {
		s.logger.Error(err)
		return true
	}
	opt.ID, opt.Owner, opt.Repo = repo.ID, repo.Owner, repo.Repo

	_, contents, resp, err := s.client.Repositories.GetContents(ctx, opt.Owner, opt.Repo, "", &github.RepositoryContentGetOptions{})
	s.logger.Debugf("RepositoryIsEmpty: %+v", *opt)
	s.logger.Debugf("RepositoryIsEmpty: err=%v", err)
	s.logger.Debugf("RepositoryIsEmpty: (err != nil && %d == 404) || (err == nil && %d == 0) == %t", resp.StatusCode, len(contents), (err != nil && resp.StatusCode == 404) || (err == nil && len(contents) == 0))

	// GitHub returns 404 both when repository does not exist and when it is empty with no commits.
	// If there are commits but no contents, GitHub returns no error and an empty slice for directory contents.
	// We want to return true if error is 404 or there is no error and no contents, otherwise false.
	return (err != nil && resp.StatusCode == 404) || (err == nil && len(contents) == 0)
}

// CreateCourse creates repositories for a new course.
func (s *GithubSCM) CreateCourse(ctx context.Context, opt *CourseOptions) ([]*Repository, error) {
	const op Op = "CreateCourse"
	m := M("failed to create course")
	if !opt.valid() {
		return nil, E(op, m, fmt.Errorf("missing fields: %+v", *opt))
	}
	// Get and check the organization's suitability for the course
	org, err := s.GetOrganization(ctx, &OrganizationOptions{ID: opt.OrganizationID, Username: opt.CourseCreator, NewCourse: true})
	if err != nil {
		// We want to preserve user errors from GetOrganization, so we return the error as is.
		return nil, err
	}
	// Set restrictions to prevent students from creating new repositories and prevent access
	// to organization repositories. This will not affect organization owners (teachers).
	if _, _, err = s.client.Organizations.Edit(ctx, org.GetScmOrganizationName(), &github.Organization{
		DefaultRepoPermission: github.String("none"),
		MembersCanCreateRepos: github.Bool(false),
		// required to allow forking the assignments repository
		MembersCanForkPrivateRepos: github.Bool(true),
	}); err != nil {
		return nil, E(op, m, fmt.Errorf("failed to update permissions for %s: %w", org.GetScmOrganizationName(), err))
	}

	// Create course repositories
	repositories := make([]*Repository, 0, len(RepoPaths)+1)
	for path, private := range RepoPaths {
		repoOptions := &CreateRepositoryOptions{
			Repo:    path,
			Owner:   org.GetScmOrganizationName(),
			Private: private,
		}
		repo, err := s.createRepository(ctx, repoOptions)
		if err != nil {
			return nil, err
		}
		repositories = append(repositories, repo)
	}

	// Create student repository for the course creator
	repo, err := s.createStudentRepo(ctx, org.GetScmOrganizationName(), opt.CourseCreator)
	if err != nil {
		return nil, err
	}
	repositories = append(repositories, repo)
	return repositories, nil
}

// UpdateEnrollment updates organization membership and creates user repositories.
func (s *GithubSCM) UpdateEnrollment(ctx context.Context, opt *UpdateEnrollmentOptions) (*Repository, error) {
	const op Op = "UpdateEnrollment"
	m := M("failed to update enrollment")
	if !opt.valid() {
		return nil, E(op, m, fmt.Errorf("missing fields: %+v", *opt))
	}
	org, err := s.GetOrganization(ctx, &OrganizationOptions{Name: opt.Organization})
	if err != nil {
		return nil, E(op, m, err)
	}
	switch opt.Status {
	case qf.Enrollment_STUDENT:
		m = M("failed to enroll %s as student in %s", opt.User, org.GetScmOrganizationName())
		if err := s.addUser(ctx, org.GetScmOrganizationName(), qf.AssignmentsRepo, opt.User, pullAccess); err != nil {
			return nil, E(op, m, err)
		}
		repo, err := s.createStudentRepo(ctx, org.GetScmOrganizationName(), opt.User)
		if err != nil {
			return nil, E(op, m, err)
		}
		// Promote user to organization member
		if err := s.updatePermission(ctx, opt.User, org.GetScmOrganizationName(), member); err != nil {
			return nil, E(op, m, err)
		}
		return repo, nil

	case qf.Enrollment_TEACHER:
		m = M("failed to enroll %s as teacher in %s", opt.User, org.GetScmOrganizationName())
		// Promote user to organization admin
		if err := s.updatePermission(ctx, opt.User, org.GetScmOrganizationName(), admin); err != nil {
			return nil, E(op, m, err)
		}
		// Teacher's private (student) repo should already exist
		return nil, nil
	}
	// Only student and teacher enrollments are allowed (NONE and PENDING are not relevant here)
	return nil, E(op, m, fmt.Errorf("invalid enrollment status: %s", opt.Status))
}

// RejectEnrollment removes user's repository and revokes user's membership in the course organization.
// If the user was already removed from the organization an error is returned, and the repository deletion is skipped.
func (s *GithubSCM) RejectEnrollment(ctx context.Context, opt *RejectEnrollmentOptions) error {
	const op Op = "RejectEnrollment"
	m := M("failed to reject enrollment")
	if !opt.valid() {
		return E(op, m, fmt.Errorf("missing fields: %+v", *opt))
	}
	m = M("failed to reject enrollment for %s", opt.User)
	org, err := s.GetOrganization(ctx, &OrganizationOptions{ID: opt.OrganizationID})
	if err != nil {
		return E(op, m, err)
	}
	// If user was already removed we report the error and skip the repository deletion
	if _, err := s.client.Organizations.RemoveMember(ctx, org.GetScmOrganizationName(), opt.User); err != nil {
		return E(op, m, fmt.Errorf("failed to remove user: %w", err))
	}
	if err := s.deleteRepository(ctx, opt.RepositoryID); err != nil {
		return E(op, m, err)
	}
	return nil
}

// DemoteTeacherToStudent revokes owner status in the organization.
func (s *GithubSCM) DemoteTeacherToStudent(ctx context.Context, opt *UpdateEnrollmentOptions) error {
	const op Op = "DemoteTeacherToStudent"
	m := M("failed to demote teacher to student")
	if !opt.valid() {
		return E(op, m, fmt.Errorf("missing fields: %+v", *opt))
	}
	// Demote user to organization member
	if err := s.updatePermission(ctx, opt.User, opt.Organization, member); err != nil {
		return E(op, M("failed to demote teacher %s to student in %s", opt.User, opt.Organization), err)
	}
	return nil
}

// CreateGroup creates repository for a new group.
func (s *GithubSCM) CreateGroup(ctx context.Context, opt *GroupOptions) (*Repository, error) {
	const op Op = "CreateGroup"
	m := M("failed to create group")
	if !opt.valid() {
		return nil, E(op, m, fmt.Errorf("missing fields: %+v", *opt))
	}
	if _, err := s.GetOrganization(ctx, &OrganizationOptions{Name: opt.Organization}); err != nil {
		// organization must exist
		return nil, E(op, m, err)
	}
	if _, err := s.getRepository(ctx, &RepositoryOptions{Owner: opt.Organization, Repo: opt.GroupName}); err == nil {
		// repository must not exist
		return nil, E(op, M("%s: repository %s %w", opt.Organization, opt.GroupName, ErrAlreadyExists))
	}
	repo, err := s.createRepository(ctx, &CreateRepositoryOptions{Owner: opt.Organization, Repo: opt.GroupName, Private: true})
	if err != nil {
		return nil, E(op, m, err)
	}
	for _, user := range opt.Users {
		if err := s.addUser(ctx, opt.Organization, repo.Repo, user, pushAccess); err != nil {
			return nil, E(op, m, err)
		}
	}
	return repo, nil
}

// UpdateGroupMembers implements the SCM interface
func (s *GithubSCM) UpdateGroupMembers(ctx context.Context, opt *GroupOptions) error {
	const op Op = "UpdateGroupMembers"
	m := M("failed to update group members")
	if !opt.valid() {
		return E(op, m, fmt.Errorf("missing fields: %+v", *opt))
	}

	// find current group members
	oldUsers, _, err := s.client.Repositories.ListCollaborators(ctx, opt.Organization, opt.GroupName, nil)
	if err != nil {
		return E(op, m, fmt.Errorf("failed to get members: %w", err))
	}

	// add members that are not already in the group
	for _, user := range opt.Users {
		if err := s.addUser(ctx, opt.Organization, opt.GroupName, user, pushAccess); err != nil {
			return E(op, m, err)
		}
	}

	// remove members that are no longer in the group
	for _, repoMember := range oldUsers {
		user := repoMember.GetLogin()
		if !slices.Contains(opt.Users, user) {
			_, err = s.client.Repositories.RemoveCollaborator(ctx, opt.Organization, opt.GroupName, user)
			if err != nil {
				return E(op, m, fmt.Errorf("failed to remove %s: %w", user, err))
			}
		}
	}
	return nil
}

// DeleteGroup deletes a group's repository.
func (s *GithubSCM) DeleteGroup(ctx context.Context, id uint64) error {
	const op Op = "DeleteGroup"

	// options will be checked in deleteRepository
	if err := s.deleteRepository(ctx, id); err != nil {
		return E(op, M("failed to delete group repository"), err)
	}
	return nil
}

// getRepository fetches a repository by ID or name.
func (s *GithubSCM) getRepository(ctx context.Context, opt *RepositoryOptions) (*Repository, error) {
	const op Op = "getRepository"
	m := M("failed to get repository")
	if !opt.valid() {
		return nil, E(op, m, fmt.Errorf("missing fields: %+v", *opt))
	}

	var repo *github.Repository
	var err error
	if opt.ID > 0 {
		repo, _, err = s.client.Repositories.GetByID(ctx, int64(opt.ID))
		if err != nil {
			return nil, E(op, M("failed to get repository %d", opt.ID), err)
		}
	} else {
		repo, _, err = s.client.Repositories.Get(ctx, opt.Owner, opt.Repo)
		if err != nil {
			return nil, E(op, M("failed to get repository %s/%s", opt.Owner, opt.Repo), err)
		}
	}
	return toRepository(repo), nil
}

// createRepository creates a new repository or returns an existing repository with the given name.
func (s *GithubSCM) createRepository(ctx context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	const op Op = "createRepository"
	m := M("failed to create repository")
	if !opt.valid() {
		return nil, E(op, m, fmt.Errorf("missing fields: %+v", *opt))
	}

	// check that repo does not already exist for this user or group
	repo, resp, err := s.client.Repositories.Get(ctx, opt.Owner, opt.Repo)
	if repo != nil {
		s.logger.Debugf("CreateRepository: found existing repository (skipping creation): %s: %v", opt.Repo, repo)
		return toRepository(repo), nil
	}
	// error expected with response status code to be 404 Not Found
	if resp != nil && resp.StatusCode != http.StatusNotFound {
		s.logger.Errorf("CreateRepository: get repository %s returned unexpected status %d: %v", opt.Repo, resp.StatusCode, err)
	}

	// repo does not exist, create it
	if _, ok := RepoPaths[opt.Repo]; ok {
		// creating a default course repository
		s.logger.Debugf("CreateRepository: creating %s", opt.Repo)
		repo, _, err = s.client.Repositories.Create(ctx, opt.Owner, &github.Repository{
			Name:    github.String(opt.Repo),
			Private: github.Bool(opt.Private),
		})
	} else {
		// forking a student or group repository from the assignments repository
		s.logger.Debugf("CreateRepository: forking student/group repository %s from %s", opt.Repo, qf.AssignmentsRepo)
		_, resp, forkErr := s.client.Repositories.CreateFork(ctx, opt.Owner, qf.AssignmentsRepo, &github.RepositoryCreateForkOptions{
			Organization: opt.Owner,
			Name:         opt.Repo,
			// A forked repository retains the visibility of the parent repository.
			// Since the assignments repository is private, the fork will also be private.
		})
		// GitHub returns 202 Accepted for fork creation, which go-github treats as an error.
		// We ignore this specific error since fork creation is asynchronous.
		if forkErr != nil && (resp == nil || resp.StatusCode != http.StatusAccepted) {
			return nil, E(op, M("failed to create fork %s/%s", opt.Owner, opt.Repo), forkErr)
		}
		// GitHub creates forks asynchronously; wait for the fork to be ready
		repo, err = s.waitForRepository(ctx, opt.Owner, opt.Repo)
		if err != nil {
			return nil, E(op, M("fork %s/%s not ready", opt.Owner, opt.Repo), err)
		}
		s.logger.Debugf("CreateRepository: successfully created fork %s/%s", opt.Owner, opt.Repo)
		return toRepository(repo), nil
	}
	if err != nil {
		return nil, E(op, M("failed to create repository %s/%s", opt.Owner, opt.Repo), err)
	}
	s.logger.Debugf("CreateRepository: successfully created %s/%s", opt.Owner, opt.Repo)
	return toRepository(repo), nil
}

// deleteRepository deletes repository by name or ID.
func (s *GithubSCM) deleteRepository(ctx context.Context, id uint64) error {
	const op Op = "deleteRepository"
	m := M("failed to delete repository")
	if id == 0 {
		return E(op, m, fmt.Errorf("missing ID"))
	}

	repo, _, err := s.client.Repositories.GetByID(ctx, int64(id))
	if err != nil {
		return E(op, m, fmt.Errorf("failed to get repository %d: %w", id, err))
	}

	if _, err := s.client.Repositories.Delete(ctx, repo.GetOwner().GetLogin(), repo.GetName()); err != nil {
		return E(op, M("failed to delete repository %s/%s", repo.GetOwner().GetLogin(), repo.GetName()), err)
	}

	return nil
}

// createStudentRepo creates {username}-labs repository and provides pull/push access to it for the given student.
func (s *GithubSCM) createStudentRepo(ctx context.Context, organization, user string) (*Repository, error) {
	// create repo, or return existing repo if it already exists
	// if repo is found, it is safe to reuse it
	repo, err := s.createRepository(ctx, &CreateRepositoryOptions{
		Owner:   organization,
		Repo:    qf.StudentRepoName(user),
		Private: true,
	})
	if err != nil {
		return nil, err
	}
	if err := s.addUser(ctx, repo.Owner, repo.Repo, user, pushAccess); err != nil {
		return nil, err
	}
	return repo, nil
}

func (s *GithubSCM) updatePermission(ctx context.Context, user, org string, role *github.Membership) error {
	if _, _, err := s.client.Organizations.EditOrgMembership(ctx, user, org, role); err != nil {
		return fmt.Errorf("failed to update to %q: %w", *role.Role, err)
	}
	return nil
}

// addUser adds user to the repository with the specified access level (pull or push).
func (s *GithubSCM) addUser(ctx context.Context, org, repo, user string, access *github.RepositoryAddCollaboratorOptions) error {
	if _, _, err := s.client.Repositories.AddCollaborator(ctx, org, repo, user, access); err != nil {
		return fmt.Errorf("failed to add with %q access: %w", access.Permission, err)
	}
	return nil
}

const (
	// waitForRepoMaxAttempts is the maximum number of attempts to wait for a repository to be ready.
	waitForRepoMaxAttempts = 10
	// waitForRepoInitialDelay is the initial delay between attempts.
	waitForRepoInitialDelay = 1 * time.Second
	// waitForRepoMaxDelay is the maximum delay between attempts.
	waitForRepoMaxDelay = 5 * time.Second
)

// waitForRepository polls until the repository is accessible or max attempts is reached.
// This is necessary because GitHub creates forks asynchronously.
// Returns the repository once it's ready.
func (s *GithubSCM) waitForRepository(ctx context.Context, owner, repo string) (*github.Repository, error) {
	delay := waitForRepoInitialDelay
	for attempt := 1; attempt <= waitForRepoMaxAttempts; attempt++ {
		gotRepo, resp, err := s.client.Repositories.Get(ctx, owner, repo)
		// Repository is ready when we get a 200 OK response and the repo is not nil
		if err == nil && resp.StatusCode == http.StatusOK && gotRepo != nil {
			s.logger.Debugf("waitForRepository: %s/%s ready after %d attempts", owner, repo, attempt)
			return gotRepo, nil
		}
		// 202 Accepted means fork is still being created - continue waiting
		// 404 Not Found also means fork is not ready yet
		if resp != nil && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNotFound {
			s.logger.Warnf("waitForRepository: %s/%s unexpected status %d: %v", owner, repo, resp.StatusCode, err)
		}
		s.logger.Debugf("waitForRepository: %s/%s not ready (attempt %d/%d, status=%d), waiting %v",
			owner, repo, attempt, waitForRepoMaxAttempts, resp.StatusCode, delay)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
		// Exponential backoff with max delay
		delay = min(delay*2, waitForRepoMaxDelay)
	}
	return nil, fmt.Errorf("repository %s/%s not ready after %d attempts", owner, repo, waitForRepoMaxAttempts)
}

// Client returns GitHub client.
func (s *GithubSCM) Client() *github.Client {
	return s.client
}

// SyncFork syncs a forked repository's branch with its upstream repository.
func (s *GithubSCM) SyncFork(ctx context.Context, opt *SyncForkOptions) error {
	const op Op = "SyncFork"
	m := M("failed to sync fork")
	if !opt.valid() {
		return E(op, m, fmt.Errorf("missing fields: %+v", *opt))
	}

	_, resp, err := s.client.Repositories.MergeUpstream(ctx, opt.Organization, opt.Repository, &github.RepoMergeUpstreamRequest{
		Branch: github.String(opt.Branch),
	})
	if err != nil {
		// 409 Conflict indicates a merge conflict - log but don't fail
		if resp != nil && resp.StatusCode == 409 {
			s.logger.Warnf("SyncFork: merge conflict for %s/%s on branch %s", opt.Organization, opt.Repository, opt.Branch)
			return E(op, M("merge conflict for %s/%s", opt.Organization, opt.Repository), err)
		}
		return E(op, M("failed to sync fork %s/%s", opt.Organization, opt.Repository), err)
	}
	s.logger.Debugf("SyncFork: successfully synced %s/%s on branch %s", opt.Organization, opt.Repository, opt.Branch)
	return nil
}

func toRepository(repo *github.Repository) *Repository {
	return &Repository{
		ID:      uint64(repo.GetID()),
		Repo:    repo.GetName(),
		Owner:   repo.Owner.GetLogin(),
		HTMLURL: repo.GetHTMLURL(),
	}
}
