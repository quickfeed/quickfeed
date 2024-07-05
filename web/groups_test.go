package web_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
)

func TestNewGroup(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := web.MockClientWithOption(t, db, scm.WithMockOrgs())

	admin := qtest.CreateFakeUser(t, db)
	var course qf.Course
	// only created 1 directory, if we had created two directories ID would be 2
	course.ScmOrganizationID = 1
	course.ScmOrganizationName = "test"
	qtest.CreateCourse(t, db, admin, &course)
	user := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, user, &course)

	ctx := context.Background()
	// current user must be in the group being created
	createGroupRequest := qtest.RequestWithCookie(&qf.Group{Name: "Heins-Group", CourseID: course.ID, Users: []*qf.User{{ID: user.ID}}}, Cookie(t, tm, user))
	wantGroup, err := client.CreateGroup(ctx, createGroupRequest)
	if err != nil {
		t.Error(err)
	}
	gotGroup, err := client.GetGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{CourseID: course.ID, GroupID: wantGroup.Msg.ID}, Cookie(t, tm, user)))
	if err != nil {
		t.Error(err)
	}
	if diff := cmp.Diff(wantGroup.Msg, gotGroup.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("CreateGroup() mismatch (-wantGroup +gotGroup):\n%s", diff)
	}
}

func TestCreateGroupWithMissingFields(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := web.MockClientWithOption(t, db, scm.WithMockOrgs())

	admin := qtest.CreateFakeUser(t, db)
	var course qf.Course
	// only created 1 directory, if we had created two directories ID would be 2
	course.ScmOrganizationID = 1
	qtest.CreateCourse(t, db, admin, &course)
	user := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, user, &course)

	users := []*qf.User{{ID: user.ID}}

	ctx := context.Background()

	// current user must be in the group being created
	group_wo_course_id := qtest.RequestWithCookie(&qf.Group{Name: "Hein's Group", Users: users}, Cookie(t, tm, user))
	_, err := client.CreateGroup(ctx, group_wo_course_id)
	if err == nil {
		t.Fatal("expected CreateGroup to fail without a course ID")
	}
	group_wo_name := qtest.RequestWithCookie(&qf.Group{CourseID: course.ID, Users: users}, Cookie(t, tm, user))
	if group_wo_name.Msg.IsValid() {
		// emulate CreateGroup check without name
		t.Fatal("expected CreateGroup to fail without group name")
	}
	group_wo_users := qtest.RequestWithCookie(&qf.Group{Name: "Hein's Group", CourseID: course.ID}, Cookie(t, tm, user))
	_, err = client.CreateGroup(ctx, group_wo_users)
	if err == nil {
		t.Fatal("expected CreateGroup to fail without users")
	}
}

func TestNewGroupTeacherCreator(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := web.MockClientWithOption(t, db, scm.WithMockOrgs())

	admin := qtest.CreateFakeUser(t, db)
	var course qf.Course
	// only created 1 directory, if we had created two directories ID would be 2
	course.ScmOrganizationID = 1
	qtest.CreateCourse(t, db, admin, &course)

	teacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, teacher, &course)

	user := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, user, &course)

	users := []*qf.User{{ID: user.ID}}
	ctx := context.Background()

	// current user must be in the group being created
	wantGroup, err := client.CreateGroup(ctx, qtest.RequestWithCookie(&qf.Group{Name: "HeinsGroup", CourseID: course.ID, Users: users}, Cookie(t, tm, user)))
	if err != nil {
		t.Error(err)
	}
	// check that group member (user) can access group
	groupReq := &qf.GroupRequest{CourseID: course.ID, GroupID: wantGroup.Msg.ID}
	gotGroup, err := client.GetGroup(ctx, qtest.RequestWithCookie(groupReq, Cookie(t, tm, user)))
	if err != nil {
		t.Error(err)
	}
	// check that teacher can access group
	_, err = client.GetGroup(ctx, qtest.RequestWithCookie(groupReq, Cookie(t, tm, teacher)))
	if err != nil {
		t.Error(err)
	}
	// check that admin can access group
	_, err = client.GetGroup(ctx, qtest.RequestWithCookie(groupReq, Cookie(t, tm, admin)))
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(wantGroup.Msg, gotGroup.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("CreateGroup() mismatch (-wantGroup +gotGroup):\n%s", diff)
	}
}

func TestNewGroupStudentCreateGroupWithTeacher(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := web.MockClientWithOption(t, db, scm.WithMockOrgs())

	admin := qtest.CreateFakeUser(t, db)
	var course qf.Course
	// only created 1 directory, if we had created two directories ID would be 2
	course.ScmOrganizationID = 1
	qtest.CreateCourse(t, db, admin, &course)

	teacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, teacher, &course)

	user := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, user, &course)

	// current user must be in the group being created
	group_req := qtest.RequestWithCookie(&qf.Group{
		Name:     "HeinsGroup",
		CourseID: course.ID,
		Users:    []*qf.User{{ID: user.ID}, {ID: teacher.ID}},
	}, Cookie(t, tm, user))
	_, err := client.CreateGroup(context.Background(), group_req)
	if err != nil {
		t.Error(err)
	}
	// we now allow teacher/student groups to be created,
	// since if undesirable these can be rejected.
}

func TestStudentCreateNewGroupTeacherUpdateGroup(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := web.MockClientWithOption(t, db, scm.WithMockOrgs())

	admin := qtest.CreateFakeUser(t, db)
	course := qf.Course{ScmOrganizationID: 1, ScmOrganizationName: qtest.MockOrg}
	qtest.CreateCourse(t, db, admin, &course)

	teacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, teacher, &course)

	// create named users; needed for group creation
	user1 := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "user1"})
	user2 := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "user2"})
	user3 := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "user3"})
	qtest.EnrollStudent(t, db, user1, &course)
	qtest.EnrollStudent(t, db, user2, &course)
	qtest.EnrollStudent(t, db, user3, &course)

	// set user1 in cookie, which is a group member
	// group with two students
	createGroupRequest := qtest.RequestWithCookie(&qf.Group{
		Name:     "HeinsTwoMemberGroup",
		CourseID: course.ID,
		Users:    []*qf.User{user1, user2},
	}, Cookie(t, tm, user1))

	ctx := context.Background()
	wantGroup, err := client.CreateGroup(ctx, createGroupRequest)
	if err != nil {
		t.Error(err)
	}

	gotGroup, err := client.GetGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{
		CourseID: course.ID,
		GroupID:  wantGroup.Msg.ID,
	}, Cookie(t, tm, user1)))
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(wantGroup.Msg, gotGroup.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("CreateGroup() mismatch (-wantGroup +gotGroup):\n%s", diff)
	}

	// ******************* Teacher UpdateGroup *******************

	// set teacher in cookie
	// group with three students
	updateGroupRequest := qtest.RequestWithCookie(&qf.Group{
		ID:       gotGroup.Msg.ID,
		Name:     "Heins3MemberGroup",
		CourseID: course.ID,
		Users:    []*qf.User{user1, user2, user3},
	}, Cookie(t, tm, teacher))
	gotUpdatedGroup, err := client.UpdateGroup(ctx, updateGroupRequest)
	if err != nil {
		t.Error(err)
	}

	// check that the group have changed group membership
	userIDs := make([]uint64, 0)
	for _, usr := range updateGroupRequest.Msg.Users {
		userIDs = append(userIDs, usr.ID)
	}

	grpUsers, err := db.GetUsers(userIDs...)
	if err != nil {
		t.Fatal(err)
	}

	wantGroup = gotGroup
	wantGroup.Msg.Name = updateGroupRequest.Msg.Name
	wantGroup.Msg.Users = grpUsers
	// UpdateGroup will autoApprove group on update
	wantGroup.Msg.Status = qf.Group_APPROVED
	// Ignore enrollments in check
	gotUpdatedGroup.Msg.Enrollments = nil
	wantGroup.Msg.Enrollments = nil

	if diff := cmp.Diff(wantGroup.Msg, gotUpdatedGroup.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateGroup() mismatch (-wantGroup +gotUpdatedGroup):\n%s", diff)
	}

	// ******************* Teacher UpdateGroup *******************

	// change group to only one student
	// name must not update because group repo already exist
	updateGroupRequest1 := qtest.RequestWithCookie(&qf.Group{
		ID:       gotGroup.Msg.ID,
		Name:     "Hein's single member Group",
		CourseID: course.ID,
		Users:    []*qf.User{user1},
	}, Cookie(t, tm, teacher))
	gotUpdatedGroup, err = client.UpdateGroup(ctx, updateGroupRequest1)
	if err != nil {
		t.Error(err)
	}
	// check that the group have changed group membership
	userIDs = make([]uint64, 0)
	for _, usr := range updateGroupRequest1.Msg.Users {
		userIDs = append(userIDs, usr.ID)
	}

	grpUsers, err = db.GetUsers(userIDs...)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotUpdatedGroup.Msg.Users) != 1 {
		t.Errorf("Expected only single member group, got %d members", len(gotUpdatedGroup.Msg.Users))
	}
	wantGroup.Msg = updateGroupRequest.Msg
	wantGroup.Msg.Users = grpUsers
	// UpdateGroup will autoApprove group on update
	wantGroup.Msg.Status = qf.Group_APPROVED
	gotUpdatedGroup.Msg.Enrollments = nil
	wantGroup.Msg.Enrollments = nil

	if diff := cmp.Diff(wantGroup.Msg, gotUpdatedGroup.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateGroup() mismatch (-wantGroup +gotUpdatedGroup):\n%s", diff)
	}
}

func TestDeleteGroup(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := web.MockClientWithOption(t, db, scm.WithMockCourses())
	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "admin", Login: "admin"})

	ctx := context.Background()
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, admin, course)

	// create user and enroll as pending (teacher)
	teacher := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "teacher", Login: "teacher"})
	if _, err := client.CreateEnrollment(ctx, qtest.RequestWithCookie(&qf.Enrollment{
		UserID:   teacher.ID,
		CourseID: course.ID,
	}, Cookie(t, tm, teacher))); err != nil {
		t.Error(err)
	}

	// update enrollment from pending->student->teacher; must be done by admin
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(&qf.Enrollments{
		Enrollments: []*qf.Enrollment{
			{
				UserID:   teacher.ID,
				CourseID: course.ID,
				Status:   qf.Enrollment_STUDENT,
			},
		},
	}, Cookie(t, tm, admin))); err != nil {
		t.Error(err)
	}

	// update enrollment to teacher
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(&qf.Enrollments{
		Enrollments: []*qf.Enrollment{
			{
				UserID:   teacher.ID,
				CourseID: course.ID,
				Status:   qf.Enrollment_TEACHER,
			},
		},
	}, Cookie(t, tm, admin))); err != nil {
		t.Error(err)
	}

	// create user and enroll as pending (student)
	user := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "student", Login: "student"})
	if _, err := client.CreateEnrollment(ctx, qtest.RequestWithCookie(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
	}, Cookie(t, tm, user))); err != nil {
		t.Error(err)
	}

	// update pending enrollment to student; must be done by teacher
	if _, err := client.UpdateEnrollments(ctx, qtest.RequestWithCookie(&qf.Enrollments{
		Enrollments: []*qf.Enrollment{
			{
				UserID:   user.ID,
				CourseID: course.ID,
				Status:   qf.Enrollment_STUDENT,
			},
		},
	}, Cookie(t, tm, teacher))); err != nil {
		t.Error(err)
	}

	// create group as student user
	group := &qf.Group{Name: "TestDeleteGroup", CourseID: course.ID, Users: []*qf.User{user}}
	respGroup, err := client.CreateGroup(ctx, qtest.RequestWithCookie(group, Cookie(t, tm, user)))
	if err != nil {
		t.Fatal(err)
	}

	// delete group as teacher
	_, err = client.DeleteGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{
		GroupID:  respGroup.Msg.ID,
		CourseID: course.ID,
	}, Cookie(t, tm, teacher)))
	if err != nil {
		t.Error(err)
	}
}

func TestGetGroup(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := web.MockClientWithOption(t, db, scm.WithMockOrgs())

	course := qf.Course{
		Name:              "Distributed Systems",
		Code:              "DAT520",
		Year:              2018,
		Tag:               "Spring",
		ScmOrganizationID: 1,
	}
	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, &course)

	// create user and enroll as student
	user := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, user, &course)

	ctx := context.Background()

	group := &qf.Group{Name: "TestGroup", CourseID: course.ID, Users: []*qf.User{user}}
	wantGroup, err := client.CreateGroup(ctx, qtest.RequestWithCookie(group, Cookie(t, tm, user)))
	if err != nil {
		t.Error(err)
	}

	gotGroup, err := client.GetGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{
		CourseID: course.ID,
		GroupID:  wantGroup.Msg.ID,
	}, Cookie(t, tm, user)))
	if err != nil {
		t.Error(err)
	}
	if diff := cmp.Diff(wantGroup.Msg, gotGroup.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("CreateGroup() mismatch (-wantGroup +gotGroup):\n%s", diff)
	}
}

func TestPatchGroupStatus(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := web.MockClientWithOption(t, db, scm.WithMockOrgs())

	course := qf.Course{
		Name:                "Distributed Systems",
		Code:                "DAT520",
		Year:                2018,
		Tag:                 "Spring",
		ScmOrganizationID:   1,
		ScmOrganizationName: qtest.MockOrg,
		ID:                  1,
	}

	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, &course)

	teacher := qtest.CreateFakeUser(t, db)
	qtest.EnrollTeacher(t, db, teacher, &course)

	if err := db.UpdateUser(&qf.User{ID: teacher.ID, IsAdmin: true}); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	user1 := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "user1"})
	user2 := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "user2"})

	// enroll users in course and group
	qtest.EnrollStudent(t, db, user1, &course)
	qtest.EnrollStudent(t, db, user2, &course)

	group := &qf.Group{
		ID:       1,
		Name:     "Test Group",
		CourseID: course.ID,
		Users:    []*qf.User{user1, user2},
	}
	err := db.CreateGroup(group)
	if err != nil {
		t.Fatal(err)
	}
	// get the group as stored in db with enrollments
	wantGroup, err := db.GetGroup(group.ID)
	if err != nil {
		t.Fatal(err)
	}

	wantGroup.Status = qf.Group_APPROVED
	gotGroup, err := client.UpdateGroup(ctx, qtest.RequestWithCookie(wantGroup, Cookie(t, tm, teacher)))
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(wantGroup, gotGroup.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateGroup() mismatch (-wantGroup +gotGroup):\n%s", diff)
	}
}

func TestGetGroupByUserAndCourse(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := web.MockClientWithOption(t, db, scm.WithMockOrgs())

	course := qf.Course{
		Name:              "Distributed Systems",
		Code:              "DAT520",
		Year:              2018,
		Tag:               "Spring",
		ScmOrganizationID: 1,
		ID:                1,
	}

	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, &course)

	ctx := context.Background()

	user1 := qtest.CreateFakeUser(t, db)
	user2 := qtest.CreateFakeUser(t, db)

	// enroll users in course and group
	qtest.EnrollStudent(t, db, user1, &course)
	qtest.EnrollStudent(t, db, user2, &course)

	group := &qf.Group{
		CourseID: course.ID,
		Users:    []*qf.User{user1, user2},
	}
	err := db.CreateGroup(group)
	if err != nil {
		t.Fatal(err)
	}

	wantGroup, err := client.GetGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{
		CourseID: course.ID,
		UserID:   user1.ID,
	}, Cookie(t, tm, admin)))
	if err != nil {
		t.Error(err)
	}
	gotGroup, err := client.GetGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{
		CourseID: course.ID,
		GroupID:  group.ID,
	}, Cookie(t, tm, admin)))
	if err != nil {
		t.Error(err)
	}
	if diff := cmp.Diff(wantGroup.Msg, gotGroup.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("GetGroupByUserAndCourse() mismatch (-wantGroup +gotGroup):\n%s", diff)
	}
}

func TestDeleteApprovedGroup(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := web.MockClientWithOption(t, db, scm.WithMockOrgs())

	admin := qtest.CreateFakeUser(t, db)
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, admin, course)

	user1 := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "user1"})
	user2 := qtest.CreateFakeCustomUser(t, db, &qf.User{Login: "user2"})

	// enroll users in course and group
	qtest.EnrollStudent(t, db, user1, course)
	qtest.EnrollStudent(t, db, user2, course)

	group := &qf.Group{
		ID:       1,
		CourseID: course.ID,
		Name:     "TestGroup",
		Users:    []*qf.User{user1, user2},
	}
	// current user1 must be in the group being created
	ctx := context.Background()
	createdGroup, err := client.CreateGroup(ctx, qtest.RequestWithCookie(group, Cookie(t, tm, user1)))
	if err != nil {
		t.Error(err)
	}

	// first approve the group
	createdGroup.Msg.Status = qf.Group_APPROVED
	// current user must be teacher for the course
	if _, err = client.UpdateGroup(ctx, qtest.RequestWithCookie(createdGroup.Msg, Cookie(t, tm, admin))); err != nil {
		t.Error(err)
	}

	// then get user enrollments with group ID
	wantEnrollment1, err := db.GetEnrollmentByCourseAndUser(course.ID, user1.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantEnrollment2, err := db.GetEnrollmentByCourseAndUser(course.ID, user2.ID)
	if err != nil {
		t.Fatal(err)
	}

	// delete the group
	if _, err = client.DeleteGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{
		CourseID: course.ID,
		GroupID:  createdGroup.Msg.ID,
	}, Cookie(t, tm, admin))); err != nil {
		t.Error(err)
	}

	// get updated enrollments of group members
	gotEnrollment1, err := db.GetEnrollmentByCourseAndUser(course.ID, user1.ID)
	if err != nil {
		t.Fatal(err)
	}
	gotEnrollment2, err := db.GetEnrollmentByCourseAndUser(course.ID, user2.ID)
	if err != nil {
		t.Fatal(err)
	}

	// now nullify manually group ID for original enrollments
	wantEnrollment1.GroupID = 0
	wantEnrollment2.GroupID = 0

	// then check that new enrollments have group IDs nullified automatically
	if diff := cmp.Diff(wantEnrollment1, gotEnrollment1, protocmp.Transform()); diff != "" {
		t.Errorf("DeleteGroup() mismatch (-wantEnrollment1 +gotEnrollment1):\n%s", diff)
	}
	if diff := cmp.Diff(wantEnrollment2, gotEnrollment2, protocmp.Transform()); diff != "" {
		t.Errorf("DeleteGroup() mismatch (-wantEnrollment2 +gotEnrollment2):\n%s", diff)
	}
}

func TestGetGroups(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := web.MockClientWithOption(t, db, scm.WithMockOrgs())

	var users []*qf.User
	for i := 0; i < 10; i++ {
		user := qtest.CreateFakeUser(t, db)
		users = append(users, user)
	}
	admin := users[0]

	// admin will be enrolled as teacher because of course creation below
	course := qtest.MockCourses[1]
	qtest.CreateCourse(t, db, admin, course)

	// enroll all users in course
	for _, user := range users[1:] {
		qtest.EnrollStudent(t, db, user, course)
	}
	// place some students in groups
	// current user must be in the group being created
	ctx := context.Background()
	group1, err := client.CreateGroup(ctx, qtest.RequestWithCookie(&qf.Group{
		Name:     "Group1",
		CourseID: course.ID,
		Users:    []*qf.User{users[1], users[2]},
	}, Cookie(t, tm, users[2])))
	if err != nil {
		t.Error(err)
	}
	group2, err := client.CreateGroup(ctx, qtest.RequestWithCookie(&qf.Group{
		Name:     "Group2",
		CourseID: course.ID,
		Users:    []*qf.User{users[4], users[5]},
	}, Cookie(t, tm, users[5])))
	if err != nil {
		t.Error(err)
	}
	wantGroups := &qf.Groups{Groups: []*qf.Group{group1.Msg, group2.Msg}}
	for _, grp := range wantGroups.Groups {
		for _, grpEnrol := range grp.Enrollments {
			grpEnrol.UsedSlipDays = []*qf.UsedSlipDays{}
		}
	}

	// get groups from the database; current user is admin, which is also teacher
	gotGroups, err := client.GetGroupsByCourse(ctx, qtest.RequestWithCookie(&qf.CourseRequest{
		CourseID: course.ID,
	}, Cookie(t, tm, admin)))
	if err != nil {
		t.Error(err)
	}

	// check that the method returns expected groups
	if diff := cmp.Diff(wantGroups, gotGroups.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("GetGroupsByCourse() mismatch (-wantGroups +gotGroups):\n%s", diff)
	}
}
