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

	Enrollments []*Enrollment `json:"-"`
	Enrolled    int           `json:"enrolled" sql:"-"`

	Assignments []*Assignment `json:"assignments,omitempty"`
}
