package web_test

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestCreateAssignmentFeedback(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := web.MockClientWithOption(t, db, scm.WithMockOrgs("admin"))
	admin := qtest.CreateFakeUser(t, db)
	student := qtest.CreateFakeUser(t, db)
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, admin, course)
	qtest.EnrollStudent(t, db, student, course)

	assignment := &qf.Assignment{
		CourseID: course.GetID(),
		Order:    1,
		Name:     "Assignment 1",
	}
	qtest.CreateAssignment(t, db, assignment)

	cookie := Cookie(t, tm, student)
	ctx := context.Background()

	tests := []struct {
		name     string
		feedback *qf.AssignmentFeedback
		wantErr  error
	}{
		{
			name: "Valid feedback with user ID",
			feedback: &qf.AssignmentFeedback{
				CourseID:                course.GetID(),
				AssignmentID:            assignment.GetID(),
				UserID:                  student.GetID(),
				LikedContent:            "I liked the clear instructions and the practical examples provided.",
				ImprovementSuggestions: "Could benefit from more detailed examples in the initial setup section.",
				TimeSpent:               "3 hours",
			},
		},
		{
			name: "Valid anonymous feedback",
			feedback: &qf.AssignmentFeedback{
				CourseID:                course.GetID(),
				AssignmentID:            assignment.GetID(),
				UserID:                  0, // Anonymous
				LikedContent:            "Great assignment overall with good learning outcomes.",
				ImprovementSuggestions: "Maybe add some extra challenges for advanced students.",
				TimeSpent:               "2.5 hours",
			},
		},
		{
			name: "Missing course ID", 
			feedback: &qf.AssignmentFeedback{
				CourseID:                0, // Missing
				AssignmentID:            assignment.GetID(),
				UserID:                  student.GetID(),
				LikedContent:            "Good assignment",
				ImprovementSuggestions: "Could be better",
				TimeSpent:               "1 hour",
			},
			// This should fail with permission denied because course ID 0 means no access
			wantErr: connect.NewError(connect.CodePermissionDenied, errors.New("access denied for CreateAssignmentFeedback: required roles [3 4] not satisfied by claims")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := client.CreateAssignmentFeedback(ctx, qtest.RequestWithCookie(test.feedback, cookie))
			if test.wantErr != nil {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if connect.CodeOf(err) != connect.CodeOf(test.wantErr) {
					t.Errorf("Expected error code %v, got %v", connect.CodeOf(test.wantErr), connect.CodeOf(err))
				}
				return
			}
			if err != nil {
				t.Fatalf("CreateAssignmentFeedback() failed: %v", err)
			}

			// Verify response
			if resp.Msg.GetID() == 0 {
				t.Error("Expected feedback ID to be set")
			}
			if resp.Msg.GetCreatedAt() == nil {
				t.Error("Expected CreatedAt to be set")
			}

			// Verify feedback fields match
			got := resp.Msg
			want := test.feedback
			if got.GetCourseID() != want.GetCourseID() {
				t.Errorf("CourseID mismatch: got %d, want %d", got.GetCourseID(), want.GetCourseID())
			}
			if got.GetAssignmentID() != want.GetAssignmentID() {
				t.Errorf("AssignmentID mismatch: got %d, want %d", got.GetAssignmentID(), want.GetAssignmentID())
			}
			if got.GetUserID() != want.GetUserID() {
				t.Errorf("UserID mismatch: got %d, want %d", got.GetUserID(), want.GetUserID())
			}
			if got.GetLikedContent() != want.GetLikedContent() {
				t.Errorf("LikedContent mismatch: got %s, want %s", got.GetLikedContent(), want.GetLikedContent())
			}
			if got.GetImprovementSuggestions() != want.GetImprovementSuggestions() {
				t.Errorf("ImprovementSuggestions mismatch: got %s, want %s", got.GetImprovementSuggestions(), want.GetImprovementSuggestions())
			}
			if got.GetTimeSpent() != want.GetTimeSpent() {
				t.Errorf("TimeSpent mismatch: got %s, want %s", got.GetTimeSpent(), want.GetTimeSpent())
			}
		})
	}
}

func TestGetAssignmentFeedback(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client, tm := web.MockClientWithOption(t, db, scm.WithMockOrgs("admin"))
	admin := qtest.CreateFakeUser(t, db)
	student1 := qtest.CreateFakeUser(t, db)
	student2 := qtest.CreateFakeUser(t, db)
	course := qtest.MockCourses[0]
	qtest.CreateCourse(t, db, admin, course)
	qtest.EnrollStudent(t, db, student1, course)
	qtest.EnrollStudent(t, db, student2, course)

	assignment := &qf.Assignment{
		CourseID: course.GetID(),
		Order:    1,
		Name:     "Assignment 1",
	}
	qtest.CreateAssignment(t, db, assignment)

	// Create feedback as students
	student1Cookie := Cookie(t, tm, student1)
	student2Cookie := Cookie(t, tm, student2)
	teacherCookie := Cookie(t, tm, admin)
	ctx := context.Background()

	// Create feedback from student1
	feedback1 := &qf.AssignmentFeedback{
		CourseID:                course.GetID(),
		AssignmentID:            assignment.GetID(),
		UserID:                  student1.GetID(),
		LikedContent:            "Well structured assignment with clear goals.",
		ImprovementSuggestions: "Add more test cases for edge conditions.",
		TimeSpent:               "4 hours",
	}
	resp1, err := client.CreateAssignmentFeedback(ctx, qtest.RequestWithCookie(feedback1, student1Cookie))
	if err != nil {
		t.Fatalf("Failed to create feedback1: %v", err)
	}
	createdFeedback1 := resp1.Msg

	// Create feedback from student2
	feedback2 := &qf.AssignmentFeedback{
		CourseID:                course.GetID(),
		AssignmentID:            assignment.GetID(),
		UserID:                  student2.GetID(),
		LikedContent:            "Interesting problem to solve with good documentation.",
		ImprovementSuggestions: "Maybe provide starter code templates.",
		TimeSpent:               "5 hours",
	}
	resp2, err := client.CreateAssignmentFeedback(ctx, qtest.RequestWithCookie(feedback2, student2Cookie))
	if err != nil {
		t.Fatalf("Failed to create feedback2: %v", err)
	}
	createdFeedback2 := resp2.Msg

	tests := []struct {
		name    string
		request *qf.AssignmentFeedbackRequest
		want    *qf.AssignmentFeedback
		wantErr error
	}{
		{
			name: "Get feedback by assignment ID only (returns first found)",
			request: &qf.AssignmentFeedbackRequest{
				CourseID:     course.GetID(),
				AssignmentID: assignment.GetID(),
			},
			want: createdFeedback1, // Should return first feedback created
		},
		{
			name: "Get feedback by assignment ID and user ID",
			request: &qf.AssignmentFeedbackRequest{
				CourseID:     course.GetID(),
				AssignmentID: assignment.GetID(),
				UserID:       student2.GetID(),
			},
			want: createdFeedback2,
		},
		{
			name: "Get feedback for non-existent assignment",
			request: &qf.AssignmentFeedbackRequest{
				CourseID:     course.GetID(),
				AssignmentID: 999999,
			},
			wantErr: connect.NewError(connect.CodeNotFound, errors.New("assignment feedback not found")),
		},
		{
			name: "Get feedback for non-existent user",
			request: &qf.AssignmentFeedbackRequest{
				CourseID:     course.GetID(),
				AssignmentID: assignment.GetID(),
				UserID:       999999,
			},
			wantErr: connect.NewError(connect.CodeNotFound, errors.New("assignment feedback not found")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := client.GetAssignmentFeedback(ctx, qtest.RequestWithCookie(test.request, teacherCookie))
			if test.wantErr != nil {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if connect.CodeOf(err) != connect.CodeOf(test.wantErr) {
					t.Errorf("Expected error code %v, got %v", connect.CodeOf(test.wantErr), connect.CodeOf(err))
				}
				return
			}
			if err != nil {
				t.Fatalf("GetAssignmentFeedback() failed: %v", err)
			}

			got := resp.Msg
			want := test.want

			// Compare the feedback (ignoring timestamps and IDs for flexibility)
			if diff := cmp.Diff(want, got, protocmp.Transform(), protocmp.IgnoreFields(&qf.AssignmentFeedback{}, "CreatedAt")); diff != "" {
				t.Errorf("GetAssignmentFeedback() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}