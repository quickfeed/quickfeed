package database

import (
	"encoding/gob"
	"os"
	"sync"

	"github.com/labstack/echo"
)

// StructDB implements Database.
type StructDB struct {
	mu      sync.Mutex
	path    string
	Users   map[int]*User
	Courses map[int]*Course

	logger echo.Logger
}

// NewStructDB creates a new database which saves the whole database to a file
// on every change. If no path is set, the database will operate in memory only.
func NewStructDB(path string, truncate bool, logger echo.Logger) (*StructDB, error) {
	if path == "" {
		return &StructDB{
			// Leave path unset to indicate in memory DB.
			Users:   make(map[int]*User),
			Courses: make(map[int]*Course),
			logger:  logger,
		}, nil
	}

	newDB := truncate || !fileExists(path)

	if !newDB {
		f, err := os.OpenFile(path, os.O_RDWR, 0666)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		var db StructDB
		dec := gob.NewDecoder(f)
		if err := dec.Decode(&db); err != nil {
			return nil, err
		}
		db.path = path
		db.logger = logger

		return &db, nil
	}

	db := &StructDB{
		path:    path,
		Users:   make(map[int]*User),
		Courses: make(map[int]*Course),
		logger:  logger,
	}

	return db, db.save()
}

// Caller must hold lock on db.
func (db *StructDB) save() error {
	// Don't write to disk if in memory DB.
	if db.path == "" {
		return nil
	}

	oldPath := db.path + "_old"

	// Move existing database and continue on error if file did not exist.
	if err := os.Rename(db.path, oldPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	f, err := os.Create(db.path)
	defer f.Close()
	if err != nil {
		return err
	}

	enc := gob.NewEncoder(f)
	if err := enc.Encode(db); err != nil {
		return err
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
