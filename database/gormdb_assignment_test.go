package database_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/score"
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

	admin := qtest.CreateFakeUser(t, db, 10)
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

func TestUpdateAssignment(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &qf.Course{}
	admin := qtest.CreateFakeUser(t, db, 10)
	qtest.CreateCourse(t, db, admin, course)

	if err := db.CreateAssignment(&qf.Assignment{
		CourseID:    course.ID,
		Name:        "lab1",
		Deadline:    qtest.Timestamp(t, "2022-11-11T23:59:00"),
		AutoApprove: false,
		Order:       1,
		IsGroupLab:  false,
	}); err != nil {
		t.Fatal(err)
	}

	if err := db.CreateAssignment(&qf.Assignment{
		CourseID:    course.ID,
		Name:        "lab2",
		Deadline:    qtest.Timestamp(t, "2022-11-11T23:59:00"),
		AutoApprove: false,
		Order:       2,
		IsGroupLab:  true,
	}); err != nil {
		t.Fatal(err)
	}

	assignments, err := db.GetAssignmentsByCourse(course.ID)
	if err != nil {
		t.Error(err)
	}

	wantAssignments := make([]*qf.Assignment, len(assignments))
	for i, a := range assignments {
		// test setting various zero-value entries to check that we can read back the same value
		a.Deadline = &timestamppb.Timestamp{}
		a.ScoreLimit = 0
		a.Reviewers = 0
		a.AutoApprove = !a.AutoApprove
		a.IsGroupLab = !a.IsGroupLab
		wantAssignments[i] = (proto.Clone(assignments[i])).(*qf.Assignment)
	}

	err = db.UpdateAssignments(assignments)
	if err != nil {
		t.Error(err)
	}
	gotAssignments, err := db.GetAssignmentsByCourse(course.ID)
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

	wantStruct := &qf.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Score:        42,
		Reviews:      []*qf.Review{},
		BuildInfo: &score.BuildInfo{
			BuildDate:      qtest.Timestamp(t, "2021-01-21T18:00:00"),
			SubmissionDate: qtest.Timestamp(t, "2021-01-21T18:00:00"),
			BuildLog:       "what do you say",
			ExecTime:       50,
		},
		Scores: []*score.Score{
			{TestName: "TestBigNum", MaxScore: 100, Score: 60, Weight: 10},
			{TestName: "TestDigNum", MaxScore: 100, Score: 70, Weight: 10},
		},
	}
	if err := db.CreateSubmission(wantStruct); err != nil {
		t.Fatal(err)
	}
	submissions, err := db.GetCourseSubmissions(course.ID, qf.SubmissionRequest_ALL)
	if err != nil {
		t.Fatal(err)
	}
	wantStruct.BuildInfo = nil
	wantAssignment := (proto.Clone(assignment)).(*qf.Assignment)
	wantAssignment.Submissions = append(wantAssignment.Submissions, wantStruct)
	if diff := cmp.Diff(wantAssignment.Submissions, submissions, protocmp.Transform()); diff != "" {
		t.Errorf("GetAssignmentsWithSubmissions() mismatch (-want +got):\n%s", diff)
	}

	// Submission with Review
	wantReview := &qf.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Score:        45,
		Reviews: []*qf.Review{
			{
				ReviewerID: 1, Feedback: "SGTM!", Score: 42, Ready: true,
				GradingBenchmarks: []*qf.GradingBenchmark{
					{
						Heading: "Ding Dong", Comment: "Communication",
						Criteria: []*qf.GradingCriterion{
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
	submissions, err = db.GetCourseSubmissions(course.ID, qf.SubmissionRequest_ALL)
	if err != nil {
		t.Fatal(err)
	}
	wantAssignment = (proto.Clone(assignment)).(*qf.Assignment)
	wantAssignment.Submissions = append(wantAssignment.Submissions, wantStruct, wantReview)
	if diff := cmp.Diff(wantAssignment.Submissions, submissions, protocmp.Transform()); diff != "" {
		t.Errorf("GetAssignmentsWithSubmissions() mismatch (-want +got):\n%s", diff)
	}
}

func TestUpdateBenchmarks(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	course := &qf.Course{}
	admin := qtest.CreateFakeUser(t, db, 10)
	qtest.CreateCourse(t, db, admin, course)

	assignment := &qf.Assignment{
		CourseID:    course.ID,
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
			AssignmentID: assignment.ID,
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
			AssignmentID: assignment.ID,
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

	gotAssignments, err := db.GetAssignmentsByCourse(course.ID)
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
	gotAssignments, err = db.GetAssignmentsByCourse(course.ID)
	if err != nil {
		t.Error(err)
	}
	for i := range gotAssignments {
		if diff := cmp.Diff(assignment, gotAssignments[i], protocmp.Transform()); diff != "" {
			t.Errorf("UpdateAssignments() mismatch (-want +got):\n%s", diff)
		}
	}
}
