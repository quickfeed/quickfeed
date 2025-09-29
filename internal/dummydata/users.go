package dummydata

import (
	"fmt"
	"strings"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

func (g generator) users() error {
	// Create a new user for each name in the list
	for i, name := range qtest.Members {
		user := &qf.User{
			Name:      name,
			Login:     name,
			Email:     strings.Replace(name, " ", ".", 1) + "@example.com",
			StudentID: fmt.Sprintf("%06d", i),
		}
		if err := g.db.CreateUser(user); err != nil {
			return err
		}

		for j := range qtest.MockCourses {
			enrollment := &qf.Enrollment{
				CourseID: uint64(j + 1),
				UserID:   user.GetID(),
			}
			if err := g.db.CreateEnrollment(enrollment); err != nil {
				return err
			}

			// Set enrollment status based on the user index
			var enrollmentStatus qf.Enrollment_UserStatus
			if i < 5 {
				enrollmentStatus = qf.Enrollment_TEACHER
			} else if i < 60 {
				enrollmentStatus = qf.Enrollment_STUDENT
			} else {
				enrollmentStatus = qf.Enrollment_PENDING
			}
			enrollment.Status = enrollmentStatus
			if err := g.db.UpdateEnrollment(enrollment); err != nil {
				return err
			}
		}
	}

	return nil
}
