package interceptor

import (
	"strings"
	"testing"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// TODO: These tests will not complain if a message is added that **should** implement the interface, but does not.
// This is because the test will fail only if a message is added to the map that does **not** implement the interface.
func TestImplementsIdCleaner(t *testing.T) {
	var tests = map[protoreflect.FullName]struct {
		cleaner   bool
		validator bool
		found     bool
	}{
		"qf.Void":                        {cleaner: false, validator: true},
		"qf.User":                        {cleaner: true, validator: true},
		"qf.Users":                       {cleaner: true, validator: false},
		"qf.Submission":                  {cleaner: false, validator: false},
		"qf.Submissions":                 {cleaner: false, validator: false},
		"qf.Enrollment":                  {cleaner: true, validator: true},
		"qf.Enrollments":                 {cleaner: true, validator: true},
		"qf.Assignment":                  {cleaner: false, validator: false},
		"qf.Course":                      {cleaner: true, validator: true},
		"qf.Courses":                     {cleaner: true, validator: false},
		"qf.Group":                       {cleaner: true, validator: true},
		"qf.Groups":                      {cleaner: true, validator: false},
		"qf.Reviewers":                   {cleaner: true, validator: false},
		"qf.SubmissionLink":              {cleaner: false, validator: false},
		"qf.OrgRequest":                  {cleaner: false, validator: true},
		"qf.Repository":                  {cleaner: false, validator: false},
		"qf.UpdateSubmissionsRequest":    {cleaner: false, validator: false},
		"qf.URLRequest":                  {cleaner: false, validator: true},
		"qf.RebuildRequest":              {cleaner: false, validator: true},
		"qf.CourseRequest":               {cleaner: false, validator: true},
		"qf.PullRequest":                 {cleaner: false, validator: false},
		"qf.Assignments":                 {cleaner: false, validator: false},
		"qf.UserRequest":                 {cleaner: false, validator: true},
		"qf.Status":                      {cleaner: false, validator: false},
		"qf.GradingBenchmark":            {cleaner: false, validator: true},
		"qf.Review":                      {cleaner: false, validator: true},
		"qf.Benchmarks":                  {cleaner: false, validator: false},
		"qf.Issue":                       {cleaner: false, validator: false},
		"qf.RemoteIdentity":              {cleaner: false, validator: false},
		"qf.UpdateSubmissionRequest":     {cleaner: false, validator: true},
		"qf.UsedSlipDays":                {cleaner: false, validator: false},
		"qf.Task":                        {cleaner: false, validator: false},
		"qf.Organizations":               {cleaner: false, validator: false},
		"qf.GradingCriterion":            {cleaner: false, validator: true},
		"qf.SubmissionReviewersRequest":  {cleaner: false, validator: true},
		"qf.Repositories":                {cleaner: false, validator: false},
		"qf.CourseSubmissions":           {cleaner: true, validator: false},
		"qf.Organization":                {cleaner: false, validator: true},
		"qf.GetGroupRequest":             {cleaner: false, validator: true},
		"qf.EnrollmentStatusRequest":     {cleaner: false, validator: true},
		"qf.CourseUserRequest":           {cleaner: false, validator: true},
		"qf.SubmissionRequest":           {cleaner: false, validator: true},
		"qf.ReviewRequest":               {cleaner: false, validator: true},
		"qf.RepositoryRequest":           {cleaner: false, validator: true},
		"qf.GroupRequest":                {cleaner: false, validator: true},
		"qf.EnrollmentLink":              {cleaner: true, validator: false},
		"qf.EnrollmentRequest":           {cleaner: false, validator: true},
		"qf.SubmissionsForCourseRequest": {cleaner: false, validator: true},
		"score.Score":                    {cleaner: false, validator: false},
		"score.BuildInfo":                {cleaner: false, validator: false},
	}

	protoregistry.GlobalTypes.RangeMessages(func(desc protoreflect.MessageType) bool {
		name := desc.Descriptor().FullName()
		qfMessage := strings.HasPrefix(string(name), "qf.")
		scoreMessage := strings.HasPrefix(string(name), "score.")
		if !(qfMessage || scoreMessage) {
			return true
		}
		test, ok := tests[name]
		if !ok {
			t.Errorf("Message %s has not been added to the test struct", name)
			return true
		}
		msg := desc.Zero().Interface()

		_, ok = msg.(idCleaner)
		if test.cleaner && !ok {
			t.Errorf("Message %s does not implement idCleaner", name)
		} else if !test.cleaner && ok {
			t.Errorf("Message %s implements idCleaner, but should not", name)
		}

		_, ok = msg.(validator)
		if test.validator && !ok {
			t.Errorf("Message %s does not implement validator", name)
		} else if !test.validator && ok {
			t.Errorf("Message %s implements validator, but should not", name)
		}

		test.found = true
		tests[name] = test
		return true
	})

	for name, test := range tests {
		if !test.found {
			t.Errorf("Message %s is tested, but no longer exists", name)
		}
	}
}

// func TestImplementsValidator(t *testing.T) {
// 	protoregistry.GlobalTypes.RangeMessages(func(desc protoreflect.MessageType) bool {
// 		if _, ok := shouldImplementValidator[desc.Descriptor().FullName()]; !ok {
// 			return true
// 		}
// 		if _, ok := desc.Zero().Interface().(validator); !ok {
// 			t.Errorf("type %s should implement validator", desc.Descriptor().FullName())
// 		}
// 		return true
// 	})
// }
