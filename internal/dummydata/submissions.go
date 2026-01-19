package dummydata

import (
	"log"
	"math/rand"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

func (g *generator) submissions() error {
	log.Println("Generating submissions")
	for i := 0; i < len(qtest.MockCourses); i++ {
		var baseAssignmentID = i * assingnmentsPerCourse
		// Create a new user for each name in the list
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
		for k := assingnmentsPerCourse - groupAssignments; k <= groupSubmissionsPerAssignment; k++ {
			for j := 1; j <= len(qtest.Groups); j++ {
				submission := &qf.Submission{
					AssignmentID: uint64(k + baseAssignmentID),
					GroupID:      uint64(j),
				}
				if err := g.db.CreateSubmission(submission); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
