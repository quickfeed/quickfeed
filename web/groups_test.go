package web_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/scm"
	"github.com/autograde/quickfeed/web"
	_ "github.com/mattn/go-sqlite3"
)

func TestNewGroup(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	var course pb.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.OrganizationID = 1
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}
	user := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	ctx := withUserContext(context.Background(), admin)
	fakeProvider, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})

	_, err := fakeProvider.CreateOrganization(ctx,
		&scm.OrganizationOptions{Path: "path", Name: "name"},
	)
	if err != nil {
		t.Fatal(err)
	}

	users := make([]*pb.User, 0)
	users = append(users, &pb.User{ID: user.ID})
	group_req := &pb.Group{Name: "Hein's Group", CourseID: course.ID, Users: users}

	// current user (in context) must be in group being created
	ctx = withUserContext(context.Background(), user)
	respGroup, err := ags.CreateGroup(ctx, group_req)
	if err != nil {
		t.Fatal(err)
	}

	group, err := ags.GetGroup(ctx, &pb.GetGroupRequest{GroupID: respGroup.ID})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(&respGroup, &group) {
		t.Errorf("have response group %+v, while database has %+v", respGroup, group)
	}
}

func TestCreateGroupWithMissingFields(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	var course pb.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.OrganizationID = 1
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}
	user := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	ctx := withUserContext(context.Background(), admin)
	fakeProvider, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})

	_, err := fakeProvider.CreateOrganization(ctx,
		&scm.OrganizationOptions{Path: "path", Name: "name"},
	)
	if err != nil {
		t.Fatal(err)
	}

	users := make([]*pb.User, 0)
	users = append(users, &pb.User{ID: user.ID})
	group_wo_course_id := &pb.Group{Name: "Hein's Group", Users: users}
	group_wo_name := &pb.Group{CourseID: course.ID, Users: users}
	group_wo_users := &pb.Group{Name: "Hein's Group", CourseID: course.ID}

	// current user (in context) must be in group being created
	ctx = withUserContext(context.Background(), user)
	_, err = ags.CreateGroup(ctx, group_wo_course_id)
	if err == nil {
		t.Fatal("expected CreateGroup to fail without a course ID")
	}
	if group_wo_name.IsValid() {
		// emulate CreateGroup check without name
		t.Fatal("expected CreateGroup to fail without group name")
	}
	_, err = ags.CreateGroup(ctx, group_wo_users)
	if err == nil {
		t.Fatal("expected CreateGroup to fail without users")
	}
}

func TestNewGroupTeacherCreator(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	var course pb.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.OrganizationID = 1
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}

	teacher := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: teacher.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   teacher.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_TEACHER,
	}); err != nil {
		t.Fatal(err)
	}

	user := createFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	fakeProvider, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})

	_, err := fakeProvider.CreateOrganization(context.Background(),
		&scm.OrganizationOptions{Path: "path", Name: "name"},
	)
	if err != nil {
		t.Fatal(err)
	}

	users := make([]*pb.User, 0)
	users = append(users, &pb.User{ID: user.ID})
	group_req := &pb.Group{Name: "Hein's Group", CourseID: course.ID, Users: users}

	ctx := withUserContext(context.Background(), user)
	respGroup, err := ags.CreateGroup(ctx, group_req)
	if err != nil {
		t.Fatal(err)
	}

	// check that group member can access group
	group, err := ags.GetGroup(ctx, &pb.GetGroupRequest{GroupID: respGroup.ID})
	if err != nil {
		t.Fatal(err)
	}
	// check that teacher can access group
	ctx = withUserContext(context.Background(), teacher)
	_, err = ags.GetGroup(ctx, &pb.GetGroupRequest{GroupID: respGroup.ID})
	if err != nil {
		t.Fatal(err)
	}
	// check that admin can access group
	ctx = withUserContext(context.Background(), admin)
	_, err = ags.GetGroup(ctx, &pb.GetGroupRequest{GroupID: respGroup.ID})
	if err != nil {
		t.Fatal(err)
	}

	// JSON marshalling removes the enrollment field from respGroup,
	// so we remove group.Enrollments obtained from the database before comparing.
	//group.Enrollments = nil
	if !cmp.Equal(respGroup, group) {
		t.Errorf("have response group %+v, while database has %+v", respGroup, group)
	}
}

func TestNewGroupStudentCreateGroupWithTeacher(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	var course pb.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.OrganizationID = 1
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}

	teacher := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: teacher.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   teacher.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_TEACHER,
	}); err != nil {
		t.Fatal(err)
	}

	user := createFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	fakeProvider, scms := fakeProviderMap(t)
	ctx := withUserContext(context.Background(), user)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})

	_, err := fakeProvider.CreateOrganization(ctx,
		&scm.OrganizationOptions{Path: "path", Name: "name"},
	)
	if err != nil {
		t.Fatal(err)
	}

	users := make([]*pb.User, 0)
	users = append(users, &pb.User{ID: user.ID})
	users = append(users, &pb.User{ID: teacher.ID})
	group_req := &pb.Group{Name: "Hein's Group", CourseID: course.ID, Users: users}

	_, err = ags.CreateGroup(ctx, group_req)
	if err != nil {
		t.Fatal(err)
	}
	// we now allow teacher/student groups to be created,
	// since if undesirable these can be rejected.
}

func TestStudentCreateNewGroupTeacherUpdateGroup(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	fakeProvider, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	_, err := fakeProvider.CreateOrganization(context.Background(),
		&scm.OrganizationOptions{Path: "path", Name: "name"},
	)
	if err != nil {
		t.Fatal(err)
	}

	admin := createFakeUser(t, db, 1)
	course := pb.Course{Provider: "fake", OrganizationID: 1}
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}

	teacher := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: teacher.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   teacher.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_TEACHER,
	}); err != nil {
		t.Fatal(err)
	}

	user1 := createFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user1.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user1.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	user2 := createFakeUser(t, db, 4)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user2.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user2.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	user3 := createFakeUser(t, db, 5)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user3.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user3.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	// group with two students
	users := make([]*pb.User, 0)
	users = append(users, user1)
	users = append(users, user2)
	newGroupReq := &pb.Group{Name: "Hein's two member Group", CourseID: course.ID, Users: users}

	// set ID of user3 to context, user3 is not member of group (should fail)
	ctx := withUserContext(context.Background(), user3)
	if _, err := ags.CreateGroup(ctx, newGroupReq); err == nil {
		t.Error("expected error 'student must be member of new group'")
	}

	// set ID of user1, which is group member
	ctx = withUserContext(context.Background(), user1)
	respGroup, err := ags.CreateGroup(ctx, newGroupReq)
	if err != nil {
		t.Fatal(err)
	}

	group, err := ags.GetGroup(ctx, &pb.GetGroupRequest{GroupID: respGroup.ID})
	if err != nil {
		t.Fatal(err)
	}

	// JSON marshalling removes the enrollment field from respGroup,
	// so we remove group.Enrollments obtained from the database before comparing.
	//group.Enrollments = nil
	if !reflect.DeepEqual(respGroup, group) {
		t.Errorf("have response group %+v, while database has %+v", respGroup, group)
	}

	// ******************* Teacher UpdateGroup *******************

	// group with three students

	users1 := make([]*pb.User, 0)
	users1 = append(users1, user1)
	users1 = append(users1, user2)
	users1 = append(users1, user3)

	updateGroupReq := &pb.Group{ID: group.ID, Name: "Hein's three member Group", CourseID: course.ID, Users: users1}

	// set teacher ID in context
	ctx = withUserContext(context.Background(), teacher)
	_, err = ags.UpdateGroup(ctx, updateGroupReq)
	if err != nil {
		t.Error(err)
	}

	// check that the group have changed group membership
	haveGroup, err := db.GetGroup(group.ID)
	if err != nil {
		t.Fatal(err)
	}
	userIDs := make([]uint64, 0)
	for _, usr := range haveGroup.Users {
		userIDs = append(userIDs, usr.ID)
	}

	grpUsers, err := db.GetUsers(userIDs...)
	if err != nil {
		t.Fatal(err)
	}

	wantGroup := group
	wantGroup.Name = updateGroupReq.Name
	wantGroup.Users = grpUsers
	wantGroup.TeamID = 1
	// UpdateGroup will autoApprove group on update
	wantGroup.Status = pb.Group_APPROVED
	haveGroup.Enrollments = nil
	wantGroup.Enrollments = nil
	if !cmp.Equal(haveGroup, wantGroup) {
		t.Errorf("have group %+v", haveGroup)
		t.Errorf("want group %+v", wantGroup)
	}

	// ******************* Teacher UpdateGroup *******************

	// change group to only one student
	users2 := make([]*pb.User, 0)
	users2 = append(users2, user1)
	// name must not update because group team and repo already exist
	updateGroupReq1 := &pb.Group{ID: group.ID, Name: "Hein's single member Group", CourseID: course.ID, Users: users2}

	// set teacher ID in context
	ctx = withUserContext(context.Background(), teacher)
	_, err = ags.UpdateGroup(ctx, updateGroupReq1)
	if err != nil {
		t.Error(err)
	}
	// check that the group have changed group membership
	haveGroup, err = db.GetGroup(group.ID)
	if err != nil {
		t.Fatal(err)
	}
	userIDs = make([]uint64, 0)
	for _, usr := range updateGroupReq1.Users {
		userIDs = append(userIDs, usr.ID)
	}

	grpUsers, err = db.GetUsers(userIDs...)
	if err != nil {
		t.Fatal(err)
	}

	if len(haveGroup.Users) != 1 {
		t.Fatal("expected only single member group")
	}
	wantGroup = updateGroupReq
	wantGroup.Users = grpUsers
	wantGroup.TeamID = 1
	// UpdateGroup will autoApprove group on update
	wantGroup.Status = pb.Group_APPROVED
	haveGroup.Enrollments = nil
	wantGroup.Enrollments = nil
	if !reflect.DeepEqual(wantGroup, haveGroup) {
		t.Errorf("have group %+v", haveGroup)
		t.Errorf("want group %+v", wantGroup)
	}
}

func TestDeleteGroup(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	testCourse := pb.Course{
		Name:           "Distributed Systems",
		Code:           "DAT520",
		Year:           2018,
		Tag:            "Spring",
		Provider:       "fake",
		OrganizationID: 1,
		ID:             1,
	}
	admin := createFakeUser(t, db, 1)
	if err := db.CreateCourse(admin.ID, &testCourse); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as student
	user := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user.ID, CourseID: testCourse.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: testCourse.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	// create teacher and enroll as teacher
	teacher := createFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: teacher.ID, CourseID: testCourse.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   teacher.ID,
		CourseID: testCourse.ID,
		Status:   pb.Enrollment_TEACHER,
	}); err != nil {
		t.Fatal(err)
	}

	group := &pb.Group{Name: "Test Delete Group", CourseID: testCourse.ID, Users: []*pb.User{user}}

	_, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})

	ctx := withUserContext(context.Background(), user)
	respGroup, err := ags.CreateGroup(ctx, group)
	if err != nil {
		t.Fatal(err)
	}

	ctx = withUserContext(context.Background(), teacher)
	_, err = ags.DeleteGroup(ctx, &pb.GroupRequest{GroupID: respGroup.ID, CourseID: testCourse.ID})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetGroup(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	testCourse := pb.Course{
		Name:           "Distributed Systems",
		Code:           "DAT520",
		Year:           2018,
		Tag:            "Spring",
		Provider:       "fake",
		OrganizationID: 1,
	}
	admin := createFakeUser(t, db, 1)
	if err := db.CreateCourse(admin.ID, &testCourse); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as student
	user := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: user.ID, CourseID: testCourse.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: testCourse.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	_, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	ctx := withUserContext(context.Background(), user)

	group := &pb.Group{Name: "Test Group", CourseID: testCourse.ID, Users: []*pb.User{user}}
	respGroup, err := ags.CreateGroup(ctx, group)
	if err != nil {
		t.Fatal(err)
	}

	gotGroup, err := ags.GetGroup(ctx, &pb.GetGroupRequest{GroupID: respGroup.ID})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(gotGroup, respGroup) {
		t.Errorf("have response group %+v, while database has %+v", &gotGroup, &respGroup)
	}
}

func TestPatchGroupStatus(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	course := pb.Course{
		Name:           "Distributed Systems",
		Code:           "DAT520",
		Year:           2018,
		Tag:            "Spring",
		Provider:       "fake",
		OrganizationID: 1,
		ID:             1,
	}

	admin := createFakeUser(t, db, 1)
	err := db.CreateCourse(admin.ID, &course)
	if err != nil {
		t.Fatal(err)
	}

	teacher := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: teacher.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   teacher.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_TEACHER,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateUser(&pb.User{ID: teacher.ID, IsAdmin: true}); err != nil {
		t.Fatal(err)
	}

	fakeProvider, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	ctx := withUserContext(context.Background(), teacher)

	if _, err := fakeProvider.CreateOrganization(ctx, &scm.OrganizationOptions{
		Name: course.Code,
		Path: course.Code,
	}); err != nil {
		t.Fatal(err)
	}

	user1 := createFakeUser(t, db, 3)
	user2 := createFakeUser(t, db, 4)

	// enroll users in course and group
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID: user1.ID, CourseID: course.ID, GroupID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user1.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID: user2.ID, CourseID: course.ID, GroupID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user2.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	group := &pb.Group{
		ID:       1,
		Name:     "Test Group",
		CourseID: course.ID,
		Users:    []*pb.User{user1, user2},
		TeamID:   1,
	}
	err = db.CreateGroup(group)
	if err != nil {
		t.Fatal(err)
	}
	// get the group as stored in db with enrollments
	prePatchGroup, err := db.GetGroup(group.ID)
	if err != nil {
		t.Fatal(err)
	}

	prePatchGroup.Status = pb.Group_APPROVED
	_, err = ags.UpdateGroup(ctx, prePatchGroup)
	if err != nil {
		t.Error(err)
	}

	// check that the group didn't change
	haveGroup, err := db.GetGroup(group.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(prePatchGroup, haveGroup) {
		t.Errorf("have\n%+v\nwant\n%+v\n", haveGroup, prePatchGroup)
	}
}

func TestGetGroupByUserAndCourse(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	course := pb.Course{
		Name:           "Distributed Systems",
		Code:           "DAT520",
		Year:           2018,
		Tag:            "Spring",
		Provider:       "fake",
		OrganizationID: 1,
		ID:             1,
	}

	admin := createFakeUser(t, db, 1)
	err := db.CreateCourse(admin.ID, &course)
	if err != nil {
		t.Fatal(err)
	}

	_, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	ctx := withUserContext(context.Background(), admin)

	user1 := createFakeUser(t, db, 2)
	user2 := createFakeUser(t, db, 3)

	// enroll users in course and group
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID: user1.ID, CourseID: course.ID, GroupID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user1.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID: user2.ID, CourseID: course.ID, GroupID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user2.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	group := &pb.Group{
		ID:       1,
		CourseID: course.ID,
		Users:    []*pb.User{user1, user2},
	}
	err = db.CreateGroup(group)
	if err != nil {
		t.Fatal(err)
	}

	respGroup, err := ags.GetGroupByUserAndCourse(ctx, &pb.GroupRequest{UserID: user1.ID, CourseID: course.ID})
	if err != nil {
		t.Error(err)
	}

	dbGroup, err := ags.GetGroup(ctx, &pb.GetGroupRequest{GroupID: group.ID})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(respGroup, dbGroup) {
		t.Errorf("have response group %+v, while database has %+v", respGroup, dbGroup)
	}
}

func TestDeleteApprovedGroup(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	course := allCourses[0]
	err := db.CreateCourse(admin.ID, course)
	if err != nil {
		t.Fatal(err)
	}

	fakeProvider, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	ctx := withUserContext(context.Background(), admin)

	if _, err := fakeProvider.CreateOrganization(ctx, &scm.OrganizationOptions{
		Name: course.Code,
		Path: course.Code,
	}); err != nil {
		t.Fatal(err)
	}

	user1 := createFakeUser(t, db, 2)
	user2 := createFakeUser(t, db, 3)

	// enroll users in course and group
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID: user1.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user1.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserID: user2.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user2.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   admin.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_TEACHER,
	}); err != nil {
		t.Fatal(err)
	}

	group := &pb.Group{
		ID:       1,
		CourseID: course.ID,
		Name:     "Test Group",
		Users:    []*pb.User{user1, user2},
	}
	// current user1 (in context) must be in group being created
	ctx = withUserContext(context.Background(), user1)
	createdGroup, err := ags.CreateGroup(ctx, group)
	if err != nil {
		t.Fatal(err)
	}

	// first approve the group
	createdGroup.Status = pb.Group_APPROVED
	// current user (in context) must be teacher for the course
	ctx = withUserContext(context.Background(), admin)
	if _, err = ags.UpdateGroup(ctx, createdGroup); err != nil {
		t.Fatal(err)
	}

	// then get user enrollments with group ID
	enr1, err := db.GetEnrollmentByCourseAndUser(course.ID, user1.ID)
	if err != nil {
		t.Fatal(err)
	}
	enr2, err := db.GetEnrollmentByCourseAndUser(course.ID, user2.ID)
	if err != nil {
		t.Fatal(err)
	}

	// delete the group
	if _, err = ags.DeleteGroup(ctx, &pb.GroupRequest{CourseID: course.ID, GroupID: createdGroup.ID}); err != nil {
		t.Fatal(err)
	}

	// get updated enrollments of group members
	newEnr1, err := db.GetEnrollmentByCourseAndUser(course.ID, user1.ID)
	if err != nil {
		t.Fatal(err)
	}
	newEnr2, err := db.GetEnrollmentByCourseAndUser(course.ID, user2.ID)
	if err != nil {
		t.Fatal(err)
	}

	// now nullify manually group ID for original enrollments
	enr1.GroupID = 0
	enr2.GroupID = 0

	// then check that new enrollments have group IDs nullified automatically
	if !reflect.DeepEqual(enr1, newEnr1) {
		t.Errorf("want enrollment %+v, while database has %+v", enr1, newEnr1)
	}
	if !reflect.DeepEqual(enr2, newEnr2) {
		t.Errorf("want enrollment %+v, while database has %+v", enr2, newEnr2)
	}
}

func TestGetGroups(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	var users []*pb.User
	for _, u := range allUsers {
		user := createFakeUser(t, db, u.remoteID)
		users = append(users, user)
	}
	admin := users[0]

	_, scms := fakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	// admin will be enrolled as teacher because of course creation below
	withUserContext(context.Background(), admin)

	course := allCourses[1]
	err := db.CreateCourse(admin.ID, course)
	if err != nil {
		t.Fatal(err)
	}

	// enroll all users in course
	for _, user := range users[1:] {
		if err := db.CreateEnrollment(&pb.Enrollment{
			UserID: user.ID, CourseID: course.ID}); err != nil {
			t.Fatal(err)
		}
		if err := db.UpdateEnrollment(&pb.Enrollment{
			UserID:   user.ID,
			CourseID: course.ID,
			Status:   pb.Enrollment_STUDENT,
		}); err != nil {
			t.Fatal(err)
		}
	}
	// place some students in groups
	// current user (in context) must be in group being created
	ctx := withUserContext(context.Background(), users[2])
	group1, err := ags.CreateGroup(ctx, &pb.Group{Name: "Group 1", CourseID: course.ID, Users: []*pb.User{users[1], users[2]}})
	if err != nil {
		t.Fatal(err)
	}
	ctx = withUserContext(context.Background(), users[5])
	group2, err := ags.CreateGroup(ctx, &pb.Group{Name: "Group 2", CourseID: course.ID, Users: []*pb.User{users[4], users[5]}})
	if err != nil {
		t.Fatal(err)
	}
	wantGroups := &pb.Groups{Groups: []*pb.Group{group1, group2}}
	for _, grp := range wantGroups.Groups {
		for _, grpEnrol := range grp.Enrollments {
			grpEnrol.UsedSlipDays = []*pb.UsedSlipDays{}
		}
	}

	// check that request on non-existent course returns error
	_, err = ags.GetGroupsByCourse(ctx, &pb.CourseRequest{CourseID: 15})
	if err == nil {
		t.Error("expected error; no groups should be returned")
	}

	// get groups from the database; admin is in ctx, which is also teacher
	ctx = withUserContext(context.Background(), admin)
	gotGroups, err := ags.GetGroupsByCourse(ctx, &pb.CourseRequest{CourseID: course.ID})
	if err != nil {
		t.Fatal(err)
	}

	// check that the method returns expected groups
	if diff := cmp.Diff(wantGroups, gotGroups); diff != "" {
		t.Errorf("mismatch (-wantGroups +gotGroups):\n%s", diff)
	}
}
