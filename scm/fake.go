package scm

import (
	"context"
	"errors"
	"strconv"

	pb "github.com/autograde/quickfeed/ag"
)

// FakeSCM implements the SCM interface.
type FakeSCM struct {
	Repositories  map[uint64]*Repository
	Organizations map[uint64]*pb.Organization
	Hooks         map[uint64]int
	Teams         map[uint64]*Team
}

// NewFakeSCMClient returns a new Fake client implementing the SCM interface.
func NewFakeSCMClient() *FakeSCM {
	return &FakeSCM{
		Repositories:  make(map[uint64]*Repository),
		Organizations: make(map[uint64]*pb.Organization),
		Hooks:         make(map[uint64]int),
		Teams:         make(map[uint64]*Team),
	}
}

// CreateOrganization implements the SCM interface.
func (s *FakeSCM) CreateOrganization(_ context.Context, opt *OrganizationOptions) (*pb.Organization, error) {
	id := len(s.Organizations) + 1
	org := &pb.Organization{
		ID:     uint64(id),
		Path:   opt.Path,
		Avatar: "https://avatars3.githubusercontent.com/u/1000" + strconv.Itoa(id) + "?v=3",
	}
	s.Organizations[org.ID] = org
	return org, nil
}

// UpdateOrganization implements the SCM interface.
func (*FakeSCM) UpdateOrganization(_ context.Context, _ *OrganizationOptions) error {
	// TODO no implementation provided yet
	return nil
}

// GetOrganization implements the SCM interface.
func (s *FakeSCM) GetOrganization(_ context.Context, opt *GetOrgOptions) (*pb.Organization, error) {
	org, ok := s.Organizations[opt.ID]
	if !ok {
		return nil, errors.New("organization not found")
	}
	return org, nil
}

// CreateRepository implements the SCM interface.
func (s *FakeSCM) CreateRepository(_ context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	repo := &Repository{
		ID:      uint64(len(s.Repositories) + 1),
		Path:    opt.Path,
		WebURL:  "https://example.com/" + opt.Organization.Path + "/" + opt.Path,
		SSHURL:  "git@example.com:" + opt.Organization.Path + "/" + opt.Path,
		HTTPURL: "https://example.com/" + opt.Organization.Path + "/" + opt.Path + ".git",
		OrgID:   opt.Organization.ID,
	}
	s.Repositories[repo.ID] = repo
	return repo, nil
}

// GetRepository implements the SCM interface.
func (*FakeSCM) GetRepository(_ context.Context, _ *RepositoryOptions) (*Repository, error) {
	// TODO no implementation provided yet
	return nil, nil
}

// GetRepositories implements the SCM interface.
func (s *FakeSCM) GetRepositories(_ context.Context, org *pb.Organization) ([]*Repository, error) {
	var repos []*Repository
	for _, repo := range s.Repositories {
		if repo.OrgID == org.ID {
			repos = append(repos, repo)
		}
	}
	return repos, nil
}

// DeleteRepository implements the SCM interface.
func (s *FakeSCM) DeleteRepository(_ context.Context, opt *RepositoryOptions) error {
	if _, ok := s.Repositories[opt.ID]; !ok {
		return errors.New("repository not found")
	}
	delete(s.Repositories, opt.ID)
	return nil
}

// UpdateRepoAccess implements the SCM interface.
func (*FakeSCM) UpdateRepoAccess(_ context.Context, _ *Repository, _, _ string) error {
	// TODO no implementation provided yet
	return nil
}

// RepositoryIsEmpty implements the SCM interface
func (*FakeSCM) RepositoryIsEmpty(_ context.Context, _ *RepositoryOptions) bool {
	// TODO no implementation provided yet
	return false
}

// ListHooks implements the SCM interface.
func (*FakeSCM) ListHooks(_ context.Context, _ *Repository, _ string) ([]*Hook, error) {
	// TODO no implementation provided yet
	return nil, nil
}

// CreateHook implements the SCM interface.
func (s *FakeSCM) CreateHook(_ context.Context, opt *CreateHookOptions) error {
	if opt.Repository != nil {
		if _, ok := s.Repositories[opt.Repository.ID]; !ok {
			return errors.New("repository not found")
		}
		s.Hooks[opt.Repository.ID]++
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
	delete(s.Repositories, opt.TeamID)
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
func (s *FakeSCM) GetTeams(_ context.Context, org *pb.Organization) ([]*Team, error) {
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
	// TODO no implementation provided yet
	return nil
}

// RemoveTeamMember implements the scm interface
func (*FakeSCM) RemoveTeamMember(_ context.Context, _ *TeamMembershipOptions) error {
	// TODO no implementation provided yet
	return nil
}

// UpdateTeamMembers implements the SCM interface.
func (*FakeSCM) UpdateTeamMembers(_ context.Context, _ *UpdateTeamOptions) error {
	// TODO no implementation provided yet
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
	// TODO no implementation provided yet
	return nil
}

// RemoveMember implements the SCM interface
func (*FakeSCM) RemoveMember(_ context.Context, _ *OrgMembershipOptions) error {
	// TODO no implementation provided yet
	return nil
}

// GetUserScopes implements the SCM interface
func (*FakeSCM) GetUserScopes(_ context.Context) *Authorization {
	// TODO no implementation provided yet
	return nil
}

// AcceptRepositoryInvite implements the SCM interface
func (*FakeSCM) AcceptRepositoryInvites(_ context.Context, _ *RepositoryInvitationOptions) error {
	// TODO no implementation provided yet
	return nil
}

// CreateIssue implements the SCM interface
func (*FakeSCM) CreateIssue(_ context.Context, _ *CreateIssueOptions) (*Issue, error) {
	// TODO no implementation provided yet
	return nil, ErrNotSupported{
		SCM:    "FakeSCM",
		Method: "CreateIssue",
	}
}

// GetRepoIssue implements the SCM interface
func (*FakeSCM) GetRepoIssue(_ context.Context, _ int, _ *RepositoryOptions) (*Issue, error) {
	// TODO no implementation provided yet
	return nil, ErrNotSupported{
		SCM:    "FakeSCM",
		Method: "GetRepoIssue",
	}
}

// GetRepoIssues implements the SCM interface
func (*FakeSCM) GetRepoIssues(_ context.Context, _ *RepositoryOptions) ([]*Issue, error) {
	// TODO no implementation provided yet
	return nil, ErrNotSupported{
		SCM:    "FakeSCM",
		Method: "GetRepoIssues",
	}
}

// EditRepoIssue implements the SCM interface
func (*FakeSCM) EditRepoIssue(_ context.Context, _ int, _ *CreateIssueOptions) (*Issue, error) {
	// TODO no implementation provided yet
	return nil, ErrNotSupported{
		SCM:    "FakeSCM",
		Method: "EditRepoIssue",
	}
}

// CreateIssueComment implements the SCM interface
func (*FakeSCM) CreateIssueComment(ctx context.Context, number int, opt *IssueCommentOptions) (uint64, error) {
	// TODO no implementation provided yet
	return 0, ErrNotSupported{
		SCM:    "FakeSCM",
		Method: "CreateIssueComment",
	}
}

// EditIssueComment implements the SCM interface
func (*FakeSCM) EditIssueComment(ctx context.Context, commentID int64, opt *IssueCommentOptions) error {
	// TODO no implementation provided yet
	return ErrNotSupported{
		SCM:    "FakeSCM",
		Method: "EditIssueComment",
	}
}

// RequestReviewers implements the SCM interface
func (*FakeSCM) RequestReviewers(ctx context.Context, opt *RequestReviewersOptions) error {
	// TODO no implementation provided yet
	return ErrNotSupported{
		SCM:    "FakeSCM",
		Method: "EditRepoIssue",
	}
}
