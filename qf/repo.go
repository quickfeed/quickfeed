package qf

import (
	"fmt"
	"strings"
)

// Default repository names.
const (
	InfoRepo          = "info"
	AssignmentsRepo   = "assignments"
	TestsRepo         = "tests"
	StudentRepoSuffix = "-labs"
)

// StudentRepoName returns the name of the given student's repository.
func StudentRepoName(userName string) string {
	return userName + StudentRepoSuffix
}

type RepoURL struct {
	ProviderURL  string
	Organization string
}

func (r RepoURL) InfoRepoURL() string {
	return fmt.Sprintf("https://%s/%s/%s", r.ProviderURL, r.Organization, InfoRepo)
}

func (r RepoURL) AssignmentsRepoURL() string {
	return fmt.Sprintf("https://%s/%s/%s", r.ProviderURL, r.Organization, AssignmentsRepo)
}

func (r RepoURL) StudentRepoURL(userName string) string {
	return fmt.Sprintf("https://%s/%s/%s", r.ProviderURL, r.Organization, StudentRepoName(userName))
}

func (r RepoURL) GroupRepoURL(groupName string) string {
	return fmt.Sprintf("https://%s/%s/%s", r.ProviderURL, r.Organization, groupName)
}

func (r RepoURL) TestsRepoURL() string {
	return fmt.Sprintf("https://%s/%s/%s", r.ProviderURL, r.Organization, TestsRepo)
}

// IsCourseRepo returns true if the repository is one of the course repo types.
func (t Repository_Type) IsCourseRepo() bool {
	return t == Repository_INFO || t == Repository_TESTS || t == Repository_ASSIGNMENTS
}

// IsUserRepo returns true if the repository is a user repo.
func (t Repository_Type) IsUserRepo() bool {
	return t == Repository_USER
}

// IsGroupRepo returns true if the repository is a group repo.
func (t Repository_Type) IsGroupRepo() bool {
	return t == Repository_GROUP
}

// IsTestsRepo returns true if the repository is a 'tests' repository.
func (t *Repository) IsTestsRepo() bool {
	return t.GetRepoType() == Repository_TESTS
}

// IsAssignmentsRepo returns true if the repository is an 'assignments' repository.
func (t *Repository) IsAssignmentsRepo() bool {
	return t.GetRepoType() == Repository_ASSIGNMENTS
}

// IsStudentRepo returns true if the repository is a user or group repo type.
func (t *Repository) IsStudentRepo() bool {
	return t.GetRepoType() == Repository_USER || t.GetRepoType() == Repository_GROUP
}

// IsStudentRepo returns true if the repository is a user repo type.
func (t Repository_Type) IsStudentRepo() bool {
	return t == Repository_USER || t == Repository_GROUP
}

// IsGroupRepo returns true if the repository is a group repo type.
func (t *Repository) IsGroupRepo() bool {
	return t.GetRepoType() == Repository_GROUP
}

// IsUserRepo returns true if the repository is a user repo type.
func (t *Repository) IsUserRepo() bool {
	return t.GetRepoType() == Repository_USER
}

// GetTestURL returns the tests repository string for this repository.
// This repository can be any repository belonging to a course,
// e.g. a user or group repository.
// Using this method we can avoid a database lookup.
func (t *Repository) GetTestURL() string {
	repoURL := t.GetHTMLURL()
	return repoURL[:strings.LastIndex(repoURL, "/")+1] + TestsRepo
}

// Name returns the name of the repository.
func (t *Repository) Name() string {
	repoURL := t.GetHTMLURL()
	return repoURL[strings.LastIndex(repoURL, "/")+1:]
}

// UserName returns the user name of the repository, without the -labs suffix.
func (t *Repository) UserName() string {
	repoName := t.Name()
	return repoName[:len(repoName)-len(StudentRepoSuffix)]
}

// RepoType returns the repository type for the given path name.
func RepoType(path string) (repoType Repository_Type) {
	switch path {
	case InfoRepo:
		repoType = Repository_INFO
	case AssignmentsRepo:
		repoType = Repository_ASSIGNMENTS
	case TestsRepo:
		repoType = Repository_TESTS
	default:
		if strings.HasSuffix(path, StudentRepoSuffix) {
			repoType = Repository_USER
		} else {
			repoType = Repository_GROUP
		}
	}
	return
}
