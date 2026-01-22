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

// TODO(Joachim): attempt to make generation fault tolerant ?
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
	Teachers                        int
	Students                        int
	EnrolledStudents                int
	AssingnmentsPerCourse           int
	GroupAssignments                int
	StudentSubmissionsPerAssignment int
	GroupSubmissionsPerAssignment   int
	Verbose                         bool
}

var (
	courses = len(qtest.MockCourses)
)

func getConfig() (config, error) {
	var conf config
	f := env.Root(configName)
	bytes, err := os.ReadFile(f)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return conf, err
		}
		// Return default config if no config file is found
		log.Printf("No config file found (%s), using default config\n", f)
		return config{
			Teachers:                        2,  // teachers + students = enrolledStudents
			Students:                        10, // > enrolledStudents
			EnrolledStudents:                8,  // < students
			AssingnmentsPerCourse:           8,  // > groupAssignments
			GroupAssignments:                2,  // < assingnmentsPerCourse
			StudentSubmissionsPerAssignment: 8,
			GroupSubmissionsPerAssignment:   3,
			Verbose:                         false,
		}, nil
	}
	return conf, json.Unmarshal(bytes, &conf)
}

// NewGenerator creates a new generator instance.
func NewGenerator() (*generator, error) {
	log.Println("Initializing the generator")
	gen := &generator{}
	err := env.Load(env.RootEnv(".env"))
	if err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}
	log.Println("Reading config")
	gen.conf, err = getConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}
	if err := gen.validate(); err != nil {
		return nil, err
	}
	log.Println("Setting up database")
	dbFile := env.DatabasePath()
	if _, err := os.Stat(dbFile); err == nil {
		if err := os.Remove(dbFile); err != nil {
			return nil, fmt.Errorf("failed to remove existing database file: %v", err)
		}
	}
	gen.db, err = database.NewGormDB(dbFile, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	return gen, nil
}

// validate ensures the provides values are within the required range
// sum of related variable can't be less then zero
func (g *generator) validate() error {
	log.Println("Validating config")
	switch {
	case g.Students() < g.EnrolledStudents():
		return fmt.Errorf("number of students (%d) can't be less number of enrolled students (%d)", g.Students(), g.EnrolledStudents())
	case g.AssingnmentsPerCourse() < g.GroupAssignments():
		return fmt.Errorf("number of assingnmentsPerCourse (%d) can't be less than groupAssignments (%d)", g.AssingnmentsPerCourse(), g.GroupAssignments())
	}
	return nil
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
	if g.Verbose() {
		log.Println(v...)
	}
}

// IsGroupLab returns the assignment number where group assignments start
func (g *generator) IsGroupLab() int {
	return g.AssingnmentsPerCourse() - g.GroupAssignments()
}

// Config field getters
func (g *generator) Teachers() int              { return g.conf.Teachers }
func (g *generator) Students() int              { return g.conf.Students }
func (g *generator) EnrolledStudents() int      { return g.conf.EnrolledStudents }
func (g *generator) AssingnmentsPerCourse() int { return g.conf.AssingnmentsPerCourse }
func (g *generator) GroupAssignments() int      { return g.conf.GroupAssignments }
func (g *generator) StudentSubmissionsPerAssignment() int {
	return g.conf.StudentSubmissionsPerAssignment
}
func (g *generator) GroupSubmissionsPerAssignment() int { return g.conf.GroupSubmissionsPerAssignment }
func (g *generator) Verbose() bool                      { return g.conf.Verbose }
