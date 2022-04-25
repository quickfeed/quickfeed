package assignments

import (
	"errors"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/scm"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

// TODO(Espeland): This test doesn't currently accomplish much.
func TestAssignReviewers(t *testing.T) {
	// Reset these before testing
	teacherReviewCounter = make(map[uint64]map[uint64]int)
	groupReviewCounter = make(map[uint64]map[uint64]int)
	type testUser struct {
		login string
		role  pb.Enrollment_UserStatus
	}
	tests := map[string]struct {
		testUsers []testUser
	}{
		"Simple": {
			testUsers: []testUser{
				{login: "student1", role: pb.Enrollment_STUDENT},
				{login: "teacher1", role: pb.Enrollment_TEACHER},
			},
		},
		"No enrollments": {testUsers: []testUser{}},
	}

	logger := qtest.Logger(t)
	repo := &pb.Repository{HTMLURL: "irrelevant"}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			db, cleanup := qtest.TestDB(t)
			defer cleanup()
			admin := qtest.CreateNamedUser(t, db, 1, "admin")
			course := &pb.Course{Provider: "fake"}
			qtest.CreateCourse(t, db, admin, course)

			var nextRemoteID uint64 = 2
			for _, testUser := range tt.testUsers {
				user := qtest.CreateNamedUser(t, db, nextRemoteID, testUser.login)
				enrollment := &pb.Enrollment{UserID: user.GetID(), CourseID: course.GetID()}
				if err := db.CreateEnrollment(enrollment); err != nil {
					t.Fatal(err)
				}
				enrollment.Status = testUser.role
				if err := db.UpdateEnrollment(enrollment); err != nil {
					t.Fatal(err)
				}
				nextRemoteID++
			}
			sc, err := scm.NewSCMClient(logger, "fake", "irrelevant")
			if err != nil {
				t.Fatal(err)
			}
			if err = AssignReviewers(sc, db, course, repo, &pb.PullRequest{Number: 1}); err != nil && errors.Is(err, scm.ErrNotSupported{}) {
				t.Fatal(err)
			}
		})
	}
}

func TestGetNextReviewer(t *testing.T) {
	// We create local versions of the maps
	teacherReviewCounter := make(map[uint64]map[uint64]int)
	groupReviewCounter := make(map[uint64]map[uint64]int)
	IDs := []uint64{1, 2, 3, 4}
	teachers := []*pb.User{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5}}
	students := []*pb.User{{ID: 1}, {ID: 2}, {ID: 3}}
	for _, ID := range IDs {
		for i := 0; i < len(teachers)*5; i++ {
			gotTeacher, err := getNextReviewer(ID, teachers, teacherReviewCounter)
			if err != nil {
				t.Fatal(err)
			}
			wantTeacher := teachers[i%len(teachers)]
			if diff := cmp.Diff(wantTeacher, gotTeacher, protocmp.Transform()); diff != "" {
				t.Errorf("getNextReviewer() mismatch (-wantTeacher, +gotTeacher):\n%s", diff)
			}
		}

		// Adding a new teacher.
		// Teacher is expected to be picked as reviewer len(teachers)-1 times.
		wantTeacher := &pb.User{ID: 6}
		teachers = append(teachers, wantTeacher)
		for i := 0; i < len(teachers)-1; i++ {
			gotTeacher, err := getNextReviewer(ID, teachers, teacherReviewCounter)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(wantTeacher, gotTeacher, protocmp.Transform()); diff != "" {
				t.Errorf("getNextReviewer() mismatch (-wantTeacher, +gotTeacher):\n%s", diff)
			}
		}
		teachers = teachers[:len(teachers)-1]

		for i := 0; i < len(students)*3; i++ {
			gotStudent, err := getNextReviewer(ID, students, groupReviewCounter)
			if err != nil {
				t.Fatal(err)
			}
			wantStudent := students[i%len(students)]
			if diff := cmp.Diff(wantStudent, gotStudent, protocmp.Transform()); diff != "" {
				t.Errorf("getNextReviewer() mismatch (-wantStudent, +gotStudent):\n%s", diff)
			}
		}

		// Adding a new student
		// Student is expected to be picked as reviewer len(student)-1 times.
		wantStudent := &pb.User{ID: 4}
		students = append(students, wantStudent)
		for i := 0; i < len(students)-1; i++ {
			gotStudent, err := getNextReviewer(ID, students, groupReviewCounter)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(wantStudent, gotStudent, protocmp.Transform()); diff != "" {
				t.Errorf("getNextReviewer() mismatch (-wantStudent, +gotStudent):\n%s", diff)
			}
		}
		students = students[:len(students)-1]
	}
}
