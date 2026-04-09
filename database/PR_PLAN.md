# QuickFeed Database Refactor PR Plan

This file defines the step-by-step PR structure for the database refactor.

---

## PR Order

### 1. Migration System + Current Schema
- Goal: generate 001_initial_setup.sql from `context/models/current/`.
- No application code changes.
- Copilot should read `context/models/current/` and produce a runnable SQL schema snapshot.

### 2. Add New Schema as Forward Migration
- Goal: generate 002/003/004 SQL files from `models/revised/`.
- No code or ORM changes.
- Copilot should create SQL that modifies the schema step-by-step.

### 3. Schema Compatibility (GORM)
- Update models to match new schema.
- Fix relations.
- Minimal query fixes if required.
- Purpose: verify schema does not break functionality.

### 4. Kill Preload (Core Refactor)
- Remove ALL Preload usage.
- Replace Preload with explicit JOIN queries.
- Introduce explicit result structs.
- Group queries into repository methods:
    ```go
    func (r *Repo) GetAssignmentsWithSubmissions(ctx, courseID int64) ([]AssignmentWithSubmission, error)
    ```
- Enforce query constraints:
    ```
    // Requirements:
    // - Single SQL query
    // - No N+1 queries
    // - Explicit joins only
    // - Must match previous behavior
    ```
- No Bun yet, no architectural changes, no logic movement.

### 5. Strengthen DB Layer
- Move integrity-critical logic into DB layer.
- Add transactions for multi-step operations.
- Enforce invariants.

### 6. Introduce Bun Implementation
- Translate explicit queries from GORM to Bun.
- Example: `Joins(...) → NewSelect().Join(...)`.

### 7. Remove GORM
- Keep behavior identical.
- Replace all remaining GORM usage with Bun.

### 8. Cleanup + Optimization
- Remove temporary structs if no longer needed.
- Optimize queries for performance.
- Improve readability of DB layer.
