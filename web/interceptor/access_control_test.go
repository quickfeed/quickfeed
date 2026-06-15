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
	ctx        context.Context
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

	courseAdminCtx := client.Context(t, courseAdmin)
	groupStudentCtx := client.Context(t, groupStudent)
	studentCtx := client.Context(t, student)
	userCtx := client.Context(t, user)
	adminCtx := client.Context(t, admin)

	freeAccessTest := map[string]accessTest{
		"admin":             {ctx: courseAdminCtx, courseID: course.GetID(), wantAccess: true},
		"student":           {ctx: studentCtx, courseID: course.GetID(), wantAccess: true},
		"group student":     {ctx: groupStudentCtx, courseID: course.GetID(), wantAccess: true},
		"user":              {ctx: userCtx, courseID: course.GetID(), wantAccess: true},
		"non-teacher admin": {ctx: adminCtx, courseID: course.GetID(), wantAccess: true},
		"empty context":     {ctx: t.Context(), wantAccess: false, wantCode: connect.CodeUnauthenticated},
	}
	for name, tt := range freeAccessTest {
		t.Run("UnrestrictedAccess/"+name, func(t *testing.T) {
			_, err := client.GetUser(tt.ctx, &qf.Void{})
			checkAccess(t, "GetUser", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetCourse(tt.ctx, &qf.CourseRequest{CourseID: tt.courseID})
			checkAccess(t, "GetCourse", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetCourses(tt.ctx, &qf.Void{})
			checkAccess(t, "GetCourses", err, tt.wantCode, tt.wantAccess)
		})
	}

	userAccessTests := map[string]accessTest{
		"correct user ID":   {ctx: userCtx, userID: user.GetID(), courseID: course.GetID(), wantAccess: true},
		"incorrect user ID": {ctx: userCtx, groupID: groupStudent.GetID(), courseID: course.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
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
			_, err := client.CreateEnrollment(tt.ctx, enrol)
			checkAccess(t, "CreateEnrollment", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateCourseVisibility(tt.ctx, enrol)
			checkAccess(t, "UpdateCourseVisibility", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateUser(tt.ctx, &qf.User{ID: tt.userID})
			checkAccess(t, "UpdateUser", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetEnrollments(tt.ctx, enrolRequest)
			checkAccess(t, "GetEnrollments", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateUser(tt.ctx, &qf.User{ID: tt.userID})
			checkAccess(t, "UpdateUser", err, tt.wantCode, tt.wantAccess)
		})
	}

	studentAccessTests := map[string]accessTest{
		"course admin":                     {ctx: courseAdminCtx, userID: courseAdmin.GetID(), courseID: course.GetID(), wantAccess: true},
		"admin, not enrolled in a course":  {ctx: adminCtx, userID: admin.GetID(), courseID: course.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
		"user, not enrolled in the course": {ctx: userCtx, userID: user.GetID(), courseID: course.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
		"student":                          {ctx: studentCtx, userID: student.GetID(), courseID: course.GetID(), wantAccess: true},
		"student of another course":        {ctx: studentCtx, userID: student.GetID(), courseID: 123, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range studentAccessTests {
		t.Run("StudentAccess/"+name, func(t *testing.T) {
			_, err := client.GetSubmissions(tt.ctx, &qf.SubmissionRequest{
				CourseID: tt.courseID,
				FetchMode: &qf.SubmissionRequest_UserID{
					UserID: tt.userID,
				},
			})
			checkAccess(t, "GetSubmissions", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetAssignments(tt.ctx, &qf.CourseRequest{CourseID: tt.courseID})
			checkAccess(t, "GetAssignments", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetRepositories(tt.ctx, &qf.CourseRequest{CourseID: tt.courseID})
			checkAccess(t, "GetRepositories", err, tt.wantCode, tt.wantAccess)
		})
	}

	// Test GetSubmissions with user ID mismatch (student accessing another student's submissions)
	t.Run("GetSubmissions/UserIDMismatch", func(t *testing.T) {
		_, err := client.GetSubmissions(studentCtx, &qf.SubmissionRequest{
			CourseID: course.GetID(),
			FetchMode: &qf.SubmissionRequest_UserID{
				UserID: groupStudent.GetID(), // student trying to access groupStudent's submissions
			},
		})
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
		"group member":         {ctx: groupStudentCtx, courseID: course.GetID(), groupID: group.GetID(), wantAccess: true},
		"student not in group": {ctx: studentCtx, courseID: course.GetID(), groupID: group.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
		"teacher":              {ctx: courseAdminCtx, courseID: course.GetID(), groupID: group.GetID(), wantAccess: true},
		"admin not enrolled":   {ctx: adminCtx, courseID: course.GetID(), groupID: group.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range submissionsGroupAccessTests {
		t.Run("SubmissionsGroupAccess/"+name, func(t *testing.T) {
			_, err := client.GetSubmissions(tt.ctx, &qf.SubmissionRequest{
				CourseID: tt.courseID,
				FetchMode: &qf.SubmissionRequest_GroupID{
					GroupID: tt.groupID,
				},
			})
			checkAccess(t, "GetSubmissions", err, tt.wantCode, tt.wantAccess)
		})
	}

	groupAccessTests := map[string]accessTest{
		"student in a group":                            {ctx: groupStudentCtx, userID: groupStudent.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: true},
		"student, not in a group":                       {ctx: studentCtx, userID: student.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
		"student in a group, wrong group ID in request": {ctx: studentCtx, userID: student.GetID(), courseID: course.GetID(), groupID: 123, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range groupAccessTests {
		t.Run("GroupAccess/"+name, func(t *testing.T) {
			_, err := client.GetGroup(tt.ctx, &qf.GroupRequest{
				CourseID: tt.courseID,
				GroupID:  tt.groupID,
			})
			checkAccess(t, "GetGroup", err, tt.wantCode, tt.wantAccess)
		})
	}

	teacherAccessTests := map[string]accessTest{
		"course teacher":                    {ctx: courseAdminCtx, userID: groupStudent.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: true},
		"student":                           {ctx: studentCtx, userID: student.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
		"admin, not enrolled in the course": {ctx: adminCtx, userID: admin.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range teacherAccessTests {
		t.Run("TeacherAccess/"+name, func(t *testing.T) {
			_, err := client.GetGroup(tt.ctx, &qf.GroupRequest{
				CourseID: tt.courseID,
				GroupID:  tt.groupID,
			})
			checkAccess(t, "GetGroup", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetGroup(tt.ctx, &qf.GroupRequest{
				CourseID: tt.courseID,
				UserID:   tt.userID,
			})
			checkAccess(t, "GetGroup", err, tt.wantCode, tt.wantAccess)
			_, err = client.DeleteGroup(tt.ctx, &qf.GroupRequest{
				GroupID:  tt.groupID,
				CourseID: tt.courseID,
				UserID:   tt.userID,
			})
			checkAccess(t, "DeleteGroup", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateGroup(tt.ctx, &qf.Group{CourseID: tt.courseID})
			checkAccess(t, "UpdateGroup", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateCourse(tt.ctx, course)
			checkAccess(t, "UpdateCourse", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateEnrollments(tt.ctx, &qf.Enrollments{
				Enrollments: []*qf.Enrollment{{ID: 1, CourseID: tt.courseID}},
			})
			checkAccess(t, "UpdateEnrollments", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateAssignments(tt.ctx, &qf.CourseRequest{CourseID: tt.courseID})
			checkAccess(t, "UpdateAssignments", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateSubmission(tt.ctx, &qf.Grade{SubmissionID: 1, UserID: tt.userID})
			checkAccess(t, "UpdateSubmission", err, tt.wantCode, tt.wantAccess)
			_, err = client.RebuildSubmissions(tt.ctx, &qf.RebuildRequest{
				AssignmentID: 1,
				CourseID:     tt.courseID,
			})
			checkAccess(t, "RebuildSubmissions", err, tt.wantCode, tt.wantAccess)
			_, err = client.CreateReview(tt.ctx, &qf.ReviewRequest{
				CourseID: tt.courseID,
				Review: &qf.Review{
					SubmissionID: 1,
					ReviewerID:   1,
				},
			})
			checkAccess(t, "CreateReview", err, tt.wantCode, tt.wantAccess)
			_, err = client.UpdateReview(tt.ctx, &qf.ReviewRequest{
				CourseID: tt.courseID,
				Review: &qf.Review{
					SubmissionID: 1,
					ReviewerID:   1,
				},
			})
			checkAccess(t, "UpdateReview", err, tt.wantCode, tt.wantAccess)
			_, err = client.IsEmptyRepo(tt.ctx, &qf.RepositoryRequest{CourseID: tt.courseID})
			checkAccess(t, "IsEmptyRepo", err, tt.wantCode, tt.wantAccess)
		})
	}

	courseAdminTests := map[string]accessTest{
		"admin, not enrolled": {ctx: adminCtx, courseID: course.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range courseAdminTests {
		t.Run("CourseAdminAccess/"+name, func(t *testing.T) {
			_, err := client.GetSubmissionsByCourse(tt.ctx, &qf.SubmissionRequest{
				CourseID: tt.courseID,
			})
			checkAccess(t, "GetSubmissionsByCourse", err, tt.wantCode, tt.wantAccess)
		})
	}

	adminAccessTests := map[string]accessTest{
		"admin (accessing own info)":              {ctx: courseAdminCtx, userID: courseAdmin.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: true},
		"admin (accessing other user's info)":     {ctx: courseAdminCtx, userID: user.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: true},
		"non admin (accessing admin's info)":      {ctx: studentCtx, userID: courseAdmin.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
		"non admin (accessing other user's info)": {ctx: studentCtx, userID: user.GetID(), courseID: course.GetID(), groupID: group.GetID(), wantAccess: false, wantCode: connect.CodePermissionDenied},
	}
	for name, tt := range adminAccessTests {
		t.Run("AdminAccess/"+name, func(t *testing.T) {
			_, err := client.UpdateUser(tt.ctx, &qf.User{ID: tt.userID})
			checkAccess(t, "UpdateUser", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetUsers(tt.ctx, &qf.Void{})
			checkAccess(t, "GetUsers", err, tt.wantCode, tt.wantAccess)
		})
	}

	createGroupTests := map[string]struct {
		ctx        context.Context
		group      *qf.Group
		wantAccess bool
		wantCode   connect.Code
	}{
		"valid student, not in the request group": {ctx: studentCtx, group: &qf.Group{
			CourseID: course.GetID(),
		}, wantAccess: false, wantCode: connect.CodePermissionDenied},
		"valid student": {ctx: studentCtx, group: &qf.Group{
			Name:     "test",
			CourseID: course.GetID(),
			Users:    []*qf.User{student},
		}, wantAccess: true},
		"course teacher": {ctx: courseAdminCtx, group: &qf.Group{
			CourseID: course.GetID(),
			Users:    []*qf.User{courseAdmin},
		}, wantAccess: true},
		"admin, not a teacher": {ctx: adminCtx, group: &qf.Group{
			CourseID: course.GetID(),
		}, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}

	for name, tt := range createGroupTests {
		t.Run("CreateGroupAccess/"+name, func(t *testing.T) {
			_, err := client.CreateGroup(tt.ctx, tt.group)
			checkAccess(t, "CreateGroup", err, tt.wantCode, tt.wantAccess)
		})
	}

	adminStatusChangeTests := map[string]struct {
		ctx        context.Context
		user       *qf.User
		wantAccess bool
		wantCode   connect.Code
	}{
		"admin demoting a user": {ctx: courseAdminCtx, user: &qf.User{
			ID:      admin.GetID(),
			IsAdmin: false,
		}, wantAccess: true},
		"admin promoting a user": {ctx: courseAdminCtx, user: &qf.User{
			ID:      admin.GetID(),
			IsAdmin: true,
		}, wantAccess: true},
		"admin demoting self": {ctx: courseAdminCtx, user: &qf.User{
			ID:      courseAdmin.GetID(),
			IsAdmin: false,
		}, wantAccess: true},
		"user promoting another user": {ctx: userCtx, user: &qf.User{
			ID:      groupStudent.GetID(),
			IsAdmin: true,
		}, wantAccess: false, wantCode: connect.CodePermissionDenied},
		"user promoting self": {ctx: userCtx, user: &qf.User{
			ID:      user.GetID(),
			IsAdmin: true,
		}, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}

	for name, tt := range adminStatusChangeTests {
		t.Run("AdminStatusChange/"+name, func(t *testing.T) {
			_, err := client.UpdateUser(tt.ctx, tt.user)
			checkAccess(t, "UpdateUser", err, tt.wantCode, tt.wantAccess)
		})
	}

	adminGetEnrollmentsTests := map[string]accessTest{
		"admin, not enrolled in the course": {ctx: adminCtx, courseID: course.GetID(), userID: student.GetID(), wantAccess: true},
	}

	for name, tt := range adminGetEnrollmentsTests {
		t.Run("AdminGetEnrollments/"+name, func(t *testing.T) {
			_, err := client.GetEnrollments(tt.ctx, &qf.EnrollmentRequest{
				FetchMode: &qf.EnrollmentRequest_CourseID{
					CourseID: tt.courseID,
				},
			})
			checkAccess(t, "GetEnrollments", err, tt.wantCode, tt.wantAccess)
			_, err = client.GetEnrollments(tt.ctx, &qf.EnrollmentRequest{
				FetchMode: &qf.EnrollmentRequest_UserID{
					UserID: tt.userID,
				},
			})
			checkAccess(t, "GetEnrollments", err, tt.wantCode, tt.wantAccess)
		})
	}

	// Test student role in GetEnrollments (when student queries course enrollments, not their own user)
	studentGetEnrollmentsTests := map[string]accessTest{
		"student querying course enrollments": {ctx: studentCtx, courseID: course.GetID(), wantAccess: true},
		"student of another course":           {ctx: studentCtx, courseID: 999, wantAccess: false, wantCode: connect.CodePermissionDenied},
	}

	for name, tt := range studentGetEnrollmentsTests {
		t.Run("StudentGetEnrollments/"+name, func(t *testing.T) {
			_, err := client.GetEnrollments(tt.ctx, &qf.EnrollmentRequest{
				FetchMode: &qf.EnrollmentRequest_CourseID{
					CourseID: tt.courseID,
				},
			})
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

	ctx := client.Context(t, userA)
	// Attempt to update the submission from course1 while acting as teacher in course2
	// This should fail because the submission belongs to course1, not course2.
	_, err := client.UpdateSubmission(ctx, &qf.Grade{
		SubmissionID: submission.GetID(),
		// TODO(meling): UpdateSubmissionRequest had these fields, but Grade does not. Should we add CourseID to Grade?
		// CourseID:     course2.GetID(), // Wrong course ID
		// Score:        100,
		// Released:     true,
	})

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
	ctx := client.Context(t, user)
	_, err := client.GetUser(ctx, &qf.Void{})
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
