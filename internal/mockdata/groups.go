package mockdata

import (
	"fmt"

	"github.com/quickfeed/quickfeed/qf"
)

func (g *generator) groups() error {
	for i := 1; i < len(g.Courses); i++ {
		id := uint64(g.Teachers + 1)
		groupStatus := qf.Group_PENDING
		if i != 1 {
			groupStatus = qf.Group_APPROVED
		}
		for j := range g.EnrolledStudents / 2 {
			group := &qf.Group{
				Name:     fmt.Sprintf("%s, %d", g.Groups[j], i),
				CourseID: uint64(i),
				Status:   groupStatus,
				Users:    []*qf.User{{ID: id}, {ID: id + 1}},
			}
			id += 2
			if err := g.db.CreateGroup(group); err != nil {
				return err
			}
		}
	}
	return nil
}
