package web_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/scm"
	"github.com/autograde/aguis/web"
	"github.com/autograde/aguis/web/auth"
	"github.com/labstack/echo"
	_ "github.com/mattn/go-sqlite3"
)

func TestNewGroup(t *testing.T) {
	const route = "/courses/:cid/groups"

	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	var course models.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.DirectoryID = 1
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}
	user := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&models.Enrollment{UserID: user.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	// Only single member group for now.
	newGroupReq := web.NewGroupRequest{
		Name:     "Hein's Group",
		CourseID: course.ID,
		UserIDs:  []uint64{user.ID},
	}
	b, err := json.Marshal(newGroupReq)
	if err != nil {
		t.Fatal(err)
	}
	requestBody := bytes.NewReader(b)

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodPost, route, web.NewGroup(db))

	requestURL := fmt.Sprintf("/courses/%d/groups", course.ID)
	r := httptest.NewRequest(http.MethodPost, requestURL, requestBody)
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)

	// Prepare provider
	fakeProvider, err := scm.NewSCMClient("fake", "token")
	if err != nil {
		t.Fatal(err)
	}
	fakeProvider.CreateDirectory(c.Request().Context(),
		&scm.CreateDirectoryOptions{Path: "path", Name: "name"},
	)
	c.Set("fake", fakeProvider)

	// Prepare context with user request.
	c.Set("user", user)
	router.Find(http.MethodPost, requestURL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}
	assertCode(t, w.Code, http.StatusCreated)

	var respGroup models.Group
	if err := json.Unmarshal(w.Body.Bytes(), &respGroup); err != nil {
		t.Fatal(err)
	}

	group, err := db.GetGroup(false, respGroup.ID)
	if err != nil {
		t.Fatal(err)
	}

	// JSON marshalling removes the enrollment field from respGroup,
	// so we remove group.Enrollments obtained from the database before comparing.
	group.Enrollments = nil
	if !reflect.DeepEqual(&respGroup, group) {
		t.Errorf("have response group %+v, while database has %+v", &respGroup, group)
	}
}

func TestNewGroupTeacherCreator(t *testing.T) {
	const route = "/courses/:cid/groups"

	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	var course models.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.DirectoryID = 1
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}

	teacher := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&models.Enrollment{UserID: teacher.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollTeacher(teacher.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	user := createFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&models.Enrollment{UserID: user.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	// Only single member group for now.
	newGroupReq := web.NewGroupRequest{
		Name:     "Hein's Group",
		CourseID: course.ID,
		UserIDs:  []uint64{user.ID},
	}
	b, err := json.Marshal(newGroupReq)
	if err != nil {
		t.Fatal(err)
	}
	requestBody := bytes.NewReader(b)

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodPost, route, web.NewGroup(db))

	requestURL := fmt.Sprintf("/courses/%d/groups", course.ID)
	r := httptest.NewRequest(http.MethodPost, requestURL, requestBody)
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)

	// Prepare provider
	fakeProvider, err := scm.NewSCMClient("fake", "token")
	if err != nil {
		t.Fatal(err)
	}
	fakeProvider.CreateDirectory(c.Request().Context(),
		&scm.CreateDirectoryOptions{Path: "path", Name: "name"},
	)
	c.Set("fake", fakeProvider)

	// Prepare context with user request.
	c.Set("user", teacher)
	router.Find(http.MethodPost, requestURL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}
	assertCode(t, w.Code, http.StatusCreated)

	var respGroup models.Group
	if err := json.Unmarshal(w.Body.Bytes(), &respGroup); err != nil {
		t.Fatal(err)
	}

	group, err := db.GetGroup(false, respGroup.ID)
	if err != nil {
		t.Fatal(err)
	}

	// JSON marshalling removes the enrollment field from respGroup,
	// so we remove group.Enrollments obtained from the database before comparing.
	group.Enrollments = nil
	if !reflect.DeepEqual(&respGroup, group) {
		t.Errorf("have response group %+v, while database has %+v", &respGroup, group)
	}
}

func TestNewGroupStudentCreateGroupWithTeacher(t *testing.T) {
	const route = "/courses/:cid/groups"

	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	var course models.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.DirectoryID = 1
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}

	teacher := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&models.Enrollment{UserID: teacher.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollTeacher(teacher.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	user := createFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&models.Enrollment{UserID: user.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	// group with teacher and student; this should fail
	newGroupReq := web.NewGroupRequest{
		Name:     "Hein's Group",
		CourseID: course.ID,
		UserIDs:  []uint64{user.ID, teacher.ID},
	}
	b, err := json.Marshal(newGroupReq)
	if err != nil {
		t.Fatal(err)
	}
	requestBody := bytes.NewReader(b)

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodPost, route, web.NewGroup(db))

	requestURL := fmt.Sprintf("/courses/%d/groups", course.ID)
	r := httptest.NewRequest(http.MethodPost, requestURL, requestBody)
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)

	// Prepare provider
	fakeProvider, err := scm.NewSCMClient("fake", "token")
	if err != nil {
		t.Fatal(err)
	}
	fakeProvider.CreateDirectory(c.Request().Context(),
		&scm.CreateDirectoryOptions{Path: "path", Name: "name"},
	)
	c.Set("fake", fakeProvider)

	// Prepare context with user request.
	c.Set("user", user)
	router.Find(http.MethodPost, requestURL, c)

	// Invoke the prepared handler.
	// We want error to occure, as this is a bad request.
	if err := c.Handler()(c); err == nil {
		t.Error("Student trying to enroll teacher should not be possible!")
	}
}

func TestStudentCreateNewGroupTeacherUpdateGroup(t *testing.T) {
	const (
		newGrpRoute    = "/courses/:cid/groups"
		updateGrpRoute = "/courses/:cid/groups/:gid"
		patchGrpRoute  = "/groups/:gid"
	)

	db, cleanup := setup(t)
	defer cleanup()

	admin := createFakeUser(t, db, 1)
	var course models.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.DirectoryID = 1
	if err := db.CreateCourse(admin.ID, &course); err != nil {
		t.Fatal(err)
	}

	teacher := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&models.Enrollment{UserID: teacher.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollTeacher(teacher.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	user1 := createFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&models.Enrollment{UserID: user1.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user1.ID, course.ID); err != nil {
		t.Fatal(err)
	}
	user2 := createFakeUser(t, db, 4)
	if err := db.CreateEnrollment(&models.Enrollment{UserID: user2.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user2.ID, course.ID); err != nil {
		t.Fatal(err)
	}
	user3 := createFakeUser(t, db, 5)
	if err := db.CreateEnrollment(&models.Enrollment{UserID: user3.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user3.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	// group with two students
	newGroupReq := web.NewGroupRequest{
		Name:     "Hein's two member Group",
		CourseID: course.ID,
		UserIDs:  []uint64{user1.ID, user2.ID},
	}
	b, err := json.Marshal(newGroupReq)
	if err != nil {
		t.Fatal(err)
	}
	requestBody := bytes.NewReader(b)

	e := echo.New()
	router := echo.NewRouter(e)

	// add NewGroup route to handler.
	router.Add(http.MethodPost, newGrpRoute, web.NewGroup(db))

	requestURL := fmt.Sprintf("/courses/%d/groups", course.ID)
	r := httptest.NewRequest(http.MethodPost, requestURL, requestBody)
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)

	fakeProvider, err := scm.NewSCMClient("fake", "token")
	if err != nil {
		t.Fatal(err)
	}
	fakeProvider.CreateDirectory(c.Request().Context(),
		&scm.CreateDirectoryOptions{Path: "path", Name: "name"},
	)
	// prepare context with user3, which is not member of group (should fail)
	c.Set("fake", fakeProvider)
	c.Set(auth.UserKey, user3)
	router.Find(http.MethodPost, requestURL, c)
	if err := c.Handler()(c); err == nil {
		t.Error("expected error 'code=400, message=student must be member of new group'")
	}

	// try again with user1, which is member of group
	requestBody = bytes.NewReader(b)
	r = httptest.NewRequest(http.MethodPost, requestURL, requestBody)
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w = httptest.NewRecorder()
	c.Reset(r, w)
	c.Set("fake", fakeProvider)
	c.Set(auth.UserKey, user1)
	router.Find(http.MethodPost, requestURL, c)
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}
	assertCode(t, w.Code, http.StatusCreated)

	var respGroup models.Group
	if err := json.Unmarshal(w.Body.Bytes(), &respGroup); err != nil {
		t.Fatal(err)
	}

	group, err := db.GetGroup(false, respGroup.ID)
	if err != nil {
		t.Fatal(err)
	}
	// JSON marshalling removes the enrollment field from respGroup,
	// so we remove group.Enrollments obtained from the database before comparing.
	group.Enrollments = nil
	if !reflect.DeepEqual(&respGroup, group) {
		t.Errorf("have response group %+v, while database has %+v", &respGroup, group)
	}

	// ******************* Admin/Teacher UpdateGroup *******************

	// group with three students
	updateGroupReq := web.NewGroupRequest{
		Name:     "Hein's three member group",
		CourseID: course.ID,
		UserIDs:  []uint64{user1.ID, user2.ID, user3.ID},
	}
	b, err = json.Marshal(updateGroupReq)
	if err != nil {
		t.Fatal(err)
	}
	requestBody = bytes.NewReader(b)

	// add UpdateGroup route to handler.
	router.Add(http.MethodPut, updateGrpRoute, web.UpdateGroup(nullLogger(), db))

	requestURL = fmt.Sprintf("/courses/%d/groups/%d", course.ID, group.ID)
	r = httptest.NewRequest(http.MethodPut, requestURL, requestBody)
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w = httptest.NewRecorder()
	c = e.NewContext(r, w) // Note: c.Reset(r, w) doesn't work here
	c.Set("fake", fakeProvider)
	// set admin as the user for this context
	c.Set(auth.UserKey, admin)
	router.Find(http.MethodPut, requestURL, c)
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}
	assertCode(t, w.Code, http.StatusOK)

	// check that the group have changed group membership
	haveGroup, err := db.GetGroup(true, group.ID)
	if err != nil {
		t.Fatal(err)
	}
	grpUsers, err := db.GetUsers(true, updateGroupReq.UserIDs...)
	if err != nil {
		t.Fatal(err)
	}
	wantGroup := group
	wantGroup.Name = updateGroupReq.Name
	wantGroup.Users = grpUsers
	haveGroup.Enrollments = nil
	if !reflect.DeepEqual(wantGroup, haveGroup) {
		t.Errorf("have group %+v", haveGroup)
		t.Errorf("want group %+v", wantGroup)
	}

	// ******************* Teacher Only UpdateGroup *******************

	// change group to only one student
	updateGroupReq = web.NewGroupRequest{
		Name:     "Hein's single member group",
		CourseID: course.ID,
		UserIDs:  []uint64{user1.ID},
	}
	b, err = json.Marshal(updateGroupReq)
	if err != nil {
		t.Fatal(err)
	}
	requestBody = bytes.NewReader(b)

	// add UpdateGroup route to handler.
	router.Add(http.MethodPut, updateGrpRoute, web.UpdateGroup(nullLogger(), db))

	requestURL = fmt.Sprintf("/courses/%d/groups/%d", course.ID, group.ID)
	r = httptest.NewRequest(http.MethodPut, requestURL, requestBody)
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w = httptest.NewRecorder()
	c = e.NewContext(r, w) // Note: c.Reset(r, w) doesn't work here
	c.Set("fake", fakeProvider)
	// set teacher as the user for this context
	c.Set(auth.UserKey, teacher)
	router.Find(http.MethodPut, requestURL, c)
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}
	assertCode(t, w.Code, http.StatusOK)

	// check that the group have changed group membership
	haveGroup, err = db.GetGroup(true, group.ID)
	if err != nil {
		t.Fatal(err)
	}
	grpUsers, err = db.GetUsers(true, updateGroupReq.UserIDs...)
	if err != nil {
		t.Fatal(err)
	}
	if len(haveGroup.Users) != 1 {
		t.Fatal("expected only single member group")
	}
	wantGroup = group
	wantGroup.Name = updateGroupReq.Name
	wantGroup.Users = grpUsers
	haveGroup.Enrollments = nil
	if !reflect.DeepEqual(wantGroup, haveGroup) {
		t.Errorf("have group %+v", haveGroup)
		t.Errorf("want group %+v", wantGroup)
	}
}

func TestDeleteGroup(t *testing.T) {
	const route = "/groups/:gid"

	db, cleanup := setup(t)
	defer cleanup()

	testCourse := models.Course{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryID: 1,
	}
	admin := createFakeUser(t, db, 1)
	if err := db.CreateCourse(admin.ID, &testCourse); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as student
	user := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&models.Enrollment{UserID: user.ID, CourseID: testCourse.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, testCourse.ID); err != nil {
		t.Fatal(err)
	}

	group := models.Group{CourseID: testCourse.ID}
	if err := db.CreateGroup(&group); err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodDelete, route, web.DeleteGroup(db))

	requestURL := fmt.Sprintf("/groups/%d", group.ID)
	r := httptest.NewRequest(http.MethodDelete, requestURL, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	// Prepare context with course request.
	router.Find(http.MethodDelete, requestURL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Fatal(err)
	}

	assertCode(t, w.Code, http.StatusOK)
}

func TestGetGroup(t *testing.T) {
	const route = "/groups/:gid"

	db, cleanup := setup(t)
	defer cleanup()

	testCourse := models.Course{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryID: 1,
	}
	admin := createFakeUser(t, db, 1)
	if err := db.CreateCourse(admin.ID, &testCourse); err != nil {
		t.Fatal(err)
	}

	// create user and enroll as student
	user := createFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&models.Enrollment{UserID: user.ID, CourseID: testCourse.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user.ID, testCourse.ID); err != nil {
		t.Fatal(err)
	}

	group := models.Group{CourseID: testCourse.ID}
	if err := db.CreateGroup(&group); err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	router := echo.NewRouter(e)

	// Add the route to handler.
	router.Add(http.MethodDelete, route, web.GetGroup(db))

	requestURL := fmt.Sprintf("/groups/%d", group.ID)
	r := httptest.NewRequest(http.MethodGet, requestURL, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	// Prepare context with course request.
	router.Find(http.MethodDelete, requestURL, c)

	// Invoke the prepared handler.
	if err := c.Handler()(c); err != nil {
		t.Fatal(err)
	}
	assertCode(t, w.Code, http.StatusOK)

	var respGroup models.Group
	if err := json.Unmarshal(w.Body.Bytes(), &respGroup); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(respGroup, group) {
		t.Errorf("have response group %+v, while database has %+v", &respGroup, group)
	}
}

func TestPatchGroupStatus(t *testing.T) {
	const route = "/groups/:gid"

	db, cleanup := setup(t)
	defer cleanup()

	course := models.Course{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryID: 1,
		ID:          1,
	}

	admin := createFakeUser(t, db, 1)
	err := db.CreateCourse(admin.ID, &course)
	if err != nil {
		t.Fatal(err)
	}

	user1 := createFakeUser(t, db, 2)
	user2 := createFakeUser(t, db, 3)

	// enroll users in course and group
	if err := db.CreateEnrollment(&models.Enrollment{
		UserID: user1.ID, CourseID: course.ID, GroupID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user1.ID, course.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&models.Enrollment{
		UserID: user2.ID, CourseID: course.ID, GroupID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user2.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	group := &models.Group{
		ID:       1,
		CourseID: course.ID,
		Users:    []*models.User{user1, user2},
	}
	err = db.CreateGroup(group)
	if err != nil {
		t.Fatal(err)
	}
	// get the group as stored in db with enrollments
	prePatchGroup, err := db.GetGroup(true, group.ID)
	if err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	router := echo.NewRouter(e)

	// add the route to handler.
	router.Add(http.MethodPatch, route, web.PatchGroup(nullLogger(), db))

	// send empty request, the user should not be modified.
	emptyJSON, err := json.Marshal(&web.UpdateGroupRequest{})
	if err != nil {
		t.Fatal(err)
	}
	requestBody := bytes.NewReader(emptyJSON)

	requestURL := fmt.Sprintf("/groups/%d", group.ID)
	r := httptest.NewRequest(http.MethodPatch, requestURL, requestBody)
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	f := scm.NewFakeSCMClient()
	if _, err := f.CreateDirectory(context.Background(), &scm.CreateDirectoryOptions{
		Name: course.Code,
		Path: course.Code,
	}); err != nil {
		t.Fatal(err)
	}
	c.Set("fake", f)
	// set admin as the user for this context
	c.Set("user", admin)
	router.Find(http.MethodPatch, requestURL, c)

	// invoke the prepared handler
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}
	assertCode(t, w.Code, http.StatusOK)

	// check that the group didn't change
	haveGroup, err := db.GetGroup(true, group.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(prePatchGroup, haveGroup) {
		t.Errorf("have group %+v want %+v", haveGroup, prePatchGroup)
	}

	// send request for status change of the group
	trueJSON, err := json.Marshal(&web.UpdateGroupRequest{Status: 3})
	if err != nil {
		t.Fatal(err)
	}
	requestBody.Reset(trueJSON)

	r = httptest.NewRequest(http.MethodPatch, requestURL, requestBody)
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w = httptest.NewRecorder()
	c.Reset(r, w)
	// set admin as the user for this context
	c.Set("user", admin)
	fakeProvider, err := scm.NewSCMClient("fake", "token")
	if err != nil {
		t.Fatal(err)
	}
	fakeProvider.CreateDirectory(c.Request().Context(),
		&scm.CreateDirectoryOptions{Path: "path", Name: "name"},
	)
	c.Set("fake", fakeProvider)
	router.Find(http.MethodPatch, requestURL, c)

	// invoke the prepared handler
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}
	assertCode(t, w.Code, http.StatusOK)

	// check that the group have changed status
	haveGroup, err = db.GetGroup(true, group.ID)
	if err != nil {
		t.Fatal(err)
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

	course := models.Course{
		Name:        "Distributed Systems",
		Code:        "DAT520",
		Year:        2018,
		Tag:         "Spring",
		Provider:    "fake",
		DirectoryID: 1,
		ID:          1,
	}

	admin := createFakeUser(t, db, 1)
	err := db.CreateCourse(admin.ID, &course)
	if err != nil {
		t.Fatal(err)
	}

	user1 := createFakeUser(t, db, 2)
	user2 := createFakeUser(t, db, 3)

	// enroll users in course and group
	if err := db.CreateEnrollment(&models.Enrollment{
		UserID: user1.ID, CourseID: course.ID, GroupID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user1.ID, course.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateEnrollment(&models.Enrollment{
		UserID: user2.ID, CourseID: course.ID, GroupID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := db.EnrollStudent(user2.ID, course.ID); err != nil {
		t.Fatal(err)
	}

	group := &models.Group{
		ID:       1,
		CourseID: course.ID,
		Users:    []*models.User{user1, user2},
	}
	err = db.CreateGroup(group)
	if err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	router := echo.NewRouter(e)
	const route = "/users/:uid/courses/:cid/group"
	router.Add(http.MethodGet, route, web.GetGroupByUserAndCourse(db))
	// add the route to handler
	requestURL := fmt.Sprintf("/users/%d/courses/%d/group", user1.ID, course.ID)
	r := httptest.NewRequest(http.MethodGet, requestURL, nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	router.Find(http.MethodGet, requestURL, c)
	// invoke the prepared handler
	if err := c.Handler()(c); err != nil {
		t.Error(err)
	}
	assertCode(t, w.Code, http.StatusFound)

	var respGroup models.Group
	if err := json.Unmarshal(w.Body.Bytes(), &respGroup); err != nil {
		t.Fatal(err)
	}

	// we don't expect remote identities from GetGroupByUserAndCourse
	dbGroup, err := db.GetGroup(false, group.ID)
	if err != nil {
		t.Fatal(err)
	}
	// see models.Group; enrollment field is not transmitted over http
	// we simply ignore enrollments
	dbGroup.Enrollments = nil

	if !reflect.DeepEqual(&respGroup, dbGroup) {
		t.Errorf("have response group %+v, while database has %+v", &respGroup, dbGroup)
	}
}
