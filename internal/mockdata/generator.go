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
	"github.com/quickfeed/quickfeed/qf"
)

// TODO(Joachim): attempt to make generation fault tolerant ?

func (g *generator) loadConfig() error {
	const (
		jsonExt = ".json"
		envVar  = "QUICKFEED_MOCK_CONFIG" // TODO(Joachim): consider standardizing the mock config file, e.g. "Root"/mock_config.json
	)
	path := os.Getenv(envVar)
	config := fmt.Sprintf("%s%s", path, jsonExt)
	if config != jsonExt {
		log.Println("Loading config")
		if bytes, err := os.ReadFile(config); err == nil {
			return json.Unmarshal(bytes, &g.config)
		} else if errors.Is(err, os.ErrNotExist) {
			log.Printf("Config file: %q does not exist", config)
		} else {
			return fmt.Errorf("failed to read config file: %v", err)
		}
	}
	log.Printf("Using default configuration")
	if path == "" {
		log.Printf("  - Set %q to use custom config", envVar)
	}
	return nil
}

// NewGenerator creates a new generator instance.
func NewGenerator() (*generator, error) {
	log.Println("Initializing the generator")
	gen := &generator{
		config: defaultConfig,
	}
	err := env.Load(env.RootEnv(".env"))
	if err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}
	if err := gen.loadConfig(); err != nil {
		return nil, err
	}
	if err := gen.config.validate(); err != nil {
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
func (c *config) validate() error {
	log.Println("Validating config")
	switch {
	case c.Students < c.EnrolledStudents:
		return fmt.Errorf("number of students (%d) can't be less number of enrolled students (%d)", c.Students, c.EnrolledStudents)
	case c.AssingnmentsPerCourse < c.GroupAssignments:
		return fmt.Errorf("number of assingnmentsPerCourse (%d) can't be less than groupAssignments (%d)", c.AssingnmentsPerCourse, c.GroupAssignments)
	}
	return nil
}

func (g *generator) Mock(adminName string) error {
	start := time.Now()
	// TODO(Joachim): Consider running in goroutines for faster generation
	// need to figure out dependencies or table relations first
	log.Printf("Creating admin user: %s\n", adminName)
	if err := g.admin(adminName); err != nil {
		return fmt.Errorf("failed to create admin user: %v", err)
	}
	log.Printf("Generating...")
	for i, fn := range []func() error{
		g.courses,
		g.users,
		g.groups,
		g.submissions,
	} {
		log.Printf("  - %s", []string{"Courses", "Users", "Groups", "Submissions"}[i])
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
	if g.Verbose {
		log.Println(v...)
	}
}
