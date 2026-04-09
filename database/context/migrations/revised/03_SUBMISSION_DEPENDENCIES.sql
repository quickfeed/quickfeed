-- dependent on submissions
CREATE TABLE build_info (
    "submission_id" INTEGER PRIMARY KEY,
    "build_log" TEXT,
    "exec_time" INTEGER,
    "build_date" DATETIME,
    "submission_date" DATETIME,
    CONSTRAINT "fk_submissions_build_info" FOREIGN KEY ("submission_id") REFERENCES "submissions"("id") ON DELETE CASCADE
);
CREATE TABLE reviews (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "submission_id" INTEGER NOT NULL,
    "feedback" TEXT,
    "ready" NUMERIC,
    "score" INTEGER NOT NULL,
    "edited" DATETIME,
    CONSTRAINT "fk_submissions_reviews" FOREIGN KEY ("submission_id") REFERENCES "submissions"("id") ON DELETE CASCADE
);
-- dependent on reviews
CREATE TABLE checklist (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "review_id" INTEGER DEFAULT NULL,
    "assignment_id" INTEGER NOT NULL,
    "heading" TEXT,
    "comment" TEXT,
    CONSTRAINT "fk_reviews_checklist" FOREIGN KEY ("review_id") REFERENCES "reviews"("id") ON DELETE SET NULL,
    CONSTRAINT "fk_assignments_checklist" FOREIGN KEY ("assignment_id") REFERENCES "assignments"("id") ON DELETE CASCADE
);
-- dependent on checklist
CREATE TABLE checklist_item (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "checklist_id" INTEGER NOT NULL,
    "points" INTEGER NOT NULL,
    "description" TEXT NOT NULL,
    "grade" INTEGER NOT NULL DEFAULT 0 CHECK ("grade" IN (0,1,2)), -- enum
    "comment" TEXT,
    CONSTRAINT "fk_checklist_checklist_item" FOREIGN KEY ("checklist_id") REFERENCES "checklist"("id") ON DELETE CASCADE
);
-- dependent on enrollments and submissions
CREATE TABLE approval (
    "submission_id" INTEGER,
    "enrollment_id" INTEGER,
    "decision" INTEGER NOT NULL DEFAULT 0 CHECK ("decision" IN (0,1,2,3)), -- enum
    PRIMARY KEY ("submission_id", "enrollment_id"),
    CONSTRAINT "fk_submissions_scores" FOREIGN KEY ("submission_id") REFERENCES "submissions"("id") ON DELETE CASCADE,
    CONSTRAINT "fk_enrollments_scores" FOREIGN KEY ("enrollment_id") REFERENCES "enrollments"("id") ON DELETE CASCADE
);