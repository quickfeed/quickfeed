package interceptor_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
)

type accessTest struct {
	cookie     string
	userID     uint64
	courseID   uint64
	groupID    uint64
	wantAccess bool
	wantCode   connect.Code
}

func TestAccessControl(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs(),
		web.WithInterceptors(
			web.UserInterceptorFunc,
			web.AccessControlInterceptorFunc,
		),
	)
	ctx := context.Background()

	courseAdmin := qtest.CreateFakeUser(t, db)
	groupStudent := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "Group Student", Login: "group student"})
	student := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "Test Student", Login: "student"})
	user := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "Test User", Login: "user"})
	admin := qtest.CreateFakeUser(t, db)
	admin.IsAdmin = true
	if err := db.UpdateUser(admin); err != nil {
		t.Fatal(err)
	}

	course := &qf.Course{
		Code:            "test101",
		Year:            2022,
		CourseCreatorID: courseAdmin.GetID(),
	}
	qtest.CreateCourse(t, db, courseAdmin, course)
	qtest.EnrollStudent(t, db, groupStudent, course)
	qtest.EnrollStudent(t, db, student, course)
	group := &qf.Group{
		CourseID: course.GetID(),
		Name:     "Test",
		Users:    []*qf.User{groupStudent},
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}

	assignment := &qf.Assignment{
		CourseID: course.GetID(),
		Name:     "Test Assignment",
		Order:    1,
	}
	if err := db.CreateAssignment(assignment); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.GetID(),
		UserID:       groupStudent.GetID(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateSubmission(&qf.Submission{
		AssignmentID: assignment.GetID(),
		GroupID:      group.GetID(),
	}); err != nil {
		t.Fatal(err)
	}

	courseAdminCookie := client.Cookie(t, courseAdmin)
	groupStudentCookie := client.Cookie(t, groupStudent)
	studentCookie := client.Cookie(t, student)
	userCookie := client.Cookie(t, user)
	adminCookie := client.Cookie(t, admin)

	freeAccessTest := map[string]accessTest{
		"admin":             {cookie: courseAdminCookie, courseID: course.GetID(), wantAccess: true},
		"student":           {cookie: studentCookie, courseID: course.GetID(), wantAccess: true},
		"group student":     {cookie: groupStudentCookie, courseID: course.GetID(), wantAccess: true},
		"user":              {cookie: userCookie, courseID: course.GetID(), wantAccess: true},
		"non-teacher admin": {cookie: adminCookie, courseID: course.GetID(), wantAccess: true},
		"empty context":     {wantAccess: false, wantCode: connect.CodeUnauthenticated},
	}
	for name, tt := range freeAccessTest {
		t.Run("UnrestrictedAccess/"+name, func(t *testing.T) {
			_, err := client.GetUser(ctx, qtest.RequestWithCookie(&qf.Void{}, tt.cookie))
			checkAccess(t, "GetUser", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetCourse(ctx, qtest.RequestWithCookie(&qf.CourseRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "GetCourse", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetCourses(ctx, qtest.RequestWithCookie(&qf.Void{}, tt.cookie))
			checkAccess(t, "GetCourses", err, tt.wantCode, tt.wantAccess)
		})
	}

	userAccessTests := map[string]accessTest{
		"correct user ID":   {cookie: userCookie, userID: user.GetID(), courseID: course.GetID(), wantAccess: true},
		"incorrect user ID": {cookie: userCookie, groupID: groupStudent.GetID(), courseID: course.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range userAccessTests {
		t.Run("UserAccess/"+name, func(t *testing.T) {
			enrol := &qf.Enrollment{
				CourseID: tt.courseID,
				UserID:   tt.userID,
			}
			enrolRequest := &qf.EnrollmentRequest{
				FetchMode: &qf.EnrollmentRequest_UserID{
					UserID: tt.userID,
				},
			}
			_, err := client.CreateEnrollment(ctx, qtest.RequestWithCookie(enrol, tt.cookie))
			checkAccess(t, "CreateEnrollment", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateCourseVisibility(ctx, qtest.RequestWithCookie(enrol, tt.cookie))
			checkAccess(t, "UpdateCourseVisibility", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateUser(ctx, qtest.RequestWithCookie(&qf.User{ID: tt.userID}, tt.cookie))
			checkAccess(t, "UpdateUser", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetEnrollments(ctx, qtest.RequestWithCookie(enrolRequest, tt.cookie))
			checkAccess(t, "GetEnrollments", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateUser(ctx, qtest.RequestWithCookie(&qf.User{ID: tt.userID}, tt.cookie))
			checkAccess(t, "UpdateUser", err, tt.wantCode, tt.wantAccess)
		})
	}

	studentAccessTests := map[string]accessTest{
		"course admin":                     {cookie: courseAdminCookie, userID: courseAdmin.GetID(), courseID: course.GetID(), wantAccess: true},
		"admin, not enrolled in a course":  {cookie: adminCookie, userID: admin.GetID(), courseID: course.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
		"user, not enrolled in the course": {cookie: userCookie, userID: user.GetID(), courseID: course.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
		"student":                          {cookie: studentCookie, userID: student.GetID(), courseID: course.GetID(), wantAccess: true},
		"student of another course":        {cookie: studentCookie, userID: student.GetID(), courseID: 123, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range studentAccessTests {
		t.Run("StudentAccess/"+name, func(t *testing.T) {
			_, err := client.GetSubmissions(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
				CourseID: tt.courseID,
				FetchMode: &qf.SubmissionRequest_UserID{
					UserID: tt.userID,
				},
			}, tt.cookie))
			checkAccess(t, "GetSubmissions", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetAssignments(ctx, qtest.RequestWithCookie(&qf.CourseRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "GetAssignments", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetRepositories(ctx, qtest.RequestWithCookie(&qf.CourseRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "GetRepositories", err, tt.wantCode, tt.wantAccess)
		})
	}

	// Test GetSubmissions with user ID mismatch (student accessing another student's submissions)
	t.Run("GetSubmissions/UserIDMismatch", func(t *testing.T) {
		_, err := client.GetSubmissions(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
			CourseID: course.GetID(),
			FetchMode: &qf.SubmissionRequest_UserID{
				UserID: groupStudent.GetID(), // student trying to access groupStudent's submissions
			},
		}, studentCookie))
		if err == nil {
			t.Error("Expected access denied for student accessing another student's submissions")
		} else {
			var connErr *connect.Error
			if !errors.As(err, &connErr) || connErr.Code() != connect.CodePermissionDenied {
				t.Errorf("Expected CodePermissionDenied, got %v", err)
			}
		}
	})

	submissionsGroupAccessTests := map[string]accessTest{
		"group member":         {cookie: groupStudentCookie, courseID: course.GetID(), groupID: group.GetID(), wantAccess: true},
		"student not in group": {cookie: studentCookie, courseID: course.GetID(), groupID: group.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
		"teacher":              {cookie: courseAdminCookie, courseID: course.GetID(), groupID: group.GetID(), wantAccess: true},
		"admin not enrolled":   {cookie: adminCookie, courseID: course.GetID(), groupID: group.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range submissionsGroupAccessTests {
		t.Run("SubmissionsGroupAccess/"+name, func(t *testing.T) {
			_, err := client.GetSubmissions(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
				CourseID: tt.courseID,
				FetchMode: &qf.SubmissionRequest_GroupID{
					GroupID: tt.groupID,
				},
			}, tt.cookie))
			checkAccess(t, "GetSubmissions", err, tt.wantCode, tt.wantAccess)
		})
	}

	groupAccessTests := map[string]accessTest{
		"student in a group":                            {cookie: groupStudentCookie, userID: groupStudent.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: true},
		"student, not in a group":                       {cookie: studentCookie, userID: student.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
		"student in a group, wrong group ID in request": {cookie: studentCookie, userID: student.GetID(), courseID: course.GetID(), groupID: 123, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range groupAccessTests {
		t.Run("GroupAccess/"+name, func(t *testing.T) {
			_, err := client.GetGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{
				CourseID: tt.courseID,
				GroupID:  tt.groupID,
			}, tt.cookie))
			checkAccess(t, "GetGroup", err, tt.wantCode, tt.wantAccess)
		})
	}

	teacherAccessTests := map[string]accessTest{
		"course teacher":                    {cookie: courseAdminCookie, userID: groupStudent.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: true},
		"student":                           {cookie: studentCookie, userID: student.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
		"admin, not enrolled in the course": {cookie: adminCookie, userID: admin.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range teacherAccessTests {
		t.Run("TeacherAccess/"+name, func(t *testing.T) {
			_, err := client.GetGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{
				CourseID: tt.courseID,
				GroupID:  tt.groupID,
			}, tt.cookie))
			checkAccess(t, "GetGroup", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{
				CourseID: tt.courseID,
				UserID:   tt.userID,
			}, tt.cookie))
			checkAccess(t, "GetGroup", err, tt.wantCode, tt.wantAccess)
			_, err = client.DeleteGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{
				GroupID:  tt.groupID,
				CourseID: tt.courseID,
				UserID:   tt.userID,
			}, tt.cookie))
			checkAccess(t, "DeleteGroup", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateGroup(ctx, qtest.RequestWithCookie(&qf.Group{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "UpdateGroup", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateCourse(ctx, qtest.RequestWithCookie(course, tt.cookie))
			checkAccess(t, "UpdateCourse", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateEnrollments(ctx, qtest.RequestWithCookie(&qf.Enrollments{
				Enrollments: []*qf.Enrollment{{ID: 1, CourseID: tt.courseID}},
			}, tt.cookie))
			checkAccess(t, "UpdateEnrollments", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateAssignments(ctx, qtest.RequestWithCookie(&qf.CourseRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "UpdateAssignments", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateSubmission(ctx, qtest.RequestWithCookie(&qf.Grade{SubmissionID: 1, UserID: tt.userID}, tt.cookie))
			checkAccess(t, "UpdateSubmission", err, tt.wantCode, tt.wantAccess)
			_, err = client.RebuildSubmissions(ctx, qtest.RequestWithCookie(&qf.RebuildRequest{
				AssignmentID: 1,
				CourseID:     tt.courseID,
			}, tt.cookie))
			checkAccess(t, "RebuildSubmissions", err, tt.wantCode, tt.wantAccess)
			_, err = client.CreateReview(ctx, qtest.RequestWithCookie(&qf.ReviewRequest{
				CourseID: tt.courseID,
				Review: &qf.Review{
					SubmissionID: 1,
					ReviewerID:   1,
				},
			}, tt.cookie))
			checkAccess(t, "CreateReview", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateReview(ctx, qtest.RequestWithCookie(&qf.ReviewRequest{
				CourseID: tt.courseID,
				Review: &qf.Review{
					SubmissionID: 1,
					ReviewerID:   1,
				},
			}, tt.cookie))
			checkAccess(t, "UpdateReview", err, tt.wantCode, tt.wantAccess)
			_, err = client.IsEmptyRepo(ctx, qtest.RequestWithCookie(&qf.RepositoryRequest{CourseID: tt.courseID}, tt.cookie))
			checkAccess(t, "IsEmptyRepo", err, tt.wantCode, tt.wantAccess)
		})
	}

	courseAdminTests := map[string]accessTest{
		"admin, not enrolled": {cookie: adminCookie, courseID: course.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range courseAdminTests {
		t.Run("CourseAdminAccess/"+name, func(t *testing.T) {
			_, err := client.GetSubmissionsByCourse(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
				CourseID: tt.courseID,
			}, tt.cookie))
			checkAccess(t, "GetSubmissionsByCourse", err, tt.wantCode, tt.wantAccess)
		})
	}

	adminAccessTests := map[string]accessTest{
		"admin (accessing own info)":              {cookie: courseAdminCookie, userID: courseAdmin.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: true},
		"admin (accessing other user's info)":     {cookie: courseAdminCookie, userID: user.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: true},
		"non admin (accessing admin's info)":      {cookie: studentCookie, userID: courseAdmin.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
		"non admin (accessing other user's info)": {cookie: studentCookie, userID: user.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range adminAccessTests {
		t.Run("AdminAccess/"+name, func(t *testing.T) {
			_, err := client.UpdateUser(ctx, qtest.RequestWithCookie(&qf.User{ID: tt.userID}, tt.cookie))
			checkAccess(t, "UpdateUser", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetUsers(ctx, qtest.RequestWithCookie(&qf.Void{}, tt.cookie))
			checkAccess(t, "GetUsers", err, tt.wantCode, tt.wantAccess)
		})
	}

	createGroupTests := map[string]struct {
		cookie     string
		group      *qf.Group
		wantAccess bool
		wantCode   connect.Code
	}{
		"valid student, not in the request group": {cookie: studentCookie, group: &qf.Group{
			CourseID: course.GetID(),
		}, wantAccess: false, wantCode: connect.CodePermissionDenied},
		"valid student": {cookie: studentCookie, group: &qf.Group{
			Name:     "test",
			CourseID: course.GetID(),
			Users:    []*qf.User{student},
		}, wantAccess: true},
		"course teacher": {cookie: courseAdminCookie, group: &qf.Group{
			CourseID: course.GetID(),
			Users:    []*qf.User{courseAdmin},
		}, wantAccess: true},
		"admin, not a teacher": {cookie: adminCookie, group: &qf.Group{
			CourseID: course.GetID(),
		}, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}

	for name, tt := range createGroupTests {
		t.Run("CreateGroupAccess/"+name, func(t *testing.T) {
			_, err := client.CreateGroup(ctx, qtest.RequestWithCookie(tt.group, tt.cookie))
			checkAccess(t, "CreateGroup", err, tt.wantCode, tt.wantAccess)
		})
	}

	adminStatusChangeTests := map[string]struct {
		cookie     string
		user       *qf.User
		wantAccess bool
		wantCode   connect.Code
	}{
		"admin demoting a user": {cookie: courseAdminCookie, user: &qf.User{
			ID:      admin.GetID(),
			IsAdmin: false,
		}, wantAccess: true},
		"admin promoting a user": {cookie: courseAdminCookie, user: &qf.User{
			ID:      admin.GetID(),
			IsAdmin: true,
		}, wantAccess: true},
		"admin demoting self": {cookie: courseAdminCookie, user: &qf.User{
			ID:      courseAdmin.GetID(),
			IsAdmin: false,
		}, wantAccess: true},
		"user promoting another user": {cookie: userCookie, user: &qf.User{
			ID:      groupStudent.GetID(),
			IsAdmin: true,
		}, wantAccess: false, wantCode: connect.CodePermissionDenied},
		"user promoting self": {cookie: userCookie, user: &qf.User{
			ID:      user.GetID(),
			IsAdmin: true,
		}, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}

	for name, tt := range adminStatusChangeTests {
		t.Run("AdminStatusChange/"+name, func(t *testing.T) {
			_, err := client.UpdateUser(ctx, qtest.RequestWithCookie(tt.user, tt.cookie))
			checkAccess(t, "UpdateUser", err, tt.wantCode, tt.wantAccess)
		})
	}

	adminGetEnrollmentsTests := map[string]accessTest{
		"admin, not enrolled in the course": {cookie: adminCookie, courseID: course.GetID(), userID: student.GetID(), wantAccess: true},
	}

	for name, tt := range adminGetEnrollmentsTests {
		t.Run("AdminGetEnrollments/"+name, func(t *testing.T) {
			_, err := client.GetEnrollments(ctx, qtest.RequestWithCookie(&qf.EnrollmentRequest{
				FetchMode: &qf.EnrollmentRequest_CourseID{
					CourseID: tt.courseID,
				},
			}, tt.cookie))
			checkAccess(t, "GetEnrollments", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetEnrollments(ctx, qtest.RequestWithCookie(&qf.EnrollmentRequest{
				FetchMode: &qf.EnrollmentRequest_UserID{
					UserID: tt.userID,
				},
			}, tt.cookie))
			checkAccess(t, "GetEnrollments", err, tt.wantCode, tt.wantAccess)
		})
	}

	// Test student role in GetEnrollments (when student queries course enrollments, not their own user)
	studentGetEnrollmentsTests := map[string]accessTest{
		"student querying course enrollments": {cookie: studentCookie, courseID: course.GetID(), wantAccess: true},
		"student of another course":           {cookie: studentCookie, courseID: 999, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}

	for name, tt := range studentGetEnrollmentsTests {
		t.Run("StudentGetEnrollments/"+name, func(t *testing.T) {
			_, err := client.GetEnrollments(ctx, qtest.RequestWithCookie(&qf.EnrollmentRequest{
				FetchMode: &qf.EnrollmentRequest_CourseID{
					CourseID: tt.courseID,
				},
			}, tt.cookie))
			checkAccess(t, "GetEnrollments", err, tt.wantCode, tt.wantAccess)
		})
	}
}

func TestCrossCourseSubmissionUpdate(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	// Create admin user for course creation (first user is always admin)
	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "Admin", Login: "admin"})

	// Create two courses with admin as creator
	course1 := &qf.Course{Code: "course1", CourseCreatorID: admin.GetID()}
	course2 := &qf.Course{Code: "course2", CourseCreatorID: admin.GetID()}
	qtest.CreateCourse(t, db, admin, course1)
	qtest.CreateCourse(t, db, admin, course2)

	// Create user A (regular non-admin user)
	userA := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "User A", Login: "userA"})
	// Enroll user A as student in course1
	qtest.EnrollStudent(t, db, userA, course1)
	// Enroll user A as teacher in course2
	qtest.EnrollTeacher(t, db, userA, course2)

	assignment1 := &qf.Assignment{CourseID: course1.GetID(), Name: "Assignment 1", Order: 1}
	qtest.CreateAssignment(t, db, assignment1)

	// Create submission for user A in course1
	submission := &qf.Submission{AssignmentID: assignment1.GetID(), UserID: userA.GetID()}
	qtest.CreateSubmission(t, db, submission)

	client := web.NewMockClient(t, db, scm.WithMockOrgs(),
		web.WithInterceptors(
			web.UserInterceptorFunc,
			web.AccessControlInterceptorFunc,
		),
	)

	// Attempt to update the submission from course1 while acting as teacher in course2
	// This should fail because the submission belongs to course1, not course2.
	_, err := client.UpdateSubmission(t.Context(), qtest.RequestWithCookie(&qf.Grade{
		SubmissionID: submission.GetID(),
		// TODO(meling): UpdateSubmissionRequest had these fields, but Grade does not. Should we add CourseID to Grade?
		// CourseID:     course2.GetID(), // Wrong course ID
		// Score:        100,
		// Released:     true,
	}, client.Cookie(t, userA)))

	if err == nil {
		t.Error("Expected access denied for cross-course submission update, but access was granted")
	} else {
		var connErr *connect.Error
		if !errors.As(err, &connErr) || connErr.Code() != connect.CodePermissionDenied {
			t.Errorf("Expected CodePermissionDenied, got %v", err)
		}
	}
}

func TestAccessControlWithoutClaims(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	// Create client WITHOUT UserInterceptor to simulate missing claims in context
	client := web.NewMockClient(t, db, scm.WithMockOrgs(),
		web.WithInterceptors(
			web.AccessControlInterceptorFunc,
		),
	)

	user := qtest.CreateFakeUser(t, db)
	course := &qf.Course{
		Code:            "test101",
		Year:            2022,
		CourseCreatorID: user.GetID(),
	}
	qtest.CreateCourse(t, db, user, course)

	// Attempt to call a method without claims in context
	// This should fail with a permission denied error
	_, err := client.GetUser(t.Context(), qtest.RequestWithCookie(&qf.Void{}, client.Cookie(t, user)))
	if err == nil {
		t.Error("Expected access denied when claims are missing from context, but access was granted")
	} else {
		var connErr *connect.Error
		if !errors.As(err, &connErr) || connErr.Code() != connect.CodePermissionDenied {
			t.Errorf("Expected CodePermissionDenied, got %v", err)
		}
		if !errors.As(err, &connErr) || !strings.Contains(connErr.Message(), "failed to get claims from request context") {
			t.Errorf("Expected error message about missing claims, got: %v", err)
		}
	}
}

func checkAccess(t *testing.T, method string, err error, wantCode connect.Code, wantAccess bool) {
	t.Helper()
	if wantAccess {
		// Expect access granted: either no error, or error not due to permission denied
		if err != nil {
			var connErr *connect.Error
			if errors.As(err, &connErr) && connErr.Code() == connect.CodePermissionDenied {
				t.Errorf("%23s: access denied: %v", method, err)
			}
			// Other errors are ok, as they mean the method ran but failed for business logic reasons
		}
	} else {
		// Expect access denied with specific code
		if err == nil {
			t.Errorf("%23s: got access granted, want access denied", method)
		} else {
			var connErr *connect.Error
			if !errors.As(err, &connErr) || connErr.Code() != wantCode {
				t.Errorf("%23s: got error %v, want permission denied with code %v", method, err, wantCode)
			}
		}
	}
}
