package mockdata

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/quickfeed/quickfeed/qf"
)

func (g *generator) users() error {
	for i := 1; i < g.Students; i++ {
		name := "User " + fmt.Sprintf("%02d", i)
		if i < len(Members) {
			name = Members[i]
		}
		user := &qf.User{
			Name:        name,
			Login:       name,
			Email:       strings.Replace(name, " ", ".", 1) + "@example.com",
			StudentID:   fmt.Sprintf("%06d", i),
			AvatarURL:   fmt.Sprintf("https://i.pravatar.cc/150?img=%d", rand.Intn(70)+1),
			ScmRemoteID: uint64(i),
			// TODO(Joachim): Should RefreshToken be set?
		}
		g.log("Creating user: %q\n", user.Name)
		if err := g.db.CreateUser(user); err != nil {
			return err
		}
		/* Bypassing the Teacher/Admin acceptance of enrollments */

		// Enroll x students and all teachers in all courses
		if i < g.EnrolledStudents+g.Teachers {
			for j := range g.Courses {
				enrollment := &qf.Enrollment{
					CourseID: uint64(j + 1),
					UserID:   user.GetID(),
				}
				if err := g.db.CreateEnrollment(enrollment); err != nil {
					return err
				}
				enrollment.Status = qf.Enrollment_STUDENT
				if i < g.Teachers {
					enrollment.Status = qf.Enrollment_TEACHER
				}
				if err := g.db.UpdateEnrollment(enrollment); err != nil {
					return err
				}
			}
		} else {
			g.log("User %q not enrolled in any courses\n", user.Name)
		}
	}
	return nil
}
