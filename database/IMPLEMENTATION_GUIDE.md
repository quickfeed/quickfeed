# QuickFeed Database Refactor - Implementation Instructions

**You are Claude AI.** This document contains instructions for refactoring QuickFeed's database from GORM to Bun ORM.

## Your Task

Execute database refactor in 8 phases. Work sequentially. At each **CHECKPOINT**, stop and wait for user confirmation before proceeding.

## Critical Rules (from db_refactor_rules.md)

- **NO Preload** - Use explicit JOINs only
- **Explicit queries only** - No hidden ORM behavior
- **Preserve behavior** - Results must match exactly
- **DB layer integrity** - Enforce invariants, use transactions
- **Work function-by-function** - Test each change
- **Schema source of truth** - Use ONLY `database/context/migrations/` for schema

## Reference Files

- **`database/db_refactor_rules.md`** - Complete refactoring rules
- **`database/pr_plan.md`** - 8-phase PR structure (matches this guide)
- **`database/context/migrations/current/`** - EXACT current schema SQL (use as-is)
- **`database/context/migrations/revised/`** - EXACT target schema SQL (use as-is)
- **`database/context/bun_models/`** - Reference Bun model definitions

## Strategy (from pr_plan.md)

1. Migration System + Current Schema
2. Add New Schema as Forward Migration
3. Schema Compatibility (GORM)
4. Kill Preload (Core Refactor) - **Explicit JOINs, still GORM**
5. Strengthen DB Layer - **Validation + transactions**
6. Introduce Bun Implementation - **Parallel to GORM**
7. Remove GORM - **Switch to Bun completely**
8. Cleanup + Optimization

---

## PHASE 1: Migration System Setup

**Goal:** Set up Bun migration infrastructure and capture current schema baseline.

**Actions:**

1. Install Bun dependencies:
```bash
go get github.com/uptrace/bun github.com/uptrace/bun/dialect/sqlitedialect github.com/uptrace/bun/driver/sqliteshim github.com/uptrace/bun/migrate
```

2. Create `database/bundb.go`:
```go
package database
import (
	"database/sql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)
type BunDB struct { db *bun.DB }
func NewBunDB(path string) (*BunDB, error) {
	sqldb, err := sql.Open(sqliteshim.ShimName, path)
	if err != nil { return nil, err }
	return &BunDB{db: bun.NewDB(sqldb, sqlitedialect.New())}, nil
}
func (db *BunDB) Close() error { return db.db.Close() }
```

3. Create `database/migrations/main.go` with CLI commands: init, up, down, status (basic Bun migrate scaffolding)

4. **USE EXACT FILES:** Copy `database/context/migrations/current/*.sql` files to create `database/migrations/001_initial_setup.sql`
   - Concatenate in order: 01, 02, 03, 04, 05
   - This is the EXACT current schema - do not modify

5. Test migrations: to Revised Schema

**Goal:** Add SQL migrations to transform current schema → revised schema (from testing environment).

**Actions:**

**USE EXACT FILES from `database/context/migrations/revised/`:**

1. Copy `database/context/migrations/revised/01_USER_DEPENDENCIES.sql` → `database/migrations/002_user_dependencies.sql`
   - This adds: NOT NULL constraints, group_users table, CASCADE constraints
   - **Do not modify** - use exactly as provided

2. Copy `database/context/migrations/revised/02_ASSIGNMENT_DEPENDENCIES.sql` → `database/migrations/003_assignment_dependencies.sql`
   - This adds: assignment_feedback table, constraints, test_info unique key
   - **Do not modify** - use exactly as provided

3. Copy `database/context/migrations/revised/03_SUBMISSION_DEPENDENCIES.sql` → `database/migrations/004_submission_dependencies.sql`
   - This: renames tables (grading_benchmarks→checklist, scores→approval)
   - **Do not modify** - use exactly as provided

4. Copy `database/context/migrations/revised/04_ADD_INDEXES.sql` → `database/migrations/005_add_indexes.sql`
   - Performance indexes for common queries
   - **Do not modify** - use exactly as provided

5. Register migrations in `database/migrations/main.go`

6. Test forward migrations:
```bash
cd database/migrations
go run main.go -cmd=up
```

**Verify:**
- All 5 migrations run successfully
- Schema now matches revised schema exactly
- No data loss (verify with sample data if available)

**⏸️ CHECKPOINT** - Revised schema migrations applied. User should verify schema
   - Rename `grading_benchmarks` → `checklist`
   - Rename `grading_criterion` → `checklist_item`
   - Rename `scores` → `approval`
   - Update foreign key references

4. `database/migrations/005_add_indexes.sql`:
   - Add indexes on enrollments(user_id, course_id, group_id)
   - Add indexes on submissions(assignment_id, user_id, group_id)
   - Add indexes on reviews(submission_id)

5. Test: `go run database/migrations/main.go -cmd=up`

**Verify:** Migrations run without errors, schema matches revised

**⏸️ CHECKPOINT** - Schema migrations created. Wait for user to review SQL before continuing.

---

## PHASE 3: GORM Compatibility with Revised Schema

**Goal:** Update GORM models to work with revised schema (still using Preload temporarily).

**Actions:**

1. Update `qf/*.proto` to match revised schema:
   - Rename: `GradingBenchmark` → `Checklist`
   - Rename: `GradingCriterion` → `ChecklistItem`
   - Rename: `Score` → `Approval`
   - **Reference:** See `database/context/bun_models/revised.go` for field names/types

2. Regenerate protobuf code:
```bash
make proto
```

3. Update `database/gormdb.go` AutoMigrate list:
```go
&qf.Checklist{},      // was GradingBenchmark
&qf.ChecklistItem{},  // was GradingCriterion
&qf.Approval{},       // was Score
```

4. Update ALL Preload calls across `databa - Core Refactor)

**Goal:** Remove ALL Preload, use explicit JOINs. Still using GORM (not Bun yet).

**Critical:** Follow `database/db_refactor_rules.md` - work function-by-function, test each change
   - `Preload("Scores")` → `Preload("Approvals")`

5. Update test expectations in `database/*_test.go`

6. Run tests:
```bash
make test
```

**Verify:**
- No compilation errors
- All database tests pass
- GORM still works with renamed models

**⏸️ CHECKPOINT** - GORM compatible with revised schema. Tests passing. Wait for approval before removing Preload.

---

## PHASE 4: Kill Preload (⚠️ Largest Phase)

**Goal:** Remove ALL Preload, use explicit JOINs. Still using GORM (not Bun yet).

**Actions:**

1. Create `database/query_results.go` with flat result structs:
```go
type AssignmentWithTests struct {
	ID uint64; CourseID uint64; Name string; ...
	TestID uint64 `gorm:"column:test_id"`
	TestName string `gorm:"column:test_name"`
	// ... more test fields
}
```

2. Refactor `GetAssignmentsByCourse` in `database/gormdb_assignment.go`:
```go
func (db *GormDB) GetAssignmentsByCourse(courseID uint64) ([]*qf.Assignment, error) {
	var results []AssignmentWithTests
	err := db.conn.
		Table("assignments a").
		Select("a.id", "a.course_id", "a.name", ..., "t.id as test_id", "t.test_name", ...).
		Joins("LEFT JOIN test_info t ON t.assignment_id = a.id").
		Where("a.course_id = ?", courseID).
		Scan(&results).Error
	// Transform flat results to nested []*qf.Assignment
	return transformToAssignments(results), nil
}
```

3. Refactor `GetCourseByStatus` in `database/gormdb_course.go`:
   - Remove Preload("Assignments"), Preload("Enrollments"), etc.
   - Use explicit JOINs for each status level
   - Single query per entity type (no N+1)

4. Refactor `GetSubmission` in `database/gormdb_submission.go`:
   - Remove Preload("Reviews"), Preload("BuildInfo"), etc.
   - Use explicit JOIN to get reviews + checklist + items in one query
   - Transform flat results to nested structure

5. Refactor remaining functions in `database/gormdb_user.go`, `database/gormdb_repository.go`

6. Add transformation functions to convert flat results to nested protobuf objects

7. Run: `grep -r "Preload" database/*.go` - should return NO matches

8. Run: `make test`

**Verify:** No Preload calls remain, all tests pass, no N+1 queries

**⏸️ CHECKPOINT** - All Preload removed, explicit queries working. Wait for approval before Phase 5.

---

## PHASE 5: Strengthen DB Layer

**Actions:**

1. Create `database/validation.go`:
```go
package database
var ErrMutuallyExclusive = errors.New("user_id and group_id are mutually exclusive")
func validateSubmission(s *qf.Submission) error {
	if (s.UserID > 0 && s.GroupID > 0) || (s.UserID == 0 && s.GroupID == 0) {
		return ErrMutuallyExclusive
	}
	return nil
}
// ... more validators
```

2. Create `database/transactions.go`:
```go
func (db *GormDB) execTransaction(fn func(*gorm.DB) error) error {
	return db.conn.Transaction(fn)
}
```

3. Update `CreateSubmission` in `database/gormdb_submission.go`:
```go
func (db *GormDB) CreateSubmission(submission *qf.Submission) error {
	if err := validateSubmission(submission); err != nil { return err }
	return db.execTransaction(func(tx *gorm.DB) error {
		// Check assignment exists, check enrollment, then create
	})
}
```

4. Update `CreateGroup` to use transaction and validate users are enrolled

5. Add tests in `database/gormdb_integrity_test.go`

6. Run: `go test ./database/...`

**Verify:** Validation catches invalid data, transactions roll back on error

**⏸️ CHECKPOINT** - Validation and transactions working. Wait for approval.

---

##Goal:** Implement Bun queries alongside GORM. Both work in parallel. Use Bun model reference.

**Reference:** `database/context/bun_models/revised.go` shows proper Bun model structure.

**Actions:**

1. Create `database/bundb_assignment.go` - translate GORM JOINs to Bun:
```go
func (r *BunRepository) GetAssignmentsByCourse(ctx context.Context, courseID uint64) ([]*qf.Assignment, error) {
	var results []AssignmentWithTests
	err := r.db.NewSelect().
		TableExpr("assignments AS a").
		Column("a.id", "a.course_id", ...).
		ColumnExpr("t.id AS test_id", ...).
		Join("LEFT JOIN test_info AS t").JoinOn("t.assignment_id = a.id").
		Where("a.course_id = ?", courseID).
		Scan(ctx, &results)
	// Reuse transformToAssignments from Phase 4
	return transformToAssignments(results), nil
}
```

2. Create `database/bundb_course.go` - translate GetCourseByStatus

3. Create `database/bundb_submission.go` - translate GetSubmission

4. Create `internal/qtest/bun.go` for Bun test helpers

5. Create `database/bundb_test.go` - verify Bun produces SAME results as GORM:
```go
func TestBunVsGorm_GetAssignments(t *testing.T) {
	// Run same query on both, compare results
}
```

6. Run tests:
```bash
go test ./database/... -run TestBun
```

**Verify:**
- Bun queries produce identical results to GORM
- All Bun tests pass
- GORM tests still pass (both working)

**⏸️ CHECKPOINT** - Bun working in parallel. Ready to remove GORM. Wait for approval

**⏸️ CHECKPOINT** - Bun working alongside GORM. Wait for approval before removing GORM.

---

## PHASE 7: Remove GORM

**Actions:**

1. Update `database/database.go` interface - add `context.Context` to all methods:
```go
type Database interface {
	GetUser(ctx context.Context, userID uint64) (*qf.User, error)
	GetCourse(ctx context.Context, courseID uint64) (*qf.Course, error)
	// ... all methods with ctx
}
```

2. Implement full Database interface in `database/bundb.go`

3. Create `database/bun_logger.go` for query logging

4. Update ALL RPC handlers in `web/*.go` to pass `ctx` to database calls:
```go
// Before: course, err := s.db.GetCourse(req.Msg.GetCourseID())
// After:  course, err := s.db.GetCourse(ctx, req.Msg.GetCourseID())
```

5. Update `main.go`: Replace `database.NewGormDB(...)` with `database.NewBunDB(...)`

6. Delete all `database/gormdb*.go` files

7. Run: `go mod tidy` to remove GORM dependencies

8. Run: `make test` - ALL tests must pass

**Verify:**
- `grep -r "gorm" go.mod` returns nothing
- `grep -r "Preload" database/` returns nothing
- All tests pass

**⏸️ CHECKPOINT** - GORM removed, Bun fully working. Wait for approval.

---

## PHASE 8: Cleanup & Optimization

**Actions:**

1. Create `database/migrations/006_optimize_indexes.sql`:
```sql
CREATE INDEX idx_submissions_assignment_user ON submissions(assignment_id, user_id);
CREATE INDEX idx_submissions_assignment_group ON submissions(assignment_id, group_id);
CREATE INDEX idx_enrollments_course_status ON enrollments(course_id, status);
```

2. Clean up `database/query_results.go` - remove unused structs
Key Patterns & Rules

### Rules from db_refactor_rules.md

1. **No Preload** - Never use ORM implicit loading
2. **Explicit Queries Only** - All JOINs must be explicit
3. **Preserve Behavior** - Results must match exactly
4. **DB Layer Integrity** - Enforce invariants, use transactions
5. **Work Function-by-Function** - Test each change immediately

### Schema Source of Truth

**Current Schema:** `database/context/migrations/current/*.sql`
**Revised Schema:** `database/context/migrations/revised/*.sql`
**Bun Models Reference:** `database/context/bun_models/revised.go`

Do NOT infer schema - always read from these files.

### Explicit JOIN Pattern (GORM)
```go
db.conn.Table("courses c").
	Select("c.*, e.id as enrollment_id, u.name as user_name").
	Joins("LEFT JOIN enrollments e ON e.course_id = c.id").
	Joins("LEFT JOIN users u ON u.id = e.user_id").
	Where("c.id = ?", courseID).
	Scan(&results)
```

### Explicit JOIN Pattern (Bun)
```go
db.NewSelect().
	TableExpr("courses AS c").
	Column("c.*").
	ColumnExpr("e.id AS enrollment_id", "u.name AS user_name").
	Join("LEFT JOIN enrollments AS e").JoinOn("e.course_id = c.id").
	Join("LEFT JOIN users AS u").JoinOn("u.id = e.user_id").
	Where("c.id = ?", courseID).
	Scan(ctx, &results)
```

### Transaction Pattern (Bun)
```go
err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
	// Validate
	exists, err := tx.NewSelect().Model((*Assignment)(nil)).Where("id = ?", id).Exists(ctx)
	if err != nil { return err }
	// Insert
	_, err = tx.NewInsert().Model(submission).Exec(ctx)
	return err
})
```

### Query Design (from db_refactor_rules.md)

- Prefer flat result structures
- Avoid deeply nested ORM models
- Use LEFT JOIN unless INNER JOIN explicitly required
- Always anchor queries to migration schema filesin("LEFT JOIN users AS u").JoinOn("u.id = e.user_id").
	Where("c.id = ?", courseID).
	Scan(ctx, &results)
```

### Transaction Pattern (Bun)
```go
err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
	// Validate
	exists, err := tx.NewSelect().Model((*Assignment)(nil)).Where("id = ?", id).Exists(ctx)
	if err != nil { return err }
	// Insert
	_, err = tx.NewInsert().Model(submission).Exec(ctx)
	return err
})
```

---

## How to Use This Guide

When user says "Execute Phase N":
1. Read all steps in that phase
2. Create files as specified
3. Run verification commands
4. Stop at CHECKPOINT and report results
5. Wait for user to say "continue" or "proceed to Phase N+1"

When tests fail:
1. Show the error
2. Analyze the problem
3. Suggest a fix
4. Apply the fix when approved
5. Rerun tests

Remember:
- **NO Preload** ever
- **Explicit queries** always
- **Test frequently**
- **Preserve behavior** exactly
