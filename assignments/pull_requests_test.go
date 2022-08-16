package assignments

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetNextReviewer(t *testing.T) {
	// We create local versions of the maps
	teacherReviewCounter := make(countMap)
	groupReviewCounter := make(countMap)
	IDs := []uint64{1, 2, 3, 4}
	teachers := []*qf.User{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5}}
	students := []*qf.User{{ID: 1}, {ID: 2}, {ID: 3}}
	for _, ID := range IDs {
		for i := 0; i < len(teachers)*5; i++ {
			teacherReviewCounter.initialize(ID)
			gotTeacher := getNextReviewer(teachers, teacherReviewCounter[ID])
			wantTeacher := teachers[i%len(teachers)]
			if diff := cmp.Diff(wantTeacher, gotTeacher, protocmp.Transform()); diff != "" {
				t.Errorf("getNextReviewer() mismatch (-wantTeacher, +gotTeacher):\n%s", diff)
			}
		}

		// Adding a new teacher.
		// Teacher is expected to be picked as reviewer len(teachers)-1 times.
		wantTeacher := &qf.User{ID: 6}
		teachers = append(teachers, wantTeacher)
		for i := 0; i < len(teachers)-1; i++ {
			teacherReviewCounter.initialize(ID)
			gotTeacher := getNextReviewer(teachers, teacherReviewCounter[ID])
			if diff := cmp.Diff(wantTeacher, gotTeacher, protocmp.Transform()); diff != "" {
				t.Errorf("getNextReviewer() mismatch (-wantTeacher, +gotTeacher):\n%s", diff)
			}
		}
		teachers = teachers[:len(teachers)-1]

		for i := 0; i < len(students)*3; i++ {
			groupReviewCounter.initialize(ID)
			gotStudent := getNextReviewer(students, groupReviewCounter[ID])
			wantStudent := students[i%len(students)]
			if diff := cmp.Diff(wantStudent, gotStudent, protocmp.Transform()); diff != "" {
				t.Errorf("getNextReviewer() mismatch (-wantStudent, +gotStudent):\n%s", diff)
			}
		}

		// Adding a new student
		// Student is expected to be picked as reviewer len(student)-1 times.
		wantStudent := &qf.User{ID: 4}
		students = append(students, wantStudent)
		for i := 0; i < len(students)-1; i++ {
			groupReviewCounter.initialize(ID)
			gotStudent := getNextReviewer(students, groupReviewCounter[ID])
			if diff := cmp.Diff(wantStudent, gotStudent, protocmp.Transform()); diff != "" {
				t.Errorf("getNextReviewer() mismatch (-wantStudent, +gotStudent):\n%s", diff)
			}
		}
		students = students[:len(students)-1]
	}
}
