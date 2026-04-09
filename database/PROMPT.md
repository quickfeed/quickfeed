# Quickfeed Database Refactor - AI Implementation Prompt

## Context
- Project: `quickfeed` database system
- Current ORM: Gorm
- Tested alternatives: sqlc, Bun, Ent → **Decision: Bun**
- SQL migrations are in `context/migrations/*` for both current and revised schema
- Reference files:
  - `db_refactor_rules.md` → database refactoring rules
  - `pr_plan.md` → implementation plan

## Goals
1. Increase database layer responsibility while keeping application layer functional:
   - Add query-level validation
   - Implement DB-layer error handling (currently only in application)
   - Use transactions where appropriate
2. Make implementation reviewer-friendly via **stacked pull-requests**
3. Produce **ready-to-execute instructions**, leaving manual oversight for critical decisions

## Instructions for AI

**Task:** Generate a detailed step-by-step implementation guide for refactoring `quickfeed` database using Bun. The output must be executable instructions with file paths, function names, and concrete steps. Flag manual steps explicitly.

### Steps AI should produce

1. **Migration Strategy**
   - Compare current vs revised schema
   - Generate Bun-compatible migration code
   - Specify execution order
   - Highlight potential breaking changes and manual review points

2. **Database Layer Implementation**
   - Implement query-level validation for all tables according to `db_refactor_rules.md`
   - Add error handling for insert/update/delete operations
   - Wrap critical operations in transactions
   - Provide type-safe Bun queries

3. **Application Layer Adjustments**
   - Identify validations that can move from application to database
   - Mark application logic that must remain
   - Ensure existing API contracts are preserved

4. **Pull-Request Structure**
   - Break changes into **stacked incremental PRs**
   - Each PR should:
     - Contain a single migration or logical DB change
     - Include corresponding Bun queries and validations
     - Contain test cases for DB-layer validations and transactions
   - Provide clear commit messages explaining each change

5. **Testing and Verification**
   - Suggest unit tests for Bun queries
   - Suggest integration tests for transactions and error handling
   - Flag manual steps requiring oversight (schema conflicts, breaking API changes, complex migrations)

## Constraints
- All output steps must be actionable or explicitly flagged for manual review
- Avoid vague statements; every step should include concrete paths, functions, or commands wherever possible
- Maintain full manual control for critical decisions
