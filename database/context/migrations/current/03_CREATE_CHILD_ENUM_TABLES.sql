CREATE TABLE groups (
    "id" INTEGER PRIMARY KEY,
    "name" TEXT,
    "course_id" INTEGER NOT NULL,
    "status" INTEGER NOT NULL DEFAULT 0 CHECK ("status" IN (0, 1)), -- enum
    CONSTRAINT "fk_courses_groups" FOREIGN KEY ("course_id") REFERENCES "courses"("id"),
    UNIQUE("course_id", "name")
);
CREATE TABLE enrollments (
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
);
CREATE TABLE submissions (
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
);
CREATE TABLE grading_criterions (
    "id" INTEGER PRIMARY KEY,
    "benchmark_id" INTEGER NOT NULL,
    "course_id" INTEGER NOT NULL,
    "points" INTEGER,
    "description" TEXT,
    "grade" INTEGER NOT NULL DEFAULT 0 CHECK ("grade" IN (0,1,2)), -- enum
    "comment" TEXT,
    CONSTRAINT "fk_grading_benchmarks_criteria" FOREIGN KEY ("benchmark_id") REFERENCES "grading_benchmarks"("id"),
    CONSTRAINT "fk_courses_grading_criterions" FOREIGN KEY ("course_id") REFERENCES "courses"("id")
);