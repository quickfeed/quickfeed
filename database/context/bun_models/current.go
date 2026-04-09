package models

import (
	"time"

	"github.com/uptrace/bun"
)

// enums:
type GroupStatus int

const (
	GroupStatusPending  GroupStatus = 0
	GroupStatusApproved GroupStatus = 1
)

type EnrollmentStatus int

const (
	EnrollmentStatusNone    EnrollmentState = 0
	EnrollmentStatusPending EnrollmentState = 1
	EnrollmentStatusStudent EnrollmentState = 2
	EnrollmentStatusTeacher EnrollmentState = 3
)

type EnrollmentState int

const (
	EnrollmentStateUnset     EnrollmentState = 0
	EnrollmentStateHidden    EnrollmentState = 1
	EnrollmentStateVisivle   EnrollmentState = 2
	EnrollmentStateFavoutite EnrollmentState = 3
)

type RepositoryType int

const (
	RepoTypeNone        RepositoryType = 0
	RepoTypeInfo        RepositoryType = 1
	RepoTypeAssignments RepositoryType = 2
	RepoTypeTests       RepositoryType = 3
	RepoTypeUser        RepositoryType = 4
	RepoTypeGroup       RepositoryType = 5
)

type SubmissionStatus int

const (
	SubmissionStatusNone     SubmissionStatus = 0
	SubmissionStatusApproved SubmissionStatus = 1
	SubmissionStatusRejected SubmissionStatus = 2
	SubmissionStatusRevision SubmissionStatus = 3
)

// string = notnull GO type
// *String = nullable, can be = nil
type CurrentUser struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID           uint64 `bun:"id,pk"`
	IsAdmin      bool   `bun:"is_admin,default:0"`
	Name         string `bun:"name"`
	StudentID    string `bun:"student_id"`
	Email        string `bun:"email"`
	AvatarURL    string `bun:"avatar_url"`
	Login        string `bun:"login"`
	ScmRemoteID  uint64 `bun:"scm_remote_id"`
	UpdateToken  bool   `bun:"update_token"`
	RefreshToken string `bun:"refresh_token"`
}

type CurrentCourse struct {
	bun.BaseModel `bun:"table:courses,alias:c"`

	ID                  uint64 `bun:"id,pk"`
	Name                string `bun:"name"`
	Code                string `bun:"code"`
	Year                uint32 `bun:"year"`
	Tag                 string `bun:"tag"`
	ScmOrganizationID   uint64 `bun:"scm_organization_id"`
	ScmOrganizationName string `bun:"scm_organization_name"`
	SlipDays            uint32 `bun:"slip_days"`
	DockerfileDigest    string `bun:"dockerfile_digest"`
	CourseCreatorID     uint64 `bun:"course_creator_id"` // fk
}

type CurrentAssignment struct {
	bun.BaseModel `bun:"table:assignments,alias:a"`

	ID               uint64    `bun:"id,pk"`
	Name             string    `bun:"name"`
	Deadline         time.Time `bun:"deadline"`
	AutoApprove      bool      `bun:"auto_approve"`
	Order            uint32    `bun:"order"`
	IsGroupLab       bool      `bun:"is_group_lab"`
	ScoreLimit       uint32    `bun:"score_limit"`
	Reviewers        uint32    `bun:"reviewers"`
	ContainerTimeout uint32    `bun:"container_timeout"`
	CourseID         uint64    `bun:"course_id"` // fk
}

type CurrentGroup struct {
	bun.BaseModel `bun:"table:groups,alias:g"`

	ID       uint64 `bun:"id,pk"`
	Name     string `bun:"name"`
	CourseID uint64 `bun:"course_id"` // fk

	Status GroupStatus `bun:"status,default:0"` // enums
}

type CurrentGroupUser struct {
	bun.BaseModel `bun:"table:group_users,alias:gu"`

	GroupID uint64 `bun:"group_id,pk"` // fk
	UserID  uint64 `bun:"user_id,pk"`  // fk
}

type CurrentEnrollment struct {
	bun.BaseModel `bun:"table:enrollments,alias:e"`

	ID               uint64    `bun:"id,pk"`
	LastActivityDate time.Time `bun:"last_activity_date"`
	TotalApproved    uint64    `bun:"total_approved"`
	CourseID         uint64    `bun:"course_id"` // fk
	UserID           uint64    `bun:"user_id"`   // fk
	GroupID          uint64    `bun:"group_id"`  // fk

	Status EnrollmentStatus `bun:"status,default:0"` // enum
	State  EnrollmentState  `bun:"state,default:0"`  // enum
}

type CurrentRepository struct {
	bun.BaseModel `bun:"table:repositories,alias:r"`

	ID                uint64 `bun:"id,pk"`
	ScmOrganizationID uint64 `bun:"scm_organization_id"`
	ScmRepositoryID   uint64 `bun:"scm_repository_id"`
	HTMLURL           string `bun:"html_url"`
	UserID            uint64 `bun:"user_id"`  // fk
	GroupID           uint64 `bun:"group_id"` // fk

	RepositoryType RepositoryType `bun:"repo_type,default:0"` // enum
}

type CurrentTestInfo struct {
	bun.BaseModel `bun:"table:test_info,alias:ti"`

	ID           uint64 `bun:"id,pk"`
	TestName     string `bun:"test_name"`
	MaxScore     int32  `bun:"max_score"`
	Weight       int32  `bun:"weight"`
	Details      string `bun:"details"`
	AssignmentID uint64 `bun:"assignment_id"` // fk
}

type CurrentUsedSlipDays struct {
	bun.BaseModel `bun:"table:used_slip_days,alias:usd"`

	ID           uint64 `bun:"id,pk"`
	UsedDays     uint32 `bun:"used_days"`
	EnrollmentID uint64 `bun:"enrollment_id"` // fk
	AssignmentID uint64 `bun:"assignment_id"` // fk
}

type CurrentFeedbackReceipt struct {
	bun.BaseModel `bun:"table:feedback_receipt,alias:fr"`

	AssignmentID uint64 `bun:"assignment_id,pk"` // fk
	UserID       uint64 `bun:"user_id,pk"`       // fk
}

type CurrentSubmission struct {
	bun.BaseModel `bun:"table:submissions,alias:s"`

	ID           uint64    `bun:"id,pk"`
	Score        uint32    `bun:"score"`
	CommitHash   string    `bun:"commit_hash"`
	Released     bool      `bun:"released"`
	ApprovedDate time.Time `bun:"approved_date"`
	AssignmentID uint64    `bun:"assignment_id"` // fk
	GroupID      uint64    `bun:"group_id"`      // fk
	UserID       uint64    `bun:"user_id"`       // fk

	Status SubmissionStatus `bun:"status,default:0"` // enum
}

// struct to hold the result of the GetCourseCreatorName query
type CurrentGetCourseCreatorNameResult struct {
	GroupID         uint64 `bun:"group_id"`
	CourseID        uint64 `bun:"course_id"`
	CourseCreatorID uint64 `bun:"course_creator_id"`
	Name            string `bun:"name"`
	ID              uint64 `bun:"id"`
}

// struct to hold the result of the GetCourseByStatus query
type CurrentGetCourseByStatusResult struct {
	CourseID             uint64 `bun:"course_id"`
	CourseName           string `bun:"course_name"`
	Code                 string `bun:"code"`
	Year                 uint32 `bun:"year"`
	Tag                  string `bun:"tag"`
	ScmOrganizationID    uint64 `bun:"scm_organization_id"`
	ScmOrganizationName  string `bun:"scm_organization_name"`
	SlipDays             uint32 `bun:"slip_days"`
	DockerfileDigest     string `bun:"dockerfile_digest"`
	CourseCreatorID      uint64 `bun:"course_creator_id"`
	UserID               uint64 `bun:"user_id"`
	IsAdmin              bool   `bun:"is_admin"`
	UserName             string `bun:"user_name"`
	StudentID            string `bun:"student_id"`
	Email                string `bun:"email"`
	AvatarURL            string `bun:"avatar_url"`
	Login                string `bun:"login"`
	ScmRemoteID          uint64 `bun:"scm_remote_id"`
	UpdateToken          bool   `bun:"update_token"`
	RefreshToken         string `bun:"refresh_token"`
	EnrollmentCount      int    `bun:"enrollment_count"`
	SlipDaysCount        int    `bun:"slip_days_count"`
	FeedbackReceiptCount int    `bun:"feedback_receipt_count"`
}
