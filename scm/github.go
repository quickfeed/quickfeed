package scm

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"go.uber.org/zap"

	"github.com/google/go-github/v62/github"
	"github.com/gosimple/slug"
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
		providerURL: "github.com",
	}
}

// GetOrganization implements the SCM interface.
func (s *GithubSCM) GetOrganization(ctx context.Context, opt *OrganizationOptions) (*qf.Organization, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("missing fields: %+v", opt)
	}
	var orgNameOrID string
	var gitOrg *github.Organization
	var err error
	// priority is getting the organization by ID
	if opt.ID > 0 {
		orgNameOrID = strconv.Itoa(int(opt.ID))
		gitOrg, _, err = s.client.Organizations.GetByID(ctx, int64(opt.ID))
	} else {
		// if ID not provided, get by name
		orgNameOrID = slug.Make(opt.Name)
		gitOrg, _, err = s.client.Organizations.Get(ctx, orgNameOrID)
	}
	if err != nil || gitOrg == nil {
		return nil, ErrFailedSCM{
			Method:   "GetOrganization",
			Message:  fmt.Sprintf("could not find github organization %s. Make sure it allows third party access.", orgNameOrID), // this message is logged, never sent to user
			GitError: err,
		}
	}

	org := &qf.Organization{
		ScmOrganizationID:   uint64(gitOrg.GetID()),
		ScmOrganizationName: gitOrg.GetLogin(),
	}

	// If getting organization for the purpose of creating a new course,
	// ensure that the organization does not already contain any course repositories.
	if opt.NewCourse {
		repos, err := s.GetRepositories(ctx, org)
		if err != nil {
			return nil, err
		}
		if isDirty(repos) {
			return nil, ErrAlreadyExists
		}
	}

	// If user name is provided, return the organization only if the user is one of its owners.
	if opt.Username != "" {
		// fetch user membership in that organization, if exists
		membership, _, err := s.client.Organizations.GetOrgMembership(ctx, opt.Username, org.ScmOrganizationName)
		if err != nil {
			return nil, ErrFailedSCM{
				Method:   "GetOrganization",
				Message:  fmt.Sprintf("Failed to GetOrganization for (%q, %q)", opt.Username, org.ScmOrganizationName),
				GitError: fmt.Errorf("failed to GetOrgMembership(%q, %q): %w", opt.Username, org.ScmOrganizationName, err),
			}
		}
		// membership role must be "admin", if not, return error (possibly to show user)
		if membership.GetRole() != OrgOwner {
			return nil, fmt.Errorf("%s not owner of organization %s", opt.Username, org.ScmOrganizationName)
		}
	}
	return org, nil
}

// GetRepositories implements the SCM interface.
func (s *GithubSCM) GetRepositories(ctx context.Context, org *qf.Organization) ([]*Repository, error) {
	path := org.GetScmOrganizationName()
	if path == "" {
		return nil, errors.New("organization name must be provided")
	}
	repos, _, err := s.client.Repositories.ListByOrg(ctx, path, nil)
	if err != nil {
		return nil, ErrFailedSCM{
			GitError: err,
			Method:   "GetRepositories",
			Message:  fmt.Sprintf("failed to access repositories for organization %s", path),
		}
	}
	repositories := make([]*Repository, 0, len(repos))
	for _, repo := range repos {
		repositories = append(repositories, toRepository(repo))
	}
	return repositories, nil
}

// RepositoryIsEmpty implements the SCM interface
func (s *GithubSCM) RepositoryIsEmpty(ctx context.Context, opt *RepositoryOptions) bool {
	_, contents, resp, err := s.client.Repositories.GetContents(
		ctx,
		opt.Owner,
		opt.Path,
		"",
		&github.RepositoryContentGetOptions{},
	)
	// GitHub returns 404 both when repository does not exist and when it is empty with no commits.
	// If there are commits but no contents, GitHub returns no error and an empty slice for directory contents.
	// We want to return true if error is 404 or there is no error and no contents, otherwise false.
	return (err != nil && resp.StatusCode == 404) || (err == nil && len(contents) == 0)
}

// UpdateGroupMembers implements the SCM interface
func (s *GithubSCM) UpdateGroupMembers(ctx context.Context, opt *GroupOptions) error {
	if !opt.valid() {
		return fmt.Errorf("missing fields: %+v", opt)
	}

	// find current group members
	oldUsers, _, err := s.client.Repositories.ListCollaborators(ctx, opt.Organization, opt.GroupName, nil)
	if err != nil {
		return ErrFailedSCM{
			GitError: err,
			Method:   "UpdateGroupMembers",
			Message:  fmt.Sprintf("failed to get members for %s/%s", opt.Organization, opt.GroupName),
		}
	}

	// add members that are not already in the group
	for _, member := range opt.Users {
		_, _, err = s.client.Repositories.AddCollaborator(ctx, opt.Organization, opt.GroupName, member, nil)
		if err != nil {
			return ErrFailedSCM{
				GitError: err,
				Method:   "UpdateGroupMembers",
				Message:  fmt.Sprintf("failed to add user %s to repository %s", member, opt.GroupName),
			}
		}
	}

	// remove members that are no longer in the group
	for _, repoMember := range oldUsers {
		member := repoMember.GetLogin()
		if !slices.Contains(opt.Users, member) {
			_, err = s.client.Repositories.RemoveCollaborator(ctx, opt.Organization, opt.GroupName, member)
			if err != nil {
				return ErrFailedSCM{
					GitError: err,
					Method:   "UpdateGroupMembers",
					Message:  fmt.Sprintf("failed to remove user %s from repository %s", member, opt.GroupName),
				}
			}
		}
	}
	return nil
}

// CreateCourse creates repositories for a new course.
func (s *GithubSCM) CreateCourse(ctx context.Context, opt *CourseOptions) ([]*Repository, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("missing fields: %+v", opt)
	}
	// Get and check the organization's suitability for the course
	org, err := s.GetOrganization(ctx, &OrganizationOptions{ID: opt.OrganizationID, NewCourse: true})
	if err != nil {
		return nil, err
	}

	// Set restrictions to prevent students from creating new repositories and prevent access
	// to organization repositories. This will not affect organization owners (teachers).
	defaultPermissions := OrgNone
	createRepoPermissions := false
	if _, _, err = s.client.Organizations.Edit(ctx, org.ScmOrganizationName, &github.Organization{
		DefaultRepoPermission: &defaultPermissions,
		MembersCanCreateRepos: &createRepoPermissions,
	}); err != nil {
		return nil, fmt.Errorf("failed to update permissions for GitHub organization %s: %w", org.ScmOrganizationName, err)
	}

	// Create course repositories
	repositories := make([]*Repository, 0, len(RepoPaths)+1)
	for path, private := range RepoPaths {
		repoOptions := &CreateRepositoryOptions{
			Path:         path,
			Organization: org.ScmOrganizationName,
			Private:      private,
		}
		repo, err := s.createRepository(ctx, repoOptions)
		if err != nil {
			return nil, err
		}
		repositories = append(repositories, repo)
	}

	// Create student repository for the course creator
	repo, err := s.createStudentRepo(ctx, org.ScmOrganizationName, opt.CourseCreator)
	if err != nil {
		return nil, err
	}
	repositories = append(repositories, repo)
	return repositories, nil
}

// UpdateEnrollment updates organization membership and creates user repositories.
func (s *GithubSCM) UpdateEnrollment(ctx context.Context, opt *UpdateEnrollmentOptions) (*Repository, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("missing fields: %+v", opt)
	}
	org, err := s.GetOrganization(ctx, &OrganizationOptions{
		Name: opt.Organization,
	})
	if err != nil {
		return nil, err
	}
	switch opt.Status {
	case qf.Enrollment_STUDENT:
		// Give access to the course's info and assignments repositories
		if err := s.grantPullAccessToCourseRepos(ctx, org.ScmOrganizationName, opt.User); err != nil {
			return nil, err
		}
		repo, err := s.createStudentRepo(ctx, org.ScmOrganizationName, opt.User)
		if err != nil {
			return nil, err
		}
		// Promote user to organization member
		role := OrgMember
		if _, _, err := s.client.Organizations.EditOrgMembership(ctx, opt.User, org.ScmOrganizationName, &github.Membership{Role: &role}); err != nil {
			return nil, err
		}
		return repo, nil

	case qf.Enrollment_TEACHER:
		// Promote user to organization owner
		role := OrgOwner
		if _, _, err := s.client.Organizations.EditOrgMembership(ctx, opt.User, org.ScmOrganizationName, &github.Membership{Role: &role}); err != nil {
			return nil, err
		}
		// Teacher's private (student) repo should already exist
		return nil, nil
	}
	// Only student and teacher enrollments are allowed (NONE and PENDING are not relevant here)
	return nil, fmt.Errorf("invalid enrollment status: %v", opt.Status)
}

// RejectEnrollment removes user's repository and revokes user's membership in the course organization.
// If the user was already removed from the organization an error is returned, and the repository deletion is skipped.
func (s *GithubSCM) RejectEnrollment(ctx context.Context, opt *RejectEnrollmentOptions) error {
	if !opt.valid() {
		return fmt.Errorf("missing fields: %+v", opt)
	}
	org, err := s.GetOrganization(ctx, &OrganizationOptions{ID: opt.OrganizationID})
	if err != nil {
		return err
	}
	// If user was already removed we report the error and skip the repository deletion
	if _, err := s.client.Organizations.RemoveMember(ctx, org.ScmOrganizationName, opt.User); err != nil {
		return err
	}
	return s.deleteRepository(ctx, &RepositoryOptions{ID: opt.RepositoryID})
}

// DemoteTeacherToStudent revokes owner status in the organization.
func (s *GithubSCM) DemoteTeacherToStudent(ctx context.Context, opt *UpdateEnrollmentOptions) error {
	if !opt.valid() {
		return fmt.Errorf("missing fields: %+v", opt)
	}
	role := OrgMember
	_, _, err := s.client.Organizations.EditOrgMembership(ctx, opt.User, opt.Organization, &github.Membership{Role: &role})
	return err
}

// CreateGroup creates repository for a new group.
func (s *GithubSCM) CreateGroup(ctx context.Context, opt *GroupOptions) (*Repository, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("missing fields: %+v", opt)
	}
	orgOptions := &OrganizationOptions{Name: opt.Organization}
	org, err := s.GetOrganization(ctx, orgOptions)
	if err != nil {
		return nil, err
	}
	repoOptions := &CreateRepositoryOptions{
		Organization: opt.Organization,
		Path:         opt.GroupName,
		Private:      true,
	}
	repo, err := s.createRepository(ctx, repoOptions)
	if err != nil {
		return nil, err
	}

	for _, user := range opt.Users {
		if _, _, err := s.client.Repositories.AddCollaborator(ctx, org.ScmOrganizationName, repo.Path, user, &github.RepositoryAddCollaboratorOptions{
			Permission: RepoPush,
		}); err != nil {
			return nil, err
		}
	}

	return repo, nil
}

// DeleteGroup deletes a group's repository.
func (s *GithubSCM) DeleteGroup(ctx context.Context, opt *RepositoryOptions) error {
	return s.deleteRepository(ctx, opt)
}

// createRepository creates a new repository or returns an existing repository with the given name.
func (s *GithubSCM) createRepository(ctx context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("missing fields: %+v", opt)
	}

	// check that repo does not already exist for this user or group
	repo, _, err := s.client.Repositories.Get(ctx, opt.Organization, slug.Make(opt.Path))
	if repo != nil {
		s.logger.Debugf("CreateRepository: found existing repository (skipping creation): %s: %v", opt.Path, repo)
		return toRepository(repo), nil
	}
	// error expected to be 404 Not Found; logging here in case it's a different error
	s.logger.Debugf("CreateRepository: check for repository %s: %s", opt.Path, err)

	// repo does not exist, create it
	s.logger.Debugf("CreateRepository: creating %s", opt.Path)
	repo, _, err = s.client.Repositories.Create(ctx, opt.Organization, &github.Repository{
		Name:    &opt.Path,
		Private: &opt.Private,
	})
	if err != nil {
		return nil, ErrFailedSCM{
			Method:   "CreateRepository",
			Message:  fmt.Sprintf("failed to create repository %s, make sure it does not already exist", opt.Path),
			GitError: err,
		}
	}
	s.logger.Debugf("CreateRepository: done creating %s", opt.Path)
	return toRepository(repo), nil
}

// deleteRepository deletes repository by name or ID.
func (s *GithubSCM) deleteRepository(ctx context.Context, opt *RepositoryOptions) error {
	if !opt.valid() {
		return fmt.Errorf("missing fields: %+v", opt)
	}

	// if ID provided, get path and owner from github
	if opt.ID > 0 {
		repo, _, err := s.client.Repositories.GetByID(ctx, int64(opt.ID))
		if err != nil {
			return ErrFailedSCM{
				GitError: err,
				Method:   "DeleteRepository",
				Message:  fmt.Sprintf("failed to fetch repository %d: may not exists in the course organization", opt.ID),
			}
		}
		opt.Path = repo.GetName()
		opt.Owner = repo.Owner.GetLogin()
	}

	if _, err := s.client.Repositories.Delete(ctx, opt.Owner, opt.Path); err != nil {
		return ErrFailedSCM{
			GitError: err,
			Method:   "DeleteRepository",
			Message:  fmt.Sprintf("failed to delete repository %s", opt.Path),
		}
	}
	return nil
}

// createStudentRepo creates {username}-labs repository and provides pull/push access to it for the given student.
func (s *GithubSCM) createStudentRepo(ctx context.Context, organization string, login string) (*Repository, error) {
	// create repo, or return existing repo if it already exists
	// if repo is found, it is safe to reuse it
	repo, err := s.createRepository(ctx, &CreateRepositoryOptions{
		Organization: organization,
		Path:         qf.StudentRepoName(login),
		Private:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create repo: %w", err)
	}

	// add push access to student repo
	opt := &github.RepositoryAddCollaboratorOptions{
		Permission: RepoPush,
	}
	if _, _, err := s.client.Repositories.AddCollaborator(ctx, repo.Owner, repo.Path, login, opt); err != nil {
		return nil, fmt.Errorf("failed to grant push access to %s/%s for user %s: %w", repo.Owner, repo.Path, login, err)
	}
	return repo, nil
}

// grantPullAccessToCourseRepos gives pull access to the course's info and assignments repositories.
func (s *GithubSCM) grantPullAccessToCourseRepos(ctx context.Context, org, login string) error {
	commonRepos := []string{qf.AssignmentsRepo}
	for _, repoType := range commonRepos {
		opt := &github.RepositoryAddCollaboratorOptions{
			Permission: RepoPull,
		}
		if _, _, err := s.client.Repositories.AddCollaborator(ctx, org, repoType, login, opt); err != nil {
			return fmt.Errorf("failed to grant pull access to %s/%s for user %s: %w", org, repoType, login, err)
		}
	}
	return nil
}

// Client returns GitHub client.
func (s *GithubSCM) Client() *github.Client {
	return s.client
}

func toRepository(repo *github.Repository) *Repository {
	return &Repository{
		ID:      uint64(repo.GetID()),
		Path:    repo.GetName(),
		Owner:   repo.Owner.GetLogin(),
		HTMLURL: repo.GetHTMLURL(),
		OrgID:   uint64(repo.Organization.GetID()),
		Size:    uint64(repo.GetSize()),
	}
}
