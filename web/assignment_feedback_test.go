package web_test

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestCreateAssignmentFeedback(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"), web.WithInterceptors())
	teacher, course, assignment, student := qtest.SetupCourseAssignmentTeacherStudent(t, db)

	tests := []struct {
		name     string
		cookie   string
		feedback *qf.AssignmentFeedback
		wantErr  error
	}{
		{
			name:   "Valid anonymous feedback",
			cookie: client.Cookie(t, student),
			feedback: &qf.AssignmentFeedback{
				CourseID:               course.GetID(),
				AssignmentID:           assignment.GetID(),
				LikedContent:           "Great assignment overall with good learning outcomes.",
				ImprovementSuggestions: "Maybe add some extra challenges for advanced students.",
				TimeSpent:              150, // 2.5 hours
			},
			wantErr: nil,
		},
		{
			name:   "Valid teacher feedback",
			cookie: client.Cookie(t, teacher),
			feedback: &qf.AssignmentFeedback{
				CourseID:               course.GetID(),
				AssignmentID:           assignment.GetID(),
				LikedContent:           "Great check that also teachers can provide feedback.",
				ImprovementSuggestions: "Maybe add some extra challenges for advanced teachers.",
				TimeSpent:              150, // 2.5 hours
			},
			wantErr: nil,
		},
		{
			name:   "Missing course ID",
			cookie: client.Cookie(t, student),
			feedback: &qf.AssignmentFeedback{
				CourseID:               0, // Missing
				AssignmentID:           assignment.GetID(),
				LikedContent:           "Good assignment in invalid course",
				ImprovementSuggestions: "Could be better",
				TimeSpent:              60, // 1 hour
			},
			// This should fail with invalid payload because course ID 0 is invalid
			wantErr: connect.NewError(connect.CodeInvalidArgument, errors.New("invalid payload")),
		},
		{
			name:   "Missing assignment ID",
			cookie: client.Cookie(t, student),
			feedback: &qf.AssignmentFeedback{
				CourseID:               course.GetID(),
				AssignmentID:           0, // Missing
				LikedContent:           "Good assignment with missing ID",
				ImprovementSuggestions: "Could be better",
				TimeSpent:              60, // 1 hour
			},
			// This should fail with invalid payload because assignment ID 0 is invalid
			wantErr: connect.NewError(connect.CodeInvalidArgument, errors.New("invalid payload")),
		},
		{
			name:   "Empty liked content",
			cookie: client.Cookie(t, student),
			feedback: &qf.AssignmentFeedback{
				CourseID:               course.GetID(),
				AssignmentID:           assignment.GetID(),
				LikedContent:           "", // Missing
				ImprovementSuggestions: "Could be better",
				TimeSpent:              60, // 1 hour
			},
			// This should fail with invalid payload because liked content is empty
			wantErr: connect.NewError(connect.CodeInvalidArgument, errors.New("invalid payload")),
		},
		{
			name:   "Empty improvement suggestions",
			cookie: client.Cookie(t, student),
			feedback: &qf.AssignmentFeedback{
				CourseID:               course.GetID(),
				AssignmentID:           assignment.GetID(),
				LikedContent:           "Good assignment",
				ImprovementSuggestions: "", // Missing
				TimeSpent:              60, // 1 hour
			},
			// This should fail with invalid payload because improvement suggestions is empty
			wantErr: connect.NewError(connect.CodeInvalidArgument, errors.New("invalid payload")),
		},
		{
			name:   "Zero time spent",
			cookie: client.Cookie(t, student),
			feedback: &qf.AssignmentFeedback{
				CourseID:               course.GetID(),
				AssignmentID:           assignment.GetID(),
				LikedContent:           "Good assignment",
				ImprovementSuggestions: "Could be better",
				TimeSpent:              0, // Missing
			},
			// This should fail with invalid payload because time spent is zero
			wantErr: connect.NewError(connect.CodeInvalidArgument, errors.New("invalid payload")),
		},
		{
			name:   "Non-existing course ID",
			cookie: client.Cookie(t, student),
			feedback: &qf.AssignmentFeedback{
				CourseID:               999, // Non-existing
				AssignmentID:           assignment.GetID(),
				LikedContent:           "Good assignment for non-existing course",
				ImprovementSuggestions: "You could at least create the course man!",
				TimeSpent:              180000, // 50 hours
			},
			// This should fail with permission denied because course ID 999 does not exist
			wantErr: connect.NewError(connect.CodePermissionDenied, errors.New("access denied for CreateAssignmentFeedback: required roles [student teacher] not satisfied by claims: UserID: 2: Courses: map[1:STUDENT], Groups: []")),
		},
		{
			name:   "Non-existing assignment ID",
			cookie: client.Cookie(t, student),
			feedback: &qf.AssignmentFeedback{
				CourseID:               course.GetID(),
				AssignmentID:           99999, // Non-existing
				LikedContent:           "Good assignment for non-existing assignment",
				ImprovementSuggestions: "You could at least create the assignment",
			},
			// This should fail with permission denied because assignment ID 99999 does not exist
			wantErr: connect.NewError(connect.CodeInvalidArgument, errors.New("invalid payload")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := client.CreateAssignmentFeedback(t.Context(), qtest.RequestWithCookie(test.feedback, test.cookie))
			if err == nil && test.wantErr == nil {
				return // both nil, all good
			}
			if !qtest.CheckCode(t, err, test.wantErr) {
				t.Errorf("CreateAssignmentFeedback() unexpected error: %v, %T, want: %v, %T", err, err, test.wantErr, test.wantErr)
			}
		})
	}
}

func TestGetAssignmentFeedback(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"), web.WithInterceptors())
	teacher, course, assignment, student1 := qtest.SetupCourseAssignmentTeacherStudent(t, db)

	// Enroll an additional student
	student2 := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, student2, course)

	// Create cookies for authentication
	teacherCookie := client.Cookie(t, teacher)
	student1Cookie := client.Cookie(t, student1)
	student2Cookie := client.Cookie(t, student2)
	ctx := t.Context()

	// Create feedback from student1
	feedback1 := &qf.AssignmentFeedback{
		CourseID:               course.GetID(),
		AssignmentID:           assignment.GetID(),
		LikedContent:           "Well structured assignment with clear goals.",
		ImprovementSuggestions: "Add more test cases for edge conditions.",
		TimeSpent:              240, // 4 hours
	}
	_, err := client.CreateAssignmentFeedback(ctx, qtest.RequestWithCookie(feedback1, student1Cookie))
	if err != nil {
		t.Fatalf("Failed to create feedback1: %v", err)
	}

	// Create feedback from student2
	feedback2 := &qf.AssignmentFeedback{
		CourseID:               course.GetID(),
		AssignmentID:           assignment.GetID(),
		LikedContent:           "Interesting problem to solve with good documentation.",
		ImprovementSuggestions: "Maybe provide starter code templates.",
		TimeSpent:              300, // 5 hours
	}
	_, err = client.CreateAssignmentFeedback(ctx, qtest.RequestWithCookie(feedback2, student2Cookie))
	if err != nil {
		t.Fatalf("Failed to create feedback2: %v", err)
	}

	tests := []struct {
		name    string
		cookie  string
		request *qf.CourseRequest
		want    *qf.AssignmentFeedbacks
		wantErr error
	}{
		{
			name:   "Teacher can get feedback by course ID only",
			cookie: teacherCookie,
			request: &qf.CourseRequest{
				CourseID: course.GetID(),
			},
			want: &qf.AssignmentFeedbacks{Feedbacks: []*qf.AssignmentFeedback{feedback1, feedback2}},
		},
		{
			name:   "Student cannot get feedback once submitted",
			cookie: student1Cookie,
			request: &qf.CourseRequest{
				CourseID: course.GetID(),
			},
			wantErr: connect.NewError(connect.CodePermissionDenied, errors.New("access denied for GetAssignmentFeedback: required roles [teacher] not satisfied by claims: UserID: 2: Courses: map[1:STUDENT], Groups: []")),
		},
		{
			name:   "Teacher cannot get feedback for non-existent course",
			cookie: teacherCookie,
			request: &qf.CourseRequest{
				CourseID: 99999,
			},
			wantErr: connect.NewError(connect.CodePermissionDenied, errors.New("access denied for GetAssignmentFeedback: required roles [teacher] not satisfied by claims: UserID: 1 (admin): Courses: map[1:TEACHER], Groups: []")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := client.GetAssignmentFeedback(ctx, qtest.RequestWithCookie(test.request, test.cookie))
			if qtest.CheckCode(t, err, test.wantErr) {
				return // cannot continue since resp is invalid
			}
			got := resp.Msg
			want := test.want
			// UserID is removed in responses, so we ignore it in the comparison
			qtest.Diff(t, "GetAssignmentFeedback mismatch", got, want, protocmp.Transform(), protocmp.IgnoreFields(&qf.AssignmentFeedback{}, "ID", "CreatedAt"))
		})
	}
}

func TestFeedbackReceiptCreation(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"), web.WithInterceptors())
	teacher, course, assignment, student := qtest.SetupCourseAssignmentTeacherStudent(t, db)

	// Create a second assignment for testing multiple receipts
	assignment2 := &qf.Assignment{
		CourseID: course.GetID(),
		Order:    2,
		Name:     "lab2",
	}
	if err := db.CreateAssignment(assignment2); err != nil {
		t.Fatalf("failed to create assignment2: %v", err)
	}

	// Create a second student for testing
	student2 := qtest.CreateFakeUser(t, db)
	qtest.EnrollStudent(t, db, student2, course)

	ctx := context.Background()

	// Helper: get user's feedback receipts
	getUserReceipts := func(cookie string) []*qf.FeedbackReceipt {
		resp, err := client.GetUser(ctx, qtest.RequestWithCookie(&qf.Void{}, cookie))
		if err != nil {
			t.Fatalf("failed to get user receipts: %v", err)
		}
		return resp.Msg.GetFeedbackReceipts()
	}

	// Helper: create feedback
	createFeedback := func(cookie string, assignmentID uint64) error {
		feedback := &qf.AssignmentFeedback{
			CourseID:               course.GetID(),
			AssignmentID:           assignmentID,
			LikedContent:           "Well structured assignment with clear goals.",
			ImprovementSuggestions: "Add more test cases for edge conditions.",
			TimeSpent:              240,
		}
		_, err := client.CreateAssignmentFeedback(ctx, qtest.RequestWithCookie(feedback, cookie))
		return err
	}

	teacherID := teacher.GetID()
	studentID := student.GetID()
	student2ID := student2.GetID()

	cookies := map[uint64]string{
		studentID:  client.Cookie(t, student),
		student2ID: client.Cookie(t, student2),
		teacherID:  client.Cookie(t, teacher),
	}

	// Assertions helpers
	checkCount := func(t *testing.T, userID uint64, want int) {
		t.Helper()
		receipts := getUserReceipts(cookies[userID])
		if got := len(receipts); got != want {
			t.Fatalf("expected %d receipts, got %d", want, got)
		}
	}

	checkHas := func(t *testing.T, userID uint64, wantAssignmentIDs ...uint64) {
		t.Helper()
		receipts := getUserReceipts(cookies[userID])
		seen := make(map[uint64]bool)
		for _, r := range receipts {
			if r.GetUserID() != userID {
				t.Errorf("expected user ID %d, got %d", userID, r.GetUserID())
			}
			seen[r.GetAssignmentID()] = true
		}
		for _, id := range wantAssignmentIDs {
			if !seen[id] {
				t.Errorf("missing receipt for assignment %d", id)
			}
		}
	}

	type action int
	const (
		none action = iota
		create
		duplicate
	)

	tests := []struct {
		name         string
		do           action
		userID       uint64
		assignmentID uint64
		expectCount  map[uint64]int      // expected total receipts per user
		expectHas    map[uint64][]uint64 // expected assignment IDs present for user
		expectErr    bool                // only relevant for duplicate actions
	}{
		{
			name: "initial state",
			do:   none,
			expectCount: map[uint64]int{
				studentID:  0,
				student2ID: 0,
				teacherID:  0,
			},
		},
		{
			name:         "student1 creates feedback for assignment1 -> gets 1 receipt",
			do:           create,
			userID:       studentID,
			assignmentID: assignment.GetID(),
			expectCount: map[uint64]int{
				studentID:  1,
				student2ID: 0,
				teacherID:  0,
			},
			expectHas: map[uint64][]uint64{
				studentID: {assignment.GetID()},
			},
		},
		{
			name:         "student1 creates feedback for assignment2 -> now 2 receipts",
			do:           create,
			userID:       studentID,
			assignmentID: assignment2.GetID(),
			expectCount: map[uint64]int{
				studentID:  2,
				student2ID: 0,
				teacherID:  0,
			},
			expectHas: map[uint64][]uint64{
				studentID: {assignment.GetID(), assignment2.GetID()},
			},
		},
		{
			name:         "duplicate prevention for student1 on assignment1",
			do:           duplicate,
			userID:       studentID,
			assignmentID: assignment.GetID(),
			expectErr:    true,
			expectCount: map[uint64]int{
				studentID: 2, // unchanged
			},
			expectHas: map[uint64][]uint64{
				studentID: {assignment.GetID(), assignment2.GetID()},
			},
		},
		{
			name:         "student2 can create feedback for same assignment1",
			do:           create,
			userID:       student2ID,
			assignmentID: assignment.GetID(),
			expectCount: map[uint64]int{
				studentID:  2,
				student2ID: 1,
				teacherID:  0,
			},
			expectHas: map[uint64][]uint64{
				student2ID: {assignment.GetID()},
			},
		},
		{
			name:         "teacher creates feedback for assignment2",
			do:           create,
			userID:       teacherID,
			assignmentID: assignment2.GetID(),
			expectCount: map[uint64]int{
				teacherID:  1,
				studentID:  2,
				student2ID: 1,
			},
			expectHas: map[uint64][]uint64{
				teacherID: {assignment2.GetID()},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			switch test.do {
			case create:
				if err := createFeedback(cookies[test.userID], test.assignmentID); err != nil {
					t.Fatalf("failed to create feedback: %v", err)
				}

			case duplicate:
				err := createFeedback(cookies[test.userID], test.assignmentID)
				if test.expectErr && err == nil {
					t.Fatalf("expected error for duplicate feedback, got nil")
				}
				if !test.expectErr && err != nil {
					t.Fatalf("did not expect error, got %v", err)
				}

			case none:
				// no-op
			}

			// Verify counts
			for userID, want := range test.expectCount {
				checkCount(t, userID, want)
			}
			// Verify presence of receipts
			for userID, ids := range test.expectHas {
				checkHas(t, userID, ids...)
			}
		})
	}
}
