package database_test

import (
	"reflect"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/kit/score"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

func TestGormDBGetAssignment(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	if _, err := db.GetAssignmentsByCourse(10, false); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}

	if _, err := db.GetAssignment(&pb.Assignment{ID: 10}); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBCreateAssignmentNoRecord(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
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
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user := qtest.CreateFakeUser(t, db, 10)
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
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &pb.Course{}
	admin := qtest.CreateFakeUser(t, db, 10)
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
		if diff := cmp.Diff(wantAssignments[i], gotAssignments[i], protocmp.Transform()); diff != "" {
			t.Errorf("UpdateAssignments() mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestGetAssignmentsWithSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	// create teacher, course, user (student) and assignment
	user, course, assignment := setupCourseAssignment(t, db)

	wantStruct := &pb.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Score:        42,
		Reviews:      []*pb.Review{},
		BuildInfo: &score.BuildInfo{
			BuildDate: timestamppb.Now(),
			BuildLog:  "what do you say",
			ExecTime:  50,
		},
		Scores: []*score.Score{
			{TestName: "TestBigNum", MaxScore: 100, Score: 60, Weight: 10},
			{TestName: "TestDigNum", MaxScore: 100, Score: 70, Weight: 10},
		},
	}
	t.Logf("B   wantStruct.bdate=%v", wantStruct.BuildInfo.BuildDate)
	if err := db.CreateSubmission(wantStruct); err != nil {
		t.Fatal(err)
	}
	assignments, err := db.GetAssignmentsWithSubmissions(course.ID, pb.SubmissionsForCourseRequest_ALL, true)
	if err != nil {
		t.Fatal(err)
	}
	wantAssignment := (proto.Clone(assignment)).(*pb.Assignment)
	wantAssignment.Submissions = append(wantAssignment.Submissions, wantStruct)
	t.Logf("assignments[0].bdate=%v", assignments[0].Submissions[0].BuildInfo.BuildDate)
	t.Logf("A   wantStruct.bdate=%v", wantStruct.BuildInfo.BuildDate)
	if diff := cmp.Diff(wantAssignment, assignments[0], protocmp.Transform()); diff != "" {
		t.Errorf("GetAssignmentsWithSubmissions() mismatch (-want +got):\n%s", diff)
	}

	// TODO(meling) Remove this legacy test; it will not actually work with the OldBuildInfo format (which is "2020-02-03")
	// Legacy Submission struct with ScoreObjects and BuildInfo as string:
	wantLegacy := &pb.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Score:        42,
		ScoreObjects: `[{"Secret":"hidden","TestName":"TestLintAG","Score":3,"MaxScore":3,"Weight":5},{"Secret":"hidden","TestName":"TestSchedulersAG/FIFO/No_jobs","Score":0,"MaxScore":0,"Weight":2}]`,
		OldBuildInfo: `{"BuildID":1,"BuildDate":{"seconds":1,"nanos":666550000},"BuildLog":"log data","ExecTime":50}`,
		Reviews:      []*pb.Review{},
	}
	if err := db.CreateSubmission(wantLegacy); err != nil {
		t.Fatal(err)
	}
	assignments, err = db.GetAssignmentsWithSubmissions(course.ID, pb.SubmissionsForCourseRequest_ALL, true)
	if err != nil {
		t.Fatal(err)
	}
	wantAssignment = (proto.Clone(assignment)).(*pb.Assignment)
	wantAssignment.Submissions = append(wantAssignment.Submissions, wantStruct, wantLegacy)
	if diff := cmp.Diff(wantAssignment, assignments[0], protocmp.Transform()); diff != "" {
		t.Errorf("GetAssignmentsWithSubmissions() mismatch (-want +got):\n%s", diff)
	}

	// Submission with Review
	wantReview := &pb.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Score:        45,
		Reviews: []*pb.Review{
			{
				ReviewerID: 1, Feedback: "SGTM!", Score: 42, Ready: true,
				GradingBenchmarks: []*pb.GradingBenchmark{
					{
						Heading: "Ding Dong", Comment: "Communication",
						Criteria: []*pb.GradingCriterion{
							{Points: 50, Description: "Loads of ding"},
						},
					},
				},
			},
		},
	}
	if err := db.CreateSubmission(wantReview); err != nil {
		t.Fatal(err)
	}
	assignments, err = db.GetAssignmentsWithSubmissions(course.ID, pb.SubmissionsForCourseRequest_ALL, true)
	if err != nil {
		t.Fatal(err)
	}
	wantAssignment = (proto.Clone(assignment)).(*pb.Assignment)
	wantAssignment.Submissions = append(wantAssignment.Submissions, wantStruct, wantLegacy, wantReview)
	if diff := cmp.Diff(wantAssignment, assignments[0], protocmp.Transform()); diff != "" {
		t.Errorf("GetAssignmentsWithSubmissions() mismatch (-want +got):\n%s", diff)
	}
}

func TestUpdateBenchmarks(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &pb.Course{}
	admin := qtest.CreateFakeUser(t, db, 10)
	if err := db.CreateCourse(admin.ID, course); err != nil {
		t.Fatal(err)
	}
	assignment := &pb.Assignment{
		CourseID:    course.ID,
		Name:        "Assignment 1",
		ScriptFile:  "go.sh",
		Deadline:    "12.12.2021",
		AutoApprove: false,
		Order:       1,
		IsGroupLab:  false,
	}

	if err := db.CreateAssignment(assignment); err != nil {
		t.Fatal(err)
	}

	benchmarks := []*pb.GradingBenchmark{
		{
			ID:           1,
			AssignmentID: assignment.ID,
			Heading:      "Test benchmark 1",
			Criteria: []*pb.GradingCriterion{
				{
					ID:          1,
					Description: "Criterion 1",
					BenchmarkID: 1,
					Points:      5,
				},
				{
					ID:          2,
					Description: "Criterion 2",
					BenchmarkID: 1,
					Points:      10,
				},
			},
		},
		{
			ID:           2,
			AssignmentID: assignment.ID,
			Heading:      "Test benchmark 2",
			Criteria: []*pb.GradingCriterion{
				{
					ID:          3,
					Description: "Criterion 3",
					BenchmarkID: 2,
					Points:      1,
				},
			},
		},
	}

	for _, bm := range benchmarks {
		if err := db.CreateBenchmark(bm); err != nil {
			t.Fatal(err)
		}
	}

	gotAssignments, err := db.GetAssignmentsByCourse(course.ID, true)
	if err != nil {
		t.Error(err)
	}

	assignment.GradingBenchmarks = benchmarks
	for i := range gotAssignments {
		if diff := cmp.Diff(assignment, gotAssignments[i], protocmp.Transform()); diff != "" {
			t.Errorf("UpdateAssignments() mismatch (-want +got):\n%s", diff)
		}
	}

	for _, bm := range benchmarks {
		bm.Heading = "Updated heading"
		if err := db.UpdateBenchmark(bm); err != nil {
			t.Fatal(err)
		}
		for _, c := range bm.Criteria {
			c.Description = "Updated description"
			if err := db.UpdateCriterion(c); err != nil {
				t.Fatal(err)
			}
		}
	}
	assignment.GradingBenchmarks = benchmarks
	gotAssignments, err = db.GetAssignmentsByCourse(course.ID, true)
	if err != nil {
		t.Error(err)
	}
	for i := range gotAssignments {
		if diff := cmp.Diff(assignment, gotAssignments[i], protocmp.Transform()); diff != "" {
			t.Errorf("UpdateAssignments() mismatch (-want +got):\n%s", diff)
		}
	}
}
