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

	// Helper function to get user's feedback receipts
	getUserReceipts := func(cookie string) []*qf.FeedbackReceipt {
		resp, err := client.GetUser(ctx, qtest.RequestWithCookie(&qf.Void{}, cookie))
		if err != nil {
			t.Fatalf("failed to get user receipts: %v", err)
		}
		return resp.Msg.GetFeedbackReceipts()
	}

	// Helper function to create feedback
	createFeedback := func(cookie string, assignmentID uint64) *qf.AssignmentFeedback {
		feedback := &qf.AssignmentFeedback{
			CourseID:               course.GetID(),
			AssignmentID:           assignmentID,
			LikedContent:           "Well structured assignment with clear goals.",
			ImprovementSuggestions: "Add more test cases for edge conditions.",
			TimeSpent:              240, // 4 hours
		}
		resp, err := client.CreateAssignmentFeedback(ctx, qtest.RequestWithCookie(feedback, cookie))
		if err != nil {
			t.Fatalf("failed to create feedback: %v", err)
		}
		return resp.Msg
	}

	// Helper function to try creating feedback (expecting failure)
	tryCreateFeedback := func(cookie string, assignmentID uint64) error {
		feedback := &qf.AssignmentFeedback{
			CourseID:               course.GetID(),
			AssignmentID:           assignmentID,
			LikedContent:           "Duplicate feedback attempt",
			ImprovementSuggestions: "This should fail",
			TimeSpent:              180,
		}
		_, err := client.CreateAssignmentFeedback(ctx, qtest.RequestWithCookie(feedback, cookie))
		return err
	}

	studentCookie := client.Cookie(t, student)
	student2Cookie := client.Cookie(t, student2)
	teacherCookie := client.Cookie(t, teacher)

	// Test 1: Student starts with zero feedback receipts
	t.Run("initial state", func(t *testing.T) {
		receipts := getUserReceipts(studentCookie)
		if len(receipts) != 0 {
			t.Errorf("expected 0 receipts, got %d", len(receipts))
		}
	})

	// Test 2: Student gets receipt after submitting feedback for assignment 1
	t.Run("single receipt creation", func(t *testing.T) {
		createFeedback(studentCookie, assignment.GetID())

		receipts := getUserReceipts(studentCookie)
		if len(receipts) != 1 {
			t.Fatalf("expected 1 receipt, got %d", len(receipts))
		}

		receipt := receipts[0]
		if receipt.GetAssignmentID() != assignment.GetID() {
			t.Errorf("expected assignment ID %d, got %d", assignment.GetID(), receipt.GetAssignmentID())
		}
		if receipt.GetUserID() != student.GetID() {
			t.Errorf("expected user ID %d, got %d", student.GetID(), receipt.GetUserID())
		}
	})

	// Test 3: Student gets second receipt after submitting feedback for assignment 2
	t.Run("multiple receipts", func(t *testing.T) {
		createFeedback(studentCookie, assignment2.GetID())

		receipts := getUserReceipts(studentCookie)
		if len(receipts) != 2 {
			t.Fatalf("expected 2 receipts, got %d", len(receipts))
		}

		assignmentIDs := make(map[uint64]bool)
		for _, receipt := range receipts {
			assignmentIDs[receipt.GetAssignmentID()] = true
			if receipt.GetUserID() != student.GetID() {
				t.Errorf("expected user ID %d, got %d", student.GetID(), receipt.GetUserID())
			}
		}

		if !assignmentIDs[assignment.GetID()] {
			t.Error("missing receipt for assignment 1")
		}
		if !assignmentIDs[assignment2.GetID()] {
			t.Error("missing receipt for assignment 2")
		}
	})

	// Test 4: Duplicate feedback submission fails and doesn't create duplicate receipt
	t.Run("duplicate prevention", func(t *testing.T) {
		err := tryCreateFeedback(studentCookie, assignment.GetID())
		if err == nil {
			t.Error("expected error when creating duplicate feedback")
		}

		receipts := getUserReceipts(studentCookie)
		if len(receipts) != 2 {
			t.Errorf("expected 2 receipts after duplicate attempt, got %d", len(receipts))
		}
	})

	// Test 5: Different student can create feedback for same assignment
	t.Run("different student receipts", func(t *testing.T) {
		createFeedback(student2Cookie, assignment.GetID())

		// Verify student1 still has 2 receipts
		receipts1 := getUserReceipts(studentCookie)
		if len(receipts1) != 2 {
			t.Errorf("student1 expected 2 receipts, got %d", len(receipts1))
		}

		// Verify student2 has 1 receipt
		receipts2 := getUserReceipts(student2Cookie)
		if len(receipts2) != 1 {
			t.Errorf("student2 expected 1 receipt, got %d", len(receipts2))
		}

		receipt := receipts2[0]
		if receipt.GetAssignmentID() != assignment.GetID() {
			t.Errorf("expected assignment ID %d for student2, got %d", assignment.GetID(), receipt.GetAssignmentID())
		}
		if receipt.GetUserID() != student2.GetID() {
			t.Errorf("expected user ID %d for student2, got %d", student2.GetID(), receipt.GetUserID())
		}
	})

	// Test 6: Teacher can create feedback and get receipt
	t.Run("teacher receipts", func(t *testing.T) {
		createFeedback(teacherCookie, assignment2.GetID())

		receipts := getUserReceipts(teacherCookie)
		if len(receipts) != 1 {
			t.Errorf("teacher expected 1 receipt, got %d", len(receipts))
		}

		receipt := receipts[0]
		if receipt.GetAssignmentID() != assignment2.GetID() {
			t.Errorf("expected assignment ID %d for teacher, got %d", assignment2.GetID(), receipt.GetAssignmentID())
		}
		if receipt.GetUserID() != teacher.GetID() {
			t.Errorf("expected user ID %d for teacher, got %d", teacher.GetID(), receipt.GetUserID())
		}
	})
}
