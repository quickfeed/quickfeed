package scm

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
func (s *GithubSCM) GetOrganization(ctx context.Context, opt *GetOrgOptions) (*qf.Organization, error) {
	if !opt.valid() {
		return nil, ErrMissingFields{
			Method:  "GetOrganization",
			Message: fmt.Sprintf("%+v", opt),
		}
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
		ID:          uint64(gitOrg.GetID()),
		Name:        gitOrg.GetLogin(),
		Avatar:      gitOrg.GetAvatarURL(),
		PaymentPlan: gitOrg.GetPlan().GetName(),
	}

	// If getting organization for a new course, make sure id does not have any corse repositories.
	if opt.NewCourse {
		repos, err := s.GetRepositories(ctx, org)
		if err != nil {
			return nil, err
		}
		if isDirty(repos) {
			return nil, ErrAlreadyExists
		}
	}

	// if user name is provided, return the found organization only if the user is one of its owners
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

// CreateRepository implements the SCM interface.
func (s *GithubSCM) CreateRepository(ctx context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	if !opt.valid() {
		return nil, ErrMissingFields{
			Method:  "CreateRepository",
			Message: fmt.Sprintf("%+v", opt),
		}
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

// GetRepositories implements the SCM interface.
func (s *GithubSCM) GetRepositories(ctx context.Context, org *qf.Organization) ([]*Repository, error) {
	if !org.IsValid() {
		return nil, ErrMissingFields{
			Method:  "GetRepositories",
			Message: fmt.Sprintf("%+v", org),
		}
	}
	var path string
	if org.Name != "" {
		path = org.Name
	} else {
		opt := &GetOrgOptions{
			ID: org.ID,
		}
		org, err := s.GetOrganization(ctx, opt)
		if err != nil {
			return nil, err
		}
		path = org.Name
	}

	repos, _, err := s.client.Repositories.ListByOrg(ctx, path, nil)
	if err != nil {
		return nil, ErrFailedSCM{
			GitError: err,
			Method:   "GetRepositories",
			Message:  fmt.Sprintf("failed to access repositories for organization %s", path),
		}
	}

	var repositories []*Repository
	for _, repo := range repos {
		repositories = append(repositories, toRepository(repo))
	}
	return repositories, nil
}

// DeleteRepository implements the SCM interface.
func (s *GithubSCM) DeleteRepository(ctx context.Context, opt *RepositoryOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %+v", opt)
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

// RepositoryIsEmpty implements the SCM interface
func (s *GithubSCM) RepositoryIsEmpty(ctx context.Context, opt *RepositoryOptions) bool {
	_, _, err := s.client.Repositories.Get(ctx, opt.Owner, opt.Path)
	if err != nil {
		return false
	}

	// test to check how repo commits look like
	_, _, err = s.client.Repositories.ListCommits(ctx, opt.Owner, opt.Path, nil)
	if err != nil {
		if strings.Contains(err.Error(), "Git Repository is empty") {
			return true
		}
	}
	return false
}

// CreateTeam implements the SCM interface.
func (s *GithubSCM) CreateTeam(ctx context.Context, opt *NewTeamOptions) (*Team, error) {
	if !opt.valid() || opt.TeamName == "" || opt.Organization == "" {
		return nil, ErrMissingFields{
			Method:  "CreateTeam",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

	// check that the team name does not already exist for this organization
	team, _, err := s.client.Teams.GetTeamBySlug(ctx, slug.Make(opt.Organization), slug.Make(opt.TeamName))
	if err != nil {
		// error expected to be 404 Not Found; logging here in case it's a different error
		s.logger.Debugf("CreateTeam: check for team %s: %s", opt.TeamName, err)
	}

	if team == nil {
		s.logger.Debugf("CreateTeam: creating %s", opt.TeamName)
		team, _, err = s.client.Teams.CreateTeam(ctx, opt.Organization, github.NewTeam{
			Name: opt.TeamName,
		})
		if err != nil {
			if opt.TeamName != TeachersTeam && opt.TeamName != StudentsTeam {
				return nil, ErrFailedSCM{
					Method:   "CreateTeam",
					Message:  fmt.Sprintf("failed to create GitHub team %s, make sure it does not already exist", opt.TeamName),
					GitError: fmt.Errorf("failed to create GitHub team %s: %w", opt.TeamName, err),
				}
			}
			// continue if it is one of standard teacher/student teams. Such teams can be safely reused
			s.logger.Debugf("Team %s already exists on organization %s", opt.TeamName, opt.Organization)
		}
		s.logger.Debugf("CreateTeam: done creating %s", opt.TeamName)
	}
	for _, user := range opt.Users {
		s.logger.Debugf("CreateTeam: adding user %s to %s", user, opt.TeamName)
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

// DeleteTeam implements the SCM interface.
func (s *GithubSCM) DeleteTeam(ctx context.Context, opt *TeamOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "DeleteTeam",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

	var err error
	if opt.TeamID > 0 {
		_, err = s.client.Teams.DeleteTeamByID(ctx, int64(opt.OrganizationID), int64(opt.TeamID))
	} else {
		_, err = s.client.Teams.DeleteTeamBySlug(ctx, slug.Make(opt.Organization), slug.Make(opt.TeamName))
	}

	if err != nil {
		return ErrFailedSCM{
			Method:   "DeleteTeam",
			Message:  fmt.Sprintf("failed to delete GitHub team '%s'", opt.TeamName),
			GitError: fmt.Errorf("failed to get GitHub team '%s': %w", opt.TeamName, err),
		}
	}
	return err
}

// RemoveTeamMember implements the scm interface
func (s *GithubSCM) RemoveTeamMember(ctx context.Context, opt *TeamMembershipOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "RemoveTeamMember",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

	var err error
	if opt.TeamID < 1 {
		_, err = s.client.Teams.RemoveTeamMembershipBySlug(ctx, opt.Organization, opt.TeamName, opt.Username)
	} else {
		_, err = s.client.Teams.RemoveTeamMembershipByID(ctx, int64(opt.OrganizationID), int64(opt.TeamID), opt.Username)
	}

	if err != nil {
		err = ErrFailedSCM{
			GitError: err,
			Method:   "RemoveTeamMember",
			Message:  fmt.Sprintf("failed to remove user %s from team ID %d", opt.Username, opt.TeamID),
		}
	}
	return err
}

// UpdateTeamMembers implements the SCM interface
func (s *GithubSCM) UpdateTeamMembers(ctx context.Context, opt *UpdateTeamOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "UpdateTeamMembers",
			Message: fmt.Sprintf("%+v", opt),
		}
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

// AddTeamRepo implements the SCM interface.
func (s *GithubSCM) AddTeamRepo(ctx context.Context, opt *AddTeamRepoOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "AddTeamRepo",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

	_, err := s.client.Teams.AddTeamRepoByID(ctx, int64(opt.OrganizationID), int64(opt.TeamID), opt.Owner, opt.Repo,
		&github.TeamAddTeamRepoOptions{
			Permission: opt.Permission, // make sure users can pull and push
		})
	if err != nil {
		return ErrFailedSCM{
			GitError: fmt.Errorf("failed to make GitHub repository '%s' a team repository for team %d: %w", opt.Repo, opt.TeamID, err),
			Method:   "AddTeamRepo",
			Message:  fmt.Sprintf("failed to make GitHub repository '%s' a team repository", opt.Repo),
		}
	}
	return nil
}

// UpdateOrgMembership implements the SCM interface
func (s *GithubSCM) UpdateOrgMembership(ctx context.Context, opt *OrgMembershipOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "UpdateOrgMembership",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	newMembership, _, err := s.client.Organizations.EditOrgMembership(ctx, opt.Username, opt.Organization, &github.Membership{Role: &opt.Role})
	if err != nil || newMembership.GetRole() != opt.Role {
		// Note: the error here is potentially nil
		return ErrFailedSCM{
			GitError: fmt.Errorf("failed to update membership for user %s in organization %s: %w", opt.Username, opt.Organization, err),
			Method:   "UpdateOrgMembership",
			Message:  fmt.Sprintf("failed to update membership for user %s", opt.Username),
		}
	}
	return nil
}

// RemoveMember implements the SCM interface
func (s *GithubSCM) RemoveMember(ctx context.Context, opt *OrgMembershipOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "RemoveMember",
			Message: fmt.Sprintf("%+v", opt),
		}
	}

	// remove user from the organization and all teams
	_, err := s.client.Organizations.RemoveMember(ctx, opt.Organization, opt.Username)
	if err != nil {
		return ErrFailedSCM{
			Method:   "RemoveMember",
			GitError: fmt.Errorf("failed to remove user %s from organization %s: %w", opt.Username, opt.Organization, err),
			Message:  fmt.Sprintf("failed to remove user %s from the organization", opt.Username),
		}
	}
	return nil
}

// CreateIssue implements the SCM interface
func (s *GithubSCM) CreateIssue(ctx context.Context, opt *IssueOptions) (*Issue, error) {
	if !opt.valid() {
		return nil, ErrMissingFields{
			Method:  "CreateIssue",
			Message: fmt.Sprintf("%+v", opt),
		}
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
		return nil, ErrMissingFields{
			Method:  "UpdateIssue",
			Message: fmt.Sprintf("%+v", opt),
		}
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
		return nil, ErrMissingFields{
			Method:  "GetIssue",
			Message: fmt.Sprintf("%+v", opt),
		}
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
		return nil, ErrMissingFields{
			Method:  "GetIssues",
			Message: fmt.Sprintf("%+v", opt),
		}
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
		return ErrMissingFields{
			Method:  "RequestReviewers",
			Message: fmt.Sprintf("%+v", opt),
		}
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
		return 0, ErrMissingFields{
			Method:  "CreateIssueComment",
			Message: fmt.Sprintf("%+v", opt),
		}
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
		return ErrMissingFields{
			Method:  "UpdateIssueComment",
			Message: fmt.Sprintf("%+v", opt),
		}
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
func (s *GithubSCM) CreateCourse(ctx context.Context, opt *NewCourseOptions) ([]*Repository, error) {
	org, err := s.GetOrganization(ctx, &GetOrgOptions{ID: opt.OrganizationID, NewCourse: true})
	if err != nil {
		return nil, err
	}

	// Restrict ability to create new repositories and default access to the organization repositories
	// for students. This will not affect organization owners (teachers).
	DefaultPermissions := OrgNone
	CreateRepoPermissions := false

	if _, _, err = s.client.Organizations.Edit(ctx, org.Name, &github.Organization{
		DefaultRepoPermission: &DefaultPermissions,
		MembersCanCreateRepos: &CreateRepoPermissions,
	}); err != nil {
		return nil, fmt.Errorf("failed to update permissions for GitHub organization %s: %w", org.Name, err)
	}
	repositories := make([]*Repository, len(RepoPaths)+1)
	// create course repositories
	i := 0
	for path, private := range RepoPaths {
		repoOptions := &CreateRepositoryOptions{
			Path:         path,
			Organization: org.Name,
			Private:      private,
		}
		repo, err := s.CreateRepository(ctx, repoOptions)
		if err != nil {
			return nil, err
		}
		repositories[i] = repo
		i++
	}
	// create teacher team with course creator
	teamOpt := &NewTeamOptions{
		Organization: org.Name,
		TeamName:     TeachersTeam,
		Users:        []string{opt.CourseCreator},
	}
	if _, err = s.CreateTeam(ctx, teamOpt); err != nil {
		return nil, fmt.Errorf("failed to create teachers team: %w", err)
	}
	// create student team without any members
	studOpt := &NewTeamOptions{Organization: org.Name, TeamName: StudentsTeam}
	if _, err = s.CreateTeam(ctx, studOpt); err != nil {
		return nil, fmt.Errorf("failed to create students team: %w", err)
	}
	// add student repo for the course creator
	repo, err := s.createStudentRepo(ctx, org, opt.CourseCreator)
	if err != nil {
		return nil, err
	}
	repositories[i] = repo
	return repositories, nil
}

// UpdateEnrollment updates organization and team membership and creates user repositories.
func (s *GithubSCM) UpdateEnrollment(ctx context.Context, opt *UpdateEnrollmentOptions) (*Repository, error) {
	if !opt.valid() {
		return nil, ErrMissingFields{
			Method:  "UpdateEnrollment",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	org, err := s.GetOrganization(ctx, &GetOrgOptions{
		Name: opt.Organization,
	})
	if err != nil {
		return nil, err
	}
	switch opt.Status {
	case qf.Enrollment_STUDENT:
		// give access to the course's info and assignments repositories
		if err := s.grantAccessToCourseRepos(ctx, org.Name, opt.User); err != nil {
			return nil, err
		}
		// add student to the organization's "students" team
		if err := s.addUserToStudentsTeam(ctx, org.Name, opt.User); err != nil {
			return nil, err
		}
		return s.createStudentRepo(ctx, org, opt.User)
	case qf.Enrollment_TEACHER:
		// promote to organization owners
		if err := s.UpdateOrgMembership(ctx, &OrgMembershipOptions{
			Organization: org.Name,
			Username:     opt.User,
			Role:         OrgOwner,
		}); err != nil {
			return nil, err
		}
		err = s.promoteToTeachers(ctx, org.Name, opt.User)
	}
	return nil, err
}

// RejectEnrollment removes user's repository and revokes user's membersip in the course organization.
func (s *GithubSCM) RejectEnrollment(ctx context.Context, opt *RejectEnrollmentOptions) error {
	if !opt.valid() {
		return ErrMissingFields{
			Method:  "RejectEnrollment",
			Message: fmt.Sprintf("%+v", opt),
		}
	}
	org, err := s.GetOrganization(ctx, &GetOrgOptions{
		ID: opt.OrganizationID,
	})
	if err != nil {
		return err
	}
	if err := s.RemoveMember(ctx, &OrgMembershipOptions{
		Organization: org.Name,
		Username:     opt.User,
	}); err != nil {
		return err
	}
	return s.DeleteRepository(ctx, &RepositoryOptions{ID: opt.RepositoryID})
}

// RevokeTeacherStatus removes user from teachers team, revokes owner status in the organization.
func (s *GithubSCM) RevokeTeacherStatus(ctx context.Context, opt *UpdateEnrollmentOptions) error {
	teamOpts := &TeamMembershipOptions{
		Organization: opt.Organization,
		TeamName:     TeachersTeam,
		Username:     opt.User,
	}
	if err := s.RemoveTeamMember(ctx, teamOpts); err != nil {
		return err
	}
	teamOpts.TeamName = StudentsTeam
	if _, _, err := s.client.Teams.AddTeamMembershipBySlug(ctx, opt.Organization, StudentsTeam, opt.User,
		&github.TeamAddTeamMembershipOptions{Role: TeamMember}); err != nil {
		return err
	}
	return s.UpdateOrgMembership(ctx, &OrgMembershipOptions{
		Organization: opt.Organization,
		Username:     opt.User,
		Role:         OrgMember,
	})
}

// createStudentRepo creates {username}-labs repository and provides pull/push access to it for the given student.
func (s *GithubSCM) createStudentRepo(ctx context.Context, org *qf.Organization, login string) (*Repository, error) {
	// create repo, or return existing repo if it already exists
	// if repo is found, it is safe to reuse it
	repo, err := s.CreateRepository(ctx, &CreateRepositoryOptions{
		Organization: org.Name,
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
		return nil, fmt.Errorf("failed to update repo push access: %w", err)
	}
	return repo, nil
}

func (s *GithubSCM) grantAccessToCourseRepos(ctx context.Context, org, login string) error {
	commonRepos := []string{qf.InfoRepo, qf.AssignmentsRepo}

	for _, repoType := range commonRepos {
		opt := &github.RepositoryAddCollaboratorOptions{
			Permission: RepoPull,
		}
		if _, _, err := s.client.Repositories.AddCollaborator(ctx, org, repoType, login, opt); err != nil {
			return fmt.Errorf("updateReposAndTeams: failed to update repo access to repo %s for user %s: %w ", repoType, login, err)
		}
	}
	return nil
}

// addUserToStudentsTeam adds user to the organization's "students" team.
func (s *GithubSCM) addUserToStudentsTeam(ctx context.Context, org, login string) error {
	_, _, err := s.client.Teams.AddTeamMembershipBySlug(ctx, org, StudentsTeam, login,
		&github.TeamAddTeamMembershipOptions{Role: TeamMember})
	return err
}

// add user to the organization's "teachers" team, and remove user from "students" team.
func (s *GithubSCM) promoteToTeachers(ctx context.Context, org, login string) error {
	studentsTeam := &TeamMembershipOptions{
		Organization: org,
		Username:     login,
		TeamName:     StudentsTeam,
	}
	if err := s.RemoveTeamMember(ctx, studentsTeam); err != nil {
		return err
	}
	_, _, err := s.client.Teams.AddTeamMembershipBySlug(ctx, org, TeachersTeam, login,
		&github.TeamAddTeamMembershipOptions{Role: TeamMaintainer})
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
