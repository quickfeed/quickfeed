# QuickFeed Database Refactor Rules

## Purpose
This file defines strict rules for refactoring QuickFeed's database layer. All code generation and modifications must follow these rules.

---

## Source of Truth
- Database schema is defined **ONLY** in `/database/context/migrations/`.
- Do **NOT** infer schema from ORM models.
- Do **NOT** guess relationships.
- If unsure, read migration files first.

---

## Core Rules

1. **No Preload**
   - All implicit ORM relation loading is prohibited.
   - Must use explicit JOIN queries.
   - Bad: `db.Preload("Submissions").Find(&assignments)`
   - Good: `db.Table("assignments").Select(...).Joins(...).Scan(...)`

2. **Explicit Queries Only**
   - Use explicit JOINs and SELECT fields.
   - Avoid hidden ORM behavior.
   - Queries must be predictable.

3. **Preserve Behavior**
   - Refactors must NOT change QuickFeed logic.
   - All results must match original functionality.

4. **DB Layer Integrity**
   - Enforce invariants and atomic operations.
   - Use transactions for multi-step operations.
   - DB layer handles integrity logic, not HTTP or permissions.

5. **Refactor Discipline**
   - Work function-by-function.
   - Avoid large batch changes.
   - Test each change before moving to the next.

---

## Query Design Rules
- Prefer flat result structures.
- Avoid deeply nested ORM models.
- Use LEFT JOIN unless INNER JOIN is explicitly required.
- Always anchor queries to migration schema.

---

## Copilot Usage Guidelines
- Follow this file and `/database/pr_plan.md`.
- Phase-specific instructions must be obeyed.
- Do **NOT** invent schema.
- Do **NOT** reintroduce Preload.
- Always produce explicit queries.
- Transformations only, not new implementations.

---

## Workflow Per Function
1. Read migration schema in `/database/context/migrations/`.
2. Understand original function in QuickFeed.
3. Rewrite function using explicit queries.
4. Verify behavior matches original.
5. Proceed to next function.

---

## Migration Generation Guidelines

- Generate one SQL file per migration step.
- Name sequentially:
    001_initial_setup.sql
    002_drop_redundant_tables.sql
    003_add_join_tables.sql
    004_modify_constraints.sql
- Use the current schema as the base for 001_initial_setup.sql.
- Use the revised schema to generate forward migrations.
- Each migration must be valid SQL, runnable on PostgreSQL (or your DB).
- Include comments for clarity, e.g.,
    -- Drop redundant table X
    -- Add join table Y
- Preserve data integrity where applicable.
- Do NOT include application code—migrations are SQL only.
