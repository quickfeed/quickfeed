package models

// Course represents a course backed by a directory.
type Course struct {
	ID uint64 `json:"id"`

	Name string `json:"name"`
	Code string `json:"code"`
	Year uint   `json:"year"`
	Tag  string `json:"tag"`

	Provider    string `json:"provider"`
	DirectoryID uint64 `json:"directoryid"`

	Users []User `gorm:"many2many:user_courses;" json:"users,omitempty"`

	Assignments []Assignment `json:"assignments,omitempty"`
}

// Assignment represents a single assignment
type Assignment struct {
	ID       uint64 `json:"id"`
	CourseID uint64 `json:"courseid"`
	Name     string `json:"name"`
}
