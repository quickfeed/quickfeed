package web_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestApproveSubmission(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm, _ := MockClientWithUser(t, db)

	admin := qtest.CreateFakeUser(t, db)
	course := qtest.MockCourses[0]
	err := db.CreateCourse(admin.ID, course)
	if err != nil {
		t.Fatal(err)
	}

	student := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, student, course)

	lab := &qf.Assignment{
		CourseID: course.ID,
		Name:     "test lab",
		Order:    1,
	}
	if err = db.CreateAssignment(lab); err != nil {
		t.Fatal(err)
	}

	wantSubmission := &qf.Submission{
		AssignmentID: lab.ID,
		UserID:       student.ID,
		Score:        17,
	}
	if err = db.CreateSubmission(wantSubmission); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	cookie := Cookie(t, tm, admin)

	if _, err = client.UpdateSubmission(ctx, qtest.RequestWithCookie(&qf.UpdateSubmissionRequest{
		SubmissionID: wantSubmission.ID,
		CourseID:     course.ID,
		Status:       qf.Submission_APPROVED,
	}, cookie)); err != nil {
		t.Error(err)
	}

	gotApprovedSubmission, err := db.GetSubmission(&qf.Submission{ID: wantSubmission.ID})
	if err != nil {
		t.Fatal(err)
	}
	wantSubmission.Status = qf.Submission_APPROVED
	wantSubmission.ApprovedDate = gotApprovedSubmission.ApprovedDate

	if diff := cmp.Diff(wantSubmission, gotApprovedSubmission, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateSubmission(approve) mismatch (-wantSubmission, +gotApprovedSubmission):\n%s", diff)
	}

	if _, err = client.UpdateSubmission(ctx, qtest.RequestWithCookie(&qf.UpdateSubmissionRequest{
		SubmissionID: wantSubmission.ID,
		CourseID:     course.ID,
		Status:       qf.Submission_REJECTED,
	}, cookie)); err != nil {
		t.Error(err)
	}

	gotRejectedSubmission, err := db.GetSubmission(&qf.Submission{ID: wantSubmission.ID})
	if err != nil {
		t.Fatal(err)
	}
	wantSubmission.Status = qf.Submission_REJECTED
	// Note that the approved date is not set when the submission is rejected

	if diff := cmp.Diff(wantSubmission, gotRejectedSubmission, protocmp.Transform()); diff != "" {
		t.Errorf("UpdateSubmission(reject) mismatch (-wantSubmission, +gotRejectedSubmission):\n%s", diff)
	}
}

func TestGetSubmissionsByCourse(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm, _ := MockClientWithUser(t, db)

	admin := qtest.CreateFakeUser(t, db)
	course := qtest.MockCourses[2]
	qtest.CreateCourse(t, db, admin, course)
	student1 := qtest.CreateFakeUser(t, db)
	student2 := qtest.CreateFakeUser(t, db)
	student3 := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, student1, course)
	qtest.EnrollStudent(t, db, student2, course)
	qtest.EnrollStudent(t, db, student3, course)

	ctx := context.Background()
	cookie := Cookie(t, tm, admin)

	enrols, err := client.GetEnrollments(ctx, qtest.RequestWithCookie(&qf.EnrollmentRequest{
		FetchMode: &qf.EnrollmentRequest_CourseID{
			CourseID: course.ID,
		},
	}, cookie))
	if err != nil {
		t.Error(err)
	}
	if len(enrols.Msg.Enrollments) != 4 {
		t.Errorf("expected 4 enrollments, got %d", len(enrols.Msg.Enrollments))
	}

	group, err := client.CreateGroup(ctx, qtest.RequestWithCookie(&qf.Group{
		CourseID: course.ID,
		Name:     "group1",
		Users:    []*qf.User{student1, student3},
		Status:   qf.Group_APPROVED,
	}, cookie))
	if err != nil {
		t.Error(err)
	}
	group2, err := client.CreateGroup(ctx, qtest.RequestWithCookie(&qf.Group{
		CourseID: course.ID,
		Name:     "group2",
		Users:    []*qf.User{student2},
		Status:   qf.Group_APPROVED,
	}, cookie))
	if err != nil {
		t.Error(err)
	}

	lab1 := &qf.Assignment{
		CourseID: course.ID,
		Name:     "lab 1",
		Deadline: qtest.Timestamp(t, "2020-02-23T18:00:00"),
		Order:    1,
	}
	lab2 := &qf.Assignment{
		CourseID:   course.ID,
		Name:       "lab 2",
		Deadline:   qtest.Timestamp(t, "2020-02-23T18:00:00"),
		Order:      2,
		IsGroupLab: true,
	}
	if err = db.CreateAssignment(lab1); err != nil {
		t.Fatal(err)
	}
	if err = db.CreateAssignment(lab2); err != nil {
		t.Fatal(err)
	}
	submission1 := &qf.Submission{
		UserID:       student1.ID,
		AssignmentID: lab1.ID,
		Score:        44,
	}
	submission2 := &qf.Submission{
		UserID:       student2.ID,
		AssignmentID: lab1.ID,
		Score:        66,
	}
	submission3 := &qf.Submission{
		GroupID:      group.Msg.ID,
		AssignmentID: lab2.ID,
		Score:        16,
	}
	submission4 := &qf.Submission{
		GroupID:      group2.Msg.ID,
		AssignmentID: lab2.ID,
		Score:        29,
	}
	if err = db.CreateSubmission(submission1); err != nil {
		t.Fatal(err)
	}
	if err = db.CreateSubmission(submission2); err != nil {
		t.Fatal(err)
	}
	if err = db.CreateSubmission(submission3); err != nil {
		t.Fatal(err)
	}
	if err = db.CreateSubmission(submission4); err != nil {
		t.Fatal(err)
	}

	// submission3 appears before submission2 because the allSubmissions.Links ([]*EnrollmentLink)
	// are returned in the order of enrollments, not the order of submission inserts.
	// Similarly, submission3 also appear at the end because student3 (last to enroll) is in submission3's group.
	wantAllSubmissions := map[uint64]*qf.Submissions{
		1: {}, // admin has no submissions
		2: {Submissions: []*qf.Submission{submission1, submission3}},
		3: {Submissions: []*qf.Submission{submission2, submission4}},
		4: {Submissions: []*qf.Submission{submission3}},
	}
	// wantAllSubmissions := []*qf.Submission{submission1, submission3, submission2, submission4, submission3}
	// wantIndividualSubmissions := []*qf.Submission{submission1, submission2}
	wantIndividualSubmissions := map[uint64]*qf.Submissions{
		1: {}, // admin has no submissions
		2: {Submissions: []*qf.Submission{submission1}},
		3: {Submissions: []*qf.Submission{submission2}},
		4: {}, // student3 has no individual submissions
	}

	// wantGroupSubmissions := []*qf.Submission{submission3, submission4}
	wantGroupSubmissions := map[uint64]*qf.Submissions{
		1: {Submissions: []*qf.Submission{submission3}},
		2: {Submissions: []*qf.Submission{submission4}},
	}

	// get all submissions
	allSubmissions, err := client.GetSubmissionsByCourse(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
		CourseID: course.ID,
		FetchMode: &qf.SubmissionRequest_Type{
			Type: qf.SubmissionRequest_ALL,
		},
	}, cookie))
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(wantAllSubmissions, allSubmissions.Msg.GetSubmissions(), protocmp.Transform()); diff != "" {
		t.Errorf("TestGetSubmissionsByCourse() mismatch (-wantAllSubmissions +gotAllSubmissions):\n%s\n%d:%d", diff, len(wantAllSubmissions), len(allSubmissions.Msg.GetSubmissions()))
	}

	// get only individual submissions
	individualSubmissions, err := client.GetSubmissionsByCourse(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
		CourseID: course.ID,
		FetchMode: &qf.SubmissionRequest_Type{
			Type: qf.SubmissionRequest_USER,
		},
	}, cookie))
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(wantIndividualSubmissions, individualSubmissions.Msg.GetSubmissions(), protocmp.Transform()); diff != "" {
		t.Errorf("TestGetSubmissionsByCourse() mismatch (-wantIndividualSubmissions +gotIndividualSubmissions):\n%s", diff)
	}

	// get only group submissions
	groupSubmissions, err := client.GetSubmissionsByCourse(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
		CourseID: course.ID,
		FetchMode: &qf.SubmissionRequest_Type{
			Type: qf.SubmissionRequest_GROUP,
		},
	}, cookie))
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(wantGroupSubmissions, groupSubmissions.Msg.GetSubmissions(), protocmp.Transform()); diff != "" {
		t.Errorf("TestGetSubmissionsByCourse() mismatch (-wantGroupSubmissions +gotGroupSubmissions):\n%s", diff)
	}
}

func TestGetCourseLabSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm, _ := MockClientWithUser(t, db)

	admin := qtest.CreateFakeUser(t, db)

	course1 := qtest.MockCourses[2]
	course2 := qtest.MockCourses[3]
	if err := db.CreateCourse(admin.ID, course1); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateCourse(admin.ID, course2); err != nil {
		t.Fatal(err)
	}

	student := qtest.CreateFakeUser(t, db)
	enrolC1 := qtest.EnrollUser(t, db, student, course1, qf.Enrollment_STUDENT)
	enrolC2 := qtest.EnrollUser(t, db, student, course2, qf.Enrollment_STUDENT)

	// make labs with similar lab names for both courses
	lab1c1 := &qf.Assignment{
		CourseID:          course1.ID,
		Name:              "lab 1",
		Deadline:          qtest.Timestamp(t, "2020-02-23T18:00:00"),
		Order:             1,
		GradingBenchmarks: []*qf.GradingBenchmark{},
	}

	lab2c1 := &qf.Assignment{
		CourseID:          course1.ID,
		Name:              "lab 2",
		Deadline:          qtest.Timestamp(t, "2020-03-23T18:00:00"),
		Order:             2,
		GradingBenchmarks: []*qf.GradingBenchmark{},
	}
	lab1c2 := &qf.Assignment{
		CourseID:          course2.ID,
		Name:              "lab 1",
		Deadline:          qtest.Timestamp(t, "2020-04-23T18:00:00"),
		Order:             1,
		GradingBenchmarks: []*qf.GradingBenchmark{},
	}
	lab2c2 := &qf.Assignment{
		CourseID:          course2.ID,
		Name:              "lab 2",
		Deadline:          qtest.Timestamp(t, "2020-05-23T18:00:00"),
		Order:             2,
		GradingBenchmarks: []*qf.GradingBenchmark{},
	}
	if err := db.CreateAssignment(lab1c1); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateAssignment(lab2c1); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateAssignment(lab1c2); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateAssignment(lab2c2); err != nil {
		t.Fatal(err)
	}

	buildInfo1 := &score.BuildInfo{
		BuildDate: qtest.Timestamp(t, "2020-02-23T18:00:00"),
		BuildLog:  "runtime error",
		ExecTime:  3,
	}

	buildInfo2 := &score.BuildInfo{
		BuildDate: qtest.Timestamp(t, "2020-02-23T18:00:00"),
		BuildLog:  "runtime error",
		ExecTime:  3,
	}

	wantSubmission1 := &qf.Submission{
		UserID:       student.ID,
		AssignmentID: lab1c1.ID,
		Score:        44,
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
		BuildInfo:    buildInfo1,
	}
	wantSubmission2 := &qf.Submission{
		UserID:       student.ID,
		AssignmentID: lab2c2.ID,
		Score:        66,
		Reviews:      []*qf.Review{},
		Scores:       []*score.Score{},
		BuildInfo:    buildInfo2,
	}
	if err := db.CreateSubmission(wantSubmission1); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateSubmission(wantSubmission2); err != nil {
		t.Fatal(err)
	}

	wantSubmission1.BuildInfo = nil
	wantSubmission2.BuildInfo = nil

	// check that all assignments were saved for the correct courses
	wantAssignments1 := []*qf.Assignment{lab1c1, lab2c1}
	wantAssignments2 := []*qf.Assignment{lab1c2, lab2c2}

	ctx := context.Background()
	cookie := Cookie(t, tm, admin)

	assignments1, err := client.GetAssignments(ctx, qtest.RequestWithCookie(&qf.CourseRequest{
		CourseID: course1.ID,
	}, cookie))
	if err != nil {
		t.Error(err)
	}
	gotAssignments1 := assignments1.Msg.GetAssignments()
	if diff := cmp.Diff(wantAssignments1, gotAssignments1, protocmp.Transform()); diff != "" {
		t.Errorf("GetAssignments() mismatch (-wantAssignments1, +gotAssignments1):\n%s", diff)
	}

	assignments2, err := client.GetAssignments(ctx, qtest.RequestWithCookie(&qf.CourseRequest{
		CourseID: course2.ID,
	}, cookie))
	if err != nil {
		t.Error(err)
	}
	gotAssignments2 := assignments2.Msg.GetAssignments()
	if diff := cmp.Diff(wantAssignments2, gotAssignments2, protocmp.Transform()); diff != "" {
		t.Errorf("GetAssignments() mismatch (-wantAssignments2, +gotAssignments2):\n%s", diff)
	}

	// check that all submissions were saved for the correct labs
	labsForCourse1, err := client.GetSubmissionsByCourse(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
		CourseID:  course1.ID,
		FetchMode: &qf.SubmissionRequest_Type{},
	}, cookie))
	if err != nil {
		t.Error(err)
	}

	labMap := labsForCourse1.Msg.GetSubmissions()
	t.Log(enrolC1)
	if submissions, ok := labMap[enrolC1.ID]; !ok {
		t.Fatalf("GetSubmissionsByCourse() did not return submissions for enrollment ID %d", enrolC1.ID)
	} else {
		labs := submissions.GetSubmissions()
		if len(labs) != 1 {
			t.Fatalf("Expected 1 submission for course 1, got %d", len(labs))
		}
		gotSubmission1 := labs[0]
		if diff := cmp.Diff(wantSubmission1, gotSubmission1, protocmp.Transform()); diff != "" {
			t.Errorf("GetSubmissionsByCourse() mismatch (-wantSubmission1 +gotSubmission1):\n%s", diff)
		}
	}

	labsForCourse2, err := client.GetSubmissionsByCourse(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
		CourseID: course2.ID,
		FetchMode: &qf.SubmissionRequest_Type{
			Type: qf.SubmissionRequest_ALL,
		},
	}, cookie))
	if err != nil {
		t.Error(err)
	}
	labMap = labsForCourse2.Msg.GetSubmissions()
	if submissions, ok := labMap[enrolC2.ID]; !ok {
		t.Fatalf("GetSubmissionsByCourse() did not return submissions for enrollment ID %d", enrolC2.ID)
	} else {
		labs := submissions.GetSubmissions()
		if len(labs) != 1 {
			t.Fatalf("Expected 1 submission for course 2, got %d", len(labs))
		}
		gotSubmission2 := labs[0]
		if diff := cmp.Diff(wantSubmission2, gotSubmission2, protocmp.Transform()); diff != "" {
			t.Errorf("GetSubmissionsByCourse() mismatch (-wantSubmission2 +gotSubmission2):\n%s", diff)
		}
	}

	// check that buildInformation is not included when not requested
	labsForCourse3, err := client.GetSubmissionsByCourse(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
		CourseID: course1.ID,
		FetchMode: &qf.SubmissionRequest_Type{
			Type: qf.SubmissionRequest_ALL,
		},
	}, cookie))
	if err != nil {
		t.Error(err)
	}
	for _, labLink := range labsForCourse3.Msg.GetSubmissions() {
		for _, submission := range labLink.GetSubmissions() {
			if submission.GetBuildInfo() != nil {
				t.Errorf("Expected build info to be nil, got %+v", submission.GetBuildInfo())
			}
		}
	}

	labsForCourse4, err := client.GetSubmissionsByCourse(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
		CourseID: course2.ID,
		FetchMode: &qf.SubmissionRequest_Type{
			Type: qf.SubmissionRequest_ALL,
		},
	}, cookie))
	if err != nil {
		t.Error(err)
	}
	for _, labLink := range labsForCourse4.Msg.GetSubmissions() {
		for _, submission := range labLink.GetSubmissions() {
			if submission != nil {
				if submission.GetBuildInfo() != nil {
					t.Errorf("Expected build info to be nil, got %+v", submission.GetBuildInfo())
				}
			}
		}
	}

	// check that no submissions will be returned for a wrong course ID
	if _, err = client.GetSubmissionsByCourse(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
		CourseID: 234,
	}, cookie)); err == nil {
		t.Error("Expected 'no submissions found'")
	}
}

func TestCreateApproveList(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm, _ := MockClientWithUser(t, db)

	admin := qtest.CreateFakeUser(t, db)

	course := qtest.MockCourses[2]
	if err := db.CreateCourse(admin.ID, course); err != nil {
		t.Fatal(err)
	}
	student1 := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "Leslie Lamport", Login: "Leslie Lamport"})
	student2 := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "Hein Meling", Login: "Hein Meling"})
	student3 := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "John Doe", Login: "John Doe"})
	enrollStudent1 := qtest.EnrollUser(t, db, student1, course, qf.Enrollment_STUDENT)
	enrollStudent2 := qtest.EnrollUser(t, db, student2, course, qf.Enrollment_STUDENT)
	enrollStudent3 := qtest.EnrollUser(t, db, student3, course, qf.Enrollment_STUDENT)

	assignments := []*qf.Assignment{
		{
			CourseID: course.ID,
			Name:     "lab 1",
			Deadline: qtest.Timestamp(t, "2020-02-23T18:00:00"),
			Order:    1,
		},
		{
			CourseID: course.ID,
			Name:     "lab 2",
			Deadline: qtest.Timestamp(t, "2020-03-23T18:00:00"),
			Order:    2,
		},
		{
			CourseID: course.ID,
			Name:     "lab 3",
			Deadline: qtest.Timestamp(t, "2020-04-23T18:00:00"),
			Order:    3,
		},
		{
			CourseID: course.ID,
			Name:     "lab 4",
			Deadline: qtest.Timestamp(t, "2020-05-23T18:00:00"),
			Order:    4,
		},
	}
	for _, a := range assignments {
		if err := db.CreateAssignment(a); err != nil {
			t.Fatal(err)
		}
	}

	submissions := []*qf.Submission{
		{
			UserID:       student1.ID,
			AssignmentID: assignments[0].ID,
			Status:       qf.Submission_APPROVED,
		},
		{
			UserID:       student1.ID,
			AssignmentID: assignments[1].ID,
			Status:       qf.Submission_APPROVED,
		},
		{
			UserID:       student1.ID,
			AssignmentID: assignments[2].ID,
			Status:       qf.Submission_APPROVED,
		},
		{
			UserID:       student1.ID,
			AssignmentID: assignments[3].ID,
			Status:       qf.Submission_APPROVED,
		},
		{
			UserID:       student2.ID,
			AssignmentID: assignments[0].ID,
			Status:       qf.Submission_APPROVED,
		},
		{
			UserID:       student2.ID,
			AssignmentID: assignments[2].ID,
			Status:       qf.Submission_APPROVED,
		},
		{
			UserID:       student2.ID,
			AssignmentID: assignments[3].ID,
			Status:       qf.Submission_APPROVED,
		},
		{
			UserID:       student3.ID,
			AssignmentID: assignments[0].ID,
			Status:       qf.Submission_APPROVED,
		},
		{
			UserID:       student3.ID,
			AssignmentID: assignments[1].ID,
			Status:       qf.Submission_REJECTED,
		},
		{
			UserID:       student3.ID,
			AssignmentID: assignments[2].ID,
			Status:       qf.Submission_REVISION,
		},
	}
	for _, s := range submissions {
		if err := db.CreateSubmission(s); err != nil {
			t.Fatal(err)
		}
	}

	testCases := []struct {
		student          *qf.Enrollment
		minNumApproved   int
		expectedApproved bool
	}{
		{
			student:          enrollStudent1,
			minNumApproved:   4,
			expectedApproved: true,
		},
		{
			student:          enrollStudent1,
			minNumApproved:   3,
			expectedApproved: true,
		},
		{
			student:          enrollStudent2,
			minNumApproved:   4,
			expectedApproved: false,
		},
		{
			student:          enrollStudent2,
			minNumApproved:   3,
			expectedApproved: true,
		},
		{
			student:          enrollStudent2,
			minNumApproved:   2,
			expectedApproved: true,
		},
		{
			student:          enrollStudent3,
			minNumApproved:   4,
			expectedApproved: false,
		},
		{
			student:          enrollStudent3,
			minNumApproved:   3,
			expectedApproved: false,
		},
		{
			student:          enrollStudent3,
			minNumApproved:   2,
			expectedApproved: false,
		},
		{
			student:          enrollStudent3,
			minNumApproved:   1,
			expectedApproved: true,
		},
	}

	ctx := context.Background()
	cookie := Cookie(t, tm, admin)

	gotSubmissions, err := client.GetSubmissionsByCourse(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
		CourseID: course.ID,
		FetchMode: &qf.SubmissionRequest_Type{
			Type: qf.SubmissionRequest_ALL,
		},
	}, cookie))
	if err != nil {
		t.Error(err)
	}
	for id, submissions := range gotSubmissions.Msg.GetSubmissions() {
		if id == admin.ID {
			continue
		}
		approved := make([]bool, len(submissions.Submissions))
		for i, s := range submissions.Submissions {
			approved[i] = s.IsApproved()
		}
		for _, test := range testCases {
			if test.student.ID == id {
				got := isApproved(test.minNumApproved, approved)
				if got != test.expectedApproved {
					t.Errorf("isApproved(%d, %v) = %t, expected %t", test.minNumApproved, approved, got, test.expectedApproved)
				}
			}
		}
		t.Logf("%d\t%t", id, isApproved(4, approved))
	}
}

func TestReleaseApproveAll(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm, _ := MockClientWithUser(t, db)

	admin := qtest.CreateFakeUser(t, db)

	course := qtest.MockCourses[2]
	if err := db.CreateCourse(admin.ID, course); err != nil {
		t.Fatal(err)
	}
	student1 := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "Leslie Lamport", Login: "Leslie Lamport"})
	student2 := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "Hein Meling", Login: "Hein Meling"})
	student3 := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "John Doe", Login: "John Doe"})
	qtest.EnrollStudent(t, db, student1, course)
	qtest.EnrollStudent(t, db, student2, course)
	qtest.EnrollStudent(t, db, student3, course)

	assignments := []*qf.Assignment{
		{
			CourseID:  course.ID,
			Name:      "lab 1",
			Deadline:  qtest.Timestamp(t, "2020-02-23T18:00:00"),
			Order:     1,
			Reviewers: 1,
		},
		{
			CourseID:  course.ID,
			Name:      "lab 2",
			Deadline:  qtest.Timestamp(t, "2020-03-23T18:00:00"),
			Order:     2,
			Reviewers: 1,
		},
	}

	for _, assignment := range assignments {
		if err := db.CreateAssignment(assignment); err != nil {
			t.Fatal(err)
		}
	}

	benchmarks := []*qf.GradingBenchmark{
		{
			AssignmentID: assignments[0].ID,
			Heading:      "lab 1",
			Criteria: []*qf.GradingCriterion{
				{
					BenchmarkID: 1,
					Description: "Test 1",
					Points:      10,
				},
				{
					BenchmarkID: 2,
					Description: "Test 2",
					Points:      10,
				},
			},
		},
		{
			AssignmentID: assignments[1].ID,
			Heading:      "lab 2",
			Criteria: []*qf.GradingCriterion{
				{
					BenchmarkID: 3,
					Description: "Test 3",
				},
				{
					BenchmarkID: 4,
					Description: "Test 4",
				},
			},
		},
	}

	for _, benchmark := range benchmarks {
		if err := db.CreateBenchmark(benchmark); err != nil {
			t.Fatal(err)
		}
	}

	submissions := []*qf.Submission{
		{
			UserID:       student1.ID,
			AssignmentID: assignments[0].ID,
		},
		{
			UserID:       student1.ID,
			AssignmentID: assignments[1].ID,
		},
		{
			UserID:       student2.ID,
			AssignmentID: assignments[0].ID,
		},
		{
			UserID:       student2.ID,
			AssignmentID: assignments[1].ID,
		},
		{
			UserID:       student3.ID,
			AssignmentID: assignments[0].ID,
		},
		{
			UserID:       student3.ID,
			AssignmentID: assignments[1].ID,
		},
	}

	ctx := context.Background()
	cookie := Cookie(t, tm, admin)

	reviews := []*qf.Review{}
	for _, s := range submissions {
		if err := db.CreateSubmission(s); err != nil {
			t.Fatal(err)
		}
		review, err := client.CreateReview(ctx, qtest.RequestWithCookie(&qf.ReviewRequest{
			CourseID: course.ID,
			Review: &qf.Review{
				SubmissionID: s.ID,
				ReviewerID:   admin.GetID(),
			},
		}, cookie))
		if err != nil {
			t.Error(err)
		}
		reviews = append(reviews, review.Msg)
	}

	for _, r := range reviews {
		for _, benchmark := range r.GradingBenchmarks {
			for _, criterion := range benchmark.Criteria {
				criterion.Grade = qf.GradingCriterion_PASSED
			}
		}

		// Update the review. This will also update the submission score for the related submission.
		_, err := client.UpdateReview(ctx, qtest.RequestWithCookie(&qf.ReviewRequest{
			CourseID: uint64(course.ID),
			Review:   r,
		}, cookie))
		if err != nil {
			t.Error(err)
		}
	}

	gotSubmissions1, err := db.GetSubmissions(&qf.Submission{
		AssignmentID: assignments[0].ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, submission := range gotSubmissions1 {
		// All submissions should have a score of 20
		if submission.Score != 20 {
			t.Errorf("Expected score 20, got %d", submission.Score)
		}
	}

	gotSubmissions2, err := db.GetSubmissions(&qf.Submission{
		AssignmentID: assignments[1].ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, submission := range gotSubmissions2 {
		// All submissions should have a score of 100
		if submission.Score != 100 {
			t.Errorf("Expected score 100, got %d", submission.Score)
		}
	}

	// Attempt to release all submissions with score >= 80
	if _, err = client.UpdateSubmissions(ctx, qtest.RequestWithCookie(&qf.UpdateSubmissionsRequest{
		CourseID:     course.ID,
		AssignmentID: assignments[0].ID,
		Release:      true,
		ScoreLimit:   80,
	}, cookie)); err != nil {
		t.Error(err)
	}

	gotSubmissions3, err := db.GetSubmissions(&qf.Submission{
		AssignmentID: assignments[0].ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Only submissions with score >= 80 should be released
	// All submissions for assignment 1 should have score == 20, and not be released
	for _, submission := range gotSubmissions3 {
		if submission.Released {
			t.Errorf("Expected submission to not be released")
		}
	}

	// We want to make sure that submissions received by the student do not leak data
	studentCookie := Cookie(t, tm, student1)

	gotStudentSubmissions, err := client.GetSubmissions(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
		CourseID: course.ID,
		FetchMode: &qf.SubmissionRequest_UserID{
			UserID: student1.GetID(),
		},
	}, studentCookie))
	if err != nil {
		t.Error(err)
	}

	for _, submission := range gotStudentSubmissions.Msg.Submissions {
		// For submissions that have not been released
		// the score should be 0, and any reviews should be nil
		if submission.Released || submission.Score > 0 || submission.Reviews != nil || submission.Status != qf.Submission_NONE {
			t.Errorf("Expected submission to not be released, have score, and have no reviews")
		}
	}

	// Attempt to release all submissions with score >= 80
	if _, err = client.UpdateSubmissions(ctx, qtest.RequestWithCookie(&qf.UpdateSubmissionsRequest{
		CourseID:     course.ID,
		AssignmentID: assignments[1].ID,
		Release:      true,
		ScoreLimit:   80,
	}, cookie)); err != nil {
		t.Error(err)
	}

	// All submissions for assignment 2 should have score == 100, and be released
	gotSubmissions4, err := db.GetSubmissions(&qf.Submission{
		AssignmentID: assignments[1].ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, submission := range gotSubmissions4 {
		if !submission.Released {
			t.Errorf("Expected submission to be released")
		}
	}

	// Approve all submissions for assignment 1 with score >= 80
	if _, err = client.UpdateSubmissions(ctx, qtest.RequestWithCookie(&qf.UpdateSubmissionsRequest{
		CourseID:     course.ID,
		AssignmentID: assignments[1].ID,
		Approve:      true,
		ScoreLimit:   80,
	}, cookie)); err != nil {
		t.Error(err)
	}

	gotSubmissions5, err := db.GetSubmissions(&qf.Submission{
		AssignmentID: assignments[1].ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, submission := range gotSubmissions5 {
		// Check that all submissions for assignment 1 have been approved
		if submission.Status != qf.Submission_APPROVED {
			t.Errorf("Expected submission to be approved")
		}
	}

	gotStudentSubmissions, err = client.GetSubmissions(ctx, qtest.RequestWithCookie(&qf.SubmissionRequest{
		CourseID: course.ID,
		FetchMode: &qf.SubmissionRequest_UserID{
			UserID: student1.GetID(),
		},
	}, studentCookie))
	if err != nil {
		t.Error(err)
	}

	for _, submission := range gotStudentSubmissions.Msg.Submissions {
		// Submissions for assignment 1 should not be released, have score, or reviews.
		if submission.ID == assignments[0].ID && (submission.Released || submission.Score > 0 || submission.Reviews != nil) {
			t.Errorf("Expected submission to not be released, have score, and have no reviews")
		}

		// Submissions for assignment 2 should be released, have score, and have reviews
		if submission.ID == assignments[1].ID && !(submission.Released || submission.Score > 0 || submission.Reviews != nil || submission.Status != qf.Submission_NONE) {
			t.Error("Expected submission to be released, have score, and have reviews", submission.Score, submission.Reviews, submission.Released)
		}
	}
}

func isApproved(requirements int, approved []bool) bool {
	for _, a := range approved {
		if a {
			requirements--
		}
	}
	return requirements <= 0
}
