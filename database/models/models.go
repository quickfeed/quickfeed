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

type CriterionGrade int

const (
	CriterionGradeIncline CriterionGrade = 0
	CriterionGradeFail    CriterionGrade = 1
	CriterionGradePass    CriterionGrade = 2
)

type PullRequestStage int

const (
	PullRequestStageNone     PullRequestStage = 0
	PullRequestStageReview   PullRequestStage = 1
	PullRequestStageApproved PullRequestStage = 2
	PullRequestStageMerged   PullRequestStage = 3
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID           uint64 `bun:"id,pk"`
	IsAdmin      bool   `bun:"is_admin"`
	Name         string `bun:"name"`
	StudentID    string `bun:"student_id"`
	Email        string `bun:"email"`
	AvatarURL    string `bun:"avatar_url"`
	Login        string `bun:"login"`
	UpdateToken  bool   `bun:"update_token"`
	ScmRemoteID  uint64 `bun:"scm_remote_id"`
	RefreshToken string `bun:"refresh_token"`

	Enrollments []*Enrollment `bun:"rel:has-many,join:id=user_id"`
}

type Course struct {
	bun.BaseModel `bun:"table:courses,alias:c"`

	ID                  uint64 `bun:"id,pk"`
	CourseCreatorID     uint64 `bun:"course_creator_id"`
	Name                string `bun:"name"`
	Code                string `bun:"code"`
	Year                uint32 `bun:"year"`
	Tag                 string `bun:"tag"`
	ScmOrganizationID   uint64 `bun:"scm_organization_id"`
	ScmOrganizationName string `bun:"scm_organization_name"`
	SlipDays            uint32 `bun:"slip_days"`
	DockerfileDigest    string `bun:"dockerfile_digest"`

	Assignments []*Assignment `bun:"rel:has-many,join:id=course_id"`
	Enrollments []*Enrollment `bun:"rel:has-many,join:id=course_id"`
}

type Assignment struct {
	bun.BaseModel `bun:"table:assignments,alias:a"`

	ID               uint64    `bun:"id,pk"`
	CourseID         uint64    `bun:"course_id"`
	Name             string    `bun:"name"`
	Deadline         time.Time `bun:"deadline"`
	AutoApprove      bool      `bun:"auto_approve"`
	Order            uint32    `bun:"order"`
	IsGroupLab       bool      `bun:"is_group_lab"`
	ScoreLimit       uint32    `bun:"score_limit"`
	Reviewers        uint32    `bun:"reviewers"`
	ContainerTimeout uint32    `bun:"container_timeout"`

	Course       *Course        `bun:"rel:belongs-to,join:course_id=id"`
	TestInfos    []*TestInfo    `bun:"rel:has-many,join:id=assignment_id"`
	Tasks        []*Task        `bun:"rel:has-many,join:id=assignment_id"`
	UsedSlipDays []*UsedSlipDay `bun:"rel:has-many,join:id=assignment_id"`
}

type Group struct {
	bun.BaseModel `bun:"table:groups,alias:g"`

	ID       uint64      `bun:"id,pk"`
	Name     string      `bun:"name"`
	CourseID uint64      `bun:"course_id"`
	Status   GroupStatus `bun:"status"`

	Course *Course `bun:"rel:belongs-to,join:course_id=id"`
	// group and use has m2m, through group_user table:
	Users       []*User       `bun:"m2m:group_users,join:Group=User"`
	Enrollments []*Enrollment `bun:"rel:has-many,join:id=group_id"`
}

type GroupUser struct {
	bun.BaseModel `bun:"table:group_users,alias:gu"`

	GroupID uint64 `bun:"group_id,pk"`
	UserID  uint64 `bun:"user_id,pk"`
	Group   *Group `bun:"rel:belongs-to,join:group_id=id"`
	User    *User  `bun:"rel:belongs-to,join:user_id=id"`
}

type Repository struct {
	bun.BaseModel `bun:"table:repositories,alias:r"`

	ID                uint64         `bun:"id,pk"`
	ScmOrganizationID uint64         `bun:"scm_organization_id"`
	ScmRepositoryID   uint64         `bun:"scm_repository_id"`
	UserID            uint64         `bun:"user_id"`
	GroupID           uint64         `bun:"group_id"`
	HTMLURL           string         `bun:"html_url"`
	RepositoryType    RepositoryType `bun:"repo_type"`

	User   *User    `bun:"rel:belongs-to,join:user_id=id"`
	Group  *Group   `bun:"rel:belongs-to,join:group_id=id"`
	Issues []*Issue `bun:"rel:has-many,join:id=repository_id"`
}

type Enrollment struct {
	bun.BaseModel `bun:"table:enrollments,alias:e"`

	ID               uint64           `bun:"id,pk"`
	CourseID         uint64           `bun:"course_id"`
	UserID           uint64           `bun:"user_id"`
	GroupID          uint64           `bun:"group_id"`
	Status           EnrollmentStatus `bun:"status"`
	State            EnrollmentState  `bun:"state"`
	LastActivityDate time.Time        `bun:"last_activity_date"`
	TotalApproved    uint64           `bun:"total_approved"`

	Course       *Course        `bun:"rel:belongs-to,join:course_id=id"`
	User         *User          `bun:"rel:belongs-to,join:user_id=id"`
	Group        *Group         `bun:"rel:belongs-to,join:group_id=id"`
	UsedSlipDays []*UsedSlipDay `bun:"rel:has-many,join:id=enrollment_id"`
}

type Submission struct {
	bun.BaseModel `bun:"table:submissions,alias:s"`

	ID           uint64           `bun:"id,pk"`
	AssignmentID uint64           `bun:"assignment_id"`
	UserID       uint64           `bun:"user_id"`
	GroupID      uint64           `bun:"group_id"`
	Score        uint32           `bun:"score"`
	CommitHash   string           `bun:"commit_hash"`
	Released     bool             `bun:"released"`
	Status       SubmissionStatus `bun:"status"`
	ApprovedDate time.Time        `bun:"approved_date"`

	Assignment *Assignment `bun:"rel:belongs-to,join:assignment_id=id"`
	User       *User       `bun:"rel:belongs-to,join:user_id=id"`
	Group      *Group      `bun:"rel:belongs-to,join:group_id=id"`
	Reviews    []*Review   `bun:"rel:has-many,join:id=submission_id"`
	Scores     []*Score    `bun:"rel:has-many,join:id=submission_id"`
	Grades     []*Grade    `bun:"rel:has-many,join:id=submission_id"`
	BuildInfo  *BuildInfo  `bun:"rel:has-one,join:id=submission_id"`
}

type UsedSlipDay struct {
	bun.BaseModel `bun:"table:used_slip_days,alias:usd"`

	ID           uint64 `bun:"id,pk"`
	EnrollmentID uint64 `bun:"enrollment_id"`
	AssignmentID uint64 `bun:"assignment_id"`
	UsedDays     uint32 `bun:"used_days"`

	Enrollment *Enrollment `bun:"rel:belongs-to,join:enrollment_id=id"`
	Assignment *Assignment `bun:"rel:belongs-to,join:assignment_id=id"`
}

type Review struct {
	bun.BaseModel `bun:"table:reviews,alias:rv"`

	ID           uint64    `bun:"id,pk"`
	SubmissionID uint64    `bun:"submission_id"`
	ReviewerID   uint64    `bun:"reviewer_id"`
	Feedback     string    `bun:"feedback"`
	Ready        bool      `bun:"ready"`
	Score        int32     `bun:"score"`
	Edited       time.Time `bun:"edited"`

	Submission        *Submission         `bun:"rel:belongs-to,join:submission_id=id"`
	Reviewer          *User               `bun:"rel:belongs-to,join:reviewer_id=id"`
	GradingBenchmarks []*GradingBenchmark `bun:"rel:has-many,join:id=review_id"`
}

type GradingBenchmark struct {
	bun.BaseModel `bun:"table:grading_benchmarks,alias:gb"`

	ID           uint64 `bun:"id,pk"`
	CourseID     uint64 `bun:"course_id"`
	AssignmentID uint64 `bun:"assignment_id"`
	ReviewID     uint64 `bun:"review_id"`
	Heading      string `bun:"heading"`
	Comment      string `bun:"comment"`

	Course     *Course             `bun:"rel:belongs-to,join:course_id=id"`
	Assignment *Assignment         `bun:"rel:belongs-to,join:assignment_id=id"`
	Review     *Review             `bun:"rel:belongs-to,join:review_id=id"`
	Criteria   []*GradingCriterion `bun:"rel:has-many,join:id=benchmark_id"`
}

type GradingCriterion struct {
	bun.BaseModel `bun:"table:grading_criterions,alias:gc"`

	ID          uint64         `bun:"id,pk"`
	BenchmarkID uint64         `bun:"benchmark_id"`
	CourseID    uint64         `bun:"course_id"`
	Points      int32          `bun:"points"`
	Description string         `bun:"description"`
	Grade       CriterionGrade `bun:"grade"`
	Comment     string         `bun:"comment"`

	Benchmark *GradingBenchmark `bun:"rel:belongs-to,join:benchmark_id=id"`
	Course    *Course           `bun:"rel:belongs-to,join:course_id=id"`
}

type Task struct {
	bun.BaseModel `bun:"table:tasks,alias:t"`

	ID              uint64 `bun:"id,pk"`
	AssignmentID    uint64 `bun:"assignment_id"`
	AssignmentOrder uint32 `bun:"assignment_order"`
	Title           string `bun:"title"`
	Body            string `bun:"body"`
	Name            string `bun:"name"`

	Assignment *Assignment `bun:"rel:belongs-to,join:assignment_id=id"`
	Issues     []*Issue    `bun:"rel:has-many,join:id=task_id"`
}

type Issue struct {
	bun.BaseModel `bun:"table:issues,alias:i"`

	ID             uint64 `bun:"id,pk"`
	RepositoryID   uint64 `bun:"repository_id"`
	TaskID         uint64 `bun:"task_id"`
	ScmIssueNumber int    `bun:"scm_issue_number"`

	Repository   *Repository    `bun:"rel:belongs-to,join:repository_id=id"`
	Task         *Task          `bun:"rel:belongs-to,join:task_id=id"`
	PullRequests []*PullRequest `bun:"rel:has-many,join:id=issue_id"`
}

type PullRequest struct {
	bun.BaseModel `bun:"table:pull_requests,alias:pr"`

	ID                     uint64           `bun:"id,pk"`
	ScmRepositoryID        uint64           `bun:"scm_repository_id"`
	TaskID                 uint64           `bun:"task_id"`
	IssueID                uint64           `bun:"issue_id"`
	UserID                 uint64           `bun:"user_id"`
	ScmCommentID           uint64           `bun:"scm_comment_id"`
	SourceBranch           string           `bun:"source_branch"`
	ImprovementSuggestions string           `bun:"improvement_suggestions"`
	Number                 int              `bun:"number"`
	Stage                  PullRequestStage `bun:"stage"`

	User  *User  `bun:"rel:belongs-to,join:user_id=id"`
	Task  *Task  `bun:"rel:belongs-to,join:task_id=id"`
	Issue *Issue `bun:"rel:belongs-to,join:issue_id=id"`
}

type BuildInfo struct {
	bun.BaseModel `bun:"table:build_infos,alias:bi"`

	ID             uint64    `bun:"id,pk"`
	SubmissionID   uint64    `bun:"submission_id"`
	BuildLog       string    `bun:"build_log"`
	ExecTime       int64     `bun:"exec_time"`
	BuildDate      time.Time `bun:"build_date"`
	SubmissionDate time.Time `bun:"submission_date"`

	Submission *Submission `bun:"rel:belongs-to,join:submission_id=id"`
}

type Score struct {
	bun.BaseModel `bun:"table:scores,alias:sc"`

	ID           uint64 `bun:"id,pk"`
	SubmissionID uint64 `bun:"submission_id"`
	TestName     string `bun:"test_name"`
	TaskName     string `bun:"task_name"`
	Score        int32  `bun:"score"`
	MaxScore     int32  `bun:"max_score"`
	Weight       int32  `bun:"weight"`
	TestDetails  string `bun:"test_details"`

	Submission *Submission `bun:"rel:belongs-to,join:submission_id=id"`
}

type Grade struct {
	bun.BaseModel `bun:"table:grades,alias:gr"`

	SubmissionID uint64           `bun:"submission_id,pk"`
	UserID       uint64           `bun:"user_id,pk"`
	Status       SubmissionStatus `bun:"status"`

	Submission *Submission `bun:"rel:belongs-to,join:submission_id=id"`
	User       *User       `bun:"rel:belongs-to,join:user_id=id"`
}

type TestInfo struct {
	bun.BaseModel `bun:"table:test_infos,alias:ti"`

	ID           uint64 `bun:"id,pk"`
	AssignmentID uint64 `bun:"assignment_id"`
	TestName     string `bun:"test_name"`
	MaxScore     int32  `bun:"max_score"`
	Weight       int32  `bun:"weight"`
	Details      string `bun:"details"`

	Assignment *Assignment `bun:"rel:belongs-to,join:assignment_id=id"`
}

type AssignmentFeedback struct {
	bun.BaseModel `bun:"table:assignment_feedbacks,alias:af"`

	ID                     uint64    `bun:"id,pk"`
	AssignmentID           uint64    `bun:"assignment_id"`
	CourseID               uint64    `bun:"course_id"`
	LikedContent           string    `bun:"liked_content"`
	ImprovementSuggestions string    `bun:"improvement_suggestions"`
	TimeSpent              int32     `bun:"time_spent"`
	CreatedAt              time.Time `bun:"created_at"`

	Assignment *Assignment `bun:"rel:belongs-to,join:assignment_id=id"`
	Course     *Course     `bun:"rel:belongs-to,join:course_id=id"`
}

type FeedbackReceipt struct {
	bun.BaseModel `bun:"table:feedback_receipts,alias:fr"`

	AssignmentID uint64 `bun:"assignment_id,pk"`
	UserID       uint64 `bun:"user_id,pk"`

	Assignment *Assignment `bun:"rel:belongs-to,join:assignment_id=id"`
	User       *User       `bun:"rel:belongs-to,join:user_id=id"`
}
