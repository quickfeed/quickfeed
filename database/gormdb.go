package database

import (
	"log/slog"

	"github.com/quickfeed/quickfeed/kit/score"
	"github.com/quickfeed/quickfeed/qf"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// GormDB implements the Database interface.
type GormDB struct {
	conn *gorm.DB
}

// NewGormDB creates a new gorm database using the provided driver.
func NewGormDB(path string, logger *slog.Logger) (*GormDB, error) {
	// We are conservative and use transactions for create/update/delete operations.
	conn, err := gorm.Open(sqlite.Open(path), &gorm.Config{ // skipcq: GO-W1004
		Logger:                 NewGORMLogger(logger),
		SkipDefaultTransaction: false,
	})
	if err != nil {
		return nil, err
	}

	schema.RegisterSerializer("timestamp", &TimestampSerializer{})

	if err := conn.AutoMigrate(
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
	); err != nil {
		return nil, err
	}

	return &GormDB{conn}, nil
}

func (db *GormDB) Close() error {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
