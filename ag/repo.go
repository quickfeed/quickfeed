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
func (t Repository_RepoType) IsCourseRepo() bool {
	return t == Repository_CourseInfo || t == Repository_Tests || t == Repository_Solution || t == Repository_Assignment
}

// IsTestsRepo returns true if the repository is a 'tests' type.
func (t Repository) IsTestsRepo() bool {
	return t.RepoType == Repository_Tests
}

// IsStudentRepo returns true if the repository is a user or group repo type.
func (t Repository) IsStudentRepo() bool {
	return t.RepoType == Repository_User
}

// IsStudentRepo returns true if the repository is a user or group repo type.
func (t Repository_RepoType) IsStudentRepo() bool {
	return t == Repository_User
}

// RepoType returns the repository type for the given path name.
func RepoType(path string) (repoType Repository_RepoType) {
	switch path {
	case InfoRepo:
		repoType = Repository_CourseInfo
	case AssignmentRepo:
		repoType = Repository_Assignment
	case TestsRepo:
		repoType = Repository_Tests
	case SolutionsRepo:
		repoType = Repository_Solution
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
