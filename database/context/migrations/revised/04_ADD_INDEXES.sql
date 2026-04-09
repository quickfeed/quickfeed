-- Commented indexes are created by UNIQUE contraints in migrations, but i added
-- them here for visibility

-- users, used EVERY login and joins
CREATE INDEX idx_users_login ON "users"("login"); -- added if needed for login, but no idea what it is?
CREATE INDEX idx_users_email ON "users"("email"); 
-- CREATE INDEX idx_users_scm_remote_id ON "users"("scm_remote_id");
-- CREATE INDEX idx_users_code_year_composite ON "users"("student_id", "email");

-- used_slip_days
CREATE INDEX idx_used_slip_days_enrollment ON "used_slip_days"("enrollment_id");
CREATE INDEX idx_used_slip_days_assignment ON "used_slip_days"("assignment_id");

-- courses
-- CREATE INDEX idx_courses_code_year_composite ON "courses"("code", "year");

-- groups
CREATE INDEX idx_groups_course ON "groups"("course_id");
CREATE INDEX idx_groups_id_status_composite ON "groups"("status", "id");
-- CREATE INDEX idx_groups_course_name_composite ON "groups"("course_id", "name");

-- group_users, needed to find user/group from group_user
CREATE INDEX idx_group_users_group ON "group_users"("group_id");
CREATE INDEX idx_group_users_user ON "group_users"("user_id");
-- CREATE INDEX idx_group_users_group_user_composite ON "group_users"("group_id", "user_id");

-- enrollments, Speeds up updating enrollments
CREATE INDEX idx_enrollments_id_user_composite ON "enrollments"("id", "user_id");
CREATE INDEX idx_enrollments_user ON "enrollments"("user_id");
CREATE INDEX idx_enrollments_course ON "enrollments"("course_id");
CREATE INDEX idx_enrollments_status ON "enrollments"("status");
-- CREATE INDEX idx_enrollments_user_course_composite ON "enrollments"("user_id", "course_id");

-- assignments
CREATE INDEX idx_assignments_course ON "assignments"("course_id");

-- assignment_feedback
CREATE INDEX idx_assignment_feedback_assignment ON "assignment_feedback"("assignment_id");
CREATE INDEX idx_assignment_feedback_course ON "assignment_feedback"("course_id");

-- feedback_receipt
CREATE INDEX idx_feedback_receipt_assignment ON "feedback_receipt"("assignment_id");
CREATE INDEX idx_feedback_receipt_user ON "feedback_receipt"("user_id");

-- test_info, for displaying test name on webpage 
CREATE INDEX idx_test_info_assignment ON "test_info"("assignment_id");
-- CREATE INDEX idx_test_info_assignment_test_name_composite ON "test_info"("assignment_id", "test_name"); 

-- submissions, for frequent queries
CREATE INDEX idx_submissions_assignment ON "submissions"("assignment_id");
CREATE INDEX idx_submissions_user ON "submissions"("user_id");
CREATE INDEX idx_submissions_group ON "submissions"("group_id");

-- reviews
CREATE INDEX idx_reviews_submission ON "reviews"("submission_id");

-- checklist
CREATE INDEX idx_checklist_review ON "checklist"("review_id");
CREATE INDEX idx_checklist_assignment ON "checklist"("assignment_id");

-- checklist_item
CREATE INDEX idx_checklist_item_checklist ON "checklist_item"("checklist_id");

-- approval
CREATE INDEX idx_approval_submission ON "approval"("submission_id");
CREATE INDEX idx_approval_enrollment ON "approval"("enrollment_id");

-- repositories
CREATE INDEX idx_repositories_group ON "repositories"("group_id");
CREATE INDEX idx_repositories_enrollment ON "repositories"("enrollment_id");
CREATE INDEX idx_repositories_repo_type ON "repositories"("repo_type");