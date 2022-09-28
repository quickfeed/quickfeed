package scm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/quickfeed/quickfeed/qf"
)

var testOrgs = []*qf.Organization{
	{
		ID:   1,
		Name: "qfTestOrg",
	},
	{
		ID:   2,
		Name: "DAT320",
	},
	{
		ID:   3,
		Name: "DATx20-2019",
	},
	{
		ID:   4,
		Name: "DATx20-2020",
	},
}

// MockSCM implements the SCM interface.
// TODO(meling) many of the methods below are not implemented.
type MockSCM struct {
	Repositories  map[uint64]*Repository
	Organizations map[uint64]*qf.Organization
	Hooks         map[uint64]*Hook
	Teams         map[uint64]*Team
}

// NewMockSCMClient returns a new mock client implementing the SCM interface.
func NewMockSCMClient() *MockSCM {
	s := &MockSCM{
		Repositories:  make(map[uint64]*Repository),
		Organizations: make(map[uint64]*qf.Organization),
		Hooks:         make(map[uint64]*Hook),
		Teams:         make(map[uint64]*Team),
	}
	// initialize four test course organizations
	for _, org := range testOrgs {
		s.Organizations[org.ID] = org
	}
	return s
}

func (MockSCM) Clone(_ context.Context, opt *CloneOptions) (string, error) {
	cloneDir := filepath.Join(opt.DestDir, repoDir(opt))
	// This is a hack to make sure the lab1 directory exists,
	// required by the web/rebuild_test.go:TestRebuildSubmissions()
	lab1Dir := filepath.Join(cloneDir, "lab1")
	err := os.MkdirAll(lab1Dir, 0o700)
	if err != nil {
		return "", err
	}
	return cloneDir, err
}

// UpdateOrganization implements the SCM interface.
func (s *MockSCM) UpdateOrganization(ctx context.Context, opt *OrganizationOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %+v", opt)
	}
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Name}); err != nil {
		return errors.New("organization not found")
	}
	return nil
}

// GetOrganization implements the SCM interface.
func (s *MockSCM) GetOrganization(_ context.Context, opt *GetOrgOptions) (*qf.Organization, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %+v", opt)
	}
	if opt.ID < 1 {
		for _, org := range s.Organizations {
			if org.Name == opt.Name {
				return org, nil
			}
		}
	}
	org, ok := s.Organizations[opt.ID]
	if !ok {
		return nil, errors.New("organization not found")
	}
	return org, nil
}

// CreateRepository implements the SCM interface.
func (s *MockSCM) CreateRepository(ctx context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %+v", opt)
	}
	org, err := s.GetOrganization(ctx, &GetOrgOptions{
		ID:   opt.Organization.ID,
		Name: opt.Organization.Name,
	})
	if err != nil {
		return nil, err
	}
	repo := &Repository{
		ID:      GenerateID(s.Repositories),
		Path:    opt.Path,
		Owner:   org.Name,
		HTMLURL: "https://example.com/" + opt.Organization.Name + "/" + opt.Path,
		OrgID:   opt.Organization.ID,
	}
	s.Repositories[repo.ID] = repo
	return repo, nil
}

// GetRepository implements the SCM interface.
func (s *MockSCM) GetRepository(_ context.Context, opt *RepositoryOptions) (*Repository, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %+v", opt)
	}
	if opt.ID > 0 {
		repo, ok := s.Repositories[opt.ID]
		if !ok {
			return nil, errors.New("repository not found")
		}
		return repo, nil
	}
	for _, repo := range s.Repositories {
		if repo.Path == opt.Path && repo.Owner == opt.Owner {
			return repo, nil
		}
	}
	return nil, errors.New("repository not found")
}

// GetRepositories implements the SCM interface.
func (s *MockSCM) GetRepositories(_ context.Context, org *qf.Organization) ([]*Repository, error) {
	var repos []*Repository
	for _, repo := range s.Repositories {
		if repo.OrgID == org.ID {
			repos = append(repos, repo)
		}
	}
	return repos, nil
}

// DeleteRepository implements the SCM interface.
func (s *MockSCM) DeleteRepository(_ context.Context, opt *RepositoryOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %+v", opt)
	}
	if _, ok := s.Repositories[opt.ID]; !ok {
		return errors.New("repository not found")
	}
	delete(s.Repositories, opt.ID)
	return nil
}

// UpdateRepoAccess implements the SCM interface.
func (s *MockSCM) UpdateRepoAccess(ctx context.Context, repo *Repository, _, _ string) error {
	if !repo.valid() {
		return fmt.Errorf("invalid argument: %+v", repo)
	}
	_, err := s.GetRepository(ctx, &RepositoryOptions{
		ID:    repo.ID,
		Path:  repo.Path,
		Owner: repo.Owner,
	})
	return err
}

// RepositoryIsEmpty implements the SCM interface
func (*MockSCM) RepositoryIsEmpty(_ context.Context, _ *RepositoryOptions) bool {
	return false
}

// ListHooks implements the SCM interface.
func (s *MockSCM) ListHooks(_ context.Context, _ *Repository, _ string) ([]*Hook, error) {
	var hooks []*Hook
	for _, v := range s.Hooks {
		hooks = append(hooks, v)
	}
	return hooks, nil
}

// CreateHook implements the SCM interface.
func (s *MockSCM) CreateHook(_ context.Context, opt *CreateHookOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %+v", opt)
	}
	hook := &Hook{
		ID:   GenerateID(s.Hooks),
		Name: opt.Organization,
	}
	s.Hooks[hook.ID] = hook
	return nil
}

// CreateTeam implements the SCM interface.
func (s *MockSCM) CreateTeam(_ context.Context, opt *NewTeamOptions) (*Team, error) {
	newTeam := &Team{
		ID:           GenerateID(s.Teams),
		Name:         opt.TeamName,
		Organization: opt.Organization,
	}
	s.Teams[newTeam.ID] = newTeam
	return newTeam, nil
}

// DeleteTeam implements the SCM interface.
func (s *MockSCM) DeleteTeam(_ context.Context, opt *TeamOptions) error {
	delete(s.Teams, opt.TeamID)
	return nil
}

// GetTeam implements the SCM interface
func (s *MockSCM) GetTeam(_ context.Context, opt *TeamOptions) (*Team, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %+v", opt)
	}
	if opt.TeamID > 0 {
		team, ok := s.Teams[opt.TeamID]
		if !ok {
			return nil, errors.New("team not found")
		}
		return team, nil
	}
	for _, team := range s.Teams {
		if team.Name == opt.TeamName && team.Organization == opt.Organization {
			return team, nil
		}
	}
	return nil, errors.New("team not found")
}

// GetTeams implements the SCM interface
func (s *MockSCM) GetTeams(_ context.Context, org *qf.Organization) ([]*Team, error) {
	var teams []*Team
	for _, team := range s.Teams {
		if team.Organization == org.Name {
			teams = append(teams, team)
		}
	}
	return teams, nil
}

// AddTeamMember implements the scm interface
func (s *MockSCM) AddTeamMember(_ context.Context, opt *TeamMembershipOptions) error {
	if !s.teamExists(opt.TeamID, opt.TeamName, opt.Organization) {
		return errors.New("team not found add")
	}
	return nil
}

// RemoveTeamMember implements the scm interface
func (s *MockSCM) RemoveTeamMember(_ context.Context, opt *TeamMembershipOptions) error {
	if !s.teamExists(opt.TeamID, opt.TeamName, opt.Organization) {
		return errors.New("team not found")
	}
	return nil
}

// UpdateTeamMembers implements the SCM interface.
func (s *MockSCM) UpdateTeamMembers(_ context.Context, opt *UpdateTeamOptions) error {
	if !s.teamExists(opt.TeamID, "", "") {
		return errors.New("team not found")
	}
	return nil
}

// CreateCloneURL implements the SCM interface.
func (*MockSCM) CreateCloneURL(_ *URLPathOptions) string {
	return ""
}

// AddTeamRepo implements the SCM interface.
func (s *MockSCM) AddTeamRepo(_ context.Context, opt *AddTeamRepoOptions) error {
	repo := &Repository{
		ID:    GenerateID(s.Repositories),
		Path:  opt.Repo,
		Owner: opt.Owner,
		OrgID: opt.OrganizationID,
	}
	s.Repositories[repo.ID] = repo
	return nil
}

// GetUserName implements the SCM interface.
func (*MockSCM) GetUserName(_ context.Context) (string, error) {
	return "", nil
}

// GetUserNameByID implements the SCM interface.
func (*MockSCM) GetUserNameByID(_ context.Context, _ uint64) (string, error) {
	return "", nil
}

// UpdateOrgMembership implements the SCM interface
func (s *MockSCM) UpdateOrgMembership(ctx context.Context, opt *OrgMembershipOptions) error {
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Organization}); err != nil {
		return errors.New("organization not found")
	}
	return nil
}

// RemoveMember implements the SCM interface
func (s *MockSCM) RemoveMember(ctx context.Context, opt *OrgMembershipOptions) error {
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Organization}); err != nil {
		return errors.New("organization not found")
	}
	return nil
}

// CreateIssue implements the SCM interface
func (*MockSCM) CreateIssue(_ context.Context, _ *IssueOptions) (*Issue, error) {
	return nil, ErrNotSupported{
		SCM:    "MockSCM",
		Method: "CreateIssue",
	}
}

// UpdateIssue implements the SCM interface
func (*MockSCM) UpdateIssue(_ context.Context, _ *IssueOptions) (*Issue, error) {
	return nil, ErrNotSupported{
		SCM:    "MockSCM",
		Method: "UpdateIssue",
	}
}

// GetIssue implements the SCM interface
func (*MockSCM) GetIssue(_ context.Context, _ *RepositoryOptions, _ int) (*Issue, error) {
	return nil, ErrNotSupported{
		SCM:    "MockSCM",
		Method: "GetIssue",
	}
}

// GetIssues implements the SCM interface
func (*MockSCM) GetIssues(_ context.Context, _ *RepositoryOptions) ([]*Issue, error) {
	return nil, ErrNotSupported{
		SCM:    "MockSCM",
		Method: "GetIssues",
	}
}

func (*MockSCM) DeleteIssue(_ context.Context, _ *RepositoryOptions, _ int) error {
	return nil
}

func (*MockSCM) DeleteIssues(_ context.Context, _ *RepositoryOptions) error {
	return nil
}

// CreateIssueComment implements the SCM interface
func (*MockSCM) CreateIssueComment(_ context.Context, _ *IssueCommentOptions) (int64, error) {
	return 0, ErrNotSupported{
		SCM:    "MockSCM",
		Method: "CreateIssueComment",
	}
}

// UpdateIssueComment implements the SCM interface
func (*MockSCM) UpdateIssueComment(_ context.Context, _ *IssueCommentOptions) error {
	return ErrNotSupported{
		SCM:    "MockSCM",
		Method: "UpdateIssueComment",
	}
}

// RequestReviewers implements the SCM interface
func (*MockSCM) RequestReviewers(_ context.Context, _ *RequestReviewersOptions) error {
	return ErrNotSupported{
		SCM:    "MockSCM",
		Method: "RequestReviewers",
	}
}

// AcceptRepositoryInvite implements the SCMInvite interface
func (*MockSCM) AcceptRepositoryInvites(_ context.Context, _ *RepositoryInvitationOptions) error {
	return ErrNotSupported{
		SCM:    "MockSCM",
		Method: "AcceptRepositoryInvites",
	}
}

// teamExists checks teams by ID, or by team and organization name.
func (s *MockSCM) teamExists(id uint64, team, org string) bool {
	if id > 0 {
		if _, ok := s.Teams[id]; ok {
			return true
		}
	} else {
		for _, t := range s.Teams {
			if t.Name == team && t.Organization == org {
				return true
			}
		}
	}
	return false
}

// GenerateID generates a new, unused map key to use as ID in tests.
func GenerateID[T any](data map[uint64]T) uint64 {
	id := uint64(1)
	_, ok := data[id]
	for ok {
		id++
		_, ok = data[id]
	}
	return id
}
