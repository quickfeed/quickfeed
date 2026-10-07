package database

import (
	"fmt"

	"gorm.io/gorm"
)

// retiredTables are tables from removed features; AutoMigrate never drops them.
var retiredTables = []string{"pull_requests", "tasks", "issues"}

// retiredColumns are columns from removed features, keyed by table name.
var retiredColumns = map[string][]string{
	"scores": {"task_name"},
}

// dropRetired removes tables and columns left behind by features that no longer exist.
func dropRetired(conn *gorm.DB) error {
	m := conn.Migrator()
	for _, table := range retiredTables {
		if !m.HasTable(table) {
			continue
		}
		if err := m.DropTable(table); err != nil {
			return fmt.Errorf("dropping table %s: %w", table, err)
		}
	}
	for table, columns := range retiredColumns {
		if !m.HasTable(table) {
			continue
		}
		for _, column := range columns {
			if !m.HasColumn(table, column) {
				continue
			}
			// The GORM migrator drops a column by recreating the table from the
			// model's schema, which no longer knows the column; use SQL instead.
			if err := conn.Exec(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", table, column)).Error; err != nil {
				return fmt.Errorf("dropping column %s.%s: %w", table, column, err)
			}
		}
	}
	return nil
}

// needsScoreBackfill reports whether the database predates the separate test
// and review scores, and must be backfilled by backfillScores after AutoMigrate.
func needsScoreBackfill(conn *gorm.DB) bool {
	m := conn.Migrator()
	return m.HasTable("submissions") && !m.HasColumn("submissions", "review_score")
}

// backfillScores copies each submission's score into its test or review score,
// and gives assignments with reviewers a review weight of 100. Tests were never
// run for such assignments, so this keeps every submission's score unchanged
// until the tests repository is read again.
func backfillScores(conn *gorm.DB) error {
	return conn.Transaction(func(tx *gorm.DB) error {
		for _, stmt := range []string{
			"UPDATE assignments SET review_weight = CASE WHEN reviewers > 0 THEN 100 ELSE 0 END",
			"UPDATE submissions SET review_score = score WHERE assignment_id IN (SELECT id FROM assignments WHERE reviewers > 0)",
			"UPDATE submissions SET test_score = score WHERE assignment_id IN (SELECT id FROM assignments WHERE reviewers = 0)",
		} {
			if err := tx.Exec(stmt).Error; err != nil {
				return fmt.Errorf("backfilling scores: %w", err)
			}
		}
		return nil
	})
}
