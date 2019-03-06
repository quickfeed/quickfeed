package web

import "time"

// MaxWait is the maximum time a request is allowed to stay open before
// aborting.
const MaxWait = 10 * time.Minute

//TODO(meling) consider to move these to models along with the RepoType etc.

// Default repository names.
const (
	InfoRepo          = "course-info"
	AssignmentRepo    = "assignments"
	TestsRepo         = "tests"
	SolutionsRepo     = "solutions"
	StudentRepoSuffix = "-labs"
)

// StudentRepoName returns the name of the given student's repository.
func StudentRepoName(userName string) string {
	return userName + StudentRepoSuffix
}
