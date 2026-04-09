package models

import (
	"time"

	"github.com/uptrace/bun"
)

// Models for the revised database schema

// Enums for revised schema
type RevisedGroupStatus int

const (
	RevisedGroupStatusPending  RevisedGroupStatus = 0
	RevisedGroupStatusApproved RevisedGroupStatus = 1
)

type RevisedEnrollmentStatus int

const (
	RevisedEnrollmentStatusNone    RevisedEnrollmentStatus = 0
	RevisedEnrollmentStatusPending RevisedEnrollmentStatus = 1
	RevisedEnrollmentStatusStudent RevisedEnrollmentStatus = 2
	RevisedEnrollmentStatusTeacher RevisedEnrollmentStatus = 3
)

type RevisedEnrollmentState int

const (
	RevisedEnrollmentStateUnset     RevisedEnrollmentState = 0
	RevisedEnrollmentStateHidden    RevisedEnrollmentState = 1
	RevisedEnrollmentStateVisible   RevisedEnrollmentState = 2
	RevisedEnrollmentStateFavourite RevisedEnrollmentState = 3
)

type RevisedRepositoryType int

const (
	RevisedRepoTypeNone        RevisedRepositoryType = 0
	RevisedRepoTypeInfo        RevisedRepositoryType = 1
	RevisedRepoTypeAssignments RevisedRepositoryType = 2
	RevisedRepoTypeTests       RevisedRepositoryType = 3
	RevisedRepoTypeUser        RevisedRepositoryType = 4
	RevisedRepoTypeGroup       RevisedRepositoryType = 5
)

type RevisedGrade int

const (
	RevisedGradeNone   RevisedGrade = 0
	RevisedGradePass   RevisedGrade = 1
	RevisedGradeFailed RevisedGrade = 2
)

type RevisedDecision int

const (
	RevisedDecisionNone     RevisedDecision = 0
	RevisedDecisionApproved RevisedDecision = 1
	RevisedDecisionRejected RevisedDecision = 2
	RevisedDecisionRevision RevisedDecision = 3
)

type RevisedUser struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID           uint64 `bun:"id,pk,autoincrement"`
	IsAdmin      bool   `bun:"is_admin"`
	Name         string `bun:"name"`
	StudentID    string `bun:"student_id"`
	Email        string `bun:"email"`
	AvatarURL    string `bun:"avatar_url"`
	Login        string `bun:"login"`
	ScmRemoteID  uint64 `bun:"scm_remote_id"`
	UpdateToken  bool   `bun:"update_token"`
	RefreshToken string `bun:"refresh_token"`
}

type RevisedCourse struct {
	bun.BaseModel `bun:"table:courses,alias:c"`

	ID                  uint64 `bun:"id,pk,autoincrement"`
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

type RevisedGroup struct {
	bun.BaseModel `bun:"table:groups,alias:g"`

	ID       uint64 `bun:"id,pk,autoincrement"`
	Name     string `bun:"name"`
	CourseID uint64 `bun:"course_id"` // fk

	Status RevisedGroupStatus `bun:"status,default:0"` // enum
}

type RevisedGroupUser struct {
	bun.BaseModel `bun:"table:group_users,alias:gu"`

	GroupID uint64 `bun:"group_id,pk"` // fk
	UserID  uint64 `bun:"user_id,pk"`  // fk
}

type RevisedEnrollment struct {
	bun.BaseModel `bun:"table:enrollments,alias:e"`

	ID               uint64    `bun:"id,pk,autoincrement"`
	UserID           uint64    `bun:"user_id"`   // fk
	CourseID         uint64    `bun:"course_id"` // fk
	GroupID          uint64    `bun:"group_id"`  // fk
	LastActivityDate time.Time `bun:"last_activity_date"`
	TotalApproved    uint64    `bun:"total_approved"`

	Status RevisedEnrollmentStatus `bun:"status,default:0"` // enum
	State  RevisedEnrollmentState  `bun:"state,default:0"`  // enum
}

type RevisedRepository struct {
	bun.BaseModel `bun:"table:repositories,alias:r"`

	ID              uint64 `bun:"id,pk,autoincrement"`
	ScmRepositoryID uint64 `bun:"scm_repository_id"`
	HTMLURL         string `bun:"html_url"`
	EnrollmentsID   uint64 `bun:"enrollments_id"` // fk
	GroupID         uint64 `bun:"group_id"`       // fk

	RepoType RevisedRepositoryType `bun:"repo_type,default:0"` // enum
}

type RevisedAssignment struct {
	bun.BaseModel `bun:"table:assignments,alias:a"`

	ID               uint64    `bun:"id,pk,autoincrement"`
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

type RevisedAssignmentFeedback struct {
	bun.BaseModel `bun:"table:assignment_feedback,alias:af"`

	ID                     uint64    `bun:"id,pk,autoincrement"`
	LikedContent           string    `bun:"liked_content"`
	ImprovementSuggestions string    `bun:"improvement_suggestions"`
	TimeSpent              uint32    `bun:"time_spent"`
	CreatedAt              time.Time `bun:"created_at"`
	AssignmentID           uint64    `bun:"assignment_id"` // fk
	CourseID               uint64    `bun:"course_id"`     // fk
}

type RevisedTestInfo struct {
	bun.BaseModel `bun:"table:test_info,alias:ti"`

	ID           uint64 `bun:"id,pk,autoincrement"`
	TestName     string `bun:"test_name"`
	MaxScore     int32  `bun:"max_score"`
	Weight       int32  `bun:"weight"`
	Details      string `bun:"details"`
	AssignmentID uint64 `bun:"assignment_id"` // fk
}

type RevisedSubmission struct {
	bun.BaseModel `bun:"table:submissions,alias:s"`

	ID           uint64    `bun:"id,pk,autoincrement"`
	Score        uint32    `bun:"score"`
	CommitHash   string    `bun:"commit_hash"`
	Released     bool      `bun:"released"`
	ApprovedDate time.Time `bun:"approved_date"`
	AssignmentID uint64    `bun:"assignment_id"` // fk
	GroupID      uint64    `bun:"group_id"`      // fk
	UserID       uint64    `bun:"user_id"`       // fk
}

type RevisedUsedSlipDays struct {
	bun.BaseModel `bun:"table:used_slip_days,alias:usd"`

	AssignmentID uint64 `bun:"assignment_id,pk"` // fk
	EnrollmentID uint64 `bun:"enrollment_id,pk"` // fk
	UsedDays     uint32 `bun:"used_days"`
}

type RevisedFeedbackReceipt struct {
	bun.BaseModel `bun:"table:feedback_receipt,alias:fr"`

	AssignmentID uint64 `bun:"assignment_id,pk"` // fk
	UserID       uint64 `bun:"user_id,pk"`       // fk
}

type RevisedBuildInfo struct {
	bun.BaseModel `bun:"table:build_info,alias:bi"`

	BuildLog       string    `bun:"build_log"`
	ExecTime       int64     `bun:"exec_time"`
	BuildDate      time.Time `bun:"build_date"`
	SubmissionDate time.Time `bun:"submission_date"`
	SubmissionID   uint64    `bun:"submission_id,pk"` // fk
}

type RevisedReview struct {
	bun.BaseModel `bun:"table:reviews,alias:r"`

	ID           uint64    `bun:"id,pk,autoincrement"`
	Feedback     string    `bun:"feedback"`
	Ready        bool      `bun:"ready"`
	Score        uint32    `bun:"score"`
	Edited       time.Time `bun:"edited"`
	SubmissionID uint64    `bun:"submission_id"` // fk
}

type RevisedChecklist struct {
	bun.BaseModel `bun:"table:checklist,alias:cl"`

	ID           uint64 `bun:"id,pk,autoincrement"`
	Heading      string `bun:"heading"`
	Comment      string `bun:"comment"`
	ReviewID     uint64 `bun:"review_id"`     // fk
	AssignmentID uint64 `bun:"assignment_id"` // fk
}

type RevisedChecklistItem struct {
	bun.BaseModel `bun:"table:checklist_item,alias:cli"`

	ID          uint64 `bun:"id,pk,autoincrement"`
	Points      uint64 `bun:"points"`
	Description string `bun:"description"`
	Comment     string `bun:"comment"`
	ChecklistID uint64 `bun:"checklist_id"` // fk

	Grade RevisedGrade `bun:"grade,default:0"` // enum
}

type RevisedApproval struct {
	bun.BaseModel `bun:"table:approval,alias:ap"`

	SubmissionID uint64 `bun:"submission_id,pk"` // fk
	EnrollmentID uint64 `bun:"enrollment_id,pk"` // fk

	Decision RevisedDecision `bun:"decision,default:0"` // enum
}

// struct to hold the result of the GetCourseCreatorName query
type RevisedGetCourseCreatorNameResult struct {
	GroupID         uint64 `bun:"group_id"`
	CourseID        uint64 `bun:"course_id"`
	CourseCreatorID uint64 `bun:"course_creator_id"`
	Name            string `bun:"name"`
	ID              uint64 `bun:"id"`
}

// struct to hold the result of the GetCourseByStatus query
type RevisedGetCourseByStatusResult struct {
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
