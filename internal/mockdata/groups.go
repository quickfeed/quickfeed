package mockdata

import (
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

func (g *generator) groups() error {
	for i := 1; i <= courses; i++ {
		id := uint64(teachers + 1)
		groupStatus := qf.Group_PENDING
		if i != 1 {
			groupStatus = qf.Group_APPROVED
		}
		for j := range enrolledStudents / 2 {
			group := &qf.Group{
				Name:     qtest.Groups[j],
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
