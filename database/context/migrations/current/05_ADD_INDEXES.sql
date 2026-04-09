CREATE UNIQUE INDEX idx_course ON "courses"("code","year");
CREATE UNIQUE INDEX idx_group ON "groups"("name","course_id");
CREATE UNIQUE INDEX idx_enrollment ON "enrollments"("course_id","user_id");
CREATE UNIQUE INDEX idx_repository ON "repositories"("scm_organization_id","user_id","group_id","repo_type");
CREATE UNIQUE INDEX idx_grade ON "grades"("submission_id","user_id");