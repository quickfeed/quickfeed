CREATE TABLE users (
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
);
CREATE TABLE courses (
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
);
CREATE TABLE assignments (
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
);
