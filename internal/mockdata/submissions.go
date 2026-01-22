package mockdata

import (
	"math/rand"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

func (g *generator) submissions() error {
	for i := range qtest.MockCourses {
		var baseAssignmentID = i * assingnmentsPerCourse
		g.studentSubs(baseAssignmentID)
		g.groupSubs(baseAssignmentID)
	}
	return nil
}

func (g *generator) studentSubs(baseAssignmentID int) error {
	for k := 1; k <= studentSubmissionsPerAssignment; k++ {
		for j := 1; j <= enrolledStudents; j++ {
			submission := &qf.Submission{
				AssignmentID: uint64(k + baseAssignmentID),
				UserID:       uint64(j),
				Score:        uint32(rand.Intn(100) + 1),
			}
			if err := g.db.CreateSubmission(submission); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *generator) groupSubs(baseAssignmentID int) error {
	for k := assingnmentsPerCourse; k <= groupSubmissionsPerAssignment; k++ {
		for j := range qtest.Groups {
			submission := &qf.Submission{
				AssignmentID: uint64(k + baseAssignmentID),
				GroupID:      uint64(j),
			}
			if err := g.db.CreateSubmission(submission); err != nil {
				return err
			}
		}
	}
	return nil
}
