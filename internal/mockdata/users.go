package mockdata

import (
	"fmt"
	"log"
	"strings"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

func (g *generator) users() error {
	for i := 1; i < g.Students(); i++ {
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
		g.log("Creating user: %q\n", user.Name)
		if err := g.db.CreateUser(user); err != nil {
			return err
		}
		/* Bypassing the Teacher/Admin acceptance of enrollments */

		// TODO(Joachim): Move expression to a getter with a more descriptive name
		if i < g.EnrolledStudents()+g.Teachers() {
			for j := 1; j <= courses; j++ {
				enrollment := &qf.Enrollment{
					CourseID: uint64(j),
					UserID:   user.GetID(),
				}
				if err := g.db.CreateEnrollment(enrollment); err != nil {
					return err
				}
				enrollment.Status = qf.Enrollment_STUDENT
				if i < g.Teachers() {
					enrollment.Status = qf.Enrollment_TEACHER
				}
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
