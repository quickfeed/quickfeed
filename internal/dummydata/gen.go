package dummydata

import (
	"fmt"
	"os"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/env"
)

type generator struct {
	db database.Database
}

// NewGenerator creates a new generator instance.
func NewGenerator() (*generator, error) {
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
	return &generator{
		db: db,
	}, nil
}

func (g *generator) Data(adminName string) error {
	fncs := []func() error{
		func() error { return g.admin(adminName) },
		g.courses,
		g.users,
	}
	for _, fnc := range fncs {
		if err := fnc(); err != nil {
			return err
		}
	}
	return nil
}
