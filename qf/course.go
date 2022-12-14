package qf

import (
	"path/filepath"

	"github.com/quickfeed/quickfeed/internal/env"
)

// HasUpdatedDockerfile returns true if the given dockerfile is different
// from the course's previous Dockerfile.
func (course *Course) HasUpdatedDockerfile(dockerfile string) bool {
	return dockerfile != "" && dockerfile != course.Dockerfile
}

func (course *Course) CloneDir() string {
	return filepath.Join(env.RepositoryPath(), course.GetScmOrganizationName())
}

func (course *Course) TeacherEnrollments() []*Enrollment {
	enrolledTeachers := []*Enrollment{}
	for _, enrollment := range course.Enrollments {
		if enrollment.IsTeacher() {
			enrolledTeachers = append(enrolledTeachers, enrollment)
		}
	}
	return enrolledTeachers
}

// Dummy implementation of the interceptor.userIDs interface.
// Marks this message type to be evaluated for token refresh.
func (*Course) UserIDs() []uint64 {
	return []uint64{}
}
