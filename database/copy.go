package database

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// copyBatchSize is the number of rows read and written at a time.
const copyBatchSize = 500

// TableCount is the number of rows in a table, in the source and in the
// destination database, after a copy.
type TableCount struct {
	Table  string
	Source int64
	Dest   int64
}

// CopyTo copies all QuickFeed data from db into dst, table by table,
// keeping all IDs. It is used to move an existing SQLite database to
// PostgreSQL. The destination must be empty; CopyTo refuses to copy into
// a database that already contains data. It returns the row count of each
// table in both databases, so the caller can verify that nothing is missing.
func (db *GormDB) CopyTo(dst *GormDB) ([]TableCount, error) {
	tables, err := db.tables()
	if err != nil {
		return nil, err
	}
	for _, t := range tables {
		var n int64
		if err := dst.conn.Table(t.name).Count(&n).Error; err != nil {
			return nil, fmt.Errorf("counting rows in destination table %s: %w", t.name, err)
		}
		if n > 0 {
			return nil, fmt.Errorf("destination table %s is not empty (%d rows); copy only into a new, empty database", t.name, n)
		}
	}

	err = dst.conn.Transaction(func(tx *gorm.DB) error {
		for _, t := range tables {
			if err := t.copy(db.conn, tx); err != nil {
				return fmt.Errorf("copying table %s: %w", t.name, err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	counts := make([]TableCount, 0, len(tables))
	for _, t := range tables {
		c := TableCount{Table: t.name}
		if err := db.conn.Table(t.name).Count(&c.Source).Error; err != nil {
			return nil, err
		}
		if err := dst.conn.Table(t.name).Count(&c.Dest).Error; err != nil {
			return nil, err
		}
		counts = append(counts, c)
	}
	return counts, nil
}

// table is one table to copy: either a model's table, or a join table
// without its own model (model is nil), such as group_users.
type table struct {
	name  string
	model any
}

// tables returns the tables of all models, followed by their join tables.
func (db *GormDB) tables() ([]table, error) {
	var tables, joinTables []table
	seen := map[string]bool{}
	for _, m := range models {
		stmt := &gorm.Statement{DB: db.conn}
		if err := stmt.Parse(m); err != nil {
			return nil, err
		}
		tables = append(tables, table{name: stmt.Schema.Table, model: m})
		for _, rel := range stmt.Schema.Relationships.Many2Many {
			if name := rel.JoinTable.Table; !seen[name] {
				seen[name] = true
				joinTables = append(joinTables, table{name: name})
			}
		}
	}
	return append(tables, joinTables...), nil
}

// copy copies all rows of the table from src to dst in batches.
// Rows are read in SQLite's storage order (rowid), which every QuickFeed table has.
func (t table) copy(src, dst *gorm.DB) error {
	for offset := 0; ; offset += copyBatchSize {
		query := src.Table(t.name).Order("rowid").Limit(copyBatchSize).Offset(offset)
		var n int
		if t.model == nil {
			// Join table: copy the plain column values.
			var rows []map[string]any
			if err := query.Find(&rows).Error; err != nil {
				return err
			}
			if n = len(rows); n > 0 {
				if err := dst.Table(t.name).Create(&rows).Error; err != nil {
					return err
				}
			}
		} else {
			// Model table: read into the model type, so that GORM converts
			// each column (booleans, timestamps) to the destination's types.
			// Associations are omitted, since each is copied with its own table.
			sliceType := reflect.SliceOf(reflect.TypeOf(t.model))
			rows := reflect.New(sliceType)
			if err := query.Model(t.model).Find(rows.Interface()).Error; err != nil {
				return err
			}
			if n = rows.Elem().Len(); n > 0 {
				if err := dst.Omit(clause.Associations).Create(rows.Interface()).Error; err != nil {
					return err
				}
			}
		}
		if n < copyBatchSize {
			return nil
		}
	}
}
