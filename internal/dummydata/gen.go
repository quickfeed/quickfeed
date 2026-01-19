package dummydata

import (
	"fmt"
	"log"
	"os"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qtest"
)

// TODO(Joachim): make it configurable via a file
// TODO(Joachim): add verbose logging option
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

type generator struct {
	db database.Database
}

var (
	courses = len(qtest.MockCourses)
)

const (
	teachers                        = 2
	students                        = 10
	enrolledStudents                = 6
	assingnmentsPerCourse           = 8
	groupAssignments                = 2
	studentSubmissionsPerAssignment = 8
	groupSubmissionsPerAssignment   = 3
)

// NewGenerator creates a new generator instance.
func NewGenerator() (*generator, error) {
	if enrolledStudents > students {
		return nil, fmt.Errorf("number of enrolled students (%d) cannot be larger than total number of students (%d)", enrolledStudents, students)
	}

	if err := env.Load(env.RootEnv(".env")); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}
	dbFile := env.DatabasePath()
	if _, err := os.Stat(dbFile); err == nil {
		if err := os.Remove(dbFile); err != nil {
			return nil, fmt.Errorf("failed to remove existing database file: %v", err)
		}
	}
	db, err := database.NewGormDB(dbFile, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	return &generator{db}, nil
}

func (g *generator) Data(adminName string) error {
	fncs := []func() error{
		func() error { return g.admin(adminName) },
		g.courses,
		g.users,
		g.groups,
		g.submissions,
	}
	for _, fnc := range fncs {
		if err := fnc(); err != nil {
			return err
		}
	}
	log.Println("Dummy data generation complete")
	return nil
}
