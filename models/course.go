package models

// Course represents a course backed by a directory.
type Course struct {
	ID uint64

	Name     string
	Year     uint
	Semester string

	Provider    string
	DirectoryID uint64
}
