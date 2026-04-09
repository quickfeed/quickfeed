CREATE TABLE pull_requests (
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
);
CREATE TABLE repositories (
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
);