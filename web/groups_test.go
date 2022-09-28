package web_test

import (
	"context"
	"os"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

func TestNewGroup(t *testing.T) {
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	var course qf.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.OrganizationID = 1
	course.OrganizationName = "test"
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}
	user := qtest.CreateFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: user.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	createGroupRequest := connect.NewRequest(&qf.Group{Name: "Heins-Group", CourseID: course.ID, Users: []*qf.User{{ID: user.ID}}})
	// current user (in context) must be in group being created
	ctx := qtest.WithUserContext(context.Background(), user)
	wantGroup, err := ags.CreateGroup(ctx, createGroupRequest)
	if err != nil {
		t.Fatal(err)
	}
	gotGroup, err := ags.GetGroup(ctx, connect.NewRequest(&qf.GetGroupRequest{GroupID: wantGroup.Msg.ID}))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantGroup.Msg, gotGroup.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("ags.CreateGroup() mismatch (-wantGroup +gotGroup):\n%s", diff)
	}
}

func TestCreateGroupWithMissingFields(t *testing.T) {
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	var course qf.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.OrganizationID = 1
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}
	user := qtest.CreateFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: user.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	users := []*qf.User{{ID: user.ID}}
	group_wo_course_id := connect.NewRequest(&qf.Group{Name: "Hein's Group", Users: users})
	group_wo_name := connect.NewRequest(&qf.Group{CourseID: course.ID, Users: users})
	group_wo_users := connect.NewRequest(&qf.Group{Name: "Hein's Group", CourseID: course.ID})

	// current user (in context) must be in group being created
	ctx := qtest.WithUserContext(context.Background(), user)
	_, err := ags.CreateGroup(ctx, group_wo_course_id)
	if err == nil {
		t.Fatal("expected CreateGroup to fail without a course ID")
	}
	if group_wo_name.Msg.IsValid() {
		// emulate CreateGroup check without name
		t.Fatal("expected CreateGroup to fail without group name")
	}
	_, err = ags.CreateGroup(ctx, group_wo_users)
	if err == nil {
		t.Fatal("expected CreateGroup to fail without users")
	}
}

func TestNewGroupTeacherCreator(t *testing.T) {
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	var course qf.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.OrganizationID = 1
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}

	teacher := qtest.CreateFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: teacher.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   teacher.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_TEACHER,
	}); err != nil {
		t.Fatal(err)
	}

	user := qtest.CreateFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: user.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	users := []*qf.User{{ID: user.ID}}
	createGroupRequest := connect.NewRequest(&qf.Group{Name: "HeinsGroup", CourseID: course.ID, Users: users})

	ctx := qtest.WithUserContext(context.Background(), user)
	wantGroup, err := ags.CreateGroup(ctx, createGroupRequest)
	if err != nil {
		t.Fatal(err)
	}

	// check that gotGroup member can access gotGroup
	gotGroup, err := ags.GetGroup(ctx, connect.NewRequest(&qf.GetGroupRequest{GroupID: wantGroup.Msg.ID}))
	if err != nil {
		t.Fatal(err)
	}
	// check that teacher can access group
	ctx = qtest.WithUserContext(context.Background(), teacher)
	_, err = ags.GetGroup(ctx, connect.NewRequest(&qf.GetGroupRequest{GroupID: wantGroup.Msg.ID}))
	if err != nil {
		t.Fatal(err)
	}
	// check that admin can access group
	ctx = qtest.WithUserContext(context.Background(), admin)
	_, err = ags.GetGroup(ctx, connect.NewRequest(&qf.GetGroupRequest{GroupID: wantGroup.Msg.ID}))
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantGroup.Msg, gotGroup.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("ags.CreateGroup() mismatch (-wantGroup +gotGroup):\n%s", diff)
	}
}

func TestNewGroupStudentCreateGroupWithTeacher(t *testing.T) {
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	var course qf.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.OrganizationID = 1
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}

	teacher := qtest.CreateFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: teacher.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   teacher.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_TEACHER,
	}); err != nil {
		t.Fatal(err)
	}

	user := qtest.CreateFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: user.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	ctx := qtest.WithUserContext(context.Background(), user)

	group_req := connect.NewRequest(&qf.Group{
		Name:     "HeinsGroup",
		CourseID: course.ID,
		Users:    []*qf.User{{ID: user.ID}, {ID: teacher.ID}},
	})
	_, err := ags.CreateGroup(ctx, group_req)
	if err != nil {
		t.Fatal(err)
	}
	// we now allow teacher/student groups to be created,
	// since if undesirable these can be rejected.
}

func TestStudentCreateNewGroupTeacherUpdateGroup(t *testing.T) {
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	course := qf.Course{Provider: "fake", OrganizationID: 1, OrganizationName: "test"}
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}

	teacher := qtest.CreateFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: teacher.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   teacher.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_TEACHER,
	}); err != nil {
		t.Fatal(err)
	}

	user1 := qtest.CreateFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: user1.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   user1.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	user2 := qtest.CreateFakeUser(t, db, 4)
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: user2.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   user2.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	user3 := qtest.CreateFakeUser(t, db, 5)
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: user3.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   user3.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	// group with two students
	createGroupRequest := connect.NewRequest(&qf.Group{
		Name:     "HeinsTwoMemberGroup",
		CourseID: course.ID,
		Users:    []*qf.User{user1, user2},
	})

	// set ID of user1, which is group member
	ctx := qtest.WithUserContext(context.Background(), user1)
	wantGroup, err := ags.CreateGroup(ctx, createGroupRequest)
	if err != nil {
		t.Fatal(err)
	}

	gotGroup, err := ags.GetGroup(ctx, connect.NewRequest(&qf.GetGroupRequest{
		GroupID: wantGroup.Msg.ID,
	},
	))
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantGroup.Msg, gotGroup.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("ags.CreateGroup() mismatch (-wantGroup +gotGroup):\n%s", diff)
	}

	// ******************* Teacher UpdateGroup *******************

	// group with three students
	updateGroupRequest := connect.NewRequest(&qf.Group{
		ID:       gotGroup.Msg.ID,
		Name:     "Heins3MemberGroup",
		CourseID: course.ID,
		Users:    []*qf.User{user1, user2, user3},
	})

	// set teacher ID in context
	ctx = qtest.WithUserContext(context.Background(), teacher)
	gotUpdatedGroup, err := ags.UpdateGroup(ctx, updateGroupRequest)
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
	wantGroup.Msg.TeamID = 1
	// UpdateGroup will autoApprove group on update
	wantGroup.Msg.Status = qf.Group_APPROVED
	// Ignore enrollments in check
	gotUpdatedGroup.Msg.Enrollments = nil
	wantGroup.Msg.Enrollments = nil

	if diff := cmp.Diff(wantGroup.Msg, gotUpdatedGroup.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("ags.UpdateGroup() mismatch (-wantGroup +gotUpdatedGroup):\n%s", diff)
	}

	// ******************* Teacher UpdateGroup *******************

	// change group to only one student
	// name must not update because group team and repo already exist
	updateGroupRequest1 := connect.NewRequest(&qf.Group{
		ID:       gotGroup.Msg.ID,
		Name:     "Hein's single member Group",
		CourseID: course.ID,
		Users:    []*qf.User{user1},
	})

	// set teacher ID in context
	ctx = qtest.WithUserContext(context.Background(), teacher)
	gotUpdatedGroup, err = ags.UpdateGroup(ctx, updateGroupRequest1)
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
	wantGroup.Msg.TeamID = 1
	// UpdateGroup will autoApprove group on update
	wantGroup.Msg.Status = qf.Group_APPROVED
	gotUpdatedGroup.Msg.Enrollments = nil
	wantGroup.Msg.Enrollments = nil

	if diff := cmp.Diff(wantGroup.Msg, gotUpdatedGroup.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("ags.UpdateGroup() mismatch (-wantGroup +gotUpdatedGroup):\n%s", diff)
	}
}

func TestDeleteGroup(t *testing.T) {
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()

	testCourse := qf.Course{
		Name:             "Distributed Systems",
		Code:             "DAT520",
		Year:             2018,
		Tag:              "Spring",
		Provider:         "fake",
		OrganizationID:   1,
		OrganizationName: "test",
		ID:               1,
	}
	admin := qtest.CreateFakeUser(t, db, 1)

	ctx := qtest.WithUserContext(context.Background(), admin)
	if _, err := ags.CreateCourse(ctx, connect.NewRequest(&testCourse)); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as pending (teacher)
	teacher := qtest.CreateFakeUser(t, db, 3)
	ctx = qtest.WithUserContext(context.Background(), teacher)
	if _, err := ags.CreateEnrollment(ctx, connect.NewRequest(&qf.Enrollment{
		UserID:   teacher.ID,
		CourseID: testCourse.ID,
	})); err != nil {
		t.Fatal(err)
	}

	if os.Getenv("TODO") == "" {
		t.Skip("See TODO description")
	}
	// TODO(meling) Calling UpdateEnrollments will trigger enrollStudent and acceptRepositoryInvites
	// This requires a real or fake SCMInvite App that implements invite handling
	// Specifically, the Config.ExchangeToken() method should have a fake implementation.

	// update enrollment from pending->student->teacher; must be done by admin
	ctx = qtest.WithUserContext(context.Background(), admin)
	if _, err := ags.UpdateEnrollments(ctx, connect.NewRequest(&qf.Enrollments{
		Enrollments: []*qf.Enrollment{
			{
				UserID:   teacher.ID,
				CourseID: testCourse.ID,
				Status:   qf.Enrollment_STUDENT,
			},
		},
	})); err != nil {
		t.Fatal(err)
	}

	// update enrollment to teacher
	if _, err := ags.UpdateEnrollments(ctx, connect.NewRequest(&qf.Enrollments{
		Enrollments: []*qf.Enrollment{
			{
				UserID:   teacher.ID,
				CourseID: testCourse.ID,
				Status:   qf.Enrollment_TEACHER,
			},
		},
	})); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as pending (student)
	user := qtest.CreateFakeUser(t, db, 2)
	ctx = qtest.WithUserContext(context.Background(), user)
	if _, err := ags.CreateEnrollment(ctx, connect.NewRequest(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: testCourse.ID,
	})); err != nil {
		t.Fatal(err)
	}

	// update pending enrollment to student; must be done by teacher
	ctx = qtest.WithUserContext(context.Background(), teacher)
	if _, err := ags.UpdateEnrollments(ctx, connect.NewRequest(&qf.Enrollments{
		Enrollments: []*qf.Enrollment{
			{
				UserID:   user.ID,
				CourseID: testCourse.ID,
				Status:   qf.Enrollment_STUDENT,
			},
		},
	})); err != nil {
		t.Fatal(err)
	}

	// create group as student user
	group := &qf.Group{Name: "TestDeleteGroup", CourseID: testCourse.ID, Users: []*qf.User{user}}
	ctx = qtest.WithUserContext(context.Background(), user)
	respGroup, err := ags.CreateGroup(ctx, connect.NewRequest(group))
	if err != nil {
		t.Fatal(err)
	}

	// delete group as teacher
	ctx = qtest.WithUserContext(context.Background(), teacher)
	_, err = ags.DeleteGroup(ctx, connect.NewRequest(&qf.GroupRequest{
		GroupID:  respGroup.Msg.ID,
		CourseID: testCourse.ID,
	}))
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetGroup(t *testing.T) {
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()

	testCourse := qf.Course{
		Name:           "Distributed Systems",
		Code:           "DAT520",
		Year:           2018,
		Tag:            "Spring",
		Provider:       "fake",
		OrganizationID: 1,
	}
	admin := qtest.CreateFakeUser(t, db, 1)
	if err := db.CreateCourse(admin.ID, &testCourse); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as student
	user := qtest.CreateFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: user.ID, CourseID: testCourse.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   user.ID,
		CourseID: testCourse.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	ctx := qtest.WithUserContext(context.Background(), user)

	group := &qf.Group{Name: "TestGroup", CourseID: testCourse.ID, Users: []*qf.User{user}}
	wantGroup, err := ags.CreateGroup(ctx, connect.NewRequest(group))
	if err != nil {
		t.Fatal(err)
	}

	gotGroup, err := ags.GetGroup(ctx, connect.NewRequest(&qf.GetGroupRequest{
		GroupID: wantGroup.Msg.ID,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantGroup.Msg, gotGroup.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("ags.CreateGroup() mismatch (-wantGroup +gotGroup):\n%s", diff)
	}
}

func TestPatchGroupStatus(t *testing.T) {
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()

	course := qf.Course{
		Name:             "Distributed Systems",
		Code:             "DAT520",
		Year:             2018,
		Tag:              "Spring",
		Provider:         "fake",
		OrganizationID:   1,
		OrganizationName: "test",
		ID:               1,
	}

	admin := qtest.CreateFakeUser(t, db, 1)
	err := db.CreateCourse(admin.ID, &course)
	if err != nil {
		t.Fatal(err)
	}

	teacher := qtest.CreateFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&qf.Enrollment{UserID: teacher.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   teacher.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_TEACHER,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateUser(&qf.User{ID: teacher.ID, IsAdmin: true}); err != nil {
		t.Fatal(err)
	}

	ctx := qtest.WithUserContext(context.Background(), teacher)

	user1 := qtest.CreateFakeUser(t, db, 3)
	user2 := qtest.CreateFakeUser(t, db, 4)

	// enroll users in course and group
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID: user1.ID, CourseID: course.ID, GroupID: 1,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   user1.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID: user2.ID, CourseID: course.ID, GroupID: 1,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   user2.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	group := &qf.Group{
		ID:       1,
		Name:     "Test Group",
		CourseID: course.ID,
		Users:    []*qf.User{user1, user2},
		TeamID:   1,
	}
	err = db.CreateGroup(group)
	if err != nil {
		t.Fatal(err)
	}
	// get the group as stored in db with enrollments
	wantGroup, err := db.GetGroup(group.ID)
	if err != nil {
		t.Fatal(err)
	}

	wantGroup.Status = qf.Group_APPROVED
	gotGroup, err := ags.UpdateGroup(ctx, connect.NewRequest(wantGroup))
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(wantGroup, gotGroup.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("ags.UpdateGroup() mismatch (-wantGroup +gotGroup):\n%s", diff)
	}
}

func TestGetGroupByUserAndCourse(t *testing.T) {
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()

	course := qf.Course{
		Name:           "Distributed Systems",
		Code:           "DAT520",
		Year:           2018,
		Tag:            "Spring",
		Provider:       "fake",
		OrganizationID: 1,
		ID:             1,
	}

	admin := qtest.CreateFakeUser(t, db, 1)
	err := db.CreateCourse(admin.ID, &course)
	if err != nil {
		t.Fatal(err)
	}

	ctx := qtest.WithUserContext(context.Background(), admin)

	user1 := qtest.CreateFakeUser(t, db, 2)
	user2 := qtest.CreateFakeUser(t, db, 3)

	// enroll users in course and group
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID: user1.ID, CourseID: course.ID, GroupID: 1,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   user1.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID: user2.ID, CourseID: course.ID, GroupID: 1,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   user2.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	group := &qf.Group{
		ID:       1,
		CourseID: course.ID,
		Users:    []*qf.User{user1, user2},
	}
	err = db.CreateGroup(group)
	if err != nil {
		t.Fatal(err)
	}

	wantGroup, err := ags.GetGroupByUserAndCourse(ctx, connect.NewRequest(&qf.GroupRequest{
		UserID:   user1.ID,
		CourseID: course.ID,
	}))
	if err != nil {
		t.Error(err)
	}
	gotGroup, err := ags.GetGroup(ctx, connect.NewRequest(&qf.GetGroupRequest{
		GroupID: group.ID,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantGroup.Msg, gotGroup.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("ags.GetGroupByUserAndCourse() mismatch (-wantGroup +gotGroup):\n%s", diff)
	}
}

func TestDeleteApprovedGroup(t *testing.T) {
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	course := qtest.MockCourses[0]
	err := db.CreateCourse(admin.ID, course)
	if err != nil {
		t.Fatal(err)
	}

	user1 := qtest.CreateFakeUser(t, db, 2)
	user2 := qtest.CreateFakeUser(t, db, 3)

	// enroll users in course and group
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID: user1.ID, CourseID: course.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   user1.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&qf.Enrollment{
		UserID: user2.ID, CourseID: course.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   user2.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&qf.Enrollment{
		UserID:   admin.ID,
		CourseID: course.ID,
		Status:   qf.Enrollment_TEACHER,
	}); err != nil {
		t.Fatal(err)
	}

	group := &qf.Group{
		ID:       1,
		CourseID: course.ID,
		Name:     "TestGroup",
		Users:    []*qf.User{user1, user2},
	}
	// current user1 (in context) must be in group being created
	ctx := qtest.WithUserContext(context.Background(), user1)
	createdGroup, err := ags.CreateGroup(ctx, connect.NewRequest(group))
	if err != nil {
		t.Fatal(err)
	}

	// first approve the group
	createdGroup.Msg.Status = qf.Group_APPROVED
	// current user (in context) must be teacher for the course
	ctx = qtest.WithUserContext(context.Background(), admin)
	if _, err = ags.UpdateGroup(ctx, connect.NewRequest(createdGroup.Msg)); err != nil {
		t.Fatal(err)
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
	if _, err = ags.DeleteGroup(ctx, connect.NewRequest(&qf.GroupRequest{
		CourseID: course.ID,
		GroupID:  createdGroup.Msg.ID,
	})); err != nil {
		t.Fatal(err)
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
		t.Errorf("ags.DeleteGroup() mismatch (-wantEnrollment1 +gotEnrollment1):\n%s", diff)
	}
	if diff := cmp.Diff(wantEnrollment2, gotEnrollment2, protocmp.Transform()); diff != "" {
		t.Errorf("ags.DeleteGroup() mismatch (-wantEnrollment2 +gotEnrollment2):\n%s", diff)
	}
}

func TestGetGroups(t *testing.T) {
	db, cleanup, _, ags := testQuickFeedService(t)
	defer cleanup()

	var users []*qf.User
	for _, u := range allUsers {
		user := qtest.CreateFakeUser(t, db, u.remoteID)
		users = append(users, user)
	}
	admin := users[0]

	// admin will be enrolled as teacher because of course creation below
	course := qtest.MockCourses[1]
	err := db.CreateCourse(admin.ID, course)
	if err != nil {
		t.Fatal(err)
	}

	// enroll all users in course
	for _, user := range users[1:] {
		if err := db.CreateEnrollment(&qf.Enrollment{
			UserID: user.ID, CourseID: course.ID,
		}); err != nil {
			t.Fatal(err)
		}
		if err := db.UpdateEnrollment(&qf.Enrollment{
			UserID:   user.ID,
			CourseID: course.ID,
			Status:   qf.Enrollment_STUDENT,
		}); err != nil {
			t.Fatal(err)
		}
	}
	// place some students in groups
	// current user (in context) must be in group being created
	ctx := qtest.WithUserContext(context.Background(), users[2])
	group1, err := ags.CreateGroup(ctx, connect.NewRequest(&qf.Group{
		Name:     "Group1",
		CourseID: course.ID,
		Users:    []*qf.User{users[1], users[2]},
	}))
	if err != nil {
		t.Fatal(err)
	}
	ctx = qtest.WithUserContext(context.Background(), users[5])
	group2, err := ags.CreateGroup(ctx, connect.NewRequest(&qf.Group{
		Name:     "Group2",
		CourseID: course.ID,
		Users:    []*qf.User{users[4], users[5]},
	}))
	if err != nil {
		t.Fatal(err)
	}
	wantGroups := &qf.Groups{Groups: []*qf.Group{group1.Msg, group2.Msg}}
	for _, grp := range wantGroups.Groups {
		for _, grpEnrol := range grp.Enrollments {
			grpEnrol.UsedSlipDays = []*qf.UsedSlipDays{}
		}
	}

	// get groups from the database; admin is in ctx, which is also teacher
	ctx = qtest.WithUserContext(context.Background(), admin)
	gotGroups, err := ags.GetGroupsByCourse(ctx, connect.NewRequest(&qf.CourseRequest{
		CourseID: course.ID,
	}))
	if err != nil {
		t.Fatal(err)
	}

	// check that the method returns expected groups
	if diff := cmp.Diff(wantGroups, gotGroups.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("ags.GetGroupsByCourse() mismatch (-wantGroups +gotGroups):\n%s", diff)
	}
}
