package mockdata

import "github.com/quickfeed/quickfeed/database"

// TODO(Joachim): Evaluate if we need this. Can be helpful if a certain amount of dummy data is needed
// We can have a json file containing options for what to generate.

// type courseGenOptions struct {
// 	enrolledUsers int
// }

// var courseMap = map[string]courseGenOptions{
// 	qtest.DAT520: {
// 		enrolledUsers: 2,
// 	},
// 	qtest.DAT320: {
// 		enrolledUsers: 2,
// 	},
// 	qtest.DATx20: {
// 		enrolledUsers: 2,
// 	},
// 	qtest.QF104: {
// 		enrolledUsers: 2,
// 	},
// }

type generator struct {
	db database.Database
	config
}

type config struct {
	Teachers                        int
	Students                        int
	EnrolledStudents                int
	AssingnmentsPerCourse           int
	GroupAssignments                int
	StudentSubmissionsPerAssignment int
	GroupSubmissionsPerAssignment   int
	ContainerTimeout                int
	Verbose                         bool
	Organizations                   []string
	Members                         []string
	Courses                         []string
	Groups                          []string
}
