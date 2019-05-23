package models

import (
	"errors"
	"strconv"
)

// RepoType represents the type of repsitory.
type RepoType uint

//TODO(meling) Figure out how this breaks the database content; automigrate only handles adding fields
// I believe the schema should remain the same, but the database content will change depending on the RepoType.
//TODO(meling) RepoType seems to be missing GroupRepo; decide if we need both.
//TODO(meling) Add None to the iota to avoid UserRepo = 0

// The available repository types.
const (
	UserRepo RepoType = iota
	AssignmentsRepo
	TestsRepo
	SolutionsRepo
	CourseInfoRepo
)

// RepoTypeFromString returns the repo type for the provided string identifier.
func RepoTypeFromString(repoStrType string) (repoType RepoType, err error) {
	repoUint, err := strconv.ParseUint(repoStrType, 10, 64)
	if err != nil {
		//TODO(meling) should not use 0 (UserRepo); introduce RepoType None
		return 0, err
	}
	switch repoUint {
	case 0:
		repoType = UserRepo
	case 1:
		repoType = AssignmentsRepo
	case 2:
		repoType = TestsRepo
	case 3:
		repoType = SolutionsRepo
	case 4:
		repoType = CourseInfoRepo
	default:
		err = errors.New("unknown repository type")
	}
	return
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

// IsStudentRepo returns true if the repository is a user or group repo type.
func (t RepoType) IsStudentRepo() bool {
	return t == UserRepo
}

// IsCourseRepo returns true if the repository is one of the course repo types.
func (t RepoType) IsCourseRepo() bool {
	return t == CourseInfoRepo || t == TestsRepo || t == SolutionsRepo || t == AssignmentsRepo
}
