package dummydata

import (
	"fmt"
	"log"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

func (g *generator) courses() error {
	log.Println("Generating courses...")
	for _, course := range qtest.MockCourses {
		if err := g.db.CreateCourse(course.GetCourseCreatorID(), course); err != nil {
			return err
		}
		for i := 1; i <= assingnmentsPerCourse; i++ {
			assignment := &qf.Assignment{
				Order:    uint32(i),
				CourseID: course.GetID(),
				Name:     fmt.Sprintf("Lab %d", i),
			}
			if i > assingnmentsPerCourse-groupAssignments {
				assignment.IsGroupLab = true
			}
			if err := g.db.CreateAssignment(assignment); err != nil {
				return err
			}
		}
	}
	return nil
}
