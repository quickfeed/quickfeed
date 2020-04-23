package database_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jinzhu/gorm"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
)

var createGroupTests = []struct {
	name        string
	desc        string
	getGroup    func(uint64, ...uint64) *pb.Group
	enrollments []uint
	err         error
}{
	{
		name: "course id not set with users",
		desc: "Should fail with ErrRecordNotFound; cannot create a group that's not connected to a course.",
		getGroup: func(_ uint64, uids ...uint64) *pb.Group {
			var users []*pb.User
			for _, uid := range uids {
				users = append(users, &pb.User{ID: uid})
			}
			return &pb.Group{
				Users: users,
			}
		},
		enrollments: []uint{uint(pb.Enrollment_PENDING), uint(pb.Enrollment_PENDING)},
		err:         gorm.ErrRecordNotFound,
	},
	{
		name: "course not found with users",
		desc: "Should fail with ErrRecordNotFound; cannot create a group that's not connected to a course.",
		getGroup: func(_ uint64, uids ...uint64) *pb.Group {
			var users []*pb.User
			for _, uid := range uids {
				users = append(users, &pb.User{ID: uid})
			}
			return &pb.Group{
				CourseID: 999,
				Users:    users,
			}
		},
		enrollments: []uint{uint(pb.Enrollment_PENDING), uint(pb.Enrollment_PENDING)},
		err:         gorm.ErrRecordNotFound,
	},
	{
		name: "course found but without users",
		desc: "Should fail with ErrEmptyGroup; cannot create a group without any users.",
		getGroup: func(cid uint64, _ ...uint64) *pb.Group {
			return &pb.Group{CourseID: cid}
		},
		err: database.ErrEmptyGroup,
	},
	{
		name: "with non existing users",
		desc: "Should fail with ErrUpdateGroup; cannot create group with users that doesn't exist.",
		getGroup: func(cid uint64, _ ...uint64) *pb.Group {
			return &pb.Group{
				CourseID: cid,
				Users: []*pb.User{
					{ID: 101},
					{ID: 102},
				},
			}
		},
		enrollments: []uint{uint(pb.Enrollment_PENDING), uint(pb.Enrollment_PENDING)},
		err:         database.ErrUpdateGroup,
	},
	{
		name: "with users but without enrollments",
		desc: "Should fail with ErrUpdateGroup; cannot create group with users not enrolled in the course.",
		getGroup: func(cid uint64, uids ...uint64) *pb.Group {
			var users []*pb.User
			for _, uid := range uids {
				users = append(users, &pb.User{ID: uid})
			}
			return &pb.Group{
				CourseID: cid,
				Users:    users,
			}
		},
		enrollments: []uint{uint(pb.Enrollment_PENDING), uint(pb.Enrollment_PENDING)},
		err:         database.ErrUpdateGroup,
	},
	{
		name: "with users and pending enrollments",
		desc: "Should fail with ErrUpdateGroup; cannot create group with users not enrolled in the course.",
		getGroup: func(cid uint64, uids ...uint64) *pb.Group {
			var users []*pb.User
			for _, uid := range uids {
				users = append(users, &pb.User{ID: uid})
			}
			return &pb.Group{
				CourseID: cid,
				Users:    users,
			}
		},
		enrollments: []uint{uint(pb.Enrollment_PENDING), uint(pb.Enrollment_PENDING)},
		err:         database.ErrUpdateGroup,
	},
	{
		name: "with users and rejected enrollments",
		desc: "Should fail with ErrUpdateGroup; cannot create group with users not enrolled in the course.",
		getGroup: func(cid uint64, uids ...uint64) *pb.Group {
			var users []*pb.User
			for _, uid := range uids {
				users = append(users, &pb.User{ID: uid})
			}
			return &pb.Group{
				CourseID: cid,
				Users:    users,
			}
		},
		enrollments: []uint{uint(pb.Enrollment_NONE), uint(pb.Enrollment_NONE)},
		err:         database.ErrUpdateGroup,
	},
	{
		name: "with user and accepted enrollment",
		desc: "Should pass as the user exists and is enrolled in the course.",
		getGroup: func(cid uint64, uids ...uint64) *pb.Group {
			var users []*pb.User
			for _, uid := range uids {
				users = append(users, &pb.User{ID: uid})
			}
			return &pb.Group{
				CourseID: cid,
				Users:    users,
			}
		},
		enrollments: []uint{uint(pb.Enrollment_STUDENT)},
	},
	{
		name: "with users and accepted enrollments",
		desc: "Should pass as the users exists and are enrolled in the course.",
		getGroup: func(cid uint64, uids ...uint64) *pb.Group {
			var users []*pb.User
			for _, uid := range uids {
				users = append(users, &pb.User{ID: uid})
			}
			return &pb.Group{
				CourseID: cid,
				Users:    users,
			}
		},
		enrollments: []uint{uint(pb.Enrollment_STUDENT), uint(pb.Enrollment_STUDENT)},
	},
}

func TestGormDBCreateAndGetGroup(t *testing.T) {
	for _, test := range createGroupTests {
		t.Run(test.name, func(t *testing.T) {
			db, cleanup := setup(t)

			teacher := createFakeUser(t, db, 10)
			var course pb.Course
			if err := db.CreateCourse(teacher.ID, &course); err != nil {
				t.Fatal(err)
			}

			var uids []uint64
			// create as many users as the desired number of enrollments
			for i, enrollment := range test.enrollments {
				user := createFakeUser(t, db, uint64(i))
				uids = append(uids, user.ID)
				if enrollment == uint(pb.Enrollment_PENDING) {
					continue
				}

				// enroll user in course
				if err := db.CreateEnrollment(&pb.Enrollment{
					CourseID: course.ID,
					UserID:   user.GetID(),
				}); err != nil {
					t.Fatal(err)
				}
				err := errors.New("enrollment status not implemented")
				switch test.enrollments[i] {
				case uint(pb.Enrollment_NONE):
					err = db.RejectEnrollment(user.GetID(), course.ID)
				case uint(pb.Enrollment_STUDENT):
					query := &pb.Enrollment{
						UserID:   user.ID,
						CourseID: course.ID,
						Status:   pb.Enrollment_STUDENT,
					}
					err = db.UpdateEnrollment(query)
				}
				if err != nil {
					t.Fatal(err)
				}
			}

			// Test.
			group := test.getGroup(course.ID, uids...)
			if err := db.CreateGroup(group); err != test.err {
				t.Errorf("have error '%v' want '%v'", err, test.err)
			}
			if test.err != nil {
				return
			}

			// Verify.
			enrollments, err := db.GetEnrollmentsByCourse(course.ID, pb.Enrollment_STUDENT)
			if err != nil {
				t.Fatal(err)
			}
			if len(group.Users) > 0 && len(enrollments) != len(group.Users) {
				t.Errorf("have %d enrollments want %d", len(enrollments), len(group.Users))
			}
			sorted := make(map[uint64]*pb.Enrollment)
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
			if len(uids) > 0 {
				group.Users, err = db.GetUsers(uids...)
				if err != nil {
					t.Fatal(err)
				}
			}
			group.Enrollments = enrollments
			for _, usr := range have.Users {
				usr.Enrollments = nil
			}
			for _, e := range have.Enrollments {
				e.User.Enrollments = nil
			}

			have.RemoveRemoteID()
			group.RemoveRemoteID()
			if diff := cmp.Diff(group, have); diff != "" {
				t.Errorf("mismatch (-group +have):\n%s", diff)
			}
			cleanup()
		})
	}
}

func TestGormDBCreateGroupTwice(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	teacher := createFakeUser(t, db, 10)
	var course pb.Course
	if err := db.CreateCourse(teacher.ID, &course); err != nil {
		t.Fatal(err)
	}
	var users []*pb.User
	enrollments := []pb.Enrollment_UserStatus{pb.Enrollment_STUDENT, pb.Enrollment_STUDENT}
	// create as many users as the desired number of enrollments
	for i := 0; i < len(enrollments); i++ {
		user := createFakeUser(t, db, uint64(i))
		users = append(users, user)
		if enrollments[i] == pb.Enrollment_PENDING {
			continue
		}

		// enroll users in course
		if err := db.CreateEnrollment(&pb.Enrollment{
			CourseID: course.ID,
			UserID:   users[i].ID,
		}); err != nil {
			t.Fatal(err)
		}
		err := errors.New("enrollment status not implemented")
		switch enrollments[i] {
		case pb.Enrollment_STUDENT:
			query := &pb.Enrollment{
				UserID:   users[i].ID,
				CourseID: course.ID,
				Status:   pb.Enrollment_STUDENT,
			}
			err = db.UpdateEnrollment(query)
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	// Try to create two identical groups. The first should succeed while
	// further attempts should fail with ErrDuplicateGroup.
	identical := &pb.Group{
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
	db, cleanup := setup(t)
	defer cleanup()

	teacher := createFakeUser(t, db, 10)
	var course pb.Course
	if err := db.CreateCourse(teacher.ID, &course); err != nil {
		t.Fatal(err)
	}
	var users []*pb.User
	enrollments := []pb.Enrollment_UserStatus{
		pb.Enrollment_STUDENT,
		pb.Enrollment_STUDENT,
		pb.Enrollment_STUDENT,
		pb.Enrollment_STUDENT,
		pb.Enrollment_STUDENT,
	}
	// create as many users as the desired number of enrollments
	for i := 0; i < len(enrollments); i++ {
		user := createFakeUser(t, db, uint64(i))
		users = append(users, user)
		if enrollments[i] == pb.Enrollment_PENDING {
			continue
		}

		// enroll users in course
		if err := db.CreateEnrollment(&pb.Enrollment{
			CourseID: course.ID,
			UserID:   users[i].ID,
		}); err != nil {
			t.Fatal(err)
		}
		err := errors.New("enrollment status not implemented")
		switch enrollments[i] {
		case pb.Enrollment_STUDENT:
			query := &pb.Enrollment{
				UserID:   users[i].ID,
				CourseID: course.ID,
				Status:   pb.Enrollment_STUDENT,
			}
			err = db.UpdateEnrollment(query)
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	group := &pb.Group{Name: "MyGroup", CourseID: course.ID, Users: users[0:2]}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}
	group2 := &pb.Group{Name: "MyOtherGroup", CourseID: course.ID, Users: users[2:5]}
	if err := db.CreateGroup(group2); err != nil {
		t.Fatal(err)
	}

	group2.Status = pb.Group_APPROVED
	if err := db.UpdateGroupStatus(group2); err != nil {
		t.Fatal(err)
	}

	// must return both groups
	groups, err := db.GetGroupsByCourse(course.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(groups[0].GetUsers(), group.GetUsers()) {
		t.Errorf("have %#v want %#v", groups[0].GetUsers(), group.GetUsers())
	}
	if !reflect.DeepEqual(groups[1].GetUsers(), group2.GetUsers()) {
		t.Errorf("have %#v want %#v", groups[1].GetUsers(), group2.GetUsers())
	}

	pendingGroups, err := db.GetGroupsByCourse(course.ID, pb.Group_PENDING)
	if err != nil {
		t.Fatal(err)
	}
	approvedGroups, err := db.GetGroupsByCourse(course.ID, pb.Group_APPROVED)
	if err != nil {
		t.Fatal(err)
	}
	if len(pendingGroups) != 1 || len(approvedGroups) != 1 {
		t.Errorf("Expected one pending and one approved group, got %d pending, %d approved", len(pendingGroups), len(approvedGroups))
	}
}
