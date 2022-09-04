package interceptor

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/quickfeed/quickfeed/qf/qfconnect"
)

// TestAccessControlMethods checks that all QuickFeedService methods have an entry
// in the access control list.
func TestAccessControlQuickFeedServiceMethods(t *testing.T) {
	service := reflect.TypeOf(qfconnect.UnimplementedQuickFeedServiceHandler{})
	serviceMethods := make(map[string]bool)
	for i := 0; i < service.NumMethod(); i++ {
		serviceMethods[service.Method(i).Name] = true
	}
	if err := checkAccessControlMethods(serviceMethods); err != nil {
		t.Error(err)
	}
}

func TestAccessControlMethodsChecker(t *testing.T) {
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
	if err := checkAccessControlMethods(serviceMethods); err != nil {
		t.Error(err)
	}

	// Disable CreateCourse method in the serviceMethods map;
	// make it appear as if it was removed from the service interface.
	serviceMethods["CreateCourse"] = false
	err := checkAccessControlMethods(serviceMethods)
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
	err = checkAccessControlMethods(serviceMethods)
	expectedErr = "missing required method(s) in access control table: [Dummy]"
	if err == nil {
		t.Errorf("Expected error: %q, got nil", expectedErr)
	}
	if err.Error() != expectedErr {
		t.Errorf("Expected error: %q, got: %q", expectedErr, err.Error())
	}
}

func has(method string) bool {
	_, ok := accessRolesFor[method]
	return ok
}

func checkAccessControlMethods(expectedMethodNames map[string]bool) error {
	missingMethods := []string{}
	superfluousMethods := []string{}
	for method := range expectedMethodNames {
		if !has(method) {
			missingMethods = append(missingMethods, method)
		}
	}
	for method := range accessRolesFor {
		if !expectedMethodNames[method] {
			superfluousMethods = append(superfluousMethods, method)
		}
	}
	if len(missingMethods) > 0 {
		return fmt.Errorf("missing required method(s) in access control table: %v", missingMethods)
	}
	if len(superfluousMethods) > 0 {
		return fmt.Errorf("superfluous method(s) in access control table: %v", superfluousMethods)
	}
	return nil
}
