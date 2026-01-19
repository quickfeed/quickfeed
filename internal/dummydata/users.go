package dummydata

import (
	"fmt"
	"log"
	"strings"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

func (g generator) users() error {
	log.Println("Generating users...")
	// Create a new user for each name in the list
	for i := 1; i < students; i++ {
		name := "User " + fmt.Sprintf("%02d", i)
		if i < len(qtest.Members) {
			name = qtest.Members[i]
		}
		user := &qf.User{
			Name:      name,
			Login:     name,
			Email:     strings.Replace(name, " ", ".", 1) + "@example.com",
			StudentID: fmt.Sprintf("%06d", i),
		}
		log.Printf("Creating user: %q\n", user.Name)
		if err := g.db.CreateUser(user); err != nil {
			return err
		}
		/* Bypassing the Teacher/Admin acceptance of enrollments */

		if i < enrolledStudents+teachers {
			for j := 1; j <= courses; j++ {
				log.Printf("Enrolling user %q in course %q\n", user.Name, qtest.MockCourses[j-1].Name)
				enrollment := &qf.Enrollment{
					CourseID: uint64(j),
					UserID:   user.GetID(),
				}
				if err := g.db.CreateEnrollment(enrollment); err != nil {
					return err
				}
				enrollment.Status = qf.Enrollment_STUDENT
				if i < teachers {
					enrollment.Status = qf.Enrollment_TEACHER
				}
				log.Printf("Enrolling user %q in course %q as %s\n", user.Name, qtest.MockCourses[j-1].Name, enrollment.Status.String())
				if err := g.db.UpdateEnrollment(enrollment); err != nil {
					return err
				}
			}
		} else {
			log.Printf("User %q not enrolled in any courses\n", user.Name)
		}
	}
	return nil
}
