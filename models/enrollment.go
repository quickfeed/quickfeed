package models

// Enrollment status.
const (
	Pending uint = iota
	Rejected
	Accepted
	None = -1
)

// Enrollment represents the status of a users enrollment into a course.
type Enrollment struct {
	ID uint64 `json:"id"`

	Course   *Course `json:"course,omitempty"`
	CourseID uint64  `json:"courseid"`

	User   *User  `json:"user,omitempty"`
	UserID uint64 `json:"userid"`

	Status int `json:"status"`
}
