package interceptor

import (
	"testing"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

var shouldImplementIdCleaner = map[protoreflect.FullName]struct{}{
	"qf.User":              {},
	"qf.Users":             {},
	"qf.Group":             {},
	"qf.Groups":            {},
	"qf.Enrollment":        {},
	"qf.Enrollments":       {},
	"qf.Course":            {},
	"qf.Courses":           {},
	"qf.EnrollmentLink":    {},
	"qf.CourseSubmissions": {},
	"qf.Reviewers":         {},
}

var shouldImplementValidator = map[protoreflect.FullName]struct{}{
	"qf.Void":                        {},
	"qf.User":                        {},
	"qf.Group":                       {},
	"qf.Course":                      {},
	"qf.UserRequest":                 {},
	"qf.Enrollment":                  {},
	"qf.Enrollments":                 {},
	"qf.CourseRequest":               {},
	"qf.OrgRequest":                  {},
	"qf.URLRequest":                  {},
	"qf.EnrollmentStatusRequest":     {},
	"qf.RepositoryRequest":           {},
	"qf.SubmissionRequest":           {},
	"qf.UpdateSubmissionRequest":     {},
	"qf.GetGroupRequest":             {},
	"qf.GroupRequest":                {},
	"qf.EnrollmentRequest":           {},
	"qf.SubmissionsForCourseRequest": {},
	"qf.RebuildRequest":              {},
	"qf.SubmissionReviewersRequest":  {},
	"qf.ReviewRequest":               {},
	"qf.CourseUserRequest":           {},
	"qf.Organization":                {},
	"qf.Review":                      {},
	"qf.GradingBenchmark":            {},
	"qf.GradingCriterion":            {},
}

func TestImplementsIdCleaner(t *testing.T) {
	protoregistry.GlobalTypes.RangeMessages(func(desc protoreflect.MessageType) bool {
		if _, ok := shouldImplementIdCleaner[desc.Descriptor().FullName()]; !ok {
			return true
		}
		if _, ok := desc.Zero().Interface().(idCleaner); !ok {
			t.Errorf("type %s should implement idCleaner", desc.Descriptor().FullName())
		}
		return true
	})
}

func TestImplementsValidator(t *testing.T) {
	protoregistry.GlobalTypes.RangeMessages(func(desc protoreflect.MessageType) bool {
		if _, ok := shouldImplementValidator[desc.Descriptor().FullName()]; !ok {
			return true
		}
		if _, ok := desc.Zero().Interface().(validator); !ok {
			t.Errorf("type %s should implement validator", desc.Descriptor().FullName())
		}
		return true
	})
}
