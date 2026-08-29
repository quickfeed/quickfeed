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
