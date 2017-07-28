package models

import "time"

// Assignment represents a single assignment
type Assignment struct {
	ID          uint64      `json:"id"`
	CourseID    uint64      `json:"courseid"`
	Name        string      `json:"name"`
	Language    string      `json:"language"`
	Deadline    time.Time   `json:"deadline"`
	AutoApprove bool        `json:"autoapprove" sql:"DEFAULT:false"`
	Order       uint        `json:"order"`
	Submission  *Submission `json:"submission,omitempty"`
}

// Submission represents a single submission
type Submission struct {
	ID           uint64 `json:"id"`
	AssignmentID uint64 `json:"assignmentid"`
	UserID       uint64 `json:"userid"`
	GroupID      uint64 `json:"groupid"`
	Score        uint8  `json:"score"`
	ScoreObjects string `json:"scoreobjects" sql:"type:text"`
	BuildInfo    string `json:"buildinfo" sql:"type:text"`
}
