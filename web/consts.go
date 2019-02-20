package web

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
