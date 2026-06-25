package database_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

var createBunGroupTests = []struct {
	name        string
	desc        string
	getGroup    func(uint64, ...uint64) *qf.Group
	enrollments []qf.Enrollment_UserStatus
	err         error
}{
	{
		name: "course id not set with users",
		desc: "Should fail with sql.ErrNoRows; cannot create a group not connected to a course.",
		getGroup: func(_ uint64, userIDs ...uint64) *qf.Group {
			var users []*qf.User
			for _, uid := range userIDs {
				users = append(users, &qf.User{ID: uid})
			}
			return &qf.Group{Users: users}
		},
		enrollments: []qf.Enrollment_UserStatus{qf.Enrollment_PENDING, qf.Enrollment_PENDING},
		err:         sql.ErrNoRows,
	},
	{
		name: "course not found with users",
		desc: "Should fail with sql.ErrNoRows; cannot create a group not connected to a course.",
		getGroup: func(_ uint64, userIDs ...uint64) *qf.Group {
			var users []*qf.User
			for _, uid := range userIDs {
				users = append(users, &qf.User{ID: uid})
			}
			return &qf.Group{CourseID: 999, Users: users}
		},
		enrollments: []qf.Enrollment_UserStatus{qf.Enrollment_PENDING, qf.Enrollment_PENDING},
		err:         sql.ErrNoRows,
	},
	{
		name: "course found but without users",
		desc: "Should fail with ErrEmptyGroup; cannot create a group without any users.",
		getGroup: func(cid uint64, _ ...uint64) *qf.Group {
			return &qf.Group{CourseID: cid}
		},
		err: database.ErrEmptyGroup,
	},
	{
		name: "with non existing users",
		desc: "Should fail with ErrUpdateGroup; cannot create group with users that don't exist.",
		getGroup: func(cid uint64, _ ...uint64) *qf.Group {
			return &qf.Group{
				CourseID: cid,
				Users:    []*qf.User{{ID: 101}, {ID: 102}},
			}
		},
		enrollments: []qf.Enrollment_UserStatus{qf.Enrollment_PENDING, qf.Enrollment_PENDING},
		err:         database.ErrUpdateGroup,
	},
	{
		name:        "with users but without enrollments",
		getGroup:    bunGroupWithUsers,
		enrollments: []qf.Enrollment_UserStatus{qf.Enrollment_PENDING, qf.Enrollment_PENDING},
		err:         database.ErrUpdateGroup,
	},
	{
		name:        "with users and pending enrollments",
		getGroup:    bunGroupWithUsers,
		enrollments: []qf.Enrollment_UserStatus{qf.Enrollment_PENDING, qf.Enrollment_PENDING},
		err:         database.ErrUpdateGroup,
	},
	{
		name:        "with users and rejected enrollments",
		getGroup:    bunGroupWithUsers,
		enrollments: []qf.Enrollment_UserStatus{qf.Enrollment_NONE, qf.Enrollment_NONE},
		err:         database.ErrUpdateGroup,
	},
	{
		name:        "with user and accepted enrollment",
		getGroup:    bunGroupWithUsers,
		enrollments: []qf.Enrollment_UserStatus{qf.Enrollment_STUDENT},
	},
	{
		name:        "with users and accepted enrollments",
		getGroup:    bunGroupWithUsers,
		enrollments: []qf.Enrollment_UserStatus{qf.Enrollment_STUDENT, qf.Enrollment_STUDENT},
	},
}

var bunGroupWithUsers = func(cid uint64, userIDs ...uint64) *qf.Group {
	var users []*qf.User
	for _, uid := range userIDs {
		users = append(users, &qf.User{ID: uid})
	}
	return &qf.Group{CourseID: cid, Users: users}
}

func TestBunDBCreateAndGetGroup(t *testing.T) {
	for _, test := range createBunGroupTests {
		t.Run(test.name, func(t *testing.T) {
			db, cleanup := qtest.TestBunDB(t)

			admin := qtest.CreateFakeUser(t, db)
			course := &qf.Course{}
			qtest.CreateCourse(t, db, admin, course)

			var userIDs []uint64
			for i, enrollment := range test.enrollments {
				user := qtest.CreateFakeUser(t, db)
				userIDs = append(userIDs, user.GetID())
				if enrollment == qf.Enrollment_PENDING {
					continue
				}

				if err := db.CreateEnrollment(&qf.Enrollment{
					CourseID: course.GetID(),
					UserID:   user.GetID(),
				}); err != nil {
					t.Fatal(err)
				}
				err := errors.New("enrollment status not implemented")
				switch test.enrollments[i] {
				case qf.Enrollment_NONE:
					err = db.RejectEnrollment(user.GetID(), course.GetID())
				case qf.Enrollment_STUDENT:
					query, err1 := db.GetEnrollmentByCourseAndUser(course.GetID(), user.GetID())
					if err1 != nil {
						t.Fatal(err1)
					}
					query.Status = qf.Enrollment_STUDENT
					err = db.UpdateEnrollment(query)
				}
				if err != nil {
					t.Fatal(err)
				}
			}

			group := test.getGroup(course.GetID(), userIDs...)
			if err := db.CreateGroup(group); err != test.err {
				t.Errorf("have error '%v' want '%v'", err, test.err)
			}
			if test.err != nil {
				return
			}

			enrollments, err := db.GetEnrollmentsByCourse(course.GetID(), qf.Enrollment_STUDENT)
			if err != nil {
				t.Fatal(err)
			}
			if len(group.GetUsers()) > 0 && len(enrollments) != len(group.GetUsers()) {
				t.Errorf("have %d enrollments want %d", len(enrollments), len(group.GetUsers()))
			}
			sorted := make(map[uint64]*qf.Enrollment)
			for _, enrollment := range enrollments {
				enrollment.Course = nil
				enrollment.Group = nil
				sorted[enrollment.GetUserID()] = enrollment
			}
			for _, user := range group.GetUsers() {
				if _, ok := sorted[user.GetID()]; !ok {
					t.Errorf("have no enrollment for user %d", user.GetID())
				}
			}

			have, err := db.GetGroup(group.GetID())
			if err != nil {
				t.Fatal(err)
			}
			if len(userIDs) > 0 {
				group.Users, err = db.GetUsers(userIDs...)
				if err != nil {
					t.Fatal(err)
				}
			}
			group.Enrollments = enrollments

			have.RemoveRemoteID()
			group.RemoveRemoteID()
			if diff := cmp.Diff(group, have, protocmp.Transform()); diff != "" {
				t.Errorf("mismatch (-group +have):\n%s", diff)
			}
			cleanup()
		})
	}
}

func TestBunDBCreateGroupTwice(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	var users []*qf.User
	enrollments := []qf.Enrollment_UserStatus{qf.Enrollment_STUDENT, qf.Enrollment_STUDENT}
	for i := 0; i < len(enrollments); i++ {
		user := qtest.CreateFakeUser(t, db)
		users = append(users, user)
		if enrollments[i] == qf.Enrollment_PENDING {
			continue
		}

		if err := db.CreateEnrollment(&qf.Enrollment{
			CourseID: course.GetID(),
			UserID:   users[i].GetID(),
		}); err != nil {
			t.Fatal(err)
		}
		err := errors.New("enrollment status not implemented")
		if enrollments[i] == qf.Enrollment_STUDENT {
			query, err1 := db.GetEnrollmentByCourseAndUser(course.GetID(), users[i].GetID())
			if err1 != nil {
				t.Fatal(err1)
			}
			query.Status = qf.Enrollment_STUDENT
			err = db.UpdateEnrollment(query)
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	identical := &qf.Group{
		Name:     "SameNameGroup",
		CourseID: course.GetID(),
		Users:    users,
	}
	if err := db.CreateGroup(identical); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateGroup(identical); err != database.ErrDuplicateGroup {
		t.Fatalf("expected error '%v' have '%v'", database.ErrDuplicateGroup, err)
	}
}

func TestBunGetGroupsByCourse(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	var users []*qf.User
	enrollments := []qf.Enrollment_UserStatus{
		qf.Enrollment_STUDENT,
		qf.Enrollment_STUDENT,
		qf.Enrollment_STUDENT,
		qf.Enrollment_STUDENT,
		qf.Enrollment_STUDENT,
	}
	for i := 0; i < len(enrollments); i++ {
		user := qtest.CreateFakeUser(t, db)
		users = append(users, user)
		if enrollments[i] == qf.Enrollment_PENDING {
			continue
		}

		if err := db.CreateEnrollment(&qf.Enrollment{
			CourseID: course.GetID(),
			UserID:   user.GetID(),
		}); err != nil {
			t.Fatal(err)
		}
		err := errors.New("enrollment status not implemented")
		if enrollments[i] == qf.Enrollment_STUDENT {
			query, err1 := db.GetEnrollmentByCourseAndUser(course.GetID(), user.GetID())
			if err1 != nil {
				t.Fatal(err1)
			}
			query.Status = qf.Enrollment_STUDENT
			err = db.UpdateEnrollment(query)
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	group := &qf.Group{Name: "MyGroup", CourseID: course.GetID(), Users: users[0:2]}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}
	group2 := &qf.Group{Name: "MyOtherGroup", CourseID: course.GetID(), Users: users[2:5]}
	if err := db.CreateGroup(group2); err != nil {
		t.Fatal(err)
	}

	group2.Status = qf.Group_APPROVED
	if err := db.UpdateGroupStatus(group2); err != nil {
		t.Fatal(err)
	}

	groups, err := db.GetGroupsByCourse(course.GetID())
	if err != nil {
		t.Fatal(err)
	}
	wantUsers, gotUsers := groups[0].GetUsers(), group.GetUsers()
	if diff := cmp.Diff(wantUsers, gotUsers, protocmp.Transform()); diff != "" {
		t.Errorf("group users mismatch (-wantUsers +gotUsers):\n%s", diff)
	}
	wantUsers, gotUsers = groups[1].GetUsers(), group2.GetUsers()
	if diff := cmp.Diff(wantUsers, gotUsers, protocmp.Transform()); diff != "" {
		t.Errorf("group users mismatch (-wantUsers +gotUsers):\n%s", diff)
	}

	pendingGroups, err := db.GetGroupsByCourse(course.GetID(), qf.Group_PENDING)
	if err != nil {
		t.Fatal(err)
	}
	approvedGroups, err := db.GetGroupsByCourse(course.GetID(), qf.Group_APPROVED)
	if err != nil {
		t.Fatal(err)
	}
	if len(pendingGroups) != 1 || len(approvedGroups) != 1 {
		t.Errorf("Expected one pending and one approved group, got %d pending, %d approved", len(pendingGroups), len(approvedGroups))
	}
}

func TestBunDeleteGroupAssociations(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	if err := db.CreateCourse(admin.GetID(), course); err != nil {
		t.Fatal(err)
	}

	student1 := qtest.CreateFakeUser(t, db)
	student2 := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, student1, course)
	qtest.EnrollStudent(t, db, student2, course)

	group := &qf.Group{
		Name:     "Test Group",
		CourseID: course.GetID(),
		Users:    []*qf.User{student1, student2},
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}

	groupID := group.GetID()

	gotGroup, err := db.GetGroup(groupID)
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}
	if len(gotGroup.GetUsers()) != 2 {
		t.Fatalf("expected 2 users before delete, got %d", len(gotGroup.GetUsers()))
	}
	if len(gotGroup.GetEnrollments()) != 2 {
		t.Fatalf("expected 2 enrollments before delete, got %d", len(gotGroup.GetEnrollments()))
	}

	if err := db.DeleteGroup(groupID); err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}

	_, err = db.GetGroup(groupID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Error("expected sql.ErrNoRows when getting deleted group, got:", err)
	}

	for _, student := range []*qf.User{student1, student2} {
		enrollments, err := db.GetEnrollmentsByUser(student.GetID())
		if err != nil {
			t.Fatalf("GetEnrollmentsByUser failed: %v", err)
		}
		for _, enrollment := range enrollments {
			if enrollment.GetGroupID() != 0 {
				t.Errorf("expected enrollment.GroupID to be 0 after delete, got %d", enrollment.GetGroupID())
			}
			if enrollment.GetGroup() != nil {
				t.Errorf("expected enrollment.Group to be nil after delete, got %+v", enrollment.GetGroup())
			}
		}
	}
}

func TestBunUpdateGroupMembers(t *testing.T) {
	db, cleanup := qtest.TestBunDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	if err := db.CreateCourse(admin.GetID(), course); err != nil {
		t.Fatal(err)
	}

	student1 := qtest.CreateFakeUser(t, db)
	student2 := qtest.CreateFakeUser(t, db)
	student3 := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, student1, course)
	qtest.EnrollStudent(t, db, student2, course)
	qtest.EnrollStudent(t, db, student3, course)

	group := &qf.Group{
		Name:     "Test Group",
		CourseID: course.ID,
		Users:    []*qf.User{student1, student2},
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}

	group.Users = []*qf.User{student2, student3}
	if err := db.UpdateGroup(group); err != nil {
		t.Fatalf("UpdateGroup failed: %v", err)
	}

	updatedGroup, err := db.GetGroup(group.GetID())
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}
	if len(updatedGroup.GetUsers()) != 2 {
		t.Fatalf("expected 2 users after update, got %d: %+v", len(updatedGroup.GetUsers()), updatedGroup.GetUsers())
	}
	userIDs := make(map[uint64]bool)
	for _, user := range updatedGroup.GetUsers() {
		userIDs[user.GetID()] = true
	}
	if !userIDs[student2.GetID()] || !userIDs[student3.GetID()] {
		t.Errorf("expected group members to be student2 and student3, got %+v", updatedGroup.GetUsers())
	}

	if userIDs[student1.GetID()] {
		t.Errorf("did not expect student1 to be in group members, but found %+v", updatedGroup.GetUsers())
	}

	for _, student := range []*qf.User{student2, student3} {
		enrollments, err := db.GetEnrollmentsByUser(student.GetID())
		if err != nil {
			t.Fatalf("GetEnrollmentsByUser failed: %v", err)
		}
		for _, enrollment := range enrollments {
			if enrollment.GetGroupID() != group.GetID() {
				t.Errorf("expected enrollment.GroupID to be %d after update, got %d for student %d", group.GetID(), enrollment.GetGroupID(), student.GetID())
			}
			for _, gUser := range enrollment.GetGroup().GetUsers() {
				if gUser.GetID() != student2.GetID() && gUser.GetID() != student3.GetID() {
					t.Errorf("expected group users to be student2 and student3, got %+v", enrollment.GetGroup().GetUsers())
				}
			}
		}
	}
}
