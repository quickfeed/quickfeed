package database

import "github.com/labstack/gommon/log"

// Course represents a course.
type Course struct {
	ID           int
	Name         string
	Organization string
}

// CreateCourse creates a new course with the given name.
func (db *StructDB) CreateCourse(name, organization string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	id := len(db.Courses)
	db.Courses[id] = &Course{
		ID:           id,
		Name:         name,
		Organization: organization,
	}
	if err := db.save(); err != nil {
		delete(db.Courses, id)
		db.logger.Infoj(log.JSON{
			"name":         name,
			"organization": organization,
			"message":      "could not persist course to database",
			"err":          err.Error(),
		})
		return err
	}

	return nil
}

// GetCourses returns all the courses in the database.
func (db *StructDB) GetCourses() (map[int]*Course, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	courses := make(map[int]*Course, len(db.Courses))
	for id, course := range db.Courses {
		courses[id] = course
	}

	return courses, nil
}
