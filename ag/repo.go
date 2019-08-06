package ag

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

// IsCourseRepo returns true if the repository is one of the course repo types.
func (t Repository_Type) IsCourseRepo() bool {
	return t == Repository_COURSEINFO || t == Repository_TESTS || t == Repository_SOLUTIONS || t == Repository_ASSIGNMENTS
}

// IsTestsRepo returns true if the repository is a 'tests' type.
func (t Repository) IsTestsRepo() bool {
	return t.RepoType == Repository_TESTS
}

// IsStudentRepo returns true if the repository is a user repo type.
func (t Repository) IsStudentRepo() bool {
	return t.RepoType == Repository_USER || t.RepoType == Repository_GROUP
}

// IsStudentRepo returns true if the repository is a user repo type.
func (t Repository_Type) IsStudentRepo() bool {
	return t == Repository_USER || t == Repository_GROUP
}

// RepoType returns the repository type for the given path name.
func RepoType(path string) (repoType Repository_Type) {
	switch path {
	case InfoRepo:
		repoType = Repository_COURSEINFO
	case AssignmentRepo:
		repoType = Repository_ASSIGNMENTS
	case TestsRepo:
		repoType = Repository_TESTS
	case SolutionsRepo:
		repoType = Repository_SOLUTIONS
	}
	return
}

// GetRemoteIDFor returns the user's remote identity for the given provider.
// If no remote identity for the given provider is found, then nil is returned.
func (user User) GetRemoteIDFor(provider string) *RemoteIdentity {
	var remoteID *RemoteIdentity
	for _, v := range user.RemoteIdentities {
		if v.Provider == provider {
			remoteID = v
			break
		}
	}
	return remoteID
}
