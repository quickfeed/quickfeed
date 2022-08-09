package interceptor

type (
	role      int
	roles     []role
	requestID interface {
		FetchID(string) uint64
	}
)

const (
	// user role implies that user attempts to access information about himself.
	user role = iota
	// group role implies that the user is a course student + a member of the given group.
	group
	// student role implies that the user is enrolled in the course with any role.
	student
	// teacher: user enrolled in the course with teacher status.
	teacher
	// courseAdmin: an admin user who is also enrolled into the course.
	courseAdmin
	// admin is the user with admin privileges.
	admin
)

// If there are several roles that can call a method, a role with the least privilege must come first.
// If method is not in the map, there is no restrictions to call it.
var access = map[string]roles{
	"GetEnrollmentsByCourse":  {student, teacher},
	"UpdateUser":              {user, admin},
	"GetEnrollmentsByUser":    {user, admin},
	"GetSubmissions":          {user, group, teacher, courseAdmin},
	"GetGroupByUserAndCourse": {group, teacher},
	"CreateGroup":             {group, teacher},
	"GetGroup":                {group, teacher},
	"UpdateGroup":             {teacher},
	"DeleteGroup":             {teacher},
	"IsEmptyRepo":             {teacher},
	"GetGroupsByCourse":       {teacher},
	"UpdateCourse":            {teacher},
	"UpdateEnrollments":       {teacher},
	"UpdateSubmission":        {teacher},
	"RebuildSubmissions":      {teacher},
	"CreateBenchmark":         {teacher},
	"UpdateBenchmark":         {teacher},
	"DeleteBenchmark":         {teacher},
	"CreateCriterion":         {teacher},
	"UpdateCriterion":         {teacher},
	"DeleteCriterion":         {teacher},
	"CreateReview":            {teacher},
	"UpdateReview":            {teacher},
	"UpdateSubmissions":       {teacher},
	"GetReviewers":            {teacher},
	"UpdateAssignments":       {teacher},
	"GetSubmissionsByCourse":  {teacher, courseAdmin},
	"GetUserByCourse":         {teacher, admin},
	"GetOrganization":         {admin},
	"CreateCourse":            {admin},
}
