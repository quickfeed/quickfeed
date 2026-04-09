CREATE TABLE users (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "is_admin" NUMERIC,
    "name" TEXT,
    "student_id" TEXT NOT NULL,
    "email" TEXT NOT NULL,
    "avatar_url" TEXT NOT NULL,
    "login" TEXT NOT NULL,
    "update_token" NUMERIC,
    "scm_remote_id" INTEGER NOT NULL,
    "refresh_token" TEXT,
    UNIQUE("scm_remote_id"),
    UNIQUE("student_id")
);
-- dependent on users
CREATE TABLE courses (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "course_creator_id" INTEGER NOT NULL,
    "name" TEXT NOT NULL,
    "code" TEXT NOT NULL,
    "year" INTEGER NOT NULL,
    "tag" TEXT,
    "scm_organization_id" INTEGER NOT NULL,
    "scm_organization_name" TEXT,
    "slip_days" INTEGER,
    "dockerfile_digest" TEXT,
    CONSTRAINT "fk_users_courses" FOREIGN KEY ("course_creator_id") REFERENCES "users"("id") ON DELETE CASCADE,
    UNIQUE("code", "year") -- need not null on both to actually work because of how sqlite handles unique
);
-- dependent on courses
CREATE TABLE groups (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "name" TEXT NOT NULL,
    "course_id" INTEGER NOT NULL,
    "status" INTEGER NOT NULL DEFAULT 0 CHECK ("status" IN (0, 1)), -- enum
    CONSTRAINT "fk_courses_groups" FOREIGN KEY ("course_id") REFERENCES "courses"("id") ON DELETE CASCADE,
    UNIQUE("course_id", "name")
);
-- dependent on groups and users
CREATE TABLE group_users (
    "group_id" INTEGER,
    "user_id" INTEGER,
    PRIMARY KEY ("group_id", "user_id"),
    CONSTRAINT "fk_group_users_group" FOREIGN KEY ("group_id") REFERENCES "groups"("id"),
    CONSTRAINT "fk_group_users_user" FOREIGN KEY ("user_id") REFERENCES "users"("id")
);
-- dependent on courses, groups and users
CREATE TABLE enrollments (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "user_id" INTEGER NOT NULL,
    "course_id" INTEGER NOT NULL,
    "group_id" INTEGER,
    "status" INTEGER NOT NULL DEFAULT 0 CHECK ("status" IN (0,1,2,3)), -- enum
    "state" INTEGER NOT NULL DEFAULT 0 CHECK ("state" IN (0,1,2,3)), -- enum
    "last_activity_date" DATETIME,
    "total_approved" INTEGER,
    CONSTRAINT "fk_courses_enrollments" FOREIGN KEY ("course_id") REFERENCES "courses"("id") ON DELETE CASCADE,
    CONSTRAINT "fk_groups_enrollments" FOREIGN KEY ("group_id") REFERENCES "groups"("id") ON DELETE CASCADE,
    CONSTRAINT "fk_users_enrollments" FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE CASCADE,
    UNIQUE("user_id", "course_id")
);
-- dependent on enrollments and groups
CREATE TABLE repositories (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "enrollments_id" INTEGER,
    "group_id" INTEGER,
    "scm_repository_id" INTEGER NOT NULL,
    "html_url" TEXT, -- no idea wht this is so maybe keep nullable?
    "repo_type" INTEGER NOT NULL DEFAULT 0 CHECK ("repo_type" IN (0,1,2,3,4,5)), -- enum
    CONSTRAINT "fk_groups_repositories" FOREIGN KEY ("group_id") REFERENCES "groups"("id") ON DELETE SET NULL,
    CONSTRAINT "fk_enrollments_repositories" FOREIGN KEY ("enrollments_id") REFERENCES "enrollments"("id") ON DELETE SET NULL,
    CHECK (("group_id" != 0 AND "enrollments_id" = 0)
    OR ("group_id" = 0 AND "enrollments_id" != 0)
));