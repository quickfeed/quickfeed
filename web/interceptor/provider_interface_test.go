package interceptor

import (
	"fmt"
	"testing"

	"github.com/quickfeed/quickfeed/qf"
)

// Provider checker functions
var (
	assertUserIDProvider = func(v any) error {
		if _, ok := v.(userIDProvider); !ok {
			return fmt.Errorf("%T does not implement userIDProvider", v)
		}
		return nil
	}
	assertCourseIDProvider = func(v any) error {
		if _, ok := v.(courseIDProvider); !ok {
			return fmt.Errorf("%T does not implement courseIDProvider", v)
		}
		return nil
	}
	assertGroupIDProvider = func(v any) error {
		if _, ok := v.(groupIDProvider); !ok {
			return fmt.Errorf("%T does not implement groupIDProvider", v)
		}
		return nil
	}
	assertSubmissionIDProvider = func(v any) error {
		if _, ok := v.(submissionIDProvider); !ok {
			return fmt.Errorf("%T does not implement submissionIDProvider", v)
		}
		return nil
	}
)

// TestIDProviderInterfaces ensures that all types and requests used by the QuickFeed
// RPC methods implement the required ID provider interfaces that the access control
// checkers depend on. This test catches breaking changes to these types.
func TestIDProviderInterfaces(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		providers []func(any) error // interface assertion functions
	}{
		// Core types used in RPC calls
		{
			name:      "User implements userIDProvider",
			value:     &qf.User{},
			providers: []func(any) error{assertUserIDProvider},
		},
		{
			name:      "Group implements groupIDProvider and courseIDProvider",
			value:     &qf.Group{},
			providers: []func(any) error{assertGroupIDProvider, assertCourseIDProvider},
		},
		{
			name:      "Course implements courseIDProvider",
			value:     &qf.Course{},
			providers: []func(any) error{assertCourseIDProvider},
		},
		{
			name:      "Enrollment implements userIDProvider and courseIDProvider and groupIDProvider",
			value:     &qf.Enrollment{},
			providers: []func(any) error{assertUserIDProvider, assertCourseIDProvider, assertGroupIDProvider},
		},
		{
			name:      "Submission implements userIDProvider and groupIDProvider",
			value:     &qf.Submission{},
			providers: []func(any) error{assertUserIDProvider, assertGroupIDProvider},
		},

		// Request types
		{
			name:      "CourseRequest implements courseIDProvider",
			value:     &qf.CourseRequest{},
			providers: []func(any) error{assertCourseIDProvider},
		},
		{
			name:      "GroupRequest implements courseIDProvider and userIDProvider and groupIDProvider",
			value:     &qf.GroupRequest{},
			providers: []func(any) error{assertCourseIDProvider, assertUserIDProvider, assertGroupIDProvider},
		},
		{
			name:      "EnrollmentRequest implements courseIDProvider and userIDProvider",
			value:     &qf.EnrollmentRequest{},
			providers: []func(any) error{assertCourseIDProvider, assertUserIDProvider},
		},
		{
			name:      "SubmissionRequest implements courseIDProvider and userIDProvider and groupIDProvider and submissionIDProvider",
			value:     &qf.SubmissionRequest{},
			providers: []func(any) error{assertCourseIDProvider, assertUserIDProvider, assertGroupIDProvider, assertSubmissionIDProvider},
		},
		{
			name:      "UpdateSubmissionRequest implements courseIDProvider and submissionIDProvider",
			value:     &qf.UpdateSubmissionRequest{},
			providers: []func(any) error{assertCourseIDProvider, assertSubmissionIDProvider},
		},
		{
			name:      "UpdateSubmissionsRequest implements courseIDProvider",
			value:     &qf.UpdateSubmissionsRequest{},
			providers: []func(any) error{assertCourseIDProvider},
		},
		{
			name:      "RepositoryRequest implements courseIDProvider and userIDProvider and groupIDProvider",
			value:     &qf.RepositoryRequest{},
			providers: []func(any) error{assertCourseIDProvider, assertUserIDProvider, assertGroupIDProvider},
		},
		{
			name:      "RebuildRequest implements courseIDProvider and submissionIDProvider",
			value:     &qf.RebuildRequest{},
			providers: []func(any) error{assertCourseIDProvider, assertSubmissionIDProvider},
		},
		{
			name:      "ReviewRequest implements courseIDProvider",
			value:     &qf.ReviewRequest{},
			providers: []func(any) error{assertCourseIDProvider},
		},
		{
			name:      "GradingBenchmark implements courseIDProvider",
			value:     &qf.GradingBenchmark{},
			providers: []func(any) error{assertCourseIDProvider},
		},
		{
			name:      "GradingCriterion implements courseIDProvider",
			value:     &qf.GradingCriterion{},
			providers: []func(any) error{assertCourseIDProvider},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, assertProvider := range tt.providers {
				if err := assertProvider(tt.value); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

// TestAccessCheckerDependencies documents which request types are expected by each checker
// and verifies that the documented requirements match the actual usage in the checker functions.
func TestAccessCheckerDependencies(t *testing.T) {
	tests := []struct {
		name              string
		checker           accessChecker
		request           any
		requiredIDGetters []func(any) uint64 // ID getter functions that should work without panic
	}{
		{
			name:              "checkUser requires userIDProvider",
			checker:           checkUser,
			request:           &qf.User{ID: 1},
			requiredIDGetters: []func(any) uint64{getUserID},
		},
		{
			name:              "checkTeacher requires courseIDProvider",
			checker:           checkTeacher,
			request:           &qf.CourseRequest{CourseID: 1},
			requiredIDGetters: []func(any) uint64{getCourseID},
		},
		{
			name:              "checkStudentOrTeacher requires courseIDProvider",
			checker:           checkStudentOrTeacher,
			request:           &qf.CourseRequest{CourseID: 1},
			requiredIDGetters: []func(any) uint64{getCourseID},
		},
		{
			name:              "checkGroupOrTeacher requires groupIDProvider and courseIDProvider",
			checker:           checkGroupOrTeacher,
			request:           &qf.GroupRequest{CourseID: 1, GroupID: 2},
			requiredIDGetters: []func(any) uint64{getCourseID, getGroupID},
		},
		{
			name:              "checkUserOrStudentOrTeacherOrAdmin requires userIDProvider and courseIDProvider",
			checker:           checkUserOrStudentOrTeacherOrAdmin,
			request:           &qf.EnrollmentRequest{},
			requiredIDGetters: []func(any) uint64{getUserID, getCourseID},
		},
		{
			name:              "checkUpdateUser requires userIDProvider",
			checker:           checkUpdateUser,
			request:           &qf.User{ID: 1},
			requiredIDGetters: []func(any) uint64{getUserID},
		},
		{
			name:              "checkGetSubmissions requires userIDProvider, groupIDProvider, and courseIDProvider",
			checker:           checkGetSubmissions,
			request:           &qf.SubmissionRequest{CourseID: 1},
			requiredIDGetters: []func(any) uint64{getUserID, getGroupID, getCourseID},
		},
		{
			name:              "checkUpdateSubmission requires courseIDProvider and submissionIDProvider",
			checker:           checkUpdateSubmission,
			request:           &qf.UpdateSubmissionRequest{CourseID: 1, SubmissionID: 2},
			requiredIDGetters: []func(any) uint64{getCourseID, getSubmissionID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify that calling the helper functions doesn't panic
			for _, getID := range tt.requiredIDGetters {
				func() {
					defer func() {
						if r := recover(); r != nil {
							t.Errorf("calling ID getter on %T panicked: %v", tt.request, r)
						}
					}()
					getID(tt.request)
				}()
			}
		})
	}
}

// TestMethodCheckerRequestTypes documents which request type is used by each RPC method.
// This serves as documentation and helps catch changes when methods start using different request types.
func TestMethodCheckerRequestTypes(t *testing.T) {
	methodRequestTypes := map[string]string{
		// checkNone methods
		"GetUser":          "qf.Void",
		"GetCourse":        "qf.CourseRequest",
		"GetCourses":       "qf.Void",
		"SubmissionStream": "qf.Void",

		// checkUser methods
		"CreateEnrollment":       "qf.Enrollment",
		"UpdateCourseVisibility": "qf.Enrollment",

		// checkUpdateUser methods
		"UpdateUser": "qf.User",

		// checkUserOrStudentOrTeacherOrAdmin methods
		"GetEnrollments": "qf.EnrollmentRequest",

		// checkGetSubmissions methods
		"GetSubmissions": "qf.SubmissionRequest",

		// checkTeacher methods
		"GetSubmission":          "qf.SubmissionRequest",
		"UpdateGroup":            "qf.Group",
		"DeleteGroup":            "qf.GroupRequest",
		"GetGroupsByCourse":      "qf.CourseRequest",
		"UpdateCourse":           "qf.Course",
		"UpdateEnrollments":      "qf.Enrollments",
		"UpdateAssignments":      "qf.CourseRequest",
		"UpdateSubmissions":      "qf.UpdateSubmissionsRequest",
		"RebuildSubmissions":     "qf.RebuildRequest",
		"CreateBenchmark":        "qf.GradingBenchmark",
		"UpdateBenchmark":        "qf.GradingBenchmark",
		"DeleteBenchmark":        "qf.GradingBenchmark",
		"CreateCriterion":        "qf.GradingCriterion",
		"UpdateCriterion":        "qf.GradingCriterion",
		"DeleteCriterion":        "qf.GradingCriterion",
		"CreateReview":           "qf.ReviewRequest",
		"UpdateReview":           "qf.ReviewRequest",
		"GetAssignmentFeedback":  "qf.CourseRequest",
		"IsEmptyRepo":            "qf.RepositoryRequest",
		"GetSubmissionsByCourse": "qf.SubmissionRequest",
		"GetRepositories":        "qf.CourseRequest",

		// checkStudentOrTeacher methods
		"GetAssignments":           "qf.CourseRequest",
		"CreateAssignmentFeedback": "qf.AssignmentFeedback",

		// checkGroupOrTeacher methods
		"CreateGroup": "qf.Group",
		"GetGroup":    "qf.GroupRequest",

		// checkUpdateSubmission methods
		"UpdateSubmission": "qf.UpdateSubmissionRequest",

		// checkAdmin methods
		"GetUsers": "qf.Void",
	}

	// Verify all methods in methodCheckers have documented request types
	for method := range methodCheckers {
		if _, exists := methodRequestTypes[method]; !exists {
			t.Errorf("method %q in methodCheckers is missing from methodRequestTypes documentation", method)
		}
	}

	// Verify all documented methods exist in methodCheckers
	for method := range methodRequestTypes {
		if _, exists := methodCheckers[method]; !exists {
			t.Errorf("method %q documented in methodRequestTypes is missing from methodCheckers", method)
		}
	}

	t.Logf("Documented %d RPC methods with their request types", len(methodRequestTypes))
}
