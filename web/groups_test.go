package web_test

import (
	"context"
	"reflect"
	"testing"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
	"github.com/autograde/aguis/web"
	"github.com/autograde/aguis/web/grpc_service"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc/metadata"
)

func TestDeleteGroup(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	testCourse := pb.Course{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryId: 1,
	}
	admin := createFakeUser(t, db, 1)
	if err := db.CreateCourse(admin.Id, &testCourse); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as student
	user := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserId: user.Id, CourseId: testCourse.Id}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.Id, testCourse.Id); err != nil {
		t.Fatal(err)
	}

	group := &pb.Group{CourseId: testCourse.Id}

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(user.Id))

	respGroup, err := test_ag.CreateGroup(cont, group)
	if err != nil {
		t.Fatal(err)
	}

	_, err = test_ag.DeleteGroup(cont, respGroup)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetGroup(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	testCourse := pb.Course{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryId: 1,
	}
	admin := createFakeUser(t, db, 1)
	if err := db.CreateCourse(admin.Id, &testCourse); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as student
	user := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserId: user.Id, CourseId: testCourse.Id}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.Id, testCourse.Id); err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(user.Id))

	group := &pb.Group{CourseId: testCourse.Id}
	respGroup, err := test_ag.CreateGroup(cont, group)
	if err != nil {
		t.Fatal(err)
	}

	gotGroup, err := test_ag.GetGroup(cont, &pb.RecordRequest{Id: respGroup.Id})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(respGroup, group) {
		t.Errorf("have response group %+v, while database has %+v", &respGroup, group)
	}

	if !reflect.DeepEqual(gotGroup, respGroup) {
		t.Errorf("have response group %+v, while database has %+v", &gotGroup, &respGroup)
	}

}

func TestPatchGroupStatus(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	course := pb.Course{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryId: 1,
		Id:          1,
	}

	admin := createFakeUser(t, db, 1)
	err := db.CreateCourse(admin.Id, &course)
	if err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(admin.Id))

	f := scm.NewFakeSCMClient()
	if _, err := f.CreateDirectory(context.Background(), &scm.CreateDirectoryOptions{
		Name: course.Code,
		Path: course.Code,
	}); err != nil {
		t.Fatal(err)
	}
	testscms["token"] = f

	user1 := createFakeUser(t, db, 2)
	user2 := createFakeUser(t, db, 3)

	// enroll users in course and group
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserId: user1.Id, CourseId: course.Id, GroupId: 1}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user1.Id, course.Id); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserId: user2.Id, CourseId: course.Id, GroupId: 1}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user2.Id, course.Id); err != nil {
		t.Fatal(err)
	}

	group := &pb.Group{
		Id:       1,
		CourseId: course.Id,
		Users:    []*pb.User{user1, user2},
	}
	err = db.CreateGroup(group)
	if err != nil {
		t.Fatal(err)
	}
	// get the group as stored in db with enrollments
	prePatchGroup, err := db.GetGroup(group.Id)
	if err != nil {
		t.Fatal(err)
	}

	_, err = test_ag.UpdateGroupStatus(cont, &pb.Group{Id: prePatchGroup.Id, CourseId: prePatchGroup.CourseId, Users: prePatchGroup.Users, Status: pb.Group_APPROVED})
	if err != nil {
		t.Error(err)
	}

	// check that the group didn't change
	haveGroup, err := db.GetGroup(group.Id)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(prePatchGroup, haveGroup) {
		t.Errorf("have group %+v want %+v", haveGroup, prePatchGroup)
	}

	wantGroup := prePatchGroup
	wantGroup.Status = 3
	if !reflect.DeepEqual(wantGroup, haveGroup) {
		t.Errorf("have group %+v want %+v", haveGroup, wantGroup)
	}
}

func TestGetGroupByUserAndCourse(t *testing.T) {

	db, cleanup := setup(t)
	defer cleanup()

	course := pb.Course{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryId: 1,
		Id:          1,
	}

	admin := createFakeUser(t, db, 1)
	err := db.CreateCourse(admin.Id, &course)
	if err != nil {
		t.Fatal(err)
	}

	testscms := make(map[string]scm.SCM)
	test_ag := grpc_service.NewAutograderService(db, testscms, web.BaseHookOptions{})
	cont := metadata.AppendToOutgoingContext(context.Background(), "user", string(admin.Id))

	user1 := createFakeUser(t, db, 2)
	user2 := createFakeUser(t, db, 3)

	// enroll users in course and group
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserId: user1.Id, CourseId: course.Id, GroupId: 1}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user1.Id, course.Id); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&pb.Enrollment{
		UserId: user2.Id, CourseId: course.Id, GroupId: 1}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user2.Id, course.Id); err != nil {
		t.Fatal(err)
	}

	group := &pb.Group{
		Id:       1,
		CourseId: course.Id,
		Users:    []*pb.User{user1, user2},
	}
	err = db.CreateGroup(group)
	if err != nil {
		t.Fatal(err)
	}

	respGroup, err := test_ag.GetGroupByUserAndCourse(cont, &pb.ActionRequest{UserId: user1.Id, CourseId: course.Id})
	if err != nil {
		t.Error(err)
	}

	dbGroup, err := db.GetGroup(group.Id)
	if err != nil {
		t.Fatal(err)
	}
	// see pb.Group; enrollment field is not transmitted over http
	// we simply ignore enrollments
	//dbGroup.Enrollments = nil

	if !reflect.DeepEqual(&respGroup, dbGroup) {
		t.Errorf("have response group %+v, while database has %+v", &respGroup, dbGroup)
	}
}
