package database_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

func TestGormDBGetAssignment(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	if _, err := db.GetAssignmentsByCourse(10); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}

	if _, err := db.GetAssignment(&qf.Assignment{ID: 10}); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestGormDBCreateAssignmentNoRecord(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	assignment := qf.Assignment{
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

	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, &qf.Course{})

	gotAssignment := &qf.Assignment{
		CourseID: 1,
		Order:    1,
	}

	if err := db.CreateAssignment(gotAssignment); err != nil {
		t.Fatal(err)
	}

	assignments, err := db.GetAssignmentsByCourse(1)
	if err != nil {
		t.Fatal(err)
	}
	wantAssignment := assignments[0]

	if len(assignments) != 1 {
		t.Fatalf("have size %v wanted %v", len(assignments), 1)
	}

	if diff := cmp.Diff(wantAssignment, gotAssignment, protocmp.Transform()); diff != "" {
		t.Errorf("CreateAssignment() mismatch (-wantAssignment, +gotAssignment):\n%s", diff)
	}

	if _, err = db.GetAssignment(&qf.Assignment{ID: 1}); err != nil {
		t.Errorf("failed to get existing assignment by ID: %s", err)
	}
}

func TestUpdateAssignments(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &qf.Course{}
	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, course)

	if err := db.CreateAssignment(&qf.Assignment{
		CourseID:    course.GetID(),
		Name:        "lab1",
		Deadline:    qtest.Timestamp(t, "2022-11-11T23:59:00"),
		AutoApprove: false,
		Order:       1,
		IsGroupLab:  false,
	}); err != nil {
		t.Fatal(err)
	}

	if err := db.CreateAssignment(&qf.Assignment{
		CourseID:    course.GetID(),
		Name:        "lab2",
		Deadline:    qtest.Timestamp(t, "2022-11-11T23:59:00"),
		AutoApprove: false,
		Order:       2,
		IsGroupLab:  true,
	}); err != nil {
		t.Fatal(err)
	}

	assignments, err := db.GetAssignmentsByCourse(course.GetID())
	if err != nil {
		t.Error(err)
	}

	wantAssignments := make([]*qf.Assignment, len(assignments))
	for i, a := range assignments {
		// test setting various zero-value entries to check that we can read back the same value
		a.Deadline = &timestamppb.Timestamp{}
		a.ScoreLimit = 0
		a.Reviewers = 0
		a.ReviewWeight = 40
		a.AutoApprove = !a.GetAutoApprove()
		a.IsGroupLab = !a.GetIsGroupLab()
		wantAssignments[i] = proto.CloneOf(assignments[i])
	}

	err = db.UpdateAssignments(assignments)
	if err != nil {
		t.Error(err)
	}
	gotAssignments, err := db.GetAssignmentsByCourse(course.GetID())
	if err != nil {
		t.Error(err)
	}

	for i := range gotAssignments {
		if diff := cmp.Diff(wantAssignments[i], gotAssignments[i], protocmp.Transform()); diff != "" {
			t.Errorf("UpdateAssignments() mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestUpdateAssignmentsWithExpectedTests(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	wantAssignment := &qf.Assignment{
		CourseID:      course.GetID(),
		Name:          "LabName",
		Deadline:      qtest.Timestamp(t, "2022-11-11T23:59:00"),
		AutoApprove:   false,
		Order:         1,
		IsGroupLab:    false,
		ExpectedTests: []*qf.TestInfo{{TestName: "test1", MaxScore: 10, Weight: 10}},
	}
	if err := db.CreateAssignment(wantAssignment); err != nil {
		t.Fatal(err)
	}

	// Get single gotAssignment
	gotAssignment, err := db.GetAssignment(&qf.Assignment{Name: "LabName", CourseID: course.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantAssignment, gotAssignment, protocmp.Transform()); diff != "" {
		t.Errorf("GetAssignment() mismatch (-want +got):\n%s", diff)
	}

	// Get all course gotAssignments
	gotAssignments, err := db.GetAssignmentsByCourse(course.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if len(gotAssignments) != 1 {
		t.Errorf("expected 1 assignment, got %d", len(gotAssignments))
	}
	gotAssignment = gotAssignments[0]
	if diff := cmp.Diff(wantAssignment, gotAssignment, protocmp.Transform()); diff != "" {
		t.Errorf("GetAssignmentsByCourse() mismatch (-want +got):\n%s", diff)
	}

	// Update the assignment with new expected tests
	wantAssignment.ExpectedTests[0].MaxScore = 20
	wantAssignment.ExpectedTests[0].Weight = 20
	wantAssignment.ExpectedTests = append(wantAssignment.ExpectedTests, &qf.TestInfo{TestName: "test2", MaxScore: 15, Weight: 15})

	updateExpectedTests := []*qf.Assignment{wantAssignment}
	err = db.UpdateAssignments(updateExpectedTests)
	if err != nil {
		t.Fatal(err)
	}

	// Get single assignment
	gotAssignment, err = db.GetAssignment(&qf.Assignment{Name: "LabName", CourseID: course.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantAssignment, gotAssignment, protocmp.Transform()); diff != "" {
		t.Errorf("GetAssignment() mismatch (-want +got):\n%s", diff)
	}

	// Get all course assignments
	gotAssignments, err = db.GetAssignmentsByCourse(course.GetID())
	if err != nil {
		t.Fatal(err)
	}
	if len(gotAssignments) != 1 {
		t.Fatalf("expected 1 assignment after update, got %d", len(gotAssignments))
	}
	gotAssignment = gotAssignments[0]
	if diff := cmp.Diff(wantAssignment, gotAssignment, protocmp.Transform()); diff != "" {
		t.Errorf("GetAssignmentsByCourse() mismatch (-want +got):\n%s", diff)
	}
}

func TestUpdateExpectedTestsWithNewName(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	wantAssignment := &qf.Assignment{
		CourseID:      course.GetID(),
		Name:          "LabName",
		Deadline:      qtest.Timestamp(t, "2022-11-11T23:59:00"),
		AutoApprove:   false,
		Order:         1,
		IsGroupLab:    false,
		ExpectedTests: []*qf.TestInfo{{TestName: "test1", MaxScore: 10, Weight: 10}},
	}
	if err := db.CreateAssignment(wantAssignment); err != nil {
		t.Fatal(err)
	}
	gotAssignment, err := db.GetAssignment(&qf.Assignment{Name: "LabName", CourseID: course.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantAssignment, gotAssignment, protocmp.Transform()); diff != "" {
		t.Errorf("GetAssignment() mismatch (-want +got):\n%s", diff)
	}

	// Update the assignment with new expected test name and new scores; should overwrite all previous tests
	wantAssignment.ExpectedTests = []*qf.TestInfo{{TestName: "test2", MaxScore: 20, Weight: 50}}
	err = db.UpdateAssignments([]*qf.Assignment{wantAssignment})
	if err != nil {
		t.Fatal(err)
	}
	gotAssignment, err = db.GetAssignment(&qf.Assignment{Name: "LabName", CourseID: course.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantAssignment, gotAssignment, protocmp.Transform()); diff != "" {
		t.Errorf("GetAssignment() mismatch (-want +got):\n%s", diff)
	}
}

func TestCreateBenchmarkWithoutAssignment(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	benchmark := &qf.GradingBenchmark{
		ID:           1,
		AssignmentID: 1,
		Heading:      "Test benchmark",
		Criteria: []*qf.GradingCriterion{
			{
				ID:          1,
				Description: "Criterion 1",
				Points:      5,
			},
			{
				ID:          2,
				Description: "Criterion 2",
				Points:      10,
			},
		},
	}

	if err := db.CreateBenchmark(benchmark); err != gorm.ErrRecordNotFound {
		t.Errorf("have error '%v' wanted '%v'", err, gorm.ErrRecordNotFound)
	}
}

func TestUpdateBenchmarks(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &qf.Course{}
	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, course)

	assignment := &qf.Assignment{
		CourseID:    course.GetID(),
		Name:        "Assignment 1",
		Deadline:    qtest.Timestamp(t, "2021-12-12T19:00:00"),
		AutoApprove: false,
		Order:       1,
		IsGroupLab:  false,
	}
	if err := db.CreateAssignment(assignment); err != nil {
		t.Fatal(err)
	}

	benchmarks := []*qf.GradingBenchmark{
		{
			ID:           1,
			AssignmentID: assignment.GetID(),
			Heading:      "Test benchmark 1",
			Criteria: []*qf.GradingCriterion{
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
			AssignmentID: assignment.GetID(),
			Heading:      "Test benchmark 2",
			Criteria: []*qf.GradingCriterion{
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

	gotAssignments, err := db.GetAssignmentsByCourse(course.GetID())
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
		for _, c := range bm.GetCriteria() {
			c.Description = "Updated description"
			if err := db.UpdateCriterion(c); err != nil {
				t.Fatal(err)
			}
		}
	}
	assignment.GradingBenchmarks = benchmarks
	gotAssignments, err = db.GetAssignmentsByCourse(course.GetID())
	if err != nil {
		t.Error(err)
	}
	for i := range gotAssignments {
		if diff := cmp.Diff(assignment, gotAssignments[i], protocmp.Transform()); diff != "" {
			t.Errorf("UpdateAssignments() mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestCreateCriterion(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &qf.Course{}
	admin := qtest.CreateFakeUser(t, db)
	qtest.CreateCourse(t, db, admin, course)

	assignment := &qf.Assignment{
		CourseID: course.GetID(),
		Order:    1,
	}
	qtest.CreateAssignment(t, db, assignment)

	benchmark := &qf.GradingBenchmark{
		CourseID:     course.GetID(),
		AssignmentID: assignment.GetID(),
	}
	qtest.CreateBenchmark(t, db, benchmark)

	tests := []struct {
		name      string
		criterion *qf.GradingCriterion
		wantErr   error
	}{
		{
			name: "valid criterion",
			criterion: &qf.GradingCriterion{
				CourseID:    course.GetID(),
				BenchmarkID: benchmark.GetID(),
			},
		},
		{
			name: "invalid benchmarkID",
			criterion: &qf.GradingCriterion{
				CourseID:    course.GetID(),
				BenchmarkID: 3,
			},
			wantErr: gorm.ErrRecordNotFound,
		},
		{
			name: "invalid courseID",
			criterion: &qf.GradingCriterion{
				CourseID:    5,
				BenchmarkID: benchmark.GetID(),
			},
			wantErr: database.ErrInvalidCourseRelation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qtest.CheckError(t, db.CreateCriterion(tt.criterion), tt.wantErr)
		})
	}
}

func TestUpdateAssignmentsRemovedExpectedTests(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	assignment := &qf.Assignment{
		CourseID: course.GetID(),
		Name:     "LabName",
		Deadline: qtest.Timestamp(t, "2022-11-11T23:59:00"),
		Order:    1,
		ExpectedTests: []*qf.TestInfo{
			{TestName: "test1", MaxScore: 10, Weight: 10},
			{TestName: "test2", MaxScore: 20, Weight: 20},
		},
	}
	if err := db.CreateAssignment(assignment); err != nil {
		t.Fatal(err)
	}

	// Removing entries from tests.json must remove the corresponding
	// test_infos rows from the database (issue #1439).
	for _, expectedTests := range [][]*qf.TestInfo{nil, {}} {
		assignment.ExpectedTests = []*qf.TestInfo{{TestName: "test1", MaxScore: 10, Weight: 10}}
		if err := db.UpdateAssignments([]*qf.Assignment{assignment}); err != nil {
			t.Fatal(err)
		}

		assignment.ExpectedTests = expectedTests
		if err := db.UpdateAssignments([]*qf.Assignment{assignment}); err != nil {
			t.Fatal(err)
		}
		gotAssignment, err := db.GetAssignment(&qf.Assignment{Name: "LabName", CourseID: course.GetID()})
		if err != nil {
			t.Fatal(err)
		}
		if len(gotAssignment.GetExpectedTests()) != 0 {
			t.Errorf("expected no expected tests after removal, got %d", len(gotAssignment.GetExpectedTests()))
		}
	}
}

func TestUpdateAssignmentsNeverHadExpectedTests(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	assignment := &qf.Assignment{
		CourseID: course.GetID(),
		Name:     "LabName",
		Deadline: qtest.Timestamp(t, "2022-11-11T23:59:00"),
		Order:    1,
	}
	if err := db.CreateAssignment(assignment); err != nil {
		t.Fatal(err)
	}

	// Syncing an assignment without a tests.json multiple times must not fail.
	for range 2 {
		if err := db.UpdateAssignments([]*qf.Assignment{assignment}); err != nil {
			t.Fatal(err)
		}
	}
	gotAssignment, err := db.GetAssignment(&qf.Assignment{Name: "LabName", CourseID: course.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if len(gotAssignment.GetExpectedTests()) != 0 {
		t.Errorf("expected no expected tests, got %d", len(gotAssignment.GetExpectedTests()))
	}
}

func TestUpdateAssignmentsRemovedGradingBenchmarks(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	assignment := &qf.Assignment{
		CourseID: course.GetID(),
		Name:     "LabName",
		Deadline: qtest.Timestamp(t, "2022-11-11T23:59:00"),
		Order:    1,
	}
	if err := db.CreateAssignment(assignment); err != nil {
		t.Fatal(err)
	}
	benchmark := &qf.GradingBenchmark{
		CourseID:     course.GetID(),
		AssignmentID: assignment.GetID(),
		Heading:      "Test benchmark 1",
		Criteria: []*qf.GradingCriterion{
			{CourseID: course.GetID(), Description: "Criterion 1", Points: 5},
		},
	}
	if err := db.CreateBenchmark(benchmark); err != nil {
		t.Fatal(err)
	}
	gotBenchmarks, err := db.GetBenchmarks(&qf.Assignment{ID: assignment.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if len(gotBenchmarks) != 1 {
		t.Fatalf("expected 1 benchmark before removal, got %d", len(gotBenchmarks))
	}

	// Removing criteria.json must remove the assignment's benchmarks
	// and criteria from the database (issue #1439).
	assignment.GradingBenchmarks = nil
	if err := db.UpdateAssignments([]*qf.Assignment{assignment}); err != nil {
		t.Fatal(err)
	}
	gotBenchmarks, err = db.GetBenchmarks(&qf.Assignment{ID: assignment.GetID()})
	if err != nil {
		t.Fatal(err)
	}
	if len(gotBenchmarks) != 0 {
		t.Errorf("expected no benchmarks after removal, got %d", len(gotBenchmarks))
	}
}

// TestUpdateAssignmentsRescoresSubmissions checks that changing how much an
// assignment's reviews count recomputes its submissions' scores, while their
// grades remain.
func TestUpdateAssignmentsRescoresSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)
	assignment := &qf.Assignment{
		CourseID:     course.GetID(),
		Name:         "lab1",
		Deadline:     qtest.Timestamp(t, "2022-11-11T23:59:00"),
		Order:        1,
		Reviewers:    1,
		ReviewWeight: 40,
	}
	qtest.CreateAssignment(t, db, assignment)
	approvedUser, ungradedUser := qtest.CreateFakeUser(t, db), qtest.CreateFakeUser(t, db)
	approved := &qf.Submission{
		AssignmentID: assignment.GetID(), UserID: approvedUser.GetID(), TestScore: 90, ReviewScore: 70, Score: 82,
		Grades: []*qf.Grade{{UserID: approvedUser.GetID(), Status: qf.Submission_APPROVED}},
	}
	ungraded := &qf.Submission{AssignmentID: assignment.GetID(), UserID: ungradedUser.GetID(), TestScore: 50, Score: 30}
	qtest.CreateSubmission(t, db, approved)
	qtest.CreateSubmission(t, db, ungraded)

	update := func(reviewers, reviewWeight uint32) {
		t.Helper()
		changed := proto.CloneOf(assignment)
		changed.Reviewers, changed.ReviewWeight = reviewers, reviewWeight
		if err := db.UpdateAssignments([]*qf.Assignment{changed}); err != nil {
			t.Fatal(err)
		}
	}
	check := func(step string, wantApproved, wantUngraded uint32) {
		t.Helper()
		got := qtest.GetSubmission(t, db, &qf.Submission{ID: approved.GetID()})
		if got.GetScore() != wantApproved {
			t.Errorf("%s: approved submission's score = %d, want %d", step, got.GetScore(), wantApproved)
		}
		if status := got.GetStatusByUser(approvedUser.GetID()); status != qf.Submission_APPROVED {
			t.Errorf("%s: approved submission's status = %v, want %v", step, status, qf.Submission_APPROVED)
		}
		if got := qtest.GetSubmission(t, db, &qf.Submission{ID: ungraded.GetID()}).GetScore(); got != wantUngraded {
			t.Errorf("%s: ungraded submission's score = %d, want %d", step, got, wantUngraded)
		}
	}

	update(1, 50)
	check("review weight 50", 80, 25)
	update(0, 50)
	check("no reviewers", 90, 50)
	update(1, 100)
	check("review weight 100", 70, 0)
}
