package database_test

import (
	"reflect"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func TestGormDBGetAssignment(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	if _, err := db.GetAssignmentsByCourse(10, false); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}

	if _, err := db.GetAssignment(&pb.Assignment{ID: 10}); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBCreateAssignmentNoRecord(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	assignment := pb.Assignment{
		CourseID: 1,
		Name:     "Lab 1",
	}

	// Should fail as course 1 does not exist.
	if err := db.CreateAssignment(&assignment); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBCreateAssignment(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	user := createFakeUser(t, db, 10)
	if err := db.CreateCourse(user.ID, &pb.Course{}); err != nil {
		t.Fatal(err)
	}

	assignment := pb.Assignment{
		CourseID: 1,
		Order:    1,
	}

	if err := db.CreateAssignment(&assignment); err != nil {
		t.Fatal(err)
	}

	assignments, err := db.GetAssignmentsByCourse(1, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(assignments) != 1 {
		t.Fatalf("have size %v wanted %v", len(assignments), 1)
	}

	if !reflect.DeepEqual(assignments[0], &assignment) {
		t.Fatalf("want %v have %v", assignments[0], &assignment)
	}

	if _, err = db.GetAssignment(&pb.Assignment{ID: 1}); err != nil {
		t.Errorf("failed to get existing assignment by ID: %s", err)
	}
}

func TestUpdateAssignment(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	course := &pb.Course{}
	admin := createFakeUser(t, db, 10)
	if err := db.CreateCourse(admin.ID, course); err != nil {
		t.Fatal(err)
	}

	if err := db.CreateAssignment(&pb.Assignment{
		CourseID:    course.ID,
		Name:        "lab1",
		ScriptFile:  "go.sh",
		Deadline:    "11.11.2022",
		AutoApprove: false,
		Order:       1,
		IsGroupLab:  false,
	}); err != nil {
		t.Fatal(err)
	}

	if err := db.CreateAssignment(&pb.Assignment{
		CourseID:    course.ID,
		Name:        "lab2",
		ScriptFile:  "go.sh",
		Deadline:    "11.11.2022",
		AutoApprove: false,
		Order:       2,
		IsGroupLab:  true,
	}); err != nil {
		t.Fatal(err)
	}

	assignments, err := db.GetAssignmentsByCourse(course.ID, false)
	if err != nil {
		t.Error(err)
	}
	wantAssignments := make([]*pb.Assignment, len(assignments))
	for i, a := range assignments {
		// test setting various zero-value entries to check that we can read back the same value
		a.Deadline = ""
		a.ScoreLimit = 0
		a.Reviewers = 0
		a.AutoApprove = !a.AutoApprove
		a.IsGroupLab = !a.IsGroupLab
		wantAssignments[i] = (proto.Clone(assignments[i])).(*pb.Assignment)
	}

	err = db.UpdateAssignments(assignments)
	if err != nil {
		t.Error(err)
	}
	gotAssignments, err := db.GetAssignmentsByCourse(course.ID, false)
	if err != nil {
		t.Error(err)
	}
	for i := range gotAssignments {
		if diff := cmp.Diff(wantAssignments[i], gotAssignments[i], cmpopts.IgnoreUnexported(pb.Assignment{})); diff != "" {
			t.Errorf("UpdateAssignments() mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestGetAssignmentsWithSubmissions(t *testing.T) {
	db, cleanup := setup(t)
	defer cleanup()

	// create teacher, course, user (student) and assignment
	user, course, assignment := setupCourseAssignment(t, db)

	want := &pb.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Score:        42,
		ScoreObjects: "scores",
		BuildInfo:    "build info",
		Reviews:      []*pb.Review{},
	}
	if err := db.CreateSubmission(want); err != nil {
		t.Fatal(err)
	}

	assignments, err := db.GetAssignmentsWithSubmissions(course.ID, pb.SubmissionsForCourseRequest_ALL)
	if err != nil {
		t.Fatal(err)
	}
	wantAssignment := (proto.Clone(assignment)).(*pb.Assignment)
	wantAssignment.Submissions = append(wantAssignment.Submissions, want)
	if diff := cmp.Diff(wantAssignment, assignments[0], cmpopts.IgnoreUnexported(pb.Assignment{}, pb.Submission{}, pb.Review{})); diff != "" {
		t.Errorf("GetAssignmentsWithSubmissions() mismatch (-want +got):\n%s", diff)
	}

	want2 := &pb.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Score:        45,
		ScoreObjects: "scores 45",
		BuildInfo:    "build info 56",
		Reviews:      []*pb.Review{},
	}
	if err := db.CreateSubmission(want2); err != nil {
		t.Fatal(err)
	}

	assignments, err = db.GetAssignmentsWithSubmissions(course.ID, pb.SubmissionsForCourseRequest_ALL)
	if err != nil {
		t.Fatal(err)
	}
	wantAssignment2 := (proto.Clone(assignment)).(*pb.Assignment)
	wantAssignment2.Submissions = append(wantAssignment2.Submissions, want2)
	if diff := cmp.Diff(wantAssignment2, assignments[0], cmpopts.IgnoreUnexported(pb.Assignment{}, pb.Submission{}, pb.Review{})); diff != "" {
		t.Errorf("GetAssignmentsWithSubmissions() mismatch (-want +got):\n%s", diff)
	}
}
