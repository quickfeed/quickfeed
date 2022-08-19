package interceptor_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/web/interceptor"
)

func TestCheckAccessMethods(t *testing.T) {
	serviceMethods := map[string]bool{
		"GetUser":                 true,
		"GetCourse":               true,
		"GetCourses":              true,
		"CreateEnrollment":        true,
		"UpdateCourseVisibility":  true,
		"GetCoursesByUser":        true,
		"UpdateUser":              true,
		"GetEnrollmentsByUser":    true,
		"GetSubmissions":          true,
		"GetGroupByUserAndCourse": true,
		"CreateGroup":             true,
		"GetGroup":                true,
		"GetAssignments":          true,
		"GetEnrollmentsByCourse":  true,
		"GetRepositories":         true,
		"UpdateGroup":             true,
		"DeleteGroup":             true,
		"GetGroupsByCourse":       true,
		"UpdateCourse":            true,
		"UpdateEnrollments":       true,
		"UpdateAssignments":       true,
		"UpdateSubmission":        true,
		"UpdateSubmissions":       true,
		"RebuildSubmissions":      true,
		"CreateBenchmark":         true,
		"UpdateBenchmark":         true,
		"DeleteBenchmark":         true,
		"CreateCriterion":         true,
		"UpdateCriterion":         true,
		"DeleteCriterion":         true,
		"CreateReview":            true,
		"UpdateReview":            true,
		"GetReviewers":            true,
		"IsEmptyRepo":             true,
		"GetSubmissionsByCourse":  true,
		"GetUserByCourse":         true,
		"GetUsers":                true,
		"GetOrganization":         true,
		"CreateCourse":            true,
	}
	if err := interceptor.CheckAccessMethods(serviceMethods); err != nil {
		t.Error(err)
	}

	// Disable CreateCourse method in the serviceMethods map;
	// make it appear as if it was removed from the service interface.
	serviceMethods["CreateCourse"] = false
	err := interceptor.CheckAccessMethods(serviceMethods)
	expectedErr := "superfluous method(s) in access control table: [CreateCourse]"
	if err == nil {
		t.Errorf("Expected error: %q, got nil", expectedErr)
	}
	if err.Error() != expectedErr {
		t.Errorf("Expected error: %q, got: %q", expectedErr, err.Error())
	}

	// Add new Dummy method to the serviceMethods map;
	// make it appear as if it was added to the service interface.
	serviceMethods["Dummy"] = true
	err = interceptor.CheckAccessMethods(serviceMethods)
	expectedErr = "missing required method(s) in access control table: [Dummy]"
	if err == nil {
		t.Errorf("Expected error: %q, got nil", expectedErr)
	}
	if err.Error() != expectedErr {
		t.Errorf("Expected error: %q, got: %q", expectedErr, err.Error())
	}
}
