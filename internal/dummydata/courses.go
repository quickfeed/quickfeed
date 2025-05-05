package dummydata

import (
	"fmt"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

/*

TODO(Joachim): Evaluate if we need this. Can be helpful if a certain amount of dummy data is needed
// We can have a json file containing options for what to generate.

type courseGenOptions struct {
	enrolledUsers int
}

var courseMap = map[string]courseGenOptions{
	qtest.DAT520: {
		enrolledUsers: 2,
	},
	qtest.DAT320: {
		enrolledUsers: 2,
	},
	qtest.DATx20: {
		enrolledUsers: 2,
	},
	qtest.QF104: {
		enrolledUsers: 2,
	},
}*/

func (g *generator) courses() error {
	for _, course := range qtest.MockCourses {
		if err := g.db.CreateCourse(course.GetCourseCreatorID(), course); err != nil {
			return err
		}
		if err := g.assignments(course); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) assignments(course *qf.Course) error {
	for i := range 8 {
		assignment := &qf.Assignment{
			Order:    uint32(i + 1),
			CourseID: course.GetID(),
			Name:     fmt.Sprintf("Lab %d", i+1),
		}
		if err := g.db.CreateAssignment(assignment); err != nil {
			return err
		}
	}
	return nil
}
