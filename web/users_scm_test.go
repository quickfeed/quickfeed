package web_test

import (
	"context"
	"testing"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
)

// TestUpdateEnrollmentsAfterUpdateUserLogin verifies that when a user's login was updated
// at the SCM prior to enrollment approval, the updated login is reflected in the database.
func TestUpdateEnrollmentsAfterUpdateUserLogin(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "admin"})
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, admin, course)

	// Student with old login
	oldLogin := "old-student-login"
	newLogin := "new-student-login"
	scmID := uint64(123)
	student := qtest.CreateFakeCustomUser(t, db, &qf.User{
		Login:       oldLogin,
		ScmRemoteID: scmID,
	})

	// Enroll student as pending
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID:   student.GetID(),
		CourseID: course.GetID(),
	}); err != nil {
		t.Fatal(err)
	}

	// Setup mock SCM with new login
	scmOpt := scm.WithMockOptions(scm.WithMockCourses(), scm.WithMockOrgs("admin"))
	ghOrg := github.Organization{Login: github.String(course.GetScmOrganizationName())}
	memberOpt := scm.WithMembers(github.Membership{
		Organization: &ghOrg,
		User: &github.User{
			ID:    github.Int64(int64(scmID)),
			Login: github.String(newLogin),
		},
	})
	client := web.NewMockClient(t, db, scm.WithMockOptions(scmOpt, memberOpt), web.WithInterceptors())
	ctx := context.Background()
	adminCookie := client.Cookie(t, admin)

	// Admin approves the enrollment
	enrollment, err := db.GetEnrollmentByCourseAndUser(course.GetID(), student.GetID())
	if err != nil {
		t.Fatal(err)
	}
	enrollment.Status = qf.Enrollment_STUDENT

	req := &qf.Enrollments{
		Enrollments: []*qf.Enrollment{enrollment},
	}
	_, err = client.UpdateEnrollments(ctx, qtest.RequestWithCookie(req, adminCookie))
	if err != nil {
		t.Fatal(err)
	}

	// Verify student login is updated in DB
	updatedStudent, err := db.GetUser(student.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if updatedStudent.GetLogin() != newLogin {
		t.Errorf("expected login %q, got %q", newLogin, updatedStudent.GetLogin())
	}
}

// TestUpdateGroupAfterUpdateUserLogin verifies that when a user's login was updated
// at the SCM prior to group approval, the updated login is reflected in the database.
func TestUpdateGroupAfterUpdateUserLogin(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "admin"})
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, admin, course)

	// Student with old login
	oldLogin := "old-student-login"
	newLogin := "new-student-login"
	scmID := uint64(123)
	student := qtest.CreateFakeCustomUser(t, db, &qf.User{
		Login:       oldLogin,
		ScmRemoteID: scmID,
	})

	// Enroll and approve student
	qtest.EnrollStudent(t, db, student, course)
	e, _ := db.GetEnrollmentByCourseAndUser(course.GetID(), student.GetID())
	e.Status = qf.Enrollment_STUDENT
	db.UpdateEnrollment(e)

	// Create a group
	group := &qf.Group{
		CourseID: course.GetID(),
		Name:     "TestGroup",
		Users:    []*qf.User{student},
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}

	// Setup mock SCM with new login
	scmOpt := scm.WithMockOptions(scm.WithMockCourses(), scm.WithMockOrgs("admin"))
	ghOrg := github.Organization{Login: github.String(course.GetScmOrganizationName())}
	memberOpt := scm.WithMembers(github.Membership{
		Organization: &ghOrg,
		User: &github.User{
			ID:    github.Int64(int64(scmID)),
			Login: github.String(newLogin),
		},
	})
	client := web.NewMockClient(t, db, scm.WithMockOptions(scmOpt, memberOpt), web.WithInterceptors())
	ctx := context.Background()
	adminCookie := client.Cookie(t, admin)

	// Admin updates the group (e.g., approving it)
	group.Status = qf.Group_APPROVED
	_, err := client.UpdateGroup(ctx, qtest.RequestWithCookie(group, adminCookie))
	if err != nil {
		t.Fatal(err)
	}

	// Verify student login is updated in DB
	updatedStudent, err := db.GetUser(student.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if updatedStudent.GetLogin() != newLogin {
		t.Errorf("expected login %q, got %q", newLogin, updatedStudent.GetLogin())
	}
}
