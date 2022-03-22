package web_test

import (
	"context"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/autograde/quickfeed/kit/score"
	"github.com/autograde/quickfeed/log"
	"github.com/autograde/quickfeed/scm"
	"github.com/autograde/quickfeed/web"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestSubmissionsAccess(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)

	teacher := qtest.CreateFakeUser(t, db, 2)
	err := db.UpdateUser(&pb.User{ID: teacher.ID, IsAdmin: true})
	if err != nil {
		t.Fatal(err)
	}

	var course pb.Course
	course.Provider = "fake"
	// only created 1 directory, if we had created two directories ID would be 2
	course.OrganizationID = 1
	if err := db.CreateCourse(teacher.ID, &course); err != nil {
		t.Fatal(err)
	}

	student1 := qtest.CreateFakeUser(t, db, 3)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: student1.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   student1.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	student2 := qtest.CreateFakeUser(t, db, 4)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: student2.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   student2.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	student3 := qtest.CreateFakeUser(t, db, 5)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: student3.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}

	fakeProvider, scms := qtest.FakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	ctx := qtest.WithUserContext(context.Background(), teacher)

	_, err = fakeProvider.CreateOrganization(context.Background(), &scm.OrganizationOptions{Path: "path", Name: "name"})
	if err != nil {
		t.Fatal(err)
	}

	users := []*pb.User{student1, student2}
	group_req := &pb.Group{Name: "TestGroup", CourseID: course.ID, Users: users}

	_, err = ags.CreateGroup(ctx, group_req)
	if err != nil {
		t.Fatal(err)
	}

	// at this stage we have a course teacher, two students enrolled in the course in the same group,
	// and one student and admin not affiliated with the course

	if err = db.CreateAssignment(&pb.Assignment{
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

	if err = db.CreateAssignment(&pb.Assignment{
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

	if err = db.CreateSubmission(&pb.Submission{
		AssignmentID: 1,
		UserID:       student1.ID,
	}); err != nil {
		t.Fatal(err)
	}

	if err = db.CreateSubmission(&pb.Submission{
		AssignmentID: 2,
		GroupID:      1,
	}); err != nil {
		t.Fatal(err)
	}

	if err = db.CreateSubmission(&pb.Submission{
		AssignmentID: 1,
		UserID:       student3.ID,
	}); err != nil {
		t.Fatal(err)
	}

	// check that all three submissions have been successfully added to the database
	submission1, err := db.GetSubmission(&pb.Submission{
		AssignmentID: 1,
		UserID:       student1.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	submission2, err := db.GetSubmission(&pb.Submission{
		AssignmentID: 1,
		UserID:       student3.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	submission3, err := db.GetSubmission(&pb.Submission{
		AssignmentID: 2,
		GroupID:      1,
	})
	if err != nil {
		t.Fatal(err)
	}

	allSubmissions := []*pb.Submission{submission1, submission2, submission3}
	wantLatestSubmissions := []*pb.Submission{submission2, submission3}

	// there must be exactly three submissions for given course and assignment in the database
	if len(allSubmissions) != 3 {
		t.Errorf("Expected 3 submissions, got %d: %+v", len(allSubmissions), allSubmissions)
	}

	// teacher must be able to access all of the latest course submissions
	submissions, err := ags.GetSubmissions(ctx, &pb.SubmissionRequest{CourseID: course.ID})
	if err != nil {
		t.Fatal(err)
	}
	gotSubmissions := submissions.GetSubmissions()
	if diff := cmp.Diff(wantLatestSubmissions, gotSubmissions, protocmp.Transform()); diff != "" {
		t.Errorf("ags.GetSubmissions() mismatch (-wantLatestSubmissions, +gotSubmissions):\n%s", diff)
	}

	// admin not enrolled in the course must not be able to access any course submissions
	ctx = qtest.WithUserContext(context.Background(), admin)
	submissions, err = ags.GetSubmissions(ctx, &pb.SubmissionRequest{CourseID: course.ID})
	if err == nil {
		t.Error("Expected error: user not enrolled")
	}
	if len(submissions.GetSubmissions()) > 0 {
		t.Errorf("Not enrolled admin should not see any submissions, got submissions: %v+ ", submissions.GetSubmissions())
	}

	// enroll admin as course student
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: admin.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   admin.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	submissions, err = ags.GetSubmissions(ctx, &pb.SubmissionRequest{CourseID: course.ID})
	if err != nil {
		t.Fatal(err)
	}
	// enrolled as student, admin must be able to access all course submissions
	gotSubmissions = submissions.GetSubmissions()
	if diff := cmp.Diff(wantLatestSubmissions, gotSubmissions, protocmp.Transform()); diff != "" {
		t.Errorf("ags.GetSubmissions() mismatch (-wantLatestSubmissions, +gotSubmissions):\n%s", diff)
	}

	// the first student must be able to access own submissions as well as submissions made by group he has membership in
	ctx = qtest.WithUserContext(context.Background(), student1)

	personalSubmission, err := ags.GetSubmissions(ctx, &pb.SubmissionRequest{CourseID: course.ID, UserID: student1.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(personalSubmission.GetSubmissions()) != 1 {
		t.Error("Expected one submission, got ", len(personalSubmission.GetSubmissions()))
	}
	groupSubmission, err := ags.GetSubmissions(ctx, &pb.SubmissionRequest{CourseID: course.ID, GroupID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(groupSubmission.GetSubmissions()) != 1 {
		t.Error("Expected one submission, got ", len(groupSubmission.GetSubmissions()))
	}

	wantSubmissions := []*pb.Submission{submission1, submission3}
	gotStudent1Submissions := []*pb.Submission{personalSubmission.GetSubmissions()[0], groupSubmission.GetSubmissions()[0]}

	if diff := cmp.Diff(wantSubmissions, gotStudent1Submissions, protocmp.Transform()); diff != "" {
		t.Errorf("ags.GetSubmissions() mismatch (-wantSubmissions, +gotStudent1Submissions):\n%s", diff)
	}

	// the second student should not be able to access the submission by student1
	ctx = qtest.WithUserContext(context.Background(), student2)
	personalSubmission, err = ags.GetSubmissions(ctx, &pb.SubmissionRequest{CourseID: course.ID, UserID: student1.ID})
	if err == nil || personalSubmission != nil {
		t.Error("Expected error: only owner and teachers can get submissions")
	}

	// the second student should no longer be able to access group submissions when removed from the group

	if err = db.UpdateGroup(&pb.Group{
		ID:       1,
		CourseID: course.ID,
		Users:    []*pb.User{student1},
	}); err != nil {
		t.Fatal(err)
	}

	groupSubmission, err = ags.GetSubmissions(ctx, &pb.SubmissionRequest{CourseID: course.ID, GroupID: 1})
	if err == nil || groupSubmission != nil {
		t.Error("Expected error: only owner and teachers can get submissions")
	}

	// the third student (not enrolled in the course) should not be able to access submission even if it belongs to that student
	ctx = qtest.WithUserContext(context.Background(), student3)
	personalSubmission, err = ags.GetSubmissions(ctx, &pb.SubmissionRequest{CourseID: course.ID, UserID: student3.ID})
	if err == nil || personalSubmission != nil {
		t.Error("Expected error: only owner and teachers can get submissions")
	}
}

func TestApproveSubmission(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)

	course := allCourses[0]
	err := db.CreateCourse(admin.ID, course)
	if err != nil {
		t.Fatal(err)
	}

	student := qtest.CreateFakeUser(t, db, 2)
	if err := db.CreateEnrollment(&pb.Enrollment{UserID: student.ID, CourseID: course.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateEnrollment(&pb.Enrollment{
		UserID:   student.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}); err != nil {
		t.Fatal(err)
	}

	lab := &pb.Assignment{
		CourseID:   course.ID,
		Name:       "test lab",
		ScriptFile: "go.sh",
		Order:      1,
	}
	if err = db.CreateAssignment(lab); err != nil {
		t.Fatal(err)
	}

	wantSubmission := &pb.Submission{
		AssignmentID: lab.ID,
		UserID:       student.ID,
		Score:        17,
	}
	if err = db.CreateSubmission(wantSubmission); err != nil {
		t.Fatal(err)
	}

	fakeProvider, scms := qtest.FakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	ctx := qtest.WithUserContext(context.Background(), admin)

	_, err = fakeProvider.CreateOrganization(context.Background(), &scm.OrganizationOptions{Path: "path", Name: "name"})
	if err != nil {
		t.Fatal(err)
	}

	if _, err = ags.UpdateSubmission(ctx, &pb.UpdateSubmissionRequest{
		SubmissionID: wantSubmission.ID,
		CourseID:     course.ID,
		Status:       pb.Submission_APPROVED,
	}); err != nil {
		t.Fatal(err)
	}

	gotApprovedSubmission, err := db.GetSubmission(&pb.Submission{ID: wantSubmission.ID})
	if err != nil {
		t.Fatal(err)
	}
	wantSubmission.Status = pb.Submission_APPROVED
	wantSubmission.ApprovedDate = gotApprovedSubmission.ApprovedDate

	if diff := cmp.Diff(wantSubmission, gotApprovedSubmission, protocmp.Transform()); diff != "" {
		t.Errorf("ags.UpdateSubmission(approve) mismatch (-wantSubmission, +gotApprovedSubmission):\n%s", diff)
	}

	if _, err = ags.UpdateSubmission(ctx, &pb.UpdateSubmissionRequest{
		SubmissionID: wantSubmission.ID,
		CourseID:     course.ID,
		Status:       pb.Submission_REJECTED,
	}); err != nil {
		t.Fatal(err)
	}

	gotRejectedSubmission, err := db.GetSubmission(&pb.Submission{ID: wantSubmission.ID})
	if err != nil {
		t.Fatal(err)
	}
	wantSubmission.Status = pb.Submission_REJECTED
	// Note that the approved date is not set when the submission is rejected

	if diff := cmp.Diff(wantSubmission, gotRejectedSubmission, protocmp.Transform()); diff != "" {
		t.Errorf("ags.UpdateSubmission(reject) mismatch (-wantSubmission, +gotRejectedSubmission):\n%s", diff)
	}
}

func TestGetSubmissionsByCourse(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)
	course := allCourses[2]
	qtest.CreateCourse(t, db, admin, course)
	student1 := qtest.CreateFakeUser(t, db, 2)
	student2 := qtest.CreateFakeUser(t, db, 3)
	student3 := qtest.CreateFakeUser(t, db, 4)

	fakeProvider, scms := qtest.FakeProviderMap(t)
	ags := web.NewAutograderService(log.Zap(false), db, scms, web.BaseHookOptions{}, &ci.Local{})
	ctx := qtest.WithUserContext(context.Background(), admin)
	if _, err := fakeProvider.CreateOrganization(context.Background(), &scm.OrganizationOptions{Path: "path", Name: "name"}); err != nil {
		t.Fatal(err)
	}
	qtest.EnrollStudent(t, db, student1, course)
	qtest.EnrollStudent(t, db, student2, course)
	qtest.EnrollStudent(t, db, student3, course)

	enrols, err := ags.GetEnrollmentsByCourse(ctx, &pb.EnrollmentRequest{CourseID: course.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(enrols.Enrollments) != 4 {
		t.Errorf("expected 4 enrollments, got %d", len(enrols.Enrollments))
	}

	group, err := ags.CreateGroup(ctx, &pb.Group{
		CourseID: course.ID,
		Name:     "group1",
		Users:    []*pb.User{student1, student3},
		Status:   pb.Group_APPROVED,
	})
	if err != nil {
		t.Fatal(err)
	}
	group2, err := ags.CreateGroup(ctx, &pb.Group{
		CourseID: course.ID,
		Name:     "group2",
		Users:    []*pb.User{student2},
		Status:   pb.Group_APPROVED,
	})
	if err != nil {
		t.Fatal(err)
	}

	lab1 := &pb.Assignment{
		CourseID: course.ID,
		Name:     "lab 1",
		Deadline: "2020-02-23T18:00",
		Order:    1,
	}
	lab2 := &pb.Assignment{
		CourseID:   course.ID,
		Name:       "lab 2",
		Deadline:   "2020-02-23T18:00",
		Order:      2,
		IsGroupLab: true,
	}
	if err = db.CreateAssignment(lab1); err != nil {
		t.Fatal(err)
	}
	if err = db.CreateAssignment(lab2); err != nil {
		t.Fatal(err)
	}
	submission1 := &pb.Submission{
		UserID:       student1.ID,
		AssignmentID: lab1.ID,
		Score:        44,
	}
	submission2 := &pb.Submission{
		UserID:       student2.ID,
		AssignmentID: lab1.ID,
		Score:        66,
	}
	submission3 := &pb.Submission{
		GroupID:      group.ID,
		AssignmentID: lab2.ID,
		Score:        16,
	}
	submission4 := &pb.Submission{
		GroupID:      group2.ID,
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
	wantAllSubmissions := []*pb.Submission{submission1, submission3, submission2, submission4, submission3}
	wantIndividualSubmissions := []*pb.Submission{submission1, submission2}
	wantGroupSubmissions := []*pb.Submission{submission3, submission4}

	// default is all submissions
	submissions, err := ags.GetSubmissionsByCourse(ctx, &pb.SubmissionsForCourseRequest{CourseID: course.ID})
	if err != nil {
		t.Fatal(err)
	}
	// be specific that we want all submissions
	allSubmissions, err := ags.GetSubmissionsByCourse(ctx, &pb.SubmissionsForCourseRequest{
		CourseID: course.ID,
		Type:     pb.SubmissionsForCourseRequest_ALL,
	})
	if err != nil {
		t.Fatal(err)
	}
	// check that default and all submissions (SubmissionsForCourseRequest_ALL) are the same
	if diff := cmp.Diff(submissions, allSubmissions, protocmp.Transform()); diff != "" {
		t.Errorf("ags.TestGetSubmissionsByCourse() mismatch (-submissions +allSubmissions):\n%s", diff)
	}

	gotAllSubmissions := []*pb.Submission{}
	for _, s := range allSubmissions.Links {
		for _, subLink := range s.Submissions {
			if subLink.Submission != nil {
				gotAllSubmissions = append(gotAllSubmissions, subLink.Submission)
			}
		}
	}
	if diff := cmp.Diff(wantAllSubmissions, gotAllSubmissions, protocmp.Transform()); diff != "" {
		t.Errorf("ags.TestGetSubmissionsByCourse() mismatch (-wantAllSubmissions +gotAllSubmissions):\n%s", diff)
	}

	// get only individual submissions
	individualSubmissions, err := ags.GetSubmissionsByCourse(ctx, &pb.SubmissionsForCourseRequest{
		CourseID: course.ID,
		Type:     pb.SubmissionsForCourseRequest_INDIVIDUAL,
	})
	if err != nil {
		t.Fatal(err)
	}

	gotIndividualSubmissions := []*pb.Submission{}
	for _, s := range individualSubmissions.Links {
		for _, subLink := range s.Submissions {
			if subLink.Submission != nil {
				gotIndividualSubmissions = append(gotIndividualSubmissions, subLink.Submission)
			}
		}
	}
	if diff := cmp.Diff(wantIndividualSubmissions, gotIndividualSubmissions, protocmp.Transform()); diff != "" {
		t.Errorf("ags.TestGetSubmissionsByCourse() mismatch (-wantIndividualSubmissions +gotIndividualSubmissions):\n%s", diff)
	}

	// get only group submissions
	groupSubmissions, err := ags.GetSubmissionsByCourse(ctx, &pb.SubmissionsForCourseRequest{
		CourseID: course.ID,
		Type:     pb.SubmissionsForCourseRequest_GROUP,
	})
	if err != nil {
		t.Fatal(err)
	}

	gotGroupSubmissions := []*pb.Submission{}
	for _, s := range groupSubmissions.Links {
		for _, subLink := range s.Submissions {
			if subLink.Submission != nil {
				gotGroupSubmissions = append(gotGroupSubmissions, subLink.Submission)
			}
		}
	}
	if diff := cmp.Diff(wantGroupSubmissions, gotGroupSubmissions, protocmp.Transform()); diff != "" {
		t.Errorf("ags.TestGetSubmissionsByCourse() mismatch (-wantGroupSubmissions +gotGroupSubmissions):\n%s", diff)
	}
}

func TestGetCourseLabSubmissions(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)

	course1 := allCourses[2]
	course2 := allCourses[3]
	if err := db.CreateCourse(admin.ID, course1); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateCourse(admin.ID, course2); err != nil {
		t.Fatal(err)
	}

	student := qtest.CreateFakeUser(t, db, 2)
	qtest.EnrollStudent(t, db, student, course1)
	qtest.EnrollStudent(t, db, student, course2)

	// make labs with similar lab names for both courses
	lab1c1 := &pb.Assignment{
		CourseID:          course1.ID,
		Name:              "lab 1",
		ScriptFile:        "go.sh",
		Deadline:          "2020-02-23T18:00:00",
		Order:             1,
		GradingBenchmarks: []*pb.GradingBenchmark{},
	}

	lab2c1 := &pb.Assignment{
		CourseID:          course1.ID,
		Name:              "lab 2",
		ScriptFile:        "go.sh",
		Deadline:          "2020-03-23T18:00:00",
		Order:             2,
		GradingBenchmarks: []*pb.GradingBenchmark{},
	}
	lab1c2 := &pb.Assignment{
		CourseID:          course2.ID,
		Name:              "lab 1",
		ScriptFile:        "go.sh",
		Deadline:          "2020-04-23T18:00:00",
		Order:             1,
		GradingBenchmarks: []*pb.GradingBenchmark{},
	}
	lab2c2 := &pb.Assignment{
		CourseID:          course2.ID,
		Name:              "lab 2",
		ScriptFile:        "go.sh",
		Deadline:          "2020-05-23T18:00:00",
		Order:             2,
		GradingBenchmarks: []*pb.GradingBenchmark{},
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
		BuildDate: "2020-02-23T18:00:00",
		BuildLog:  "runtime error",
		ExecTime:  3,
	}

	buildInfo2 := &score.BuildInfo{
		BuildDate: "2020-02-23T18:00:00",
		BuildLog:  "runtime error",
		ExecTime:  3,
	}

	wantSubmission1 := &pb.Submission{
		UserID:       student.ID,
		AssignmentID: lab1c1.ID,
		Score:        44,
		Reviews:      []*pb.Review{},
		Scores:       []*score.Score{},
		BuildInfo:    buildInfo1,
	}
	wantSubmission2 := &pb.Submission{
		UserID:       student.ID,
		AssignmentID: lab2c2.ID,
		Score:        66,
		Reviews:      []*pb.Review{},
		Scores:       []*score.Score{},
		BuildInfo:    buildInfo2,
	}
	if err := db.CreateSubmission(wantSubmission1); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateSubmission(wantSubmission2); err != nil {
		t.Fatal(err)
	}

	fakeProvider, scms := qtest.FakeProviderMap(t)
	ags := web.NewAutograderService(log.Zap(false), db, scms, web.BaseHookOptions{}, &ci.Local{})
	ctx := qtest.WithUserContext(context.Background(), admin)

	_, err := fakeProvider.CreateOrganization(context.Background(), &scm.OrganizationOptions{Path: "path", Name: "name"})
	if err != nil {
		t.Fatal(err)
	}

	// check that all assignments were saved for the correct courses
	wantAssignments1 := []*pb.Assignment{lab1c1, lab2c1}
	wantAssignments2 := []*pb.Assignment{lab1c2, lab2c2}

	assignments1, err := ags.GetAssignments(ctx, &pb.CourseRequest{CourseID: course1.ID})
	if err != nil {
		t.Fatal(err)
	}
	gotAssignments1 := assignments1.GetAssignments()
	if diff := cmp.Diff(wantAssignments1, gotAssignments1, protocmp.Transform()); diff != "" {
		t.Errorf("ags.GetAssignments() mismatch (-wantAssignments1, +gotAssignments1):\n%s", diff)
	}

	assignments2, err := ags.GetAssignments(ctx, &pb.CourseRequest{CourseID: course2.ID})
	if err != nil {
		t.Fatal(err)
	}
	gotAssignments2 := assignments2.GetAssignments()
	if diff := cmp.Diff(wantAssignments2, gotAssignments2, protocmp.Transform()); diff != "" {
		t.Errorf("ags.GetAssignments() mismatch (-wantAssignments2, +gotAssignments2):\n%s", diff)
	}

	// check that all submissions were saved for the correct labs
	labsForCourse1, err := ags.GetSubmissionsByCourse(ctx, &pb.SubmissionsForCourseRequest{CourseID: course1.ID, Type: pb.SubmissionsForCourseRequest_ALL, WithBuildInfo: true})
	if err != nil {
		t.Fatal(err)
	}

	for _, enrolLink := range labsForCourse1.GetLinks() {
		if enrolLink.GetEnrollment().GetUserID() == student.ID {
			labs := enrolLink.GetSubmissions()
			if len(labs) != 2 {
				t.Fatalf("Expected 2 submission links for course 1, got %d", len(labs))
			}
			gotSubmission1 := labs[0].GetSubmission()
			if diff := cmp.Diff(wantSubmission1, gotSubmission1, protocmp.Transform()); diff != "" {
				t.Errorf("ags.GetSubmissionsByCourse() mismatch (-wantSubmission1 +gotSubmission1):\n%s", diff)
			}
		}
	}

	labsForCourse2, err := ags.GetSubmissionsByCourse(ctx, &pb.SubmissionsForCourseRequest{CourseID: course2.ID, WithBuildInfo: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, labLink := range labsForCourse2.GetLinks() {
		if labLink.GetEnrollment().GetUserID() == student.ID {
			labs := labLink.GetSubmissions()
			if len(labs) != 2 {
				t.Fatalf("Expected 2 submission for course 1, got %d", len(labs))
			}
			gotSubmission2 := labs[1].GetSubmission()
			if diff := cmp.Diff(wantSubmission2, gotSubmission2, protocmp.Transform()); diff != "" {
				t.Errorf("ags.GetSubmissionsByCourse() mismatch (-wantSubmission2 +gotSubmission2):\n%s", diff)
			}
		}
	}

	// check that buildInformation is not included when not requested
	labsForCourse3, err := ags.GetSubmissionsByCourse(ctx, &pb.SubmissionsForCourseRequest{CourseID: course1.ID, WithBuildInfo: false})
	if err != nil {
		t.Fatal(err)
	}
	for _, labLink := range labsForCourse3.GetLinks() {
		for _, submission := range labLink.GetSubmissions() {
			if submission.Submission.GetBuildInfo().GetBuildLog() != "" {
				t.Errorf("Expected build log: \"\", got %+v", submission.GetSubmission().GetBuildInfo().GetBuildLog())
			}
		}
	}

	labsForCourse4, err := ags.GetSubmissionsByCourse(ctx, &pb.SubmissionsForCourseRequest{CourseID: course2.ID, WithBuildInfo: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, labLink := range labsForCourse4.GetLinks() {
		for _, submission := range labLink.GetSubmissions() {
			if submission.GetSubmission() != nil {
				if submission.GetSubmission().GetBuildInfo().GetBuildLog() != "runtime error" {
					t.Errorf("Expected build log: \"runtime error\", got %+v", submission.GetSubmission().GetBuildInfo().GetBuildLog())
				}
			}
		}
	}

	// check that no submissions will be returned for a wrong course ID
	if _, err = ags.GetSubmissionsByCourse(ctx, &pb.SubmissionsForCourseRequest{CourseID: 234}); err == nil {
		t.Error("Expected 'no submissions found'")
	}

	// check that method fails with empty context
	if _, err = ags.GetSubmissionsByCourse(context.Background(), &pb.SubmissionsForCourseRequest{CourseID: course1.ID}); err == nil {
		t.Error("Expected 'authorization failed. please try to logout and sign in again'")
	}

	// check that method fails for unenrolled student user
	unenrolledStudent := qtest.CreateFakeUser(t, db, 3)
	ctx = qtest.WithUserContext(ctx, unenrolledStudent)
	if _, err := ags.GetSubmissionsByCourse(ctx, &pb.SubmissionsForCourseRequest{CourseID: course1.ID}); err == nil {
		t.Error("Expected 'only teachers can get all lab submissions'")
	}
	// check that method fails for non-teacher user
	ctx = qtest.WithUserContext(ctx, student)
	if _, err = ags.GetSubmissionsByCourse(ctx, &pb.SubmissionsForCourseRequest{CourseID: course1.ID}); err == nil {
		t.Error("Expected 'only teachers can get all lab submissions'")
	}
}

func TestCreateApproveList(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeUser(t, db, 1)

	course := allCourses[2]
	if err := db.CreateCourse(admin.ID, course); err != nil {
		t.Fatal(err)
	}
	student1 := qtest.CreateNamedUser(t, db, 2, "Leslie Lamport")
	student2 := qtest.CreateNamedUser(t, db, 3, "Hein Meling")
	student3 := qtest.CreateNamedUser(t, db, 4, "John Doe")
	qtest.EnrollStudent(t, db, student1, course)
	qtest.EnrollStudent(t, db, student2, course)
	qtest.EnrollStudent(t, db, student3, course)

	assignments := []*pb.Assignment{
		{
			CourseID:   course.ID,
			Name:       "lab 1",
			ScriptFile: "go.sh",
			Deadline:   "2020-02-23T18:00:00",
			Order:      1,
		},
		{
			CourseID:   course.ID,
			Name:       "lab 2",
			ScriptFile: "go.sh",
			Deadline:   "2020-03-23T18:00:00",
			Order:      2,
		},
		{
			CourseID:   course.ID,
			Name:       "lab 3",
			ScriptFile: "go.sh",
			Deadline:   "2020-04-23T18:00:00",
			Order:      3,
		},
		{
			CourseID:   course.ID,
			Name:       "lab 4",
			ScriptFile: "go.sh",
			Deadline:   "2020-05-23T18:00:00",
			Order:      4,
		},
	}
	for _, a := range assignments {
		if err := db.CreateAssignment(a); err != nil {
			t.Fatal(err)
		}
	}

	submissions := []*pb.Submission{
		{
			UserID:       student1.ID,
			AssignmentID: assignments[0].ID,
			Status:       pb.Submission_APPROVED,
		},
		{
			UserID:       student1.ID,
			AssignmentID: assignments[1].ID,
			Status:       pb.Submission_APPROVED,
		},
		{
			UserID:       student1.ID,
			AssignmentID: assignments[2].ID,
			Status:       pb.Submission_APPROVED,
		},
		{
			UserID:       student1.ID,
			AssignmentID: assignments[3].ID,
			Status:       pb.Submission_APPROVED,
		},
		{
			UserID:       student2.ID,
			AssignmentID: assignments[0].ID,
			Status:       pb.Submission_APPROVED,
		},
		{
			UserID:       student2.ID,
			AssignmentID: assignments[2].ID,
			Status:       pb.Submission_APPROVED,
		},
		{
			UserID:       student2.ID,
			AssignmentID: assignments[3].ID,
			Status:       pb.Submission_APPROVED,
		},
		{
			UserID:       student3.ID,
			AssignmentID: assignments[0].ID,
			Status:       pb.Submission_APPROVED,
		},
		{
			UserID:       student3.ID,
			AssignmentID: assignments[1].ID,
			Status:       pb.Submission_REJECTED,
		},
		{
			UserID:       student3.ID,
			AssignmentID: assignments[2].ID,
			Status:       pb.Submission_REVISION,
		},
	}
	for _, s := range submissions {
		if err := db.CreateSubmission(s); err != nil {
			t.Fatal(err)
		}
	}

	fakeProvider, scms := qtest.FakeProviderMap(t)
	ags := web.NewAutograderService(zap.NewNop(), db, scms, web.BaseHookOptions{}, &ci.Local{})
	ctx := qtest.WithUserContext(context.Background(), admin)
	_, err := fakeProvider.CreateOrganization(context.Background(), &scm.OrganizationOptions{Path: "path", Name: "name"})
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		student          *pb.User
		minNumApproved   int
		expectedApproved bool
	}{
		{
			student:          student1,
			minNumApproved:   4,
			expectedApproved: true,
		},
		{
			student:          student1,
			minNumApproved:   3,
			expectedApproved: true,
		},
		{
			student:          student2,
			minNumApproved:   4,
			expectedApproved: false,
		},
		{
			student:          student2,
			minNumApproved:   3,
			expectedApproved: true,
		},
		{
			student:          student2,
			minNumApproved:   2,
			expectedApproved: true,
		},
		{
			student:          student3,
			minNumApproved:   4,
			expectedApproved: false,
		},
		{
			student:          student3,
			minNumApproved:   3,
			expectedApproved: false,
		},
		{
			student:          student3,
			minNumApproved:   2,
			expectedApproved: false,
		},
		{
			student:          student3,
			minNumApproved:   1,
			expectedApproved: true,
		},
	}

	gotSubmissions, err := ags.GetSubmissionsByCourse(ctx, &pb.SubmissionsForCourseRequest{CourseID: course.ID, Type: pb.SubmissionsForCourseRequest_ALL})
	if err != nil {
		t.Fatal(err)
	}
	for _, el := range gotSubmissions.GetLinks() {
		if el.Enrollment.User.IsAdmin || el.Enrollment.GetHasTeacherScopes() {
			continue
		}
		approved := make([]bool, len(el.Submissions))
		for i, s := range el.Submissions {
			approved[i] = s.GetSubmission().IsApproved()
		}
		for _, test := range testCases {
			if test.student.ID == el.Enrollment.UserID {
				got := isApproved(test.minNumApproved, approved)
				if got != test.expectedApproved {
					t.Errorf("isApproved(%d, %v) = %t, expected %t", test.minNumApproved, approved, got, test.expectedApproved)
				}
			}
		}
		t.Logf("%s\t%t", el.Enrollment.User.Name, isApproved(4, approved))
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
