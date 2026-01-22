package mockdata

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (g *generator) courses() error {
	for i, course := range qtest.MockCourses {
		if err := g.db.CreateCourse(course.GetCourseCreatorID(), course); err != nil {
			return err
		}
		// orderID -> taskName -> task
		tasks := make(map[uint32]map[string]*qf.Task)
		for j := 1; j <= assingnmentsPerCourse; j++ {
			assignment := &qf.Assignment{
				ID:               uint64(i*assingnmentsPerCourse + j),
				Deadline:         timestamppb.New(time.Now().Add(time.Duration(i) * 24 * time.Hour)),
				ScoreLimit:       uint32(rand.Intn(41) + 60),
				AutoApprove:      rand.Intn(4) == 0,
				Order:            uint32(j),
				CourseID:         course.GetID(),
				Name:             fmt.Sprintf("Lab %d", j),
				ContainerTimeout: containerTimeout,
				IsGroupLab:       j > assingnmentsPerCourse-groupAssignments,
			}
			tasks[assignment.GetOrder()] = taskMap(assignment)
			if err := g.db.CreateAssignment(assignment); err != nil {
				return err
			}
		}
		g.db.SynchronizeAssignmentTasks(course, tasks)
	}
	return nil
}

// taskMap returns a map of tasks related to an assignment order
func taskMap(assignment *qf.Assignment) map[string]*qf.Task {
	tmap := make(map[string]*qf.Task)
	for i := 1; i < rand.Intn(5)+1; i++ {
		var issues []*qf.Issue
		for j := 1; j <= 3; j++ {
			issues = append(issues, &qf.Issue{
				RepositoryID:   uint64(j + i), // TODO(Joachim): sync repository IDs
				ScmIssueNumber: uint64(j),
			})
		}
		name := fmt.Sprintf("Task %d", i)
		tmap[name] = &qf.Task{
			AssignmentID:    assignment.GetID(),
			AssignmentOrder: assignment.GetOrder(),
			Title:           fmt.Sprintf("Task %d Title", i),
			Name:            name,
			Body:            fmt.Sprintf("This is the description for task %d", i),
			Issues:          issues,
		}
	}
	return tmap
}
