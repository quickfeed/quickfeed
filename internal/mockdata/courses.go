package mockdata

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (g *generator) courses() error {
	for i, name := range g.Courses {
		course := &qf.Course{
			ID:              uint64(i + 1),
			Name:            name,
			Code:            fmt.Sprintf("CODE%d", 100+i),
			Year:            uint32(time.Now().Year() + i%2),
			Tag:             map[int]string{0: "Spring", 1: "Fall"}[i%2],
			CourseCreatorID: 1, // All courses created by first user (admin)
		}
		if err := g.db.CreateCourse(course.GetCourseCreatorID(), course); err != nil {
			return err
		}
		// orderID -> taskName -> task
		tasks := make(map[uint32]map[string]*qf.Task)
		for j := 1; j <= g.AssingnmentsPerCourse; j++ {
			assignment := &qf.Assignment{
				ID:               uint64(i*g.AssingnmentsPerCourse + j),
				Deadline:         timestamppb.New(time.Now().Add(time.Duration(i) * 24 * time.Hour)),
				ScoreLimit:       uint32(rand.Intn(41) + 60),
				AutoApprove:      rand.Intn(4) == 0,
				Order:            uint32(j),
				CourseID:         course.GetID(),
				Name:             fmt.Sprintf("Lab %d", j),
				ContainerTimeout: uint32(g.config.containerTimeout),
				IsGroupLab:       j > g.AssingnmentsPerCourse-g.GroupAssignments,
				Reviewers:        uint32(rand.Intn(2)),
			}
			tasks[assignment.GetOrder()] = taskMap(assignment)
			if err := g.db.CreateAssignment(assignment); err != nil {
				return err
			}
			for k := 1; k <= 5; k++ {
				if err := g.db.CreateBenchmark(&qf.GradingBenchmark{
					CourseID:     course.GetID(),
					AssignmentID: assignment.GetID(),
					Heading:      fmt.Sprintf("Benchmark %d", k),
					Comment:      fmt.Sprintf("This is the comment for benchmark %d", k),
					Criteria: []*qf.GradingCriterion{
						{
							Description: "Criterion 1 ",
							Points:      uint64(rand.Intn(100) + 1),
						},
						{
							Description: "Criterion 2",
							Points:      uint64(rand.Intn(100) + 1),
						},
					},
				}); err != nil {
					return err
				}
			}
		}
		if _, _, err := g.db.SynchronizeAssignmentTasks(course, tasks); err != nil {
			return err
		}
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
