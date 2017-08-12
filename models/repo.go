package models

// RepoType represents a type of repsitory
type RepoType uint

// Enum for
const (
	UserRepo RepoType = iota
	AssignmentsRepo
	TestsRepo
	SolutionsRepo
	CourseInfoRepo
)

// Repository represents a git repository
type Repository struct {
	ID uint64 `json:"id"`

	DirectoryID  uint64 `json:"directoryid"`
	RepositoryID uint64 `json:"repositoryid"`
	UserID       uint64 `json:"userid"`

	// TODO: See if this have a functionality
	// Could be used if we need to get the link to the repo for the frontend
	// Or use the SCM could provide that with the use of RepositoryID
	// Name string `json:"name"`

	Type RepoType `json:"type"`
}
