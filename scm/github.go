package scm

import (
	"context"
	"fmt"
	"net/http"
	"slices"

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
			return nil, E(op, M("failed to get organization by ID: %d", opt.ID), fmt.Errorf("failed to get organization: %w", err))
		}
	} else {
		name := slug.Make(opt.Name)
		githubOrg, _, err = s.client.Organizations.Get(ctx, name)
		if err != nil {
			return nil, E(op, M("failed to get organization %s", opt.Name), fmt.Errorf("failed to get organization: %w", err))
		}
	}

	org := &qf.Organization{
		ScmOrganizationID:   uint64(githubOrg.GetID()),
		ScmOrganizationName: githubOrg.GetLogin(),
	}

	// If getting organization for the purpose of creating a new course,
	// ensure that the organization does not already contain any course repositories.
	if opt.NewCourse {
		repos, err := s.GetRepositories(ctx, org)
		if err != nil {
			// this code path can only happen if there is an issue with accessing GitHub since
			// we already checked that the organization exists; returning the underlying error.
			return nil, err
		}
		if isDirty(repos) {
			m := M("course repositories %s already exist for %s", repoNames, org.ScmOrganizationName)
			return nil, E(op, m, fmt.Errorf("%s: %w", org.ScmOrganizationName, ErrAlreadyExists))
		}
	}

	// If user name is provided, return the organization only if the user is one of its owners.
	// This is used together with NewCourse to ensure that the user has access to create a new course.
	if opt.Username != "" {
		m := M("%s: permission denied for %s", org.ScmOrganizationName, opt.Username)
		membership, _, err := s.client.Organizations.GetOrgMembership(ctx, opt.Username, org.ScmOrganizationName)
		if err != nil {
			return nil, E(op, m, fmt.Errorf("failed to get membership: %w", err))
		}
		// membership role must be "admin"
		if membership.GetRole() != OrgOwner {
			return nil, E(op, m, fmt.Errorf("%s/%s: %w", org.ScmOrganizationName, opt.Username, ErrNotOwner))
		}
	}
	return org, nil
}

// GetRepositories implements the SCM interface.
func (s *GithubSCM) GetRepositories(ctx context.Context, org *qf.Organization) ([]*Repository, error) {
	const op Op = "GetRepositories"
	orgName := org.GetScmOrganizationName()
	if orgName == "" {
		return nil, E(op, "organization name must be provided")
	}
	repos, _, err := s.client.Repositories.ListByOrg(ctx, orgName, nil)
	if err != nil {
		return nil, E(op, M("failed to get repositories for organization %s", orgName), fmt.Errorf("failed to get repositories: %w", err))
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
	opt.ID, opt.Owner, opt.Path = repo.ID, repo.Owner, repo.Path

	_, contents, resp, err := s.client.Repositories.GetContents(ctx, opt.Owner, opt.Path, "", &github.RepositoryContentGetOptions{})
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
	if _, _, err = s.client.Organizations.Edit(ctx, org.ScmOrganizationName, &github.Organization{
		DefaultRepoPermission: github.String("none"),
		MembersCanCreateRepos: github.Bool(false),
	}); err != nil {
		return nil, E(op, m, fmt.Errorf("failed to update permissions for organization %s: %w", org.ScmOrganizationName, err))
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
		m = M("failed to enroll %s as student in %s", opt.User, org.ScmOrganizationName)
		if err := s.addUser(ctx, org.ScmOrganizationName, qf.AssignmentsRepo, opt.User, pullAccess); err != nil {
			return nil, E(op, m, err)
		}
		repo, err := s.createStudentRepo(ctx, org.ScmOrganizationName, opt.User)
		if err != nil {
			return nil, E(op, m, err)
		}
		// Promote user to organization member
		if err := s.updatePermission(ctx, opt.User, org.ScmOrganizationName, member); err != nil {
			return nil, E(op, m, err)
		}
		return repo, nil

	case qf.Enrollment_TEACHER:
		m = M("failed to enroll %s as teacher in %s", opt.User, org.ScmOrganizationName)
		// Promote user to organization admin
		if err := s.updatePermission(ctx, opt.User, org.ScmOrganizationName, admin); err != nil {
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
	m := M("failed to reject enrollment for %s", opt.User)
	if !opt.valid() {
		return E(op, m, fmt.Errorf("missing fields: %+v", *opt))
	}
	org, err := s.GetOrganization(ctx, &OrganizationOptions{ID: opt.OrganizationID})
	if err != nil {
		return E(op, m, err)
	}
	// If user was already removed we report the error and skip the repository deletion
	if _, err := s.client.Organizations.RemoveMember(ctx, org.ScmOrganizationName, opt.User); err != nil {
		return E(op, m, fmt.Errorf("failed to remove user: %w", err))
	}
	if err := s.deleteRepository(ctx, &RepositoryOptions{ID: opt.RepositoryID}); err != nil {
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
	if _, err := s.getRepository(ctx, &RepositoryOptions{Owner: opt.Organization, Path: opt.GroupName}); err == nil {
		// repository must not exist
		return nil, E(op, m, fmt.Errorf("repository %s/%s already exists: %w", opt.Organization, opt.GroupName, ErrAlreadyExists))
	}
	repo, err := s.createRepository(ctx, &CreateRepositoryOptions{Organization: opt.Organization, Path: opt.GroupName, Private: true})
	if err != nil {
		return nil, E(op, m, err)
	}
	for _, user := range opt.Users {
		if err := s.addUser(ctx, opt.Organization, repo.Path, user, pushAccess); err != nil {
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
		return E(op, m, fmt.Errorf("failed to get members for %s/%s: %w", opt.Organization, opt.GroupName, err))
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
				return E(op, m, fmt.Errorf("failed to remove user %s from repository %s/%s: %w", user, opt.Organization, opt.GroupName, err))
			}
		}
	}
	return nil
}

// DeleteGroup deletes a group's repository.
func (s *GithubSCM) DeleteGroup(ctx context.Context, opt *RepositoryOptions) error {
	const op Op = "DeleteGroup"

	// options will be checked in deleteRepository
	if err := s.deleteRepository(ctx, opt); err != nil {
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
			return nil, E(op, m, fmt.Errorf("failed to get repository %d: %w", opt.ID, err))
		}
	} else {
		repo, _, err = s.client.Repositories.Get(ctx, opt.Owner, opt.Path)
		if err != nil {
			return nil, E(op, m, fmt.Errorf("failed to get repository %s/%s: %w", opt.Owner, opt.Path, err))
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
	repo, resp, err := s.client.Repositories.Get(ctx, opt.Organization, opt.Path)
	if repo != nil {
		s.logger.Debugf("CreateRepository: found existing repository (skipping creation): %s: %v", opt.Path, repo)
		return toRepository(repo), nil
	}
	// error expected with response status code to be 404 Not Found
	if resp != nil && resp.StatusCode != http.StatusNotFound {
		s.logger.Errorf("CreateRepository: get repository %s returned unexpected status %d: %v", opt.Path, resp.StatusCode, err)
	}

	// repo does not exist, create it
	s.logger.Debugf("CreateRepository: creating %s", opt.Path)
	repo, _, err = s.client.Repositories.Create(ctx, opt.Organization, &github.Repository{
		Name:    github.String(opt.Path),
		Private: github.Bool(opt.Private),
	})
	if err != nil {
		m = M("failed to create repository %s/%s", opt.Organization, opt.Path)
		return nil, E(op, m, fmt.Errorf("failed to create repository %s/%s: %w", opt.Organization, opt.Path, err))
	}
	s.logger.Debugf("CreateRepository: successfully created %s/%s", opt.Organization, opt.Path)
	return toRepository(repo), nil
}

// deleteRepository deletes repository by name or ID.
func (s *GithubSCM) deleteRepository(ctx context.Context, opt *RepositoryOptions) error {
	const op Op = "deleteRepository"
	m := M("failed to delete repository")
	if !opt.valid() {
		return E(op, m, fmt.Errorf("missing fields: %+v", *opt))
	}

	// if ID provided, get path and owner from github
	if opt.ID > 0 {
		repo, _, err := s.client.Repositories.GetByID(ctx, int64(opt.ID))
		if err != nil {
			return E(op, m, fmt.Errorf("failed to get repository %d: %w", opt.ID, err))
		}
		opt.Path = repo.GetName()
		opt.Owner = repo.Owner.GetLogin()
	}

	if _, err := s.client.Repositories.Delete(ctx, opt.Owner, opt.Path); err != nil {
		return E(op, m, fmt.Errorf("failed to delete repository %s/%s: %w", opt.Owner, opt.Path, err))
	}
	return nil
}

// createStudentRepo creates {username}-labs repository and provides pull/push access to it for the given student.
func (s *GithubSCM) createStudentRepo(ctx context.Context, organization string, user string) (*Repository, error) {
	// create repo, or return existing repo if it already exists
	// if repo is found, it is safe to reuse it
	repo, err := s.createRepository(ctx, &CreateRepositoryOptions{
		Organization: organization,
		Path:         qf.StudentRepoName(user),
		Private:      true,
	})
	if err != nil {
		return nil, err
	}
	if err := s.addUser(ctx, repo.Owner, repo.Path, user, pushAccess); err != nil {
		return nil, err
	}
	return repo, nil
}

func (s *GithubSCM) updatePermission(ctx context.Context, user, org string, role *github.Membership) error {
	if _, _, err := s.client.Organizations.EditOrgMembership(ctx, user, org, role); err != nil {
		return fmt.Errorf("failed to update %s's role to %q in organization %s: %w", user, *role.Role, org, err)
	}
	return nil
}

// addUser adds user to the repository with the specified access level (pull or push).
func (s *GithubSCM) addUser(ctx context.Context, org, repo, user string, access *github.RepositoryAddCollaboratorOptions) error {
	if _, _, err := s.client.Repositories.AddCollaborator(ctx, org, repo, user, access); err != nil {
		return fmt.Errorf("failed to add %s with %q access to %s/%s: %w", user, access.Permission, org, repo, err)
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
