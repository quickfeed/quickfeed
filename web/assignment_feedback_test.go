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
			name:   "Valid feedback with user ID",
			cookie: client.Cookie(t, student),
			feedback: &qf.AssignmentFeedback{
				CourseID:               course.GetID(),
				AssignmentID:           assignment.GetID(),
				UserID:                 student.GetID(),
				LikedContent:           "I liked the clear instructions and the practical examples provided.",
				ImprovementSuggestions: "Could benefit from more detailed examples in the initial setup section.",
				TimeSpent:              "3 hours",
			},
		},
		{
			name:   "Valid anonymous feedback",
			cookie: client.Cookie(t, student),
			feedback: &qf.AssignmentFeedback{
				CourseID:               course.GetID(),
				AssignmentID:           assignment.GetID(),
				UserID:                 0, // Anonymous
				LikedContent:           "Great assignment overall with good learning outcomes.",
				ImprovementSuggestions: "Maybe add some extra challenges for advanced students.",
				TimeSpent:              "2.5 hours",
			},
		},
		{
			name:   "Valid teacher feedback",
			cookie: client.Cookie(t, teacher),
			feedback: &qf.AssignmentFeedback{
				CourseID:               course.GetID(),
				AssignmentID:           assignment.GetID(),
				UserID:                 teacher.GetID(),
				LikedContent:           "Great check that also teachers can provide feedback.",
				ImprovementSuggestions: "Maybe add some extra challenges for advanced teachers.",
				TimeSpent:              "2.5 hours",
			},
		},
		{
			name:   "Missing course ID",
			cookie: client.Cookie(t, student),
			feedback: &qf.AssignmentFeedback{
				CourseID:               0, // Missing
				AssignmentID:           assignment.GetID(),
				UserID:                 student.GetID(),
				LikedContent:           "Good assignment in invalid course",
				ImprovementSuggestions: "Could be better",
				TimeSpent:              "1 hour",
			},
			// This should fail with permission denied because course ID 0 is invalid
			wantErr: connect.NewError(connect.CodePermissionDenied, errors.New("access denied for CreateAssignmentFeedback: required roles [student teacher] not satisfied by claims: UserID: 2: Courses: map[1:STUDENT], Groups: []")),
		},
		{
			name:   "Non-existing course ID",
			cookie: client.Cookie(t, student),
			feedback: &qf.AssignmentFeedback{
				CourseID:               999, // Non-existing
				AssignmentID:           assignment.GetID(),
				UserID:                 student.GetID(),
				LikedContent:           "Good assignment for non-existing course",
				ImprovementSuggestions: "You could at least create the course man!",
				TimeSpent:              "50 hours",
			},
			// This should fail with permission denied because course ID 999 does not exist
			wantErr: connect.NewError(connect.CodePermissionDenied, errors.New("access denied for CreateAssignmentFeedback: required roles [student teacher] not satisfied by claims: UserID: 2: Courses: map[1:STUDENT], Groups: []")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := client.CreateAssignmentFeedback(t.Context(), qtest.RequestWithCookie(test.feedback, test.cookie))
			if hasError := qtest.CheckCode(t, err, test.wantErr); hasError {
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
		UserID:                 student1.GetID(),
		LikedContent:           "Well structured assignment with clear goals.",
		ImprovementSuggestions: "Add more test cases for edge conditions.",
		TimeSpent:              "4 hours",
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
		UserID:                 student2.GetID(),
		LikedContent:           "Interesting problem to solve with good documentation.",
		ImprovementSuggestions: "Maybe provide starter code templates.",
		TimeSpent:              "5 hours",
	}
	resp2, err := client.CreateAssignmentFeedback(ctx, qtest.RequestWithCookie(feedback2, student2Cookie))
	if err != nil {
		t.Fatalf("Failed to create feedback2: %v", err)
	}
	createdFeedback2 := resp2.Msg

	tests := []struct {
		name    string
		cookie  string
		request *qf.AssignmentFeedbackRequest
		want    *qf.AssignmentFeedback
		wantErr error
	}{
		{
			name:   "Teacher can get feedback by assignment ID only (returns first found)",
			cookie: teacherCookie,
			request: &qf.AssignmentFeedbackRequest{
				CourseID:     course.GetID(),
				AssignmentID: assignment.GetID(),
			},
			want: createdFeedback1, // Should return first feedback created
		},
		{
			name:   "Teacher can get feedback by assignment ID and user ID",
			cookie: teacherCookie,
			request: &qf.AssignmentFeedbackRequest{
				CourseID:     course.GetID(),
				AssignmentID: assignment.GetID(),
				UserID:       student2.GetID(),
			},
			want: createdFeedback2,
		},
		{
			name:   "Student cannot get feedback once submitted",
			cookie: student1Cookie,
			request: &qf.AssignmentFeedbackRequest{
				CourseID:     course.GetID(),
				AssignmentID: assignment.GetID(),
				UserID:       student1.GetID(),
			},
			wantErr: connect.NewError(connect.CodePermissionDenied, errors.New("access denied for GetAssignmentFeedback: required roles [teacher] not satisfied by claims: UserID: 2: Courses: map[1:STUDENT], Groups: []")),
		},
		{
			name:   "Student cannot get other student's feedback",
			cookie: student1Cookie,
			request: &qf.AssignmentFeedbackRequest{
				CourseID:     course.GetID(),
				AssignmentID: assignment.GetID(),
				UserID:       student2.GetID(),
			},
			wantErr: connect.NewError(connect.CodePermissionDenied, errors.New("access denied for GetAssignmentFeedback: required roles [teacher] not satisfied by claims: UserID: 2: Courses: map[1:STUDENT], Groups: []")),
		},
		{
			name:   "Student cannot get feedback without specifying user ID",
			cookie: student2Cookie,
			request: &qf.AssignmentFeedbackRequest{
				CourseID:     course.GetID(),
				AssignmentID: assignment.GetID(),
			},
			wantErr: connect.NewError(connect.CodePermissionDenied, errors.New("access denied for GetAssignmentFeedback: required roles [teacher] not satisfied by claims: UserID: 3: Courses: map[1:STUDENT], Groups: []")),
		},
		{
			name:   "Teacher can get feedback for non-existent assignment",
			cookie: teacherCookie,
			request: &qf.AssignmentFeedbackRequest{
				CourseID:     course.GetID(),
				AssignmentID: 999999,
			},
			wantErr: connect.NewError(connect.CodeNotFound, errors.New("assignment feedback not found")),
		},
		{
			name:   "Teacher can get feedback for non-existent user",
			cookie: teacherCookie,
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
			resp, err := client.GetAssignmentFeedback(ctx, qtest.RequestWithCookie(test.request, test.cookie))
			if hasError := qtest.CheckCode(t, err, test.wantErr); hasError {
				return // cannot continue since resp is invalid
			}
			got := resp.Msg
			want := test.want
			qtest.Diff(t, "GetAssignmentFeedback mismatch", got, want, protocmp.Transform(), protocmp.IgnoreFields(&qf.AssignmentFeedback{}, "CreatedAt"))
		})
	}
}
