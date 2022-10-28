package scm

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/fileop"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

// MockSCM implements the SCM interface.
type MockSCM struct {
	Repositories  map[uint64]*Repository
	Organizations map[uint64]*qf.Organization
	Teams         map[uint64]*Team
	Issues        map[uint64]*Issue
	IssueComments map[uint64]string
}

// NewMockSCMClient returns a new mock client implementing the SCM interface.
func NewMockSCMClient() *MockSCM {
	s := &MockSCM{
		Repositories:  make(map[uint64]*Repository),
		Organizations: make(map[uint64]*qf.Organization),
		Teams:         make(map[uint64]*Team),
		Issues:        make(map[uint64]*Issue),
		IssueComments: make(map[uint64]string),
	}
	// initialize four test course organizations
	for _, course := range qtest.MockCourses {
		s.Organizations[course.OrganizationID] = &qf.Organization{
			ID:   course.OrganizationID,
			Name: course.OrganizationName,
		}
	}
	return s
}

// Clone copies the repository in testdata to the given destination path.
func (s MockSCM) Clone(ctx context.Context, opt *CloneOptions) (string, error) {
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{
		Name: opt.Organization,
	}); err != nil {
		return "", err
	}
	// Simulate cloning by copying the testdata repository to the destination path.
	testdataSrc := filepath.Join(env.Root(), "testdata", "courses", opt.Organization, opt.Repository)
	if err := fileop.CopyDir(testdataSrc, opt.DestDir); err != nil {
		return "", err
	}
	cloneDir := filepath.Join(opt.DestDir, opt.Repository)
	return cloneDir, nil
}

// GetOrganization implements the SCM interface.
func (s *MockSCM) GetOrganization(ctx context.Context, opt *GetOrgOptions) (*qf.Organization, error) {
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
	if opt.NewCourse {
		repos, err := s.GetRepositories(ctx, org)
		if err != nil {
			return nil, err
		}
		if isDirty(repos) {
			return nil, ErrAlreadyExists
		}
	}
	return org, nil
}

// CreateRepository implements the SCM interface.
func (s *MockSCM) CreateRepository(ctx context.Context, opt *CreateRepositoryOptions) (*Repository, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %+v", opt)
	}
	org, err := s.GetOrganization(ctx, &GetOrgOptions{
		Name: opt.Organization,
	})
	if err != nil {
		return nil, err
	}
	url, err := url.JoinPath("https://example.com", opt.Organization, opt.Path)
	if err != nil {
		return nil, err
	}
	repo := &Repository{
		ID:      generateID(s.Repositories),
		Path:    opt.Path,
		Owner:   org.Name,
		HTMLURL: url,
		OrgID:   org.ID,
	}
	s.Repositories[repo.ID] = repo
	return repo, nil
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

// RepositoryIsEmpty implements the SCM interface
func (*MockSCM) RepositoryIsEmpty(_ context.Context, _ *RepositoryOptions) bool {
	return false
}

// CreateTeam implements the SCM interface.
func (s *MockSCM) CreateTeam(_ context.Context, opt *NewTeamOptions) (*Team, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %+v", opt)
	}
	newTeam := &Team{
		ID:           generateID(s.Teams),
		Name:         opt.TeamName,
		Organization: opt.Organization,
	}
	s.Teams[newTeam.ID] = newTeam
	return newTeam, nil
}

// DeleteTeam implements the SCM interface.
func (s *MockSCM) DeleteTeam(_ context.Context, opt *TeamOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %+v", opt)
	}
	delete(s.Teams, opt.TeamID)
	return nil
}

// AddTeamMember implements the scm interface
func (s *MockSCM) AddTeamMember(_ context.Context, opt *TeamMembershipOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %+v", opt)
	}
	if !s.teamExists(opt.TeamID, opt.TeamName, opt.Organization) {
		return errors.New("team not found")
	}
	return nil
}

// RemoveTeamMember implements the scm interface
func (s *MockSCM) RemoveTeamMember(_ context.Context, opt *TeamMembershipOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %+v", opt)
	}
	if !s.teamExists(opt.TeamID, opt.TeamName, opt.Organization) {
		return errors.New("team not found")
	}
	return nil
}

// UpdateTeamMembers implements the SCM interface.
func (s *MockSCM) UpdateTeamMembers(_ context.Context, opt *UpdateTeamOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %+v", opt)
	}
	if !s.teamExists(opt.TeamID, "", "") {
		return errors.New("team not found")
	}
	return nil
}

// AddTeamRepo implements the SCM interface.
func (s *MockSCM) AddTeamRepo(_ context.Context, opt *AddTeamRepoOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %+v", opt)
	}
	if !s.teamExists(opt.TeamID, opt.Repo, opt.Owner) {
		return errors.New("team not found")
	}
	repo := &Repository{
		ID:    generateID(s.Repositories),
		Path:  opt.Repo,
		Owner: opt.Owner,
		OrgID: opt.OrganizationID,
	}
	s.Repositories[repo.ID] = repo
	return nil
}

// UpdateOrgMembership implements the SCM interface
func (s *MockSCM) UpdateOrgMembership(ctx context.Context, opt *OrgMembershipOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %+v", opt)
	}
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Organization}); err != nil {
		return errors.New("organization not found")
	}
	return nil
}

// RemoveMember implements the SCM interface
func (s *MockSCM) RemoveMember(ctx context.Context, opt *OrgMembershipOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %+v", opt)
	}
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Organization}); err != nil {
		return errors.New("organization not found")
	}
	return nil
}

// CreateIssue implements the SCM interface
func (s *MockSCM) CreateIssue(ctx context.Context, opt *IssueOptions) (*Issue, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %v", opt)
	}
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Organization}); err != nil {
		return nil, errors.New("organization not found")
	}
	if _, err := s.getRepository(&RepositoryOptions{
		Path:  opt.Repository,
		Owner: opt.Organization,
	}); err != nil {
		return nil, errors.New("repository not found")
	}
	id := generateID(s.Issues)
	issue := &Issue{
		ID:         id,
		Title:      opt.Title,
		Repository: opt.Repository,
		Body:       opt.Body,
		Number:     int(id),
	}
	if opt.Assignee != nil {
		issue.Assignee = *opt.Assignee
	}
	s.Issues[issue.ID] = issue
	return issue, nil
}

// UpdateIssue implements the SCM interface
func (s *MockSCM) UpdateIssue(ctx context.Context, opt *IssueOptions) (*Issue, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %v", opt)
	}
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Organization}); err != nil {
		return nil, errors.New("organization not found")
	}
	if _, err := s.getRepository(&RepositoryOptions{
		Path:  opt.Repository,
		Owner: opt.Organization,
	}); err != nil {
		return nil, errors.New("repository not found")
	}
	issue, ok := s.Issues[uint64(opt.Number)]
	if !ok {
		return nil, errors.New("issue not found")
	}
	issue.Title = opt.Title
	issue.Body = opt.Body
	issue.Status = opt.State
	if opt.Assignee != nil {
		issue.Assignee = *opt.Assignee
	}
	return issue, nil
}

// GetIssue implements the SCM interface
func (s *MockSCM) GetIssue(ctx context.Context, opt *RepositoryOptions, issueNumber int) (*Issue, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %v", opt)
	}
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Owner}); err != nil {
		return nil, errors.New("organization not found")
	}
	if _, err := s.getRepository(&RepositoryOptions{
		Path:  opt.Path,
		Owner: opt.Owner,
	}); err != nil {
		return nil, errors.New("repository not found")
	}
	issue, ok := s.Issues[uint64(issueNumber)]
	if !ok {
		return nil, errors.New("issue not found")
	}
	return issue, nil
}

// GetIssues implements the SCM interface
func (s *MockSCM) GetIssues(ctx context.Context, opt *RepositoryOptions) ([]*Issue, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %v", opt)
	}
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Owner}); err != nil {
		return nil, errors.New("organization not found")
	}
	if _, err := s.getRepository(&RepositoryOptions{
		Path:  opt.Path,
		Owner: opt.Owner,
	}); err != nil {
		return nil, errors.New("repository not found")
	}
	var issues []*Issue

	for _, i := range s.Issues {
		if i.Repository == opt.Path {
			issues = append(issues, i)
		}
	}
	return issues, nil
}

func (s *MockSCM) DeleteIssue(ctx context.Context, opt *RepositoryOptions, issueNumber int) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %v", opt)
	}
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Owner}); err != nil {
		return errors.New("organization not found")
	}
	if _, err := s.getRepository(&RepositoryOptions{
		Path:  opt.Path,
		Owner: opt.Owner,
	}); err != nil {
		return errors.New("repository not found")
	}
	delete(s.Issues, uint64(issueNumber))
	return nil
}

func (s *MockSCM) DeleteIssues(ctx context.Context, opt *RepositoryOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %v", opt)
	}
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Owner}); err != nil {
		return errors.New("organization not found")
	}
	if _, err := s.getRepository(&RepositoryOptions{
		Path:  opt.Path,
		Owner: opt.Owner,
	}); err != nil {
		return errors.New("repository not found")
	}
	for _, issue := range s.Issues {
		if issue.Repository == opt.Path {
			delete(s.Issues, issue.ID)
		}
	}
	return nil
}

// CreateIssueComment implements the SCM interface
func (s *MockSCM) CreateIssueComment(ctx context.Context, opt *IssueCommentOptions) (int64, error) {
	if !opt.valid() {
		return 0, fmt.Errorf("invalid argument: %v", opt)
	}
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Organization}); err != nil {
		return 0, errors.New("organization not found")
	}
	if _, err := s.getRepository(&RepositoryOptions{
		Path:  opt.Repository,
		Owner: opt.Organization,
	}); err != nil {
		return 0, errors.New("repository not found")
	}
	if _, ok := s.Issues[uint64(opt.Number)]; !ok {
		return 0, errors.New("issue not found")
	}
	id := generateID(s.IssueComments)
	s.IssueComments[id] = opt.Body
	return int64(id), nil
}

// UpdateIssueComment implements the SCM interface
func (s *MockSCM) UpdateIssueComment(ctx context.Context, opt *IssueCommentOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %v", opt)
	}
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{Name: opt.Organization}); err != nil {
		return errors.New("organization not found")
	}
	if _, err := s.getRepository(&RepositoryOptions{
		Path:  opt.Repository,
		Owner: opt.Organization,
	}); err != nil {
		return errors.New("repository not found")
	}
	if _, ok := s.Issues[uint64(opt.Number)]; !ok {
		return errors.New("issue not found")
	}
	if _, ok := s.IssueComments[uint64(opt.CommentID)]; !ok {
		return errors.New("issue comment not found")
	}
	s.IssueComments[uint64(opt.CommentID)] = opt.Body
	return nil
}

// RequestReviewers implements the SCM interface
func (*MockSCM) RequestReviewers(_ context.Context, _ *RequestReviewersOptions) error {
	return ErrNotSupported{
		SCM:    "MockSCM",
		Method: "RequestReviewers",
	}
}

// AcceptInvitations accepts course invites.
func (*MockSCM) AcceptInvitations(_ context.Context, _ *InvitationOptions) error {
	return ErrNotSupported{
		SCM:    "MockSCM",
		Method: "AcceptInvitations",
	}
}

// CreateCourse creates repositories and teams for a new course.
func (s *MockSCM) CreateCourse(ctx context.Context, opt *NewCourseOptions) ([]*Repository, error) {
	org, err := s.GetOrganization(ctx, &GetOrgOptions{ID: opt.OrganizationID, NewCourse: true})
	if err != nil {
		return nil, err
	}
	var repositories []*Repository
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
		repositories = append(repositories, repo)
	}
	labRepo, err := s.CreateRepository(ctx, &CreateRepositoryOptions{
		Path:         qf.StudentRepoName(opt.CourseCreator),
		Organization: org.Name,
		Private:      true,
	})
	if err != nil {
		return nil, err
	}
	repositories = append(repositories, labRepo)
	teams := []*NewTeamOptions{
		{
			Organization: org.Name,
			TeamName:     TeachersTeam,
			Users:        []string{opt.CourseCreator},
		},
		{
			Organization: org.Name,
			TeamName:     StudentsTeam,
		},
	}
	for _, team := range teams {
		if _, err := s.CreateTeam(ctx, team); err != nil {
			return nil, err
		}
	}
	return repositories, nil
}

func (s *MockSCM) UpdateEnrollment(ctx context.Context, opt *UpdateEnrollmentOptions) (*Repository, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %v", opt)
	}
	org, err := s.GetOrganization(ctx, &GetOrgOptions{
		ID:   opt.Course.OrganizationID,
		Name: opt.Course.OrganizationName,
	})
	if err != nil {
		return nil, errors.New("organization not found")
	}
	if opt.Status == qf.Enrollment_STUDENT {
		return s.CreateRepository(ctx, &CreateRepositoryOptions{
			Organization: org.Name,
			Path:         qf.StudentRepoName(opt.User),
			Private:      true,
			Owner:        org.Name,
		})
	}
	return nil, nil
}

// RejectEnrollment removes user's repository and revokes user's membersip in the course organization.
func (s *MockSCM) RejectEnrollment(ctx context.Context, opt *RejectEnrollmentOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %v", opt)
	}
	if _, err := s.GetOrganization(ctx, &GetOrgOptions{
		ID: opt.OrganizationID,
	}); err != nil {
		return errors.New("organization not found")
	}
	return s.DeleteRepository(ctx, &RepositoryOptions{
		ID: opt.RepositoryID,
	})
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

// generateID generates a new, unused map key to use as ID in tests.
func generateID[T any](data map[uint64]T) uint64 {
	id := uint64(1)
	_, ok := data[id]
	for ok {
		id++
		_, ok = data[id]
	}
	return id
}

// getRepository imitates the check done by GitHub when performing API calls that depend on
// existence of a certain repository.
func (s *MockSCM) getRepository(opt *RepositoryOptions) (*Repository, error) {
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
