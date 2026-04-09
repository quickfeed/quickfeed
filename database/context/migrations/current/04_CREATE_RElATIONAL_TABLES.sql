CREATE TABLE group_users (
    "group_id" INTEGER,
    "user_id" INTEGER,
    PRIMARY KEY ("group_id","user_id"),
    CONSTRAINT "fk_group_users_group" FOREIGN KEY ("group_id") REFERENCES "groups"("id"),
    CONSTRAINT "fk_group_users_user" FOREIGN KEY ("user_id") REFERENCES "users"("id")
);
CREATE TABLE used_slip_days (
    "id" INTEGER PRIMARY KEY,
    "enrollment_id" INTEGER NOT NULL,
    "assignment_id" INTEGER,
    "used_days" INTEGER,
    CONSTRAINT "fk_enrollments_used_slip_days" FOREIGN KEY ("enrollment_id") REFERENCES "enrollments"("id"),
    CONSTRAINT "fk_assignments_used_slip_days" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id")
);
CREATE TABLE reviews (
    "id" INTEGER PRIMARY KEY,
    "submission_id" INTEGER NOT NULL,
    "reviewer_id" INTEGER NOT NULL,
    "feedback" TEXT,
    "ready" NUMERIC,
    "score" INTEGER,
    "edited" DATETIME,
    CONSTRAINT "fk_submissions_reviews" FOREIGN KEY ("submission_id") REFERENCES "submissions"("id"),
    CONSTRAINT "fk_users_reviews" FOREIGN KEY ("reviewer_id") REFERENCES "users"("id")
);
CREATE TABLE grading_benchmarks (
    "id" INTEGER PRIMARY KEY,
    "course_id" INTEGER NOT NULL,
    "assignment_id" INTEGER NOT NULL,
    "review_id" INTEGER NOT NULL,
    "heading" TEXT,
    "comment" TEXT,
    CONSTRAINT "fk_assignments_grading_benchmarks" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id"),
    CONSTRAINT "fk_reviews_grading_benchmarks" FOREIGN KEY ("review_id") REFERENCES "reviews"("id"),
    CONSTRAINT "fk_courses_grading_benchmarks" FOREIGN KEY ("course_id") REFERENCES "courses"("id")
);
CREATE TABLE tasks (
    "id" INTEGER PRIMARY KEY,
    "assignment_id" INTEGER NOT NULL,
    "assignment_order" INTEGER,
    "title" TEXT,
    "body" TEXT,
    "name" TEXT,
    CONSTRAINT "fk_assignments_tasks" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id")
);
CREATE TABLE issues (
    "id" INTEGER PRIMARY KEY,
    "repository_id" INTEGER NOT NULL,
    "task_id" INTEGER NOT NULL,
    "scm_issue_number" INTEGER,
    CONSTRAINT "fk_tasks_issues" FOREIGN KEY ("task_id") REFERENCES "tasks"("id"),
    CONSTRAINT "fk_repositories_issues" FOREIGN KEY ("repository_id") REFERENCES "repositories"("id")
);
CREATE TABLE build_infos (
    "id" INTEGER PRIMARY KEY,
    "submission_id" INTEGER NOT NULL,
    "build_log" TEXT,
    "exec_time" INTEGER,
    "build_date" DATETIME,
    "submission_date" DATETIME,
    CONSTRAINT "fk_submissions_build_info" FOREIGN KEY ("submission_id") REFERENCES "submissions"("id"),
    UNIQUE("submission_id")
);
CREATE TABLE scores (
    "id" INTEGER PRIMARY KEY,
    "submission_id" INTEGER NOT NULL,
    "test_name" TEXT,
    "task_name" TEXT,
    "score" INTEGER,
    "max_score" INTEGER,
    "weight" INTEGER,
    "test_details" TEXT,
    CONSTRAINT "fk_submissions_scores" FOREIGN KEY ("submission_id") REFERENCES "submissions"("id")
);
CREATE TABLE grades (
    "submission_id" INTEGER,
    "user_id" INTEGER,
    "status" INTEGER,
    PRIMARY KEY ("submission_id", "user_id"),
    CONSTRAINT "fk_submissions_grades" FOREIGN KEY ("submission_id") REFERENCES "submissions"("id"),
    CONSTRAINT "fk_users_grades" FOREIGN KEY ("user_id") REFERENCES "users"("id")
);
CREATE TABLE test_info (
    "id" INTEGER PRIMARY KEY,
    "assignment_id" INTEGER NOT NULL,
    "test_name" TEXT NOT NULL,
    "max_score" INTEGER,
    "weight" INTEGER,
    "details" TEXT,
    CONSTRAINT "fk_assignments_test_infos" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id"),
    UNIQUE("assignment_id", "test_name")
);
CREATE TABLE assignment_feedback (
    "id" INTEGER PRIMARY KEY,
    "assignment_id" INTEGER NOT NULL,
    "course_id" INTEGER NOT NULL,
    "liked_content" TEXT,
    "improvement_suggestions" TEXT,
    "time_spent" INTEGER,
    "created_at" DATETIME,
    CONSTRAINT "fk_assignments_assignment_feedback" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id"),
    CONSTRAINT "fk_courses_assignment_feedback" FOREIGN KEY ("course_id") REFERENCES "courses"("id")
);
CREATE TABLE feedback_receipt (
    "assignment_id" INTEGER,
    "user_id" INTEGER,
    PRIMARY KEY ("assignment_id", "user_id"),
    CONSTRAINT "fk_assignments_feedback_receipt" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id"),
    CONSTRAINT "fk_users_feedback_receipt" FOREIGN KEY ("user_id") REFERENCES "users"("id")
);