package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/quickfeed/quickfeed/database/migrations"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/migrate"
	"go.uber.org/zap"
)

// BunDB wraps a bun.DB connection for database operations.
type BunDB struct {
	conn *bun.DB
}

// NewBunDB creates a new BunDB instance with SQLite driver and runs migrations.
func NewBunDB(path string, _ *zap.Logger) (*BunDB, error) {
	sqldb, err := sql.Open(sqliteshim.ShimName, path)
	if err != nil {
		return nil, err
	}
	conn := bun.NewDB(sqldb, sqlitedialect.New())

	// Register join model required for many2many group–user relation.
	conn.RegisterModel((*groupUserJoin)(nil))

	migrator := migrate.NewMigrator(conn, migrations.Migrations)
	if err := migrator.Init(context.Background()); err != nil {
		return nil, fmt.Errorf("initializing migrator: %w", err)
	}
	if _, err := migrator.Migrate(context.Background()); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return &BunDB{conn: conn}, nil
}

// Close closes the database connection.
func (db *BunDB) Close() error {
	return db.conn.Close()
}
