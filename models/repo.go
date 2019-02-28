package models

import "errors"

// RepoType represents the type of repsitory.
type RepoType uint

// TODO(meling) RepoType seems to be missing GroupRepo; decide if we need both.

// The available repository types.
const (
	UserRepo RepoType = iota
	AssignmentsRepo
	TestsRepo
	SolutionsRepo
	CourseInfoRepo
)

// IdentifyRepoTypeFromFrontEnd Identifies a repo type from int.
func IdentifyRepoTypeFromFrontEnd(repoType uint64) (RepoType, error) {
	switch repoType {
	case 0:
		return UserRepo, nil
	case 1:
		return AssignmentsRepo, nil
	case 2:
		return TestsRepo, nil
	case 3:
		return SolutionsRepo, nil
	case 4:
		return CourseInfoRepo, nil
	default:
		return 0, errors.New("Repository type not found")
	}
}

// Repository represents a git repository
type Repository struct {
	ID           uint64   `json:"id"`
	Type         RepoType `json:"type"`
	DirectoryID  uint64   `json:"directoryid"`
	RepositoryID uint64   `json:"repositoryid"`
	UserID       uint64   `json:"userid"`
	GroupID      uint64   `json:"groupid"`
	HTMLURL      string   `json:"htmlurl"`
	// TODO: See if this have a functionality
	// Could be used if we need to get the link to the repo for the frontend
	// Or use the SCM could provide that with the use of RepositoryID
	// Name string `json:"name"`
}

// IsTestsRepo returns true if the repository is a 'tests' type.
func (t Repository) IsTestsRepo() bool {
	return t.Type == TestsRepo
}

// IsStudentRepo returns true if the repository is a user or group repo type.
func (t Repository) IsStudentRepo() bool {
	return t.Type == UserRepo
}
