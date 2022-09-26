package scm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/quickfeed/quickfeed/qf"
)

// MockSCM implements the SCM interface.
// TODO(meling) many of the methods below are not implemented.
type MockSCM struct {
	Repositories  map[uint64]map[string]*Repository // [Organization ID][Repository Path]*Repository
	Organizations map[uint64]*qf.Organization
	Hooks         map[uint64]*Hook
	Teams         map[uint64]*Team
}

// NewMockSCMClient returns a new mock client implementing the SCM interface.
func NewMockSCMClient() *MockSCM {
	return &MockSCM{
		Repositories:  make(map[uint64]map[string]*Repository),
		Organizations: make(map[uint64]*qf.Organization),
		Hooks:         make(map[uint64]*Hook),
		Teams:         make(map[uint64]*Team),
	}
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

// CreateOrganization implements the SCM interface.
func (s *MockSCM) CreateOrganization(_ context.Context, opt *OrganizationOptions) (*qf.Organization, error) {
	id := len(s.Organizations) + 1
	org := &qf.Organization{
		ID:     uint64(id),
		Path:   opt.Path,
		Avatar: "https://avatars3.githubusercontent.com/u/1000" + strconv.Itoa(id) + "?v=3",
	}
	s.Organizations[org.ID] = org
	s.Repositories[org.ID] = make(map[string]*Repository)
	return org, nil
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
	if opt.ID == 0 {
		for _, org := range s.Organizations {
			if org.Path == opt.Name {
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
		Name: opt.Organization.Path,
	})
	if err != nil {
		return nil, err
	}
	if repo, ok := s.Repositories[org.ID][opt.Path]; ok {
		return repo, nil
	}
	repo := &Repository{
		ID:      uint64(len(s.Repositories) + 1),
		Path:    opt.Path,
		Owner:   org.Path,
		HTMLURL: "https://example.com/" + opt.Organization.Path + "/" + opt.Path,
		OrgID:   opt.Organization.ID,
	}
	s.Repositories[org.ID][opt.Path] = repo
	return repo, nil
}

// GetRepository implements the SCM interface.
func (s *MockSCM) GetRepository(ctx context.Context, opt *RepositoryOptions) (*Repository, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %+v", opt)
	}
	org, err := s.GetOrganization(ctx, &GetOrgOptions{ID: opt.ID})
	if err != nil {
		return nil, err
	}
	repo, ok := s.Repositories[org.ID][opt.Path]
	if !ok {
		return nil, errors.New("repository not found")
	}
	return repo, nil
}

// GetRepositories implements the SCM interface.
func (s *MockSCM) GetRepositories(_ context.Context, org *qf.Organization) ([]*Repository, error) {
	courseRepos, ok := s.Repositories[org.ID]
	if !ok {
		return nil, errors.New("organization does not have any repositories")
	}
	repos := make([]*Repository, 0)
	for _, v := range courseRepos {
		repos = append(repos, v)
	}
	return repos, nil
}

// DeleteRepository implements the SCM interface.
func (s *MockSCM) DeleteRepository(ctx context.Context, opt *RepositoryOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %+v", opt)
	}
	var org *qf.Organization
	if opt.Owner == "" {
		opt.Owner = s.Organizations[1].Path
	}
	org, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Owner})
	if err != nil {
		return err
	}
	delete(s.Repositories[org.ID], opt.Path)
	return nil
}

// UpdateRepoAccess implements the SCM interface.
func (s *MockSCM) UpdateRepoAccess(ctx context.Context, repo *Repository, _, _ string) error {
	if !repo.valid() {
		return fmt.Errorf("invalid argument: %+v", repo)
	}
	org, err := s.GetOrganization(ctx, &GetOrgOptions{ID: repo.OrgID, Name: repo.Owner})
	if err != nil {
		return err
	}
	if _, ok := s.Repositories[org.ID][repo.Path]; !ok {
		return errors.New("repository not found")
	}
	return nil
}

// RepositoryIsEmpty implements the SCM interface
func (*MockSCM) RepositoryIsEmpty(_ context.Context, _ *RepositoryOptions) bool {
	return false
}

// ListHooks implements the SCM interface.
func (s *MockSCM) ListHooks(ctx context.Context, repo *Repository, orgName string) ([]*Hook, error) {
	hooks := make([]*Hook, len(s.Hooks))
	for _, v := range s.Hooks {
		hooks = append(hooks, v)
	}
	return hooks, nil
}

// CreateHook implements the SCM interface.
func (s *MockSCM) CreateHook(ctx context.Context, opt *CreateHookOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %+v", opt)
	}
	hook := &Hook{
		Name: "test",
	}
	if opt.Organization != "" {
		org, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Organization})
		if err != nil {
			return err
		}
		s.Hooks[org.ID] = hook
	} else {
		s.Hooks[opt.Repository.ID] = hook
	}
	return nil
}

// CreateTeam implements the SCM interface.
func (s *MockSCM) CreateTeam(_ context.Context, opt *NewTeamOptions) (*Team, error) {
	newTeam := &Team{
		ID:           uint64(len(s.Teams) + 1),
		Name:         opt.TeamName,
		Organization: opt.Organization,
	}
	s.Teams[newTeam.ID] = newTeam
	return newTeam, nil
}

// DeleteTeam implements the SCM interface.
func (s *MockSCM) DeleteTeam(_ context.Context, opt *TeamOptions) error {
	if _, ok := s.Teams[opt.TeamID]; !ok {
		return errors.New("repository not found")
	}
	return nil
}

// GetTeam implements the SCM interface
func (s *MockSCM) GetTeam(_ context.Context, opt *TeamOptions) (*Team, error) {
	team, ok := s.Teams[opt.TeamID]
	if !ok {
		return nil, errors.New("team not found")
	}
	return team, nil
}

// GetTeams implements the SCM interface
func (s *MockSCM) GetTeams(_ context.Context, org *qf.Organization) ([]*Team, error) {
	var teams []*Team
	for _, team := range s.Teams {
		if team.Organization == org.Path {
			teams = append(teams, team)
		}
	}
	return teams, nil
}

// AddTeamMember implements the scm interface
func (*MockSCM) AddTeamMember(_ context.Context, _ *TeamMembershipOptions) error {
	return nil
}

// RemoveTeamMember implements the scm interface
func (*MockSCM) RemoveTeamMember(_ context.Context, _ *TeamMembershipOptions) error {
	return nil
}

// UpdateTeamMembers implements the SCM interface.
func (*MockSCM) UpdateTeamMembers(_ context.Context, _ *UpdateTeamOptions) error {
	return nil
}

// CreateCloneURL implements the SCM interface.
func (*MockSCM) CreateCloneURL(_ *URLPathOptions) string {
	return ""
}

// AddTeamRepo implements the SCM interface.
func (*MockSCM) AddTeamRepo(_ context.Context, _ *AddTeamRepoOptions) error {
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
func (*MockSCM) UpdateOrgMembership(_ context.Context, _ *OrgMembershipOptions) error {
	return nil
}

// RemoveMember implements the SCM interface
func (*MockSCM) RemoveMember(_ context.Context, _ *OrgMembershipOptions) error {
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
func (*MockSCM) CreateIssueComment(ctx context.Context, opt *IssueCommentOptions) (int64, error) {
	return 0, ErrNotSupported{
		SCM:    "MockSCM",
		Method: "CreateIssueComment",
	}
}

// UpdateIssueComment implements the SCM interface
func (*MockSCM) UpdateIssueComment(ctx context.Context, opt *IssueCommentOptions) error {
	return ErrNotSupported{
		SCM:    "MockSCM",
		Method: "UpdateIssueComment",
	}
}

// RequestReviewers implements the SCM interface
func (*MockSCM) RequestReviewers(ctx context.Context, opt *RequestReviewersOptions) error {
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
