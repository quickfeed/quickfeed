package scm

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"github.com/quickfeed/quickfeed/qf"
)

// FakeSCM implements the SCM interface.
// TODO(meling) many of the methods below are not implemented.
type FakeSCM struct {
	Repositories  map[string]map[string]*Repository
	Organizations map[uint64]*qf.Organization
	Hooks         map[string]*Hook
	Teams         map[uint64]*Team
}

// NewFakeSCMClient returns a new Fake client implementing the SCM interface.
func NewFakeSCMClient() *FakeSCM {
	return &FakeSCM{
		Repositories:  make(map[string]map[string]*Repository),
		Organizations: make(map[uint64]*qf.Organization),
		Hooks:         make(map[string]*Hook),
		Teams:         make(map[uint64]*Team),
	}
}

func (FakeSCM) Clone(_ context.Context, opt *CloneOptions) (string, error) {
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
func (s *FakeSCM) CreateOrganization(_ context.Context, opt *OrganizationOptions) (*qf.Organization, error) {
	id := len(s.Organizations) + 1
	org := &qf.Organization{
		ID:     uint64(id),
		Path:   opt.Path,
		Avatar: "https://avatars3.githubusercontent.com/u/1000" + strconv.Itoa(id) + "?v=3",
	}
	s.Organizations[org.ID] = org
	s.Repositories[opt.Path] = make(map[string]*Repository)
	return org, nil
}

// UpdateOrganization implements the SCM interface.
func (s *FakeSCM) UpdateOrganization(ctx context.Context, opt *OrganizationOptions) error {
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Name}); err != nil {
		return errors.New("organization not found")
	}
	return nil
}

// GetOrganization implements the SCM interface.
func (s *FakeSCM) GetOrganization(_ context.Context, opt *GetOrgOptions) (*qf.Organization, error) {
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
func (s *FakeSCM) CreateRepository(ctx context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Path}); err != nil {
		return nil, err
	}
	if repo, ok := s.Repositories[opt.Organization.Path][opt.Path]; ok {
		return repo, nil
	}
	repo := &Repository{
		ID:      uint64(len(s.Repositories) + 1),
		Path:    opt.Path,
		HTMLURL: "https://example.com/" + opt.Organization.Path + "/" + opt.Path,
		OrgID:   opt.Organization.ID,
	}
	s.Repositories[opt.Organization.Path][opt.Path] = repo
	return repo, nil
}

// GetRepository implements the SCM interface.
func (s *FakeSCM) GetRepository(_ context.Context, opt *RepositoryOptions) (*Repository, error) {
	repo, ok := s.Repositories[opt.Owner][opt.Path]
	if !ok {
		return nil, errors.New("repository not found")
	}
	return repo, nil
}

// GetRepositories implements the SCM interface.
func (s *FakeSCM) GetRepositories(_ context.Context, org *qf.Organization) ([]*Repository, error) {
	courseRepos, ok := s.Repositories[org.Path]
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
func (s *FakeSCM) DeleteRepository(_ context.Context, opt *RepositoryOptions) error {
	if _, ok := s.Repositories[opt.Owner]; !ok {
		return errors.New("organization does not have any repositories")
	}
	delete(s.Repositories[opt.Owner], opt.Path)
	return nil
}

// UpdateRepoAccess implements the SCM interface.
func (s *FakeSCM) UpdateRepoAccess(_ context.Context, repository *Repository, _, _ string) error {
	if _, ok := s.Repositories[repository.Owner][repository.Path]; !ok {
		return errors.New("repository not found")
	}
	return nil
}

// RepositoryIsEmpty implements the SCM interface
func (*FakeSCM) RepositoryIsEmpty(_ context.Context, _ *RepositoryOptions) bool {
	return false
}

// ListHooks implements the SCM interface.
func (s *FakeSCM) ListHooks(_ context.Context, _ *Repository, _ string) ([]*Hook, error) {
	hooks := make([]*Hook, len(s.Hooks))
	for _, v := range s.Hooks {
		hooks = append(hooks, v)
	}
	return hooks, nil
}

// CreateHook implements the SCM interface.
func (s *FakeSCM) CreateHook(_ context.Context, opt *CreateHookOptions) error {
	_, ok := s.Hooks[opt.Organization]
	if ok {
		return errors.New("hook already exists")
	}
	s.Hooks[opt.Organization] = &Hook{
		Name: opt.Organization,
	}
	return nil
}

// CreateTeam implements the SCM interface.
func (s *FakeSCM) CreateTeam(_ context.Context, opt *NewTeamOptions) (*Team, error) {
	newTeam := &Team{
		ID:           uint64(len(s.Teams) + 1),
		Name:         opt.TeamName,
		Organization: opt.Organization,
	}
	s.Teams[newTeam.ID] = newTeam
	return newTeam, nil
}

// DeleteTeam implements the SCM interface.
func (s *FakeSCM) DeleteTeam(_ context.Context, opt *TeamOptions) error {
	if _, ok := s.Teams[opt.TeamID]; !ok {
		return errors.New("repository not found")
	}
	return nil
}

// GetTeam implements the SCM interface
func (s *FakeSCM) GetTeam(_ context.Context, opt *TeamOptions) (*Team, error) {
	team, ok := s.Teams[opt.TeamID]
	if !ok {
		return nil, errors.New("team not found")
	}
	return team, nil
}

// GetTeams implements the SCM interface
func (s *FakeSCM) GetTeams(_ context.Context, org *qf.Organization) ([]*Team, error) {
	var teams []*Team
	for _, team := range s.Teams {
		if team.Organization == org.Path {
			teams = append(teams, team)
		}
	}
	return teams, nil
}

// AddTeamMember implements the scm interface
func (*FakeSCM) AddTeamMember(_ context.Context, _ *TeamMembershipOptions) error {
	return nil
}

// RemoveTeamMember implements the scm interface
func (*FakeSCM) RemoveTeamMember(_ context.Context, _ *TeamMembershipOptions) error {
	return nil
}

// UpdateTeamMembers implements the SCM interface.
func (*FakeSCM) UpdateTeamMembers(_ context.Context, _ *UpdateTeamOptions) error {
	return nil
}

// CreateCloneURL implements the SCM interface.
func (*FakeSCM) CreateCloneURL(_ *URLPathOptions) string {
	return ""
}

// AddTeamRepo implements the SCM interface.
func (*FakeSCM) AddTeamRepo(_ context.Context, _ *AddTeamRepoOptions) error {
	return nil
}

// GetUserName implements the SCM interface.
func (*FakeSCM) GetUserName(_ context.Context) (string, error) {
	return "", nil
}

// GetUserNameByID implements the SCM interface.
func (*FakeSCM) GetUserNameByID(_ context.Context, _ uint64) (string, error) {
	return "", nil
}

// UpdateOrgMembership implements the SCM interface
func (*FakeSCM) UpdateOrgMembership(_ context.Context, _ *OrgMembershipOptions) error {
	return nil
}

// RemoveMember implements the SCM interface
func (*FakeSCM) RemoveMember(_ context.Context, _ *OrgMembershipOptions) error {
	return nil
}

// CreateIssue implements the SCM interface
func (*FakeSCM) CreateIssue(_ context.Context, _ *IssueOptions) (*Issue, error) {
	return nil, ErrNotSupported{
		SCM:    "FakeSCM",
		Method: "CreateIssue",
	}
}

// UpdateIssue implements the SCM interface
func (*FakeSCM) UpdateIssue(_ context.Context, _ *IssueOptions) (*Issue, error) {
	return nil, ErrNotSupported{
		SCM:    "FakeSCM",
		Method: "UpdateIssue",
	}
}

// GetIssue implements the SCM interface
func (*FakeSCM) GetIssue(_ context.Context, _ *RepositoryOptions, _ int) (*Issue, error) {
	return nil, ErrNotSupported{
		SCM:    "FakeSCM",
		Method: "GetIssue",
	}
}

// GetIssues implements the SCM interface
func (*FakeSCM) GetIssues(_ context.Context, _ *RepositoryOptions) ([]*Issue, error) {
	return nil, ErrNotSupported{
		SCM:    "FakeSCM",
		Method: "GetIssues",
	}
}

func (*FakeSCM) DeleteIssue(_ context.Context, _ *RepositoryOptions, _ int) error {
	return nil
}

func (*FakeSCM) DeleteIssues(_ context.Context, _ *RepositoryOptions) error {
	return nil
}

// CreateIssueComment implements the SCM interface
func (*FakeSCM) CreateIssueComment(ctx context.Context, opt *IssueCommentOptions) (int64, error) {
	return 0, ErrNotSupported{
		SCM:    "FakeSCM",
		Method: "CreateIssueComment",
	}
}

// UpdateIssueComment implements the SCM interface
func (*FakeSCM) UpdateIssueComment(ctx context.Context, opt *IssueCommentOptions) error {
	return ErrNotSupported{
		SCM:    "FakeSCM",
		Method: "UpdateIssueComment",
	}
}

// RequestReviewers implements the SCM interface
func (*FakeSCM) RequestReviewers(ctx context.Context, opt *RequestReviewersOptions) error {
	return ErrNotSupported{
		SCM:    "FakeSCM",
		Method: "RequestReviewers",
	}
}

// AcceptRepositoryInvite implements the SCMInvite interface
func (*FakeSCM) AcceptRepositoryInvites(_ context.Context, _ *RepositoryInvitationOptions) error {
	return ErrNotSupported{
		SCM:    "fake",
		Method: "AcceptRepositoryInvites",
	}
}
