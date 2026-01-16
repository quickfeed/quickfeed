package interceptor

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/reflect/protoreflect"
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
		"qf.Assignment":              {cleaner: F, validator: F},
		"qf.AssignmentFeedback":      {cleaner: F, validator: T},
		"qf.AssignmentFeedbacks":     {cleaner: F, validator: F},
		"qf.Assignments":             {cleaner: F, validator: F},
		"qf.Benchmarks":              {cleaner: F, validator: F},
		"qf.Course":                  {cleaner: T, validator: T},
		"qf.CourseRequest":           {cleaner: F, validator: T},
		"qf.CourseSubmissions":       {cleaner: F, validator: F},
		"qf.Courses":                 {cleaner: T, validator: F},
		"qf.Enrollment":              {cleaner: T, validator: T},
		"qf.EnrollmentRequest":       {cleaner: F, validator: T},
		"qf.Enrollments":             {cleaner: T, validator: T},
		"qf.FeedbackReceipt":         {cleaner: F, validator: F},
		"qf.Grade":                   {cleaner: F, validator: F},
		"qf.GradingBenchmark":        {cleaner: F, validator: T},
		"qf.GradingCriterion":        {cleaner: F, validator: T},
		"qf.Group":                   {cleaner: T, validator: T},
		"qf.GroupRequest":            {cleaner: F, validator: T},
		"qf.Groups":                  {cleaner: T, validator: F},
		"qf.Issue":                   {cleaner: F, validator: F},
		"qf.Organization":            {cleaner: F, validator: T},
		"qf.PullRequest":             {cleaner: F, validator: F},
		"qf.RebuildRequest":          {cleaner: F, validator: T},
		"qf.Repositories":            {cleaner: F, validator: F},
		"qf.Repository":              {cleaner: F, validator: F},
		"qf.RepositoryRequest":       {cleaner: F, validator: T},
		"qf.Review":                  {cleaner: F, validator: T},
		"qf.ReviewRequest":           {cleaner: F, validator: T},
		"qf.Submission":              {cleaner: F, validator: F},
		"qf.SubmissionRequest":       {cleaner: F, validator: T},
		"qf.Submissions":             {cleaner: F, validator: F},
		"qf.Task":                    {cleaner: F, validator: F},
		"qf.TestInfo":                {cleaner: F, validator: F},
		"qf.UpdateSubmissionRequest": {cleaner: F, validator: T},
		"qf.UsedSlipDays":            {cleaner: F, validator: F},
		"qf.User":                    {cleaner: T, validator: T},
		"qf.Users":                   {cleaner: T, validator: F},
		"qf.Void":                    {cleaner: F, validator: T},
		"score.BuildInfo":            {cleaner: F, validator: F},
		"score.Score":                {cleaner: F, validator: F},
	}

	protoregistry.GlobalTypes.RangeMessages(func(desc protoreflect.MessageType) bool {
		name := desc.Descriptor().FullName()
		qfMessage := strings.HasPrefix(string(name), "qf.")
		scoreMessage := strings.HasPrefix(string(name), "score.")

		// GlobalTypes includes all registered messages.
		// We only want to test messages from the qf and score packages.
		if !qfMessage && !scoreMessage {
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

func TestIsValid(t *testing.T) {
	tests := map[string]struct {
		request validator
		want    bool
	}{
		"AssignmentFeedback/EmptyImprovement":      {request: &qf.AssignmentFeedback{CourseID: 1, AssignmentID: 1, LikedContent: "A", TimeSpent: 1}, want: false},
		"AssignmentFeedback/EmptyLikedContent":     {request: &qf.AssignmentFeedback{CourseID: 1, AssignmentID: 1, ImprovementSuggestions: "B", TimeSpent: 1}, want: false},
		"AssignmentFeedback/Invalid":               {request: &qf.AssignmentFeedback{}, want: false},
		"AssignmentFeedback/MissingAssignmentID":   {request: &qf.AssignmentFeedback{CourseID: 1, LikedContent: "A", ImprovementSuggestions: "B", TimeSpent: 1}, want: false},
		"AssignmentFeedback/MissingCourseID":       {request: &qf.AssignmentFeedback{AssignmentID: 1, LikedContent: "A", ImprovementSuggestions: "B", TimeSpent: 1}, want: false},
		"AssignmentFeedback/Valid":                 {request: &qf.AssignmentFeedback{CourseID: 1, AssignmentID: 1, LikedContent: "A", ImprovementSuggestions: "B", TimeSpent: 1}, want: true},
		"AssignmentFeedback/ZeroTimeSpent":         {request: &qf.AssignmentFeedback{CourseID: 1, AssignmentID: 1, LikedContent: "A", ImprovementSuggestions: "B"}, want: false},
		"Course/Invalid":                           {request: &qf.Course{}, want: false},
		"Course/Valid":                             {request: &qf.Course{Name: "A", Code: "B", ScmOrganizationID: 1, Year: 2021, Tag: "C"}, want: true},
		"CourseRequest/Invalid":                    {request: &qf.CourseRequest{CourseID: 0}, want: false},
		"CourseRequest/Valid":                      {request: &qf.CourseRequest{CourseID: 1}, want: true},
		"Enrollment/Invalid":                       {request: &qf.Enrollment{}, want: false},
		"Enrollment/Status/Invalid":                {request: &qf.Enrollment{Status: 10, UserID: 1, CourseID: 1}, want: false},
		"Enrollment/StatusNone":                    {request: &qf.Enrollment{Status: qf.Enrollment_NONE, UserID: 1, CourseID: 1}, want: true},
		"Enrollment/StatusPending":                 {request: &qf.Enrollment{Status: qf.Enrollment_PENDING, UserID: 1, CourseID: 1}, want: true},
		"Enrollment/StatusStudent":                 {request: &qf.Enrollment{Status: qf.Enrollment_STUDENT, UserID: 1, CourseID: 1}, want: true},
		"Enrollment/StatusTeacher":                 {request: &qf.Enrollment{Status: qf.Enrollment_TEACHER, UserID: 1, CourseID: 1}, want: true},
		"EnrollmentRequest/CourseID":               {request: &qf.EnrollmentRequest{FetchMode: &qf.EnrollmentRequest_CourseID{CourseID: 1}}, want: true},
		"EnrollmentRequest/CourseID/Invalid":       {request: &qf.EnrollmentRequest{FetchMode: &qf.EnrollmentRequest_CourseID{CourseID: 0}}, want: false},
		"EnrollmentRequest/Invalid":                {request: &qf.EnrollmentRequest{}, want: false},
		"EnrollmentRequest/UserID":                 {request: &qf.EnrollmentRequest{FetchMode: &qf.EnrollmentRequest_UserID{UserID: 1}}, want: true},
		"EnrollmentRequest/UserID/Invalid":         {request: &qf.EnrollmentRequest{FetchMode: &qf.EnrollmentRequest_UserID{UserID: 0}}, want: false},
		"Enrollments/DifferentCourseIDs":           {request: &qf.Enrollments{Enrollments: []*qf.Enrollment{{CourseID: 1, UserID: 1}, {CourseID: 2, UserID: 2}}}, want: false},
		"Enrollments/Invalid":                      {request: &qf.Enrollments{}, want: false},
		"Enrollments/InvalidEnrollment":            {request: &qf.Enrollments{Enrollments: []*qf.Enrollment{{CourseID: 1, UserID: 0}}}, want: false},
		"Enrollments/Valid":                        {request: &qf.Enrollments{Enrollments: []*qf.Enrollment{{CourseID: 1, UserID: 1, Status: qf.Enrollment_STUDENT}}}, want: true},
		"GradingBenchmark/EmptyHeading":            {request: &qf.GradingBenchmark{AssignmentID: 1}, want: false},
		"GradingBenchmark/Invalid":                 {request: &qf.GradingBenchmark{}, want: false},
		"GradingBenchmark/MissingAssignmentID":     {request: &qf.GradingBenchmark{Heading: "A"}, want: false},
		"GradingBenchmark/Valid":                   {request: &qf.GradingBenchmark{AssignmentID: 1, Heading: "A"}, want: true},
		"GradingCriterion/EmptyDescription":        {request: &qf.GradingCriterion{BenchmarkID: 1}, want: false},
		"GradingCriterion/Invalid":                 {request: &qf.GradingCriterion{}, want: false},
		"GradingCriterion/MissingBenchmarkID":      {request: &qf.GradingCriterion{Description: "A"}, want: false},
		"GradingCriterion/Valid":                   {request: &qf.GradingCriterion{BenchmarkID: 1, Description: "A"}, want: true},
		"Group/Invalid":                            {request: &qf.Group{}, want: false},
		"Group/Valid":                              {request: &qf.Group{Name: "A", CourseID: 1, Users: []*qf.User{{ID: 1}}}, want: true},
		"GroupRequest/GroupID":                     {request: &qf.GroupRequest{CourseID: 1, GroupID: 1}, want: true},
		"GroupRequest/Invalid":                     {request: &qf.GroupRequest{CourseID: 1, UserID: 1, GroupID: 1}, want: false},
		"GroupRequest/UserID":                      {request: &qf.GroupRequest{CourseID: 1, UserID: 1}, want: true},
		"Organization/Invalid":                     {request: &qf.Organization{}, want: false},
		"Organization/Valid":                       {request: &qf.Organization{ScmOrganizationName: "A"}, want: true},
		"RebuildRequest/Invalid":                   {request: &qf.RebuildRequest{CourseID: 1}, want: false},
		"RebuildRequest/Valid":                     {request: &qf.RebuildRequest{CourseID: 1, AssignmentID: 1}, want: true},
		"RepositoryRequest/GroupID":                {request: &qf.RepositoryRequest{CourseID: 1, GroupID: 1}, want: true},
		"RepositoryRequest/Invalid":                {request: &qf.RepositoryRequest{CourseID: 1}, want: false},
		"RepositoryRequest/UserID":                 {request: &qf.RepositoryRequest{CourseID: 1, UserID: 1}, want: true},
		"RepositoryRequest/UserID/GroupID":         {request: &qf.RepositoryRequest{CourseID: 1, UserID: 1, GroupID: 1}, want: false},
		"Review/Invalid":                           {request: &qf.Review{}, want: false},
		"Review/Valid":                             {request: &qf.Review{ReviewerID: 1, SubmissionID: 1}, want: true},
		"ReviewRequest/MissingReview":              {request: &qf.ReviewRequest{CourseID: 1}, want: false},
		"ReviewRequest/Valid":                      {request: &qf.ReviewRequest{CourseID: 1, Review: &qf.Review{ReviewerID: 1, SubmissionID: 1}}, want: true},
		"SubmissionRequest/GroupID":                {request: &qf.SubmissionRequest{CourseID: 1, FetchMode: &qf.SubmissionRequest_GroupID{GroupID: 1}}, want: true},
		"SubmissionRequest/Invalid":                {request: &qf.SubmissionRequest{CourseID: 1}, want: false},
		"SubmissionRequest/MissingCourseID":        {request: &qf.SubmissionRequest{FetchMode: &qf.SubmissionRequest_SubmissionID{SubmissionID: 1}}, want: false},
		"SubmissionRequest/SubmissionID":           {request: &qf.SubmissionRequest{CourseID: 1, FetchMode: &qf.SubmissionRequest_SubmissionID{SubmissionID: 1}}, want: true},
		"SubmissionRequest/Type":                   {request: &qf.SubmissionRequest{CourseID: 1, FetchMode: &qf.SubmissionRequest_Type{Type: qf.SubmissionRequest_ALL}}, want: true},
		"SubmissionRequest/UserID":                 {request: &qf.SubmissionRequest{CourseID: 1, FetchMode: &qf.SubmissionRequest_UserID{UserID: 1}}, want: true},
		"UpdateSubmissionRequest/Invalid":          {request: &qf.UpdateSubmissionRequest{CourseID: 1, SubmissionID: 1, AssignmentID: 1}, want: false},
		"UpdateSubmissionRequest/OnlyCourseID":     {request: &qf.UpdateSubmissionRequest{CourseID: 1}, want: false},
		"UpdateSubmissionRequest/OnlySubmissionID": {request: &qf.UpdateSubmissionRequest{SubmissionID: 1}, want: false},
		"UpdateSubmissionRequest/OnlyAssignmentID": {request: &qf.UpdateSubmissionRequest{AssignmentID: 1}, want: false},
		"UpdateSubmissionRequest/AssignmentID":     {request: &qf.UpdateSubmissionRequest{CourseID: 1, AssignmentID: 1}, want: true},
		"UpdateSubmissionRequest/SubmissionID":     {request: &qf.UpdateSubmissionRequest{CourseID: 1, SubmissionID: 1}, want: true},
		"User/Invalid":                             {request: &qf.User{ID: 0}, want: false},
		"User/Valid":                               {request: &qf.User{ID: 1}, want: true},
		"Void/Valid":                               {request: &qf.Void{}, want: true},
	}
	// Run tests in sorted order for easier reading of test results.
	for _, name := range slices.Sorted(maps.Keys(tests)) {
		test := tests[name]
		t.Run(name, func(t *testing.T) {
			if got := test.request.IsValid(); got != test.want {
				t.Errorf("IsValid() = %v, want %v", got, test.want)
			}
		})
	}
}
