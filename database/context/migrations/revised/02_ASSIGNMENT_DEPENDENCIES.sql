-- dependent on courses
CREATE TABLE assignments (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "course_id" INTEGER NOT NULL,
    "name" TEXT NOT NULL,
    "deadline" DATETIME,
    "auto_approve" NUMERIC,
    "order" INTEGER,
    "is_group_lab" NUMERIC,
    "score_limit" INTEGER,
    "reviewers" INTEGER,
    "container_timeout" INTEGER,
    CONSTRAINT "fk_courses_assignments" FOREIGN KEY ("course_id") REFERENCES "courses"("id") ON DELETE CASCADE
);
-- dependent on assignments and courses
CREATE TABLE assignment_feedback (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "assignment_id" INTEGER NOT NULL,
    "course_id" INTEGER NOT NULL,
    "liked_content" TEXT,
    "improvement_suggestions" TEXT,
    "time_spent" INTEGER,
    "created_at" DATETIME NOT NULL,
    CONSTRAINT "fk_assignments_assignment_feedback" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id") ON DELETE CASCADE,
    CONSTRAINT "fk_courses_assignment_feedback" FOREIGN KEY ("course_id") REFERENCES "courses"("id") ON DELETE CASCADE
);
-- dependent on assignments
CREATE TABLE test_info (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "assignment_id" INTEGER NOT NULL,
    "test_name" TEXT NOT NULL,
    "max_score" INTEGER NOT NULL,
    "weight" INTEGER NOT NULL,
    "details" TEXT,
    CONSTRAINT "fk_assignments_test_infos" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id") ON DELETE CASCADE,
    UNIQUE("assignment_id", "test_name")
);
-- dependent on assignments, groups, and users
CREATE TABLE submissions (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "assignment_id" INTEGER NOT NULL,
    "group_id" INTEGER,
    "user_id" INTEGER,
    "score" INTEGER,
    "commit_hash" TEXT NOT NULL,
    "released" NUMERIC NOT NULL,
    "approved_date" DATETIME,
    CONSTRAINT "fk_assignments_submissions" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id") ON DELETE CASCADE,
    CONSTRAINT "fk_groups_submissions" FOREIGN KEY ("group_id") REFERENCES "groups"("id") ON DELETE CASCADE,
    CONSTRAINT "fk_users_submissions" FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE CASCADE,
    CHECK (("group_id" != 0 AND "user_id" = 0)
    OR ("group_id" = 0 AND "user_id" != 0)
    )
);
-- dependent on assignments and enrollments
CREATE TABLE used_slip_days (
    "assignment_id" INTEGER,
    "enrollment_id" INTEGER,
    "used_days" INTEGER,
    PRIMARY KEY ("assignment_id", "enrollment_id"),
    CONSTRAINT "fk_enrollments_used_slip_days" FOREIGN KEY ("enrollment_id") REFERENCES "enrollments"("id") ON DELETE CASCADE,
    CONSTRAINT "fk_assignments_used_slip_days" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id") ON DELETE CASCADE
);
-- dependent on assignment and users
CREATE TABLE feedback_receipt (
    "assignment_id" INTEGER,
    "user_id" INTEGER,
    PRIMARY KEY ("assignment_id", "user_id"),
    CONSTRAINT "fk_assignments_feedback_receipt" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id") ON DELETE CASCADE,
    CONSTRAINT "fk_users_feedback_receipt" FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE CASCADE
);