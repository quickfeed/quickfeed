package models

// Enrollment represents the status of a users enrollment into a course.
type Enrollment struct {
	ID uint64

	Course   *Course
	CourseID uint64

	User   *User
	UserID uint64

	Status int
}
