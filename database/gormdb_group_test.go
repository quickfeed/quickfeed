package database_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

var createGroupTests = []struct {
	name        string
	desc        string
	getGroup    func(uint64, ...uint64) *qf.Group
	enrollments []qf.Enrollment_UserStatus
	err         error
}{
	{
		name: "course id not set with users",
		desc: "Should fail with ErrRecordNotFound; cannot create a group that's not connected to a course.",
		getGroup: func(_ uint64, userIDs ...uint64) *qf.Group {
			var users []*qf.User
			for _, uid := range userIDs {
				users = append(users, &qf.User{ID: uid})
			}
			return &qf.Group{
				Users: users,
			}
		},
		enrollments: []qf.Enrollment_UserStatus{qf.Enrollment_PENDING, qf.Enrollment_PENDING},
		err:         gorm.ErrRecordNotFound,
	},
	{
		name: "course not found with users",
		desc: "Should fail with ErrRecordNotFound; cannot create a group that's not connected to a course.",
		getGroup: func(_ uint64, userIDs ...uint64) *qf.Group {
			var users []*qf.User
			for _, uid := range userIDs {
				users = append(users, &qf.User{ID: uid})
			}
			return &qf.Group{
				CourseID: 999,
				Users:    users,
			}
		},
		enrollments: []qf.Enrollment_UserStatus{qf.Enrollment_PENDING, qf.Enrollment_PENDING},
		err:         gorm.ErrRecordNotFound,
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
		desc: "Should fail with ErrUpdateGroup; cannot create group with users that doesn't exist.",
		getGroup: func(cid uint64, _ ...uint64) *qf.Group {
			return &qf.Group{
				CourseID: cid,
				Users: []*qf.User{
					{ID: 101},
					{ID: 102},
				},
			}
		},
		enrollments: []qf.Enrollment_UserStatus{qf.Enrollment_PENDING, qf.Enrollment_PENDING},
		err:         database.ErrUpdateGroup,
	},
	{
		name:        "with users but without enrollments",
		desc:        "Should fail with ErrUpdateGroup; cannot create group with users not enrolled in the course.",
		getGroup:    groupWithUsers,
		enrollments: []qf.Enrollment_UserStatus{qf.Enrollment_PENDING, qf.Enrollment_PENDING},
		err:         database.ErrUpdateGroup,
	},
	{
		name:        "with users and pending enrollments",
		desc:        "Should fail with ErrUpdateGroup; cannot create group with users not enrolled in the course.",
		getGroup:    groupWithUsers,
		enrollments: []qf.Enrollment_UserStatus{qf.Enrollment_PENDING, qf.Enrollment_PENDING},
		err:         database.ErrUpdateGroup,
	},
	{
		name:        "with users and rejected enrollments",
		desc:        "Should fail with ErrUpdateGroup; cannot create group with users not enrolled in the course.",
		getGroup:    groupWithUsers,
		enrollments: []qf.Enrollment_UserStatus{qf.Enrollment_NONE, qf.Enrollment_NONE},
		err:         database.ErrUpdateGroup,
	},
	{
		name:        "with user and accepted enrollment",
		desc:        "Should pass as the user exists and is enrolled in the course.",
		getGroup:    groupWithUsers,
		enrollments: []qf.Enrollment_UserStatus{qf.Enrollment_STUDENT},
	},
	{
		name:        "with users and accepted enrollments",
		desc:        "Should pass as the users exists and are enrolled in the course.",
		getGroup:    groupWithUsers,
		enrollments: []qf.Enrollment_UserStatus{qf.Enrollment_STUDENT, qf.Enrollment_STUDENT},
	},
}

var groupWithUsers = func(cid uint64, userIDs ...uint64) *qf.Group {
	var users []*qf.User
	for _, uid := range userIDs {
		users = append(users, &qf.User{ID: uid})
	}
	return &qf.Group{
		CourseID: cid,
		Users:    users,
	}
}

func TestGormDBCreateAndGetGroup(t *testing.T) {
	for _, test := range createGroupTests {
		t.Run(test.name, func(t *testing.T) {
			db, cleanup := qtest.TestDB(t)

			admin := qtest.CreateFakeUser(t, db, 10)
			course := &qf.Course{}
			qtest.CreateCourse(t, db, admin, course)

			var userIDs []uint64
			// create as many users as the desired number of enrollments
			for i, enrollment := range test.enrollments {
				user := qtest.CreateFakeUser(t, db, uint64(i))
				userIDs = append(userIDs, user.ID)
				if enrollment == qf.Enrollment_PENDING {
					continue
				}

				// enroll user in course
				if err := db.CreateEnrollment(&qf.Enrollment{
					CourseID: course.ID,
					UserID:   user.GetID(),
				}); err != nil {
					t.Fatal(err)
				}
				err := errors.New("enrollment status not implemented")
				switch test.enrollments[i] {
				case qf.Enrollment_NONE:
					err = db.RejectEnrollment(user.GetID(), course.ID)
				case qf.Enrollment_STUDENT:
					query := &qf.Enrollment{
						UserID:   user.ID,
						CourseID: course.ID,
						Status:   qf.Enrollment_STUDENT,
					}
					err = db.UpdateEnrollment(query)
				}
				if err != nil {
					t.Fatal(err)
				}
			}

			// Test.
			group := test.getGroup(course.ID, userIDs...)
			if err := db.CreateGroup(group); err != test.err {
				t.Errorf("have error '%v' want '%v'", err, test.err)
			}
			if test.err != nil {
				return
			}

			// Verify.
			enrollments, err := db.GetEnrollmentsByCourse(course.ID, qf.Enrollment_STUDENT)
			if err != nil {
				t.Fatal(err)
			}
			if len(group.Users) > 0 && len(enrollments) != len(group.Users) {
				t.Errorf("have %d enrollments want %d", len(enrollments), len(group.Users))
			}
			sorted := make(map[uint64]*qf.Enrollment)
			for _, enrollment := range enrollments {
				enrollment.Course = nil
				enrollment.Group = nil
				sorted[enrollment.UserID] = enrollment
			}
			for _, user := range group.Users {
				if _, ok := sorted[user.ID]; !ok {
					t.Errorf("have no enrollment for user %d", user.ID)
				}
			}

			have, err := db.GetGroup(group.ID)
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

func TestGormDBCreateGroupTwice(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	var users []*qf.User
	enrollments := []qf.Enrollment_UserStatus{qf.Enrollment_STUDENT, qf.Enrollment_STUDENT}
	// create as many users as the desired number of enrollments
	for i := 0; i < len(enrollments); i++ {
		user := qtest.CreateFakeUser(t, db, uint64(i))
		users = append(users, user)
		if enrollments[i] == qf.Enrollment_PENDING {
			continue
		}

		// enroll users in course
		if err := db.CreateEnrollment(&qf.Enrollment{
			CourseID: course.ID,
			UserID:   users[i].ID,
		}); err != nil {
			t.Fatal(err)
		}
		err := errors.New("enrollment status not implemented")
		if enrollments[i] == qf.Enrollment_STUDENT {
			query := &qf.Enrollment{
				UserID:   users[i].ID,
				CourseID: course.ID,
				Status:   qf.Enrollment_STUDENT,
			}
			err = db.UpdateEnrollment(query)
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	// Try to create two identical groups. The first should succeed while
	// further attempts should fail with ErrDuplicateGroup.
	identical := &qf.Group{
		Name:     "SameNameGroup",
		CourseID: course.ID,
		Users:    users,
	}
	if err := db.CreateGroup(identical); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateGroup(identical); err != database.ErrDuplicateGroup {
		t.Fatalf("expected error '%v' have '%v'", database.ErrDuplicateGroup, err)
	}
}

func TestGetGroupsByCourse(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 10)
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
	// create as many users as the desired number of enrollments
	for i := 0; i < len(enrollments); i++ {
		user := qtest.CreateFakeUser(t, db, uint64(i))
		users = append(users, user)
		if enrollments[i] == qf.Enrollment_PENDING {
			continue
		}

		// enroll users in course
		if err := db.CreateEnrollment(&qf.Enrollment{
			CourseID: course.ID,
			UserID:   users[i].ID,
		}); err != nil {
			t.Fatal(err)
		}
		err := errors.New("enrollment status not implemented")
		if enrollments[i] == qf.Enrollment_STUDENT {
			query := &qf.Enrollment{
				UserID:   users[i].ID,
				CourseID: course.ID,
				Status:   qf.Enrollment_STUDENT,
			}
			err = db.UpdateEnrollment(query)
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	group := &qf.Group{Name: "MyGroup", CourseID: course.ID, Users: users[0:2]}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}
	group2 := &qf.Group{Name: "MyOtherGroup", CourseID: course.ID, Users: users[2:5]}
	if err := db.CreateGroup(group2); err != nil {
		t.Fatal(err)
	}

	group2.Status = qf.Group_APPROVED
	if err := db.UpdateGroupStatus(group2); err != nil {
		t.Fatal(err)
	}

	// must return both groups
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

	pendingGroups, err := db.GetGroupsByCourse(course.ID, qf.Group_PENDING)
	if err != nil {
		t.Fatal(err)
	}
	approvedGroups, err := db.GetGroupsByCourse(course.ID, qf.Group_APPROVED)
	if err != nil {
		t.Fatal(err)
	}
	if len(pendingGroups) != 1 || len(approvedGroups) != 1 {
		t.Errorf("Expected one pending and one approved group, got %d pending, %d approved", len(pendingGroups), len(approvedGroups))
	}
}
