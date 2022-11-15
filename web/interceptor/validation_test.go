package interceptor

import (
	"strings"
	"testing"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func TestImplementsValidation(t *testing.T) {
	const (
		T = true
		F = false
	)
	tests := map[protoreflect.FullName]*struct {
		cleaner   bool
		validator bool
		found     bool
	}{
		"qf.Void":                        {cleaner: F, validator: T},
		"qf.User":                        {cleaner: T, validator: T},
		"qf.Users":                       {cleaner: T, validator: F},
		"qf.Submission":                  {cleaner: F, validator: F},
		"qf.Submissions":                 {cleaner: F, validator: F},
		"qf.Enrollment":                  {cleaner: T, validator: T},
		"qf.Enrollments":                 {cleaner: T, validator: T},
		"qf.Assignment":                  {cleaner: F, validator: F},
		"qf.Course":                      {cleaner: T, validator: T},
		"qf.Courses":                     {cleaner: T, validator: F},
		"qf.Group":                       {cleaner: T, validator: T},
		"qf.Groups":                      {cleaner: T, validator: F},
		"qf.Reviewers":                   {cleaner: T, validator: F},
		"qf.SubmissionLink":              {cleaner: F, validator: F},
		"qf.OrgRequest":                  {cleaner: F, validator: T},
		"qf.Repository":                  {cleaner: F, validator: F},
		"qf.UpdateSubmissionsRequest":    {cleaner: F, validator: F},
		"qf.URLRequest":                  {cleaner: F, validator: T},
		"qf.RebuildRequest":              {cleaner: F, validator: T},
		"qf.CourseRequest":               {cleaner: F, validator: T},
		"qf.PullRequest":                 {cleaner: F, validator: F},
		"qf.Assignments":                 {cleaner: F, validator: F},
		"qf.UserRequest":                 {cleaner: F, validator: T},
		"qf.Status":                      {cleaner: F, validator: F},
		"qf.GradingBenchmark":            {cleaner: F, validator: T},
		"qf.Review":                      {cleaner: F, validator: T},
		"qf.Benchmarks":                  {cleaner: F, validator: F},
		"qf.Issue":                       {cleaner: F, validator: F},
		"qf.RemoteIdentity":              {cleaner: F, validator: F},
		"qf.UpdateSubmissionRequest":     {cleaner: F, validator: T},
		"qf.UsedSlipDays":                {cleaner: F, validator: F},
		"qf.Task":                        {cleaner: F, validator: F},
		"qf.Organizations":               {cleaner: F, validator: F},
		"qf.GradingCriterion":            {cleaner: F, validator: T},
		"qf.SubmissionReviewersRequest":  {cleaner: F, validator: T},
		"qf.Repositories":                {cleaner: F, validator: F},
		"qf.CourseSubmissions":           {cleaner: T, validator: F},
		"qf.Organization":                {cleaner: F, validator: T},
		"qf.GetGroupRequest":             {cleaner: F, validator: T},
		"qf.EnrollmentStatusRequest":     {cleaner: F, validator: T},
		"qf.SubmissionRequest":           {cleaner: F, validator: T},
		"qf.ReviewRequest":               {cleaner: F, validator: T},
		"qf.RepositoryRequest":           {cleaner: F, validator: T},
		"qf.GroupRequest":                {cleaner: F, validator: T},
		"qf.EnrollmentLink":              {cleaner: T, validator: F},
		"qf.EnrollmentRequest":           {cleaner: F, validator: T},
		"qf.SubmissionsForCourseRequest": {cleaner: F, validator: T},
		"score.Score":                    {cleaner: F, validator: F},
		"score.BuildInfo":                {cleaner: F, validator: F},
	}

	protoregistry.GlobalTypes.RangeMessages(func(desc protoreflect.MessageType) bool {
		name := desc.Descriptor().FullName()
		qfMessage := strings.HasPrefix(string(name), "qf.")
		scoreMessage := strings.HasPrefix(string(name), "score.")

		// GlobalTypes includes all registered messages.
		// We only want to test messages from the qf and score packages.
		if !(qfMessage || scoreMessage) {
			return true
		}
		test, ok := tests[name]
		if !ok {
			t.Errorf("Message %s has not been added to the test struct", name)
			return true
		}
		test.found = true
		msg := desc.Zero().Interface()

		if _, ok = msg.(idCleaner); test.cleaner && !ok {
			t.Errorf("Message %s does not implement idCleaner", name)
		} else if !test.cleaner && ok {
			t.Errorf("Message %s implements idCleaner, but should not", name)
		}

		if _, ok = msg.(validator); test.validator && !ok {
			t.Errorf("Message %s does not implement validator", name)
		} else if !test.validator && ok {
			t.Errorf("Message %s implements validator, but should not", name)
		}
		return true
	})

	for name, test := range tests {
		if !test.found {
			t.Errorf("Message %s is tested, but no longer exists", name)
		}
	}
}
