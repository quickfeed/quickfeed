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
			resp, err := client.CreateAssignmentFeedback(t.Context(), qtest.RequestWithCookie(test.feedback, test.cookie))
			if qtest.CheckCode(t, err, test.wantErr) {
				return // cannot continue since resp is invalid
			}
			if resp.Msg.GetID() == 0 {
				t.Error("Expected feedback ID to be set")
			}
			if resp.Msg.GetCreatedAt() == nil {
				t.Error("Expected CreatedAt to be set")
			}
			got := resp.Msg
			want := test.feedback
			qtest.Diff(t, "CreateAssignmentFeedback mismatch", got, want, protocmp.Transform(), protocmp.IgnoreFields(&qf.AssignmentFeedback{}, "ID", "CreatedAt"))
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
	ctx := context.Background()

	// Create feedback from student1
	feedback1 := &qf.AssignmentFeedback{
		CourseID:               course.GetID(),
		AssignmentID:           assignment.GetID(),
		LikedContent:           "Well structured assignment with clear goals.",
		ImprovementSuggestions: "Add more test cases for edge conditions.",
		TimeSpent:              240, // 4 hours
	}
	resp1, err := client.CreateAssignmentFeedback(ctx, qtest.RequestWithCookie(feedback1, student1Cookie))
	if err != nil {
		t.Fatalf("Failed to create feedback1: %v", err)
	}
	createdFeedback1 := resp1.Msg

	// Create feedback from student2
	feedback2 := &qf.AssignmentFeedback{
		CourseID:               course.GetID(),
		AssignmentID:           assignment.GetID(),
		LikedContent:           "Interesting problem to solve with good documentation.",
		ImprovementSuggestions: "Maybe provide starter code templates.",
		TimeSpent:              300, // 5 hours
	}
	resp2, err := client.CreateAssignmentFeedback(ctx, qtest.RequestWithCookie(feedback2, student2Cookie))
	if err != nil {
		t.Fatalf("Failed to create feedback2: %v", err)
	}
	createdFeedback2 := resp2.Msg

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
			want: &qf.AssignmentFeedbacks{Feedbacks: []*qf.AssignmentFeedback{createdFeedback1, createdFeedback2}},
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
			name:   "Teacher can get feedback for non-existent course",
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
			qtest.Diff(t, "GetAssignmentFeedback mismatch", got, want, protocmp.Transform(), protocmp.IgnoreFields(&qf.AssignmentFeedback{}, "CreatedAt"))
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
	studentCookie := client.Cookie(t, student)
	student2Cookie := client.Cookie(t, student2)

	// Helper function to get user's feedback receipts
	getUserReceipts := func(cookie string) ([]*qf.FeedbackReceipt, error) {
		resp, err := client.GetUser(ctx, qtest.RequestWithCookie(&qf.Void{}, cookie))
		if err != nil {
			return nil, err
		}
		return resp.Msg.GetFeedbackReceipts(), nil
	}

	// Helper function to create feedback
	createFeedback := func(cookie string, assignmentID uint64) (*qf.AssignmentFeedback, error) {
		feedback := &qf.AssignmentFeedback{
			CourseID:               course.GetID(),
			AssignmentID:           assignmentID,
			LikedContent:           "Well structured assignment with clear goals.",
			ImprovementSuggestions: "Add more test cases for edge conditions.",
			TimeSpent:              240, // 4 hours
		}
		resp, err := client.CreateAssignmentFeedback(ctx, qtest.RequestWithCookie(feedback, cookie))
		if err != nil {
			return nil, err
		}
		return resp.Msg, nil
	}

	tests := []struct {
		name          string
		setupFunc     func() error
		verifyFunc    func() error
		expectedError error
	}{
		{
			name: "Student starts with zero feedback receipts",
			setupFunc: func() error {
				return nil // No setup needed
			},
			verifyFunc: func() error {
				receipts, err := getUserReceipts(studentCookie)
				if err != nil {
					return err
				}
				if len(receipts) != 0 {
					return errors.New("expected 0 receipts")
				}
				return nil
			},
		},
		{
			name: "Student gets receipt after submitting feedback for assignment 1",
			setupFunc: func() error {
				_, err := createFeedback(studentCookie, assignment.GetID())
				return err
			},
			verifyFunc: func() error {
				receipts, err := getUserReceipts(studentCookie)
				if err != nil {
					return err
				}
				if len(receipts) != 1 {
					return errors.New("expected 1 receipt")
				}
				receipt := receipts[0]
				if receipt.GetAssignmentID() != assignment.GetID() {
					return errors.New("assignment ID mismatch")
				}

				return nil
			},
		},
		{
			name: "Student gets second receipt after submitting feedback for assignment 2",
			setupFunc: func() error {
				_, err := createFeedback(studentCookie, assignment2.GetID())
				return err
			},
			verifyFunc: func() error {
				receipts, err := getUserReceipts(studentCookie)
				if err != nil {
					return err
				}
				if len(receipts) != 2 {
					return errors.New("expected 2 receipts")
				}

				// Verify we have receipts for both assignments
				assignmentIDs := make(map[uint64]bool)
				for _, receipt := range receipts {
					assignmentIDs[receipt.GetAssignmentID()] = true
				}

				if !assignmentIDs[assignment.GetID()] {
					return errors.New("missing receipt for assignment 1")
				}
				if !assignmentIDs[assignment2.GetID()] {
					return errors.New("missing receipt for assignment 2")
				}
				return nil
			},
		},
		{
			name: "Duplicate feedback submission fails and doesn't create duplicate receipt",
			setupFunc: func() error {
				// Try to create feedback for assignment 1 again (should fail)
				_, err := createFeedback(studentCookie, assignment.GetID())
				if err == nil {
					return errors.New("expected error when creating duplicate feedback")
				}
				return nil // Expected to fail
			},
			verifyFunc: func() error {
				receipts, err := getUserReceipts(studentCookie)
				if err != nil {
					return err
				}
				if len(receipts) != 2 {
					return errors.New("expected 2 receipts (no duplicates)")
				}
				return nil
			},
		},
		{
			name: "Different student can create feedback for same assignment",
			setupFunc: func() error {
				_, err := createFeedback(student2Cookie, assignment.GetID())
				return err
			},
			verifyFunc: func() error {
				// Verify student1 still has 2 receipts
				receipts1, err := getUserReceipts(studentCookie)
				if err != nil {
					return err
				}
				if len(receipts1) != 2 {
					return errors.New("student1 expected 2 receipts")
				}

				// Verify student2 has 1 receipt
				receipts2, err := getUserReceipts(student2Cookie)
				if err != nil {
					return err
				}
				if len(receipts2) != 1 {
					return errors.New("student2 expected 1 receipt")
				}

				receipt := receipts2[0]
				if receipt.GetAssignmentID() != assignment.GetID() {
					return errors.New("assignment ID mismatch for student2")
				}
				return nil
			},
		},
		{
			name: "Teacher can create feedback and get receipt",
			setupFunc: func() error {
				teacherCookie := client.Cookie(t, teacher)
				_, err := createFeedback(teacherCookie, assignment2.GetID())
				return err
			},
			verifyFunc: func() error {
				teacherCookie := client.Cookie(t, teacher)
				receipts, err := getUserReceipts(teacherCookie)
				if err != nil {
					return err
				}
				if len(receipts) != 1 {
					return errors.New("teacher expected 1 receipt")
				}

				receipt := receipts[0]
				if receipt.GetAssignmentID() != assignment2.GetID() {
					return errors.New("assignment ID mismatch for teacher")
				}
				return nil
			},
		},
	}

	// Run tests sequentially since they build on each other
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.setupFunc(); err != nil {
				if test.expectedError == nil {
					t.Fatalf("Setup failed: %v", err)
				}
				// Expected error during setup, continue to verify
			}

			if err := test.verifyFunc(); err != nil {
				t.Errorf("Verification failed: %v", err)
			}
		})
	}
}
