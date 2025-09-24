package dummydata

import (
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

func (g generator) groups() error {
	// Create a new user for each name in the list
	for i := range qtest.MockCourses {
		id := uint64(6)
		for _, name := range qtest.Groups {
			group := &qf.Group{
				Name:     name,
				CourseID: uint64(i + 1),
				Users: []*qf.User{
					{
						ID: id,
					},
					{
						ID: id + 1,
					},
				},
			}
			id += 2
			if err := g.db.CreateGroup(group); err != nil {
				return err
			}
		}
	}

	return nil
}
