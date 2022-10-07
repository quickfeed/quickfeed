package qf

import (
	"path/filepath"

	"github.com/quickfeed/quickfeed/internal/env"
)

func (course *Course) CloneDir() string {
	return filepath.Join(env.RepositoryPath(), course.GetOrganizationName())
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
