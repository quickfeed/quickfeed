package scm

import (
	"context"
	"errors"
	"fmt"
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
	Issues        map[uint64]*Issue
	IssueComments map[uint64]string
}

// NewMockSCMClient returns a new mock client implementing the SCM interface.
func NewMockSCMClient() *MockSCM {
	s := &MockSCM{
		Repositories:  make(map[uint64]*Repository),
		Organizations: make(map[uint64]*qf.Organization),
		Issues:        make(map[uint64]*Issue),
		IssueComments: make(map[uint64]string),
	}
	// initialize four test course organizations
	for _, course := range qtest.MockCourses {
		s.Organizations[course.ScmOrganizationID] = &qf.Organization{
			ScmOrganizationID:   course.ScmOrganizationID,
			ScmOrganizationName: course.ScmOrganizationName,
		}
	}
	return s
}

// NewMockSCMClientWithCourse creates a new mock scm with default course repositories
// associated with qtest.MockOrg mock organization.
func NewMockSCMClientWithCourse() *MockSCM {
	s := NewMockSCMClient()
	s.Repositories = map[uint64]*Repository{
		1: {
			ID:    1,
			Repo:  "info",
			Owner: qtest.MockOrg,
		},
		2: {
			ID:    2,
			Repo:  "assignments",
			Owner: qtest.MockOrg,
		},
		3: {
			ID:    3,
			Repo:  "tests",
			Owner: qtest.MockOrg,
		},
		4: {
			ID:    4,
			Repo:  qf.StudentRepoName("user"),
			Owner: qtest.MockOrg,
		},
	}
	return s
}

// Clone copies the repository in testdata to the given destination path.
func (s *MockSCM) Clone(ctx context.Context, opt *CloneOptions) (string, error) {
	if _, err := s.GetOrganization(ctx, &OrganizationOptions{
		Name: opt.Organization,
	}); err != nil {
		return "", err
	}
	// Simulate cloning by copying the testdata repository to the destination path.
	testdataSrc := filepath.Join(env.TestdataPath(), opt.Organization, opt.Repository)
	if err := fileop.CopyDir(testdataSrc, opt.DestDir); err != nil {
		return "", err
	}
	cloneDir := filepath.Join(opt.DestDir, opt.Repository)
	return cloneDir, nil
}

// GetOrganization implements the SCM interface.
func (s *MockSCM) GetOrganization(ctx context.Context, opt *OrganizationOptions) (*qf.Organization, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %+v", opt)
	}
	if opt.ID < 1 {
		for _, org := range s.Organizations {
			if org.ScmOrganizationName == opt.Name {
				return org, nil
			}
		}
	}
	org, ok := s.Organizations[opt.ID]
	if !ok {
		return nil, errors.New("organization not found")
	}
	if opt.NewCourse {
		repos, err := s.GetRepositories(ctx, org.ScmOrganizationName)
		if err != nil {
			return nil, err
		}
		if isDirty(repos) {
			return nil, ErrAlreadyExists
		}
	}
	return org, nil
}

// GetRepositories implements the SCM interface.
func (s *MockSCM) GetRepositories(_ context.Context, owner string) ([]*Repository, error) {
	var repos []*Repository
	for _, repo := range s.Repositories {
		if repo.Owner == owner {
			repos = append(repos, repo)
		}
	}
	return repos, nil
}

// RepositoryIsEmpty implements the SCM interface
func (s *MockSCM) RepositoryIsEmpty(_ context.Context, opts *RepositoryOptions) bool {
	if _, err := s.getRepository(opts); err != nil {
		return true
	}
	return false
}

// UpdateGroupMembers implements the SCM interface.
func (s *MockSCM) UpdateGroupMembers(_ context.Context, opt *GroupOptions) error {
	if !opt.valid() {
		fmt.Printf("invalid argument: %+v", opt)
		return fmt.Errorf("invalid argument: %+v", opt)
	}
	// TODO: This previously checked if a team existed.
	return nil
}

// CreateIssue implements the SCM interface
func (s *MockSCM) CreateIssue(ctx context.Context, opt *IssueOptions) (*Issue, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %v", opt)
	}
	if _, err := s.GetOrganization(ctx, &OrganizationOptions{Name: opt.Organization}); err != nil {
		return nil, errors.New("organization not found")
	}
	if _, err := s.getRepository(&RepositoryOptions{
		Repo:  opt.Repository,
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
	if _, err := s.GetOrganization(ctx, &OrganizationOptions{Name: opt.Organization}); err != nil {
		return nil, errors.New("organization not found")
	}
	if _, err := s.getRepository(&RepositoryOptions{
		Repo:  opt.Repository,
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
	if _, err := s.GetOrganization(ctx, &OrganizationOptions{Name: opt.Owner}); err != nil {
		return nil, errors.New("organization not found")
	}
	if _, err := s.getRepository(&RepositoryOptions{
		Repo:  opt.Repo,
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
	if _, err := s.GetOrganization(ctx, &OrganizationOptions{Name: opt.Owner}); err != nil {
		return nil, errors.New("organization not found")
	}
	if _, err := s.getRepository(&RepositoryOptions{
		Repo:  opt.Repo,
		Owner: opt.Owner,
	}); err != nil {
		return nil, errors.New("repository not found")
	}
	var issues []*Issue

	for _, i := range s.Issues {
		if i.Repository == opt.Repo {
			issues = append(issues, i)
		}
	}
	return issues, nil
}

func (s *MockSCM) DeleteIssue(ctx context.Context, opt *RepositoryOptions, issueNumber int) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %v", opt)
	}
	if _, err := s.GetOrganization(ctx, &OrganizationOptions{Name: opt.Owner}); err != nil {
		return errors.New("organization not found")
	}
	if _, err := s.getRepository(&RepositoryOptions{
		Repo:  opt.Repo,
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
	if _, err := s.GetOrganization(ctx, &OrganizationOptions{Name: opt.Owner}); err != nil {
		return errors.New("organization not found")
	}
	if _, err := s.getRepository(&RepositoryOptions{
		Repo:  opt.Repo,
		Owner: opt.Owner,
	}); err != nil {
		return errors.New("repository not found")
	}
	for _, issue := range s.Issues {
		if issue.Repository == opt.Repo {
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
	if _, err := s.GetOrganization(ctx, &OrganizationOptions{Name: opt.Organization}); err != nil {
		return 0, errors.New("organization not found")
	}
	if _, err := s.getRepository(&RepositoryOptions{
		Repo:  opt.Repository,
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
	if _, err := s.GetOrganization(ctx, &OrganizationOptions{Name: opt.Organization}); err != nil {
		return errors.New("organization not found")
	}
	if _, err := s.getRepository(&RepositoryOptions{
		Repo:  opt.Repository,
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
func (*MockSCM) AcceptInvitations(_ context.Context, opt *InvitationOptions) (string, error) {
	if !opt.valid() {
		return "", fmt.Errorf("invalid argument: %v", opt)
	}
	return "refresh_token", nil
}

// CreateCourse creates repositories for a new course.
func (s *MockSCM) CreateCourse(ctx context.Context, opt *CourseOptions) ([]*Repository, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %v", opt)
	}
	org, err := s.GetOrganization(ctx, &OrganizationOptions{ID: opt.OrganizationID, NewCourse: true})
	if err != nil {
		return nil, err
	}
	repositories := make([]*Repository, 0, len(RepoPaths)+1)
	for path := range RepoPaths {
		id := generateID(s.Repositories)
		repo := &Repository{
			ID:    id,
			Repo:  path,
			Owner: org.ScmOrganizationName,
		}
		s.Repositories[id] = repo
		repositories = append(repositories, repo)
	}
	id := generateID(s.Repositories)
	labRepo := &Repository{
		ID:    id,
		Repo:  qf.StudentRepoName(opt.CourseCreator),
		Owner: org.ScmOrganizationName,
	}
	s.Repositories[id] = labRepo
	repositories = append(repositories, labRepo)

	return repositories, nil
}

func (s *MockSCM) UpdateEnrollment(ctx context.Context, opt *UpdateEnrollmentOptions) (*Repository, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %v", opt)
	}
	org, err := s.GetOrganization(ctx, &OrganizationOptions{
		Name: opt.Organization,
	})
	if err != nil {
		return nil, errors.New("organization not found")
	}
	var repo *Repository
	if opt.Status == qf.Enrollment_STUDENT {
		id := generateID(s.Repositories)
		repo = &Repository{
			ID:    id,
			Repo:  qf.StudentRepoName(opt.User),
			Owner: org.ScmOrganizationName,
		}
		s.Repositories[id] = repo
	}
	return repo, nil
}

// RejectEnrollment removes user's repository and revokes user's membership in the course organization.
func (s *MockSCM) RejectEnrollment(ctx context.Context, opt *RejectEnrollmentOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %v", opt)
	}
	if _, err := s.GetOrganization(ctx, &OrganizationOptions{
		ID: opt.OrganizationID,
	}); err != nil {
		return errors.New("organization not found")
	}
	delete(s.Repositories, opt.RepositoryID)
	return nil
}

// DemoteTeacherToStudent implements the SCM interface.
func (*MockSCM) DemoteTeacherToStudent(_ context.Context, _ *UpdateEnrollmentOptions) error {
	return nil
}

// CreateGroup creates a repository for a new group.
func (s *MockSCM) CreateGroup(ctx context.Context, opt *GroupOptions) (*Repository, error) {
	if !opt.valid() {
		return nil, fmt.Errorf("invalid argument: %v", opt)
	}
	if _, err := s.GetOrganization(ctx, &OrganizationOptions{
		Name: opt.Organization,
	}); err != nil {
		return nil, errors.New("organization not found")
	}

	id := generateID(s.Repositories)
	repo := &Repository{
		ID:    id,
		Repo:  opt.GroupName,
		Owner: opt.Organization,
	}
	s.Repositories[id] = repo
	return repo, nil
}

// DeleteGroup deletes repository for a group.
func (s *MockSCM) DeleteGroup(ctx context.Context, opt *RepositoryOptions) error {
	if !opt.valid() {
		return fmt.Errorf("invalid argument: %v", opt)
	}
	delete(s.Repositories, opt.ID)
	return nil
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
		if repo.Repo == opt.Repo && repo.Owner == opt.Owner {
			return repo, nil
		}
	}
	return nil, errors.New("repository not found")
}
