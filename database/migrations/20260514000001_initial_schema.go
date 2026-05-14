package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(upInitialSchema, downInitialSchema)
}

func upInitialSchema(ctx context.Context, db *bun.DB) error {
	statements := []string{
		`CREATE TABLE users (
    		"id" INTEGER PRIMARY KEY,
    		"is_admin" INTEGER NOT NULL DEFAULT 0,
    		"name" TEXT,
    		"student_id" TEXT NOT NULL,
    		"email" TEXT,
    		"avatar_url" TEXT,
    		"login" TEXT,
    		"update_token" NUMERIC,
    		"scm_remote_id" INTEGER NOT NULL,
    		"refresh_token" TEXT,
    		UNIQUE("scm_remote_id"),
    		UNIQUE("student_id")
		)`,
		`CREATE TABLE courses (
		    "id" INTEGER PRIMARY KEY,
		    "course_creator_id" INTEGER NOT NULL,
		    "name" TEXT,
		    "code" TEXT,
		    "year" INTEGER,
		    "tag" TEXT,
		    "scm_organization_id" INTEGER,
		    "scm_organization_name" TEXT,
		    "slip_days" INTEGER,
		    "dockerfile_digest" TEXT,
		    UNIQUE("code", "year")
		)`,
		`CREATE TABLE assignments (
		    "id" INTEGER PRIMARY KEY,
		    "course_id" INTEGER NOT NULL,
		    "name" TEXT,
		    "deadline" DATETIME,
		    "auto_approve" NUMERIC,
		    "order" INTEGER,
		    "is_group_lab" NUMERIC,
		    "score_limit" INTEGER,
		    "reviewers" INTEGER,
		    "container_timeout" INTEGER,
		    CONSTRAINT "fk_courses_assignments" FOREIGN KEY ("course_id") REFERENCES "courses"("id")
		)`,
		`CREATE TABLE groups (
		    "id" INTEGER PRIMARY KEY,
		    "name" TEXT,
		    "course_id" INTEGER NOT NULL,
		    "status" INTEGER NOT NULL DEFAULT 0 CHECK ("status" IN (0, 1)), -- enum
		    CONSTRAINT "fk_courses_groups" FOREIGN KEY ("course_id") REFERENCES "courses"("id"),
		    UNIQUE("course_id", "name")
		)`,
		`CREATE TABLE repositories (
		    "id" INTEGER PRIMARY KEY,
		    "scm_organization_id" INTEGER NOT NULL,
		    "scm_repository_id" INTEGER,
		    "user_id" INTEGER NOT NULL,
		    "group_id" INTEGER NOT NULL,
		    "html_url" TEXT,
		    "repo_type" INTEGER NOT NULL DEFAULT 0 CHECK ("repo_type" IN (0,1,2,3,4,5)), -- enum
		    CONSTRAINT "fk_users_repositories" FOREIGN KEY ("user_id") REFERENCES "users"("id"),
		    CONSTRAINT "fk_groups_repositories" FOREIGN KEY ("group_id") REFERENCES "groups"("id"),
		    UNIQUE("scm_organization_id", "user_id", "group_id", "repo_type")
		)`,
		`CREATE TABLE enrollments (
		    "id" INTEGER PRIMARY KEY,
		    "course_id" INTEGER NOT NULL,
		    "user_id" INTEGER NOT NULL,
		    "group_id" INTEGER,
		    "status" INTEGER NOT NULL DEFAULT 0 CHECK ("status" IN (0,1,2,3)), -- enum
		    "state" INTEGER NOT NULL DEFAULT 0 CHECK ("state" IN (0,1,2,3)), -- enum
		    "last_activity_date" DATETIME,
		    "total_approved" INTEGER,
		    CONSTRAINT "fk_courses_enrollments" FOREIGN KEY ("course_id") REFERENCES "courses"("id"),
		    CONSTRAINT "fk_groups_enrollments" FOREIGN KEY ("group_id") REFERENCES "groups"("id"),
		    CONSTRAINT "fk_users_enrollments" FOREIGN KEY ("user_id") REFERENCES "users"("id"),
		    UNIQUE("course_id", "user_id")
		)`,
		`CREATE TABLE submissions (
		    "id" INTEGER PRIMARY KEY,
		    "assignment_id" INTEGER NOT NULL,
		    "user_id" INTEGER,
		    "group_id" INTEGER,
		    "score" INTEGER,
		    "commit_hash" TEXT,
		    "released" NUMERIC,
		    "status" INTEGER NOT NULL DEFAULT 0 CHECK ("status" IN (0,1,2,3)), -- enum
		    "approved_date" DATETIME,
		    CONSTRAINT "fk_assignments_submissions" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id"),
		    CONSTRAINT "fk_users_submissions" FOREIGN KEY ("user_id") REFERENCES "users"("id"),
		    CONSTRAINT "fk_groups_submissions" FOREIGN KEY ("group_id") REFERENCES "groups"("id")
		)`,
		`CREATE TABLE group_users (
		    "group_id" INTEGER,
		    "user_id" INTEGER,
		    PRIMARY KEY ("group_id","user_id"),
		    CONSTRAINT "fk_group_users_group" FOREIGN KEY ("group_id") REFERENCES "groups"("id"),
		    CONSTRAINT "fk_group_users_user" FOREIGN KEY ("user_id") REFERENCES "users"("id")
		)`,
		`CREATE TABLE used_slip_days (
		    "id" INTEGER PRIMARY KEY,
		    "enrollment_id" INTEGER NOT NULL,
		    "assignment_id" INTEGER,
		    "used_days" INTEGER,
		    CONSTRAINT "fk_enrollments_used_slip_days" FOREIGN KEY ("enrollment_id") REFERENCES "enrollments"("id"),
		    CONSTRAINT "fk_assignments_used_slip_days" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id")
		)`,
		`CREATE TABLE reviews (
		    "id" INTEGER PRIMARY KEY,
		    "submission_id" INTEGER NOT NULL,
		    "reviewer_id" INTEGER NOT NULL,
		    "feedback" TEXT,
		    "ready" NUMERIC,
		    "score" INTEGER,
		    "edited" DATETIME,
		    CONSTRAINT "fk_submissions_reviews" FOREIGN KEY ("submission_id") REFERENCES "submissions"("id"),
		    CONSTRAINT "fk_users_reviews" FOREIGN KEY ("reviewer_id") REFERENCES "users"("id")
		)`,
		`CREATE TABLE grading_benchmarks (
		    "id" INTEGER PRIMARY KEY,
		    "course_id" INTEGER NOT NULL,
		    "assignment_id" INTEGER NOT NULL,
		    "review_id" INTEGER NOT NULL,
		    "heading" TEXT,
		    "comment" TEXT,
		    CONSTRAINT "fk_assignments_grading_benchmarks" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id"),
		    CONSTRAINT "fk_reviews_grading_benchmarks" FOREIGN KEY ("review_id") REFERENCES "reviews"("id"),
		    CONSTRAINT "fk_courses_grading_benchmarks" FOREIGN KEY ("course_id") REFERENCES "courses"("id")
		)`,
		`CREATE TABLE grading_criterions (
		    "id" INTEGER PRIMARY KEY,
		    "benchmark_id" INTEGER NOT NULL,
		    "course_id" INTEGER NOT NULL,
		    "points" INTEGER,
		    "description" TEXT,
		    "grade" INTEGER NOT NULL DEFAULT 0 CHECK ("grade" IN (0,1,2)), -- enum
		    "comment" TEXT,
		    CONSTRAINT "fk_grading_benchmarks_criteria" FOREIGN KEY ("benchmark_id") REFERENCES "grading_benchmarks"("id"),
		    CONSTRAINT "fk_courses_grading_criterions" FOREIGN KEY ("course_id") REFERENCES "courses"("id")
		)`,
		`CREATE TABLE tasks (
		    "id" INTEGER PRIMARY KEY,
		    "assignment_id" INTEGER NOT NULL,
		    "assignment_order" INTEGER,
		    "title" TEXT,
		    "body" TEXT,
		    "name" TEXT,
		    CONSTRAINT "fk_assignments_tasks" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id")
		)`,
		`CREATE TABLE issues (
		    "id" INTEGER PRIMARY KEY,
		    "repository_id" INTEGER NOT NULL,
		    "task_id" INTEGER NOT NULL,
		    "scm_issue_number" INTEGER,
		    CONSTRAINT "fk_tasks_issues" FOREIGN KEY ("task_id") REFERENCES "tasks"("id"),
		    CONSTRAINT "fk_repositories_issues" FOREIGN KEY ("repository_id") REFERENCES "repositories"("id")
		)`,

		`CREATE TABLE pull_requests (
		    "id" INTEGER PRIMARY KEY,
		    "scm_repository_id" INTEGER,
		    "task_id" INTEGER NOT NULL,
		    "issue_id" INTEGER NOT NULL,
		    "user_id" INTEGER NOT NULL,
		    "scm_comment_id" INTEGER,
		    "source_branch" TEXT,
		    "improvement_suggestions" TEXT,
		    "number" INTEGER,
		    "stage" INTEGER NOT NULL DEFAULT 0 CHECK ("stage" IN (0,1,2,3)), -- sqlite enum
		    CONSTRAINT "fk_users_pull_requests" FOREIGN KEY ("user_id") REFERENCES "users"("id"),
		    CONSTRAINT "fk_tasks_pull_requests" FOREIGN KEY ("task_id") REFERENCES "tasks"("id"),
		    CONSTRAINT "fk_issues_pull_requests" FOREIGN KEY ("issue_id") REFERENCES "issues"("id")
		)`,
		`CREATE TABLE build_infos (
		    "id" INTEGER PRIMARY KEY,
		    "submission_id" INTEGER NOT NULL,
		    "build_log" TEXT,
		    "exec_time" INTEGER,
		    "build_date" DATETIME,
		    "submission_date" DATETIME,
		    CONSTRAINT "fk_submissions_build_info" FOREIGN KEY ("submission_id") REFERENCES "submissions"("id"),
		    UNIQUE("submission_id")
		)`,
		`CREATE TABLE scores (
		    "id" INTEGER PRIMARY KEY,
		    "submission_id" INTEGER NOT NULL,
		    "test_name" TEXT,
		    "task_name" TEXT,
		    "score" INTEGER,
		    "max_score" INTEGER,
		    "weight" INTEGER,
		    "test_details" TEXT,
		    CONSTRAINT "fk_submissions_scores" FOREIGN KEY ("submission_id") REFERENCES "submissions"("id")
		)`,
		`CREATE TABLE grades (
		    "submission_id" INTEGER,
		    "user_id" INTEGER,
		    "status" INTEGER,
		    PRIMARY KEY ("submission_id", "user_id"),
		    CONSTRAINT "fk_submissions_grades" FOREIGN KEY ("submission_id") REFERENCES "submissions"("id"),
		    CONSTRAINT "fk_users_grades" FOREIGN KEY ("user_id") REFERENCES "users"("id")
		)`,
		`CREATE TABLE test_infos (
		    "id" INTEGER PRIMARY KEY,
		    "assignment_id" INTEGER NOT NULL,
		    "test_name" TEXT NOT NULL,
		    "max_score" INTEGER,
		    "weight" INTEGER,
		    "details" TEXT,
		    CONSTRAINT "fk_assignments_test_infos" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id"),
		    UNIQUE("assignment_id", "test_name")
		)`,
		`CREATE TABLE assignment_feedbacks (
		    "id" INTEGER PRIMARY KEY,
		    "assignment_id" INTEGER NOT NULL,
		    "course_id" INTEGER NOT NULL,
		    "liked_content" TEXT,
		    "improvement_suggestions" TEXT,
		    "time_spent" INTEGER,
		    "created_at" DATETIME,
		    CONSTRAINT "fk_assignments_assignment_feedbacks" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id"),
		    CONSTRAINT "fk_courses_assignment_feedbacks" FOREIGN KEY ("course_id") REFERENCES "courses"("id")
		)`,
		`CREATE TABLE feedback_receipts (
		    "assignment_id" INTEGER,
		    "user_id" INTEGER,
		    PRIMARY KEY ("assignment_id", "user_id"),
		    CONSTRAINT "fk_assignments_feedback_receipts" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id"),
		    CONSTRAINT "fk_users_feedback_receipts" FOREIGN KEY ("user_id") REFERENCES "users"("id")
		)`,
		`CREATE UNIQUE INDEX idx_course ON "courses"("code","year")`,
		`CREATE UNIQUE INDEX idx_group ON "groups"("name","course_id")`,
		`CREATE UNIQUE INDEX idx_enrollment ON "enrollments"("course_id","user_id")`,
		`CREATE UNIQUE INDEX idx_repository ON "repositories"("scm_organization_id","user_id","group_id","repo_type")`,
		`CREATE UNIQUE INDEX idx_grade ON "grades"("submission_id","user_id")`,
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func downInitialSchema(ctx context.Context, db *bun.DB) error {
	tables := []string{
		"scores",
		"build_infos",
		"feedback_receipts",
		"assignment_feedbacks",
		"reviews",
		"grading_criterions",
		"grading_benchmarks",
		"grades",
		"submissions",
		"pull_requests",
		"issues",
		"tasks",
		"test_infos",
		"assignments",
		"used_slip_days",
		"enrollments",
		"repositories",
		"courses",
		"group_users",
		"groups",
		"users",
	}

	for _, table := range tables {
		if _, err := db.NewDropTable().TableExpr(table).IfExists().Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
