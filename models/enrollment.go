package models

// Enrollment status.
const (
	Pending uint = iota
	Rejected
	Student
	Teacher
	None = -1
)

// Enrollment represents the status of a users enrollment into a course.
type Enrollment struct {
	ID uint64 `json:"id"`

	Course   *Course `json:"course,omitempty"`
	CourseID uint64  `json:"courseid"`

	User   *User  `json:"user,omitempty"`
	UserID uint64 `json:"userid"`

	Group   *Group `json:"group,omitempty"`
	GroupID uint64 `json:"groupid"`

	Status uint `json:"status"`
}
