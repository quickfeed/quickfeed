package database

import (
	"log/slog"
	"strings"

	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// models are the types stored in the database, one table each.
var models = []any{
	&qf.User{},
	&qf.Course{},
	&qf.Enrollment{},
	&qf.Assignment{},
	&qf.Submission{},
	&qf.Grade{},
	&qf.Group{},
	&qf.Repository{},
	&qf.UsedSlipDays{},
	&qf.GradingBenchmark{},
	&qf.TestInfo{},
	&qf.GradingCriterion{},
	&qf.Review{},
	&qf.Note{},
	&qf.AssignmentFeedback{},
	&qf.FeedbackReceipt{},
	&score.BuildInfo{},
	&score.Score{},
}

// GormDB implements the Database interface.
type GormDB struct {
	conn *gorm.DB
}

// NewGormDB creates a new gorm database.
//
// The source selects the database driver:
//   - a PostgreSQL connection URL (postgres://user:pass@host:5432/dbname?sslmode=disable)
//     connects to a PostgreSQL server;
//   - anything else is treated as the path to an SQLite database file.
func NewGormDB(source string, logger *slog.Logger) (*GormDB, error) {
	// We are conservative and use transactions for create/update/delete operations.
	conn, err := gorm.Open(dialector(source), &gorm.Config{ // skipcq: GO-W1004
		Logger:                 NewGORMLogger(logger),
		SkipDefaultTransaction: false,
		// QuickFeed uses 0 (not NULL) for unset references, e.g., an enrollment
		// without a group has group_id = 0. SQLite does not enforce foreign keys
		// unless enabled, so QuickFeed has never relied on them. PostgreSQL always
		// enforces them, so we do not create them, to keep the same behavior.
		DisableForeignKeyConstraintWhenMigrating: IsPostgres(source),
	})
	if err != nil {
		return nil, err
	}

	schema.RegisterSerializer("timestamp", &TimestampSerializer{})

	if IsPostgres(source) {
		if err := preparePostgres(conn); err != nil {
			return nil, err
		}
	}

	if err := conn.AutoMigrate(models...); err != nil {
		return nil, err
	}

	if err := dropRetired(conn); err != nil {
		return nil, err
	}

	return &GormDB{conn}, nil
}

// IsPostgres reports whether source is a PostgreSQL connection URL.
func IsPostgres(source string) bool {
	return strings.HasPrefix(source, "postgres://") || strings.HasPrefix(source, "postgresql://")
}

// dialector returns the GORM driver for the given database source.
func dialector(source string) gorm.Dialector {
	if IsPostgres(source) {
		return postgres.Open(source)
	}
	return sqlite.Open(source)
}

func (db *GormDB) Close() error {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
