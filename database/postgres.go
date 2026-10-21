package database

import (
	"errors"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// This file contains what QuickFeed needs to run on PostgreSQL in addition
// to SQLite. QuickFeed was written for SQLite, and a few things that SQLite
// accepts are rejected by PostgreSQL. Each function below handles one of them.

// preparePostgres runs before AutoMigrate on a PostgreSQL connection.
func preparePostgres(conn *gorm.DB) error {
	if err := createDatetimeType(conn); err != nil {
		return err
	}
	if err := conn.Callback().Create().Before("gorm:create").
		Register("quickfeed:conflict_target", addConflictTarget); err != nil {
		return err
	}
	if err := conn.Callback().Create().Before("gorm:create").
		Register("quickfeed:detect_explicit_id", detectExplicitID); err != nil {
		return err
	}
	return conn.Callback().Create().After("gorm:create").
		Register("quickfeed:sync_id_sequence", syncIDSequence)
}

// createDatetimeType makes the SQLite column type "datetime" available in PostgreSQL.
//
// The timestamp fields in the generated protobuf types carry the struct tag
// gorm:"type:datetime". SQLite accepts this type, but PostgreSQL has no
// "datetime" type. Instead of changing the generated code (which would
// also change existing SQLite databases), we define "datetime" as a
// PostgreSQL domain over "timestamp with time zone".
func createDatetimeType(conn *gorm.DB) error {
	return conn.Exec(`DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'datetime') THEN
		CREATE DOMAIN datetime AS timestamp with time zone;
	END IF;
END
$$;`).Error
}

// addConflictTarget adds the missing conflict columns to an upsert.
//
// When GORM saves nested records (e.g., the grades of a submission), it
// generates "INSERT ... ON CONFLICT DO UPDATE" using the table's primary key
// as the conflict target. Some tables, such as grades, have no primary key,
// only a unique index. SQLite accepts ON CONFLICT DO UPDATE without a target,
// but PostgreSQL requires one. We therefore use the table's unique index.
func addConflictTarget(db *gorm.DB) {
	stmt := db.Statement
	if stmt.Schema == nil {
		return
	}
	c, ok := stmt.Clauses["ON CONFLICT"]
	if !ok {
		return
	}
	onConflict, ok := c.Expression.(clause.OnConflict)
	if !ok || onConflict.DoNothing || len(onConflict.Columns) > 0 || onConflict.OnConstraint != "" {
		return
	}
	for _, idx := range stmt.Schema.ParseIndexes() {
		if idx.Class != "UNIQUE" {
			continue
		}
		for _, f := range idx.Fields {
			onConflict.Columns = append(onConflict.Columns, clause.Column{Name: f.DBName})
		}
		stmt.AddClause(onConflict)
		return
	}
}

const explicitIDKey = "quickfeed:explicit_id"

// detectExplicitID records whether a record is inserted with its ID already set.
//
// SQLite gives a new row the ID max(id)+1, so inserting a row with a chosen ID
// is safe. PostgreSQL instead takes the next value from a counter (a sequence),
// which does not notice IDs chosen by the caller. The next insert without an
// ID would then reuse an existing ID and fail with "duplicate key".
// syncIDSequence fixes the counter after such inserts.
func detectExplicitID(db *gorm.DB) {
	stmt := db.Statement
	if stmt.Schema == nil || stmt.Schema.PrioritizedPrimaryField == nil || !stmt.Schema.PrioritizedPrimaryField.AutoIncrement {
		return
	}
	pk := stmt.Schema.PrioritizedPrimaryField
	rv := reflect.Indirect(stmt.ReflectValue)
	switch rv.Kind() {
	case reflect.Struct:
		if _, zero := pk.ValueOf(stmt.Context, rv); !zero {
			db.InstanceSet(explicitIDKey, true)
		}
	case reflect.Slice, reflect.Array:
		for i := range rv.Len() {
			if _, zero := pk.ValueOf(stmt.Context, reflect.Indirect(rv.Index(i))); !zero {
				db.InstanceSet(explicitIDKey, true)
				return
			}
		}
	}
}

// syncIDSequence moves the table's ID sequence past the largest ID in use,
// if the insert used explicitly chosen IDs (see detectExplicitID).
func syncIDSequence(db *gorm.DB) {
	if db.Error != nil {
		return
	}
	if explicit, ok := db.InstanceGet(explicitIDKey); !ok || explicit != true {
		return
	}
	stmt := db.Statement
	table, column := stmt.Table, stmt.Schema.PrioritizedPrimaryField.DBName
	err := db.Session(&gorm.Session{NewDB: true}).Exec(
		"SELECT setval(pg_get_serial_sequence(?, ?), (SELECT COALESCE(MAX("+stmt.Quote(column)+"), 1) FROM "+stmt.Quote(table)+"))",
		table, column,
	).Error
	if err != nil {
		_ = db.AddError(err) // AddError returns the same error, now recorded on db
	}
}

// isUniqueViolation reports whether err is caused by a unique constraint,
// for both SQLite and PostgreSQL.
func isUniqueViolation(err error) bool {
	if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
		return true // SQLite
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" // PostgreSQL unique_violation
}
