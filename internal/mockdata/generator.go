package mockdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
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
	db   database.Database
	conf config
}

const (
	containerTimeout = 30
	configName       = "mock.json"
)

type config struct {
	teachers                        int
	students                        int
	enrolledStudents                int
	assingnmentsPerCourse           int
	groupAssignments                int
	studentSubmissionsPerAssignment int
	groupSubmissionsPerAssignment   int
	verbose                         bool
}

var (
	courses = len(qtest.MockCourses)
)

const (
	teachers                        = 2  // teachers + students = enrolledStudents
	students                        = 10 // > enrolledStudents
	enrolledStudents                = 8  // < students
	assingnmentsPerCourse           = 8  // > groupAssignments
	groupAssignments                = 2  // < assingnmentsPerCourse
	studentSubmissionsPerAssignment = 8
	groupSubmissionsPerAssignment   = 3
)

// validate ensures the provides values are within the required range
// sum of related variable can't be less then zero
func validate() error {
	switch {
	case students < enrolledStudents:
		return fmt.Errorf("number of students (%d) can't be less number of enrolled students (%d)", students, enrolledStudents)
	case assingnmentsPerCourse < groupAssignments:
		return fmt.Errorf("number of assingnmentsPerCourse (%d) can't be less than groupAssignments (%d)", assingnmentsPerCourse, groupAssignments)
	}
	return nil
}

func getConfig() (config, error) {
	var conf config
	bytes, err := os.ReadFile(env.Root(configName))
	if err != nil && errors.Is(err, os.ErrNotExist) {
		// Return default config if no config file is found
		return config{
			teachers,
			students,
			enrolledStudents,
			assingnmentsPerCourse,
			groupAssignments,
			studentSubmissionsPerAssignment,
			groupSubmissionsPerAssignment,
			false,
		}, nil
	}
	return conf, json.Unmarshal(bytes, &conf)
}

// NewGenerator creates a new generator instance.
func NewGenerator() (*generator, error) {
	log.Println("Initializing the generator")
	if err := validate(); err != nil {
		return nil, err
	}
	if err := env.Load(env.RootEnv(".env")); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}
	log.Println("Reading config")
	conf, err := getConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}
	log.Println("Setting up database")
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
	return &generator{db, conf}, nil
}

func (g *generator) Mock(adminName string) error {
	start := time.Now()
	// TODO(Joachim): Consider running in goroutines for faster generation
	// need to figure out dependencies first
	log.Printf("Creating admin user: %s\n", adminName)
	if err := g.admin(adminName); err != nil {
		return fmt.Errorf("failed to create admin user: %v", err)
	}
	fns := []func() error{
		g.courses,
		g.users,
		g.groups,
		g.submissions,
	}
	names := []string{"Courses", "Users", "Groups", "Submissions"}
	log.Printf("Generating...")
	for i, fn := range fns {
		log.Printf("  - %s", names[i])
		if err := fn(); err != nil {
			return err
		}
	}
	log.Printf("Mock database generation complete (~%vs)", math.Round(time.Since(start).Seconds()))
	return nil
}
func (g *generator) admin(username string) error {
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/users/%s", username))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("failed to fetch user info, likely wrong username")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := resp.Body.Close(); err != nil {
		fmt.Println("failed to close response body:", err)
	}
	var info *struct {
		Login     string `json:"login"`
		ID        uint64 `json:"id"`
		AvatarURL string `json:"avatar_url"`
		Name      string `json:"name"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return err
	}
	return g.db.CreateUser(&qf.User{
		ID:          1,
		Login:       info.Login,
		Name:        info.Name,
		ScmRemoteID: info.ID,
		AvatarURL:   info.AvatarURL,
		Email:       fmt.Sprintf("%s@gmail.com", username),
		StudentID:   "999999",
	})
}

// log prints a log message if verbose mode is enabled.
// should be used for cumbersome log messages only.
// msg cannot be an empty string.
func (g *generator) log(format string, v ...any) {
	if format == "" {
		panic("log message cannot be empty")
	}
	if g.conf.verbose {
		log.Println(v...)
	}
}
