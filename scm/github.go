package scm

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/google/go-github/v45/github"
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
		gitOrg, _, err = s.client.Organizations.Get(ctx, slug.Make(opt.Name))
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
		membership, _, err := s.client.Organizations.GetOrgMembership(ctx, opt.Username, slug.Make(opt.Name))
		if err != nil {
			return nil, ErrFailedSCM{
				Method:   "GetOrganization",
				Message:  fmt.Sprintf("Failed to GetOrganization for (%q, %q)", opt.Username, slug.Make(opt.Name)),
				GitError: fmt.Errorf("failed to GetOrgMembership(%q, %q): %w", opt.Username, slug.Make(opt.Name), err),
			}
		}
		// membership role must be "admin", if not, return error (possibly to show user)
		if membership.GetRole() != OrgOwner {
			return nil, ErrNotOwner
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

// UpdateTeamMembers implements the SCM interface
func (s *GithubSCM) UpdateTeamMembers(ctx context.Context, opt *UpdateTeamOptions) error {
	if !opt.valid() {
		return fmt.Errorf("missing fields: %+v", opt)
	}

	// find current team members
	oldUsers, _, err := s.client.Teams.ListTeamMembersByID(ctx, int64(opt.OrganizationID), int64(opt.TeamID), nil)
	if err != nil {
		return ErrFailedSCM{
			GitError: err,
			Method:   "UpdateTeamMember",
			Message:  fmt.Sprintf("failed to get members for team ID %d", opt.TeamID),
		}
	}

	// check whether group members are already in team; add missing members
	for _, member := range opt.Users {
		_, _, err = s.client.Teams.AddTeamMembershipByID(ctx, int64(opt.OrganizationID), int64(opt.TeamID), member, nil)
		if err != nil {
			return ErrFailedSCM{
				GitError: err,
				Method:   "UpdateTeamMember",
				Message:  fmt.Sprintf("failed to add user %s to team ID %d", member, opt.TeamID),
			}
		}
	}

	// check if all the team members are in the new group;
	for _, teamMember := range oldUsers {
		toRemove := true
		for _, groupMember := range opt.Users {
			if teamMember.GetLogin() == groupMember {
				toRemove = false
			}
		}
		if toRemove {
			_, err = s.client.Teams.RemoveTeamMembershipByID(ctx, int64(opt.OrganizationID), int64(opt.TeamID), teamMember.GetLogin())
			if err != nil {
				return ErrFailedSCM{
					GitError: err,
					Method:   "UpdateTeamMember",
					Message:  fmt.Sprintf("failed to remove user %s from team ID %d", teamMember.GetLogin(), opt.TeamID),
				}
			}
		}
	}
	return nil
}

// CreateIssue implements the SCM interface
func (s *GithubSCM) CreateIssue(ctx context.Context, opt *IssueOptions) (*Issue, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("missing fields: %+v", opt)
	}
	newIssue := &github.IssueRequest{
		Title:     &opt.Title,
		Body:      &opt.Body,
		Assignee:  opt.Assignee,
		Assignees: opt.Assignees,
	}

	s.logger.Debugf("Creating issue %q on %s", opt.Title, opt.Repository)
	issue, _, err := s.client.Issues.Create(ctx, opt.Organization, opt.Repository, newIssue)
	if err != nil {
		return nil, ErrFailedSCM{
			Method:   "CreateIssue",
			Message:  fmt.Sprintf("failed to create issue %q", opt.Title),
			GitError: err,
		}
	}
	s.logger.Debugf("Created issue %q", opt.Title)

	return toIssue(issue), nil
}

// UpdateIssue implements the SCM interface
func (s *GithubSCM) UpdateIssue(ctx context.Context, opt *IssueOptions) (*Issue, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("missing fields: %+v", opt)
	}

	issueReq := &github.IssueRequest{
		Title:     &opt.Title,
		Body:      &opt.Body,
		State:     &opt.State,
		Assignee:  opt.Assignee,
		Assignees: opt.Assignees,
	}
	s.logger.Debugf("Updating issue %d on %s", opt.Number, opt.Repository)
	issue, _, err := s.client.Issues.Edit(ctx, opt.Organization, opt.Repository, opt.Number, issueReq)
	if err != nil {
		return nil, ErrFailedSCM{
			Method:   "UpdateIssue",
			Message:  fmt.Sprintf("failed to update issue %d on %s/%s", opt.Number, opt.Organization, opt.Repository),
			GitError: err,
		}
	}
	s.logger.Debugf("Updated issue number %d", opt.Number)
	return toIssue(issue), nil
}

// GetIssue implements the SCM interface
func (s *GithubSCM) GetIssue(ctx context.Context, opt *RepositoryOptions, number int) (*Issue, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("missing fields: %+v", opt)
	}
	issue, _, err := s.client.Issues.Get(ctx, opt.Owner, opt.Path, number)
	if err != nil {
		return nil, ErrFailedSCM{
			Method:   "GetIssue",
			Message:  fmt.Sprintf("failed to get issue %d", number),
			GitError: err,
		}
	}
	return toIssue(issue), nil
}

// GetIssues implements the SCM interface
func (s *GithubSCM) GetIssues(ctx context.Context, opt *RepositoryOptions) ([]*Issue, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("missing fields: %+v", opt)
	}
	issueList, _, err := s.client.Issues.ListByRepo(ctx, opt.Owner, opt.Path, &github.IssueListByRepoOptions{})
	if err != nil {
		return nil, ErrFailedSCM{
			Method:   "GetIssues",
			Message:  fmt.Sprintf("failed to get issues for %s", opt.Path),
			GitError: err,
		}
	}
	var issues []*Issue
	for _, issue := range issueList {
		issues = append(issues, toIssue(issue))
	}

	return issues, nil
}

// RequestReviewers implements the SCM interface
func (s *GithubSCM) RequestReviewers(ctx context.Context, opt *RequestReviewersOptions) error {
	if !opt.valid() {
		return fmt.Errorf("missing fields: %+v", opt)
	}
	reviewersRequest := github.ReviewersRequest{
		Reviewers: opt.Reviewers,
	}
	if _, _, err := s.client.PullRequests.RequestReviewers(ctx, opt.Organization, opt.Repository, opt.Number, reviewersRequest); err != nil {
		return ErrFailedSCM{
			Method:   "RequestReviewers",
			Message:  fmt.Sprintf("failed to request reviewers for pull request #%d on %s/%s", opt.Number, opt.Organization, opt.Repository),
			GitError: err,
		}
	}
	return nil
}

// CreateIssueComment implements the SCM interface
func (s *GithubSCM) CreateIssueComment(ctx context.Context, opt *IssueCommentOptions) (int64, error) {
	if !opt.valid() {
		return 0, fmt.Errorf("missing fields: %+v", opt)
	}
	createdComment, _, err := s.client.Issues.CreateComment(ctx, opt.Organization, opt.Repository, opt.Number, &github.IssueComment{Body: &opt.Body})
	if err != nil {
		return 0, ErrFailedSCM{
			Method:   "CreateIssueComment",
			Message:  fmt.Sprintf("failed to create comment for issue #%d, in repository: %s, for organization: %s", opt.Number, opt.Repository, opt.Organization),
			GitError: err,
		}
	}
	return createdComment.GetID(), nil
}

// UpdateIssueComment implements the SCM interface
func (s *GithubSCM) UpdateIssueComment(ctx context.Context, opt *IssueCommentOptions) error {
	if !opt.valid() {
		return fmt.Errorf("missing fields: %+v", opt)
	}
	if _, _, err := s.client.Issues.EditComment(ctx, opt.Organization, opt.Repository, opt.CommentID, &github.IssueComment{Body: &opt.Body}); err != nil {
		return ErrFailedSCM{
			Method:   "UpdateIssueComment",
			Message:  fmt.Sprintf("failed to edit comment in repository: %s, for organization: %s", opt.Repository, opt.Organization),
			GitError: err,
		}
	}
	return nil
}

// CreateCourse creates repositories and teams for a new course.
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

	// Create teacher team with course creator
	teamOpt := &TeamOptions{
		Organization: org.ScmOrganizationName,
		TeamName:     TeachersTeam,
		Users:        []string{opt.CourseCreator},
	}
	if _, err = s.createTeam(ctx, teamOpt); err != nil {
		return nil, err
	}

	// Create student repository for the course creator
	repo, err := s.createStudentRepo(ctx, org.ScmOrganizationName, opt.CourseCreator)
	if err != nil {
		return nil, err
	}
	repositories = append(repositories, repo)
	return repositories, nil
}

// UpdateEnrollment updates organization and team membership and creates user repositories.
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
		return s.createStudentRepo(ctx, org.ScmOrganizationName, opt.User)

	case qf.Enrollment_TEACHER:
		// Promote user to organization owner
		role := OrgOwner
		if _, _, err := s.client.Organizations.EditOrgMembership(ctx, opt.User, org.ScmOrganizationName, &github.Membership{Role: &role}); err != nil {
			return nil, err
		}
		err = s.promoteToTeacher(ctx, org.ScmOrganizationName, opt.User)
	}
	return nil, err
}

// RejectEnrollment removes user's repository and revokes user's membership in the course organization.
func (s *GithubSCM) RejectEnrollment(ctx context.Context, opt *RejectEnrollmentOptions) error {
	if !opt.valid() {
		return fmt.Errorf("missing fields: %+v", opt)
	}
	org, err := s.GetOrganization(ctx, &OrganizationOptions{ID: opt.OrganizationID})
	if err != nil {
		return err
	}
	if _, err := s.client.Organizations.RemoveMember(ctx, org.ScmOrganizationName, opt.User); err != nil {
		return err
	}
	return s.deleteRepository(ctx, &RepositoryOptions{ID: opt.RepositoryID})
}

// DemoteTeacherToStudent removes user from teachers team, revokes owner status in the organization.
func (s *GithubSCM) DemoteTeacherToStudent(ctx context.Context, opt *UpdateEnrollmentOptions) error {
	if !opt.valid() {
		return fmt.Errorf("missing fields: %+v", opt)
	}
	if _, err := s.client.Teams.RemoveTeamMembershipBySlug(ctx, opt.Organization, TeachersTeam, opt.User); err != nil {
		return err
	}
	role := OrgMember
	_, _, err := s.client.Organizations.EditOrgMembership(ctx, opt.User, opt.Organization, &github.Membership{Role: &role})
	return err
}

// CreateGroup creates team and repository for a new group.
func (s *GithubSCM) CreateGroup(ctx context.Context, opt *TeamOptions) (*Repository, *Team, error) {
	if !opt.valid() {
		return nil, nil, fmt.Errorf("missing fields: %+v", opt)
	}
	orgOptions := &OrganizationOptions{Name: opt.Organization}
	org, err := s.GetOrganization(ctx, orgOptions)
	if err != nil {
		return nil, nil, err
	}
	repoOptions := &CreateRepositoryOptions{
		Organization: opt.Organization,
		Path:         opt.TeamName,
		Private:      true,
	}
	repo, err := s.createRepository(ctx, repoOptions)
	if err != nil {
		return nil, nil, err
	}

	team, err := s.createTeam(ctx, opt)
	if err != nil {
		return nil, nil, err
	}
	permissions := &github.TeamAddTeamRepoOptions{
		Permission: RepoPush, // make sure users can pull and push
	}
	if _, err := s.client.Teams.AddTeamRepoByID(ctx, int64(org.ScmOrganizationID), int64(team.ID), org.ScmOrganizationName, repo.Path, permissions); err != nil {
		return nil, nil, err
	}
	return repo, team, nil
}

// DeleteGroup deletes group's repository and team.
func (s *GithubSCM) DeleteGroup(ctx context.Context, opt *GroupOptions) error {
	if !opt.valid() {
		return fmt.Errorf("missing fields: %+v", opt)
	}
	if err := s.deleteRepository(ctx, &RepositoryOptions{ID: opt.RepositoryID}); err != nil {
		return err
	}
	_, err := s.client.Teams.DeleteTeamByID(ctx, int64(opt.OrganizationID), int64(opt.TeamID))
	return err
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

// createTeam creates a new GitHub team.
func (s *GithubSCM) createTeam(ctx context.Context, opt *TeamOptions) (*Team, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("missing fields: %+v", opt)
	}
	teamName := slug.Make(opt.TeamName)
	// check that the team name does not already exist for this organization
	team, _, err := s.client.Teams.GetTeamBySlug(ctx, opt.Organization, teamName)
	if err != nil {
		// error expected to be 404 Not Found; logging here in case it's a different error
		s.logger.Debugf("CreateTeam: check for team %s: %s", teamName, err)
	}

	if team == nil {
		s.logger.Debugf("CreateTeam: creating %s", teamName)
		team, _, err = s.client.Teams.CreateTeam(ctx, opt.Organization, github.NewTeam{
			Name: teamName,
		})
		if err != nil {
			return nil, ErrFailedSCM{
				Method:   "CreateTeam",
				Message:  fmt.Sprintf("failed to create GitHub team %s, make sure it does not already exist", opt.TeamName),
				GitError: fmt.Errorf("failed to create GitHub team %s: %w", opt.TeamName, err),
			}
		}
		s.logger.Debugf("CreateTeam: done creating %s", teamName)
	}
	for _, user := range opt.Users {
		s.logger.Debugf("CreateTeam: adding user %s to %s", user, teamName)
		_, _, err = s.client.Teams.AddTeamMembershipByID(ctx, team.GetOrganization().GetID(), team.GetID(), user, nil)
		if err != nil {
			return nil, ErrFailedSCM{
				Method:   "CreateTeam",
				Message:  fmt.Sprintf("failed to add user '%s' to GitHub team '%s'", user, team.GetName()),
				GitError: fmt.Errorf("failed to add '%s' to GitHub team '%s': %w", user, team.GetName(), err),
			}
		}
	}
	return &Team{
		ID:           uint64(team.GetID()),
		Name:         team.GetName(),
		Organization: team.GetOrganization().GetLogin(),
	}, nil
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
	commonRepos := []string{qf.InfoRepo, qf.AssignmentsRepo}
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

// promoteToTeacher adds user to the organization's "teachers" team.
func (s *GithubSCM) promoteToTeacher(ctx context.Context, org, login string) error {
	teamMaintainer := &github.TeamAddTeamMembershipOptions{Role: TeamMaintainer}
	_, _, err := s.client.Teams.AddTeamMembershipBySlug(ctx, org, TeachersTeam, login, teamMaintainer)
	return err
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

func toIssue(issue *github.Issue) *Issue {
	return &Issue{
		ID:         uint64(issue.GetID()),
		Title:      issue.GetTitle(),
		Body:       issue.GetBody(),
		Repository: issue.Repository.GetName(),
		Assignee:   issue.Assignee.GetName(),
		Number:     issue.GetNumber(),
		Status:     issue.GetState(),
	}
}
