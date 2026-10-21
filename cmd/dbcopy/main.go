// Command dbcopy copies a QuickFeed SQLite database into a PostgreSQL database.
//
// It is used once, when moving an existing QuickFeed installation from SQLite
// to PostgreSQL. Stop QuickFeed first, so that the SQLite file does not change
// during the copy. The original SQLite file is never modified: dbcopy works
// on a temporary copy of it.
//
// Usage:
//
//	QUICKFEED_DATABASE_URL='postgres://user:pass@host:5432/quickfeed?sslmode=disable' \
//	  dbcopy -from /data/qf.db
//
// The destination database must exist and be empty. dbcopy creates the tables,
// copies every row with its original ID, and prints the number of rows in each
// table in both databases. It exits with an error if any count differs.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/env"
)

func main() {
	from := flag.String("from", "/data/qf.db", "path to the SQLite database file to copy from")
	to := flag.String("to", "", "PostgreSQL URL to copy to (default: $QUICKFEED_DATABASE_URL)")
	flag.Parse()

	if *to == "" {
		*to = env.DatabaseURL()
	}
	if err := run(*from, *to); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Copy complete: all row counts match.")
}

func run(from, to string) (err error) {
	if !database.IsPostgres(to) {
		return errors.New("destination must be a PostgreSQL URL (postgres://...); set -to or QUICKFEED_DATABASE_URL")
	}
	if _, err := os.Stat(from); err != nil {
		return fmt.Errorf("source database: %w", err)
	}

	// Work on a copy, so that the original file is never changed
	// (opening a database may upgrade its tables to the current schema).
	tmpDir, err := os.MkdirTemp("", "dbcopy")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(tmpDir)) }()
	srcPath := filepath.Join(tmpDir, "qf.db")
	if err := copySQLiteFile(from, srcPath); err != nil {
		return fmt.Errorf("copying source database: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	src, err := database.NewGormDB(srcPath, logger)
	if err != nil {
		return fmt.Errorf("opening source database: %w", err)
	}
	defer func() { err = errors.Join(err, src.Close()) }()
	dst, err := database.NewGormDB(to, logger)
	if err != nil {
		return fmt.Errorf("opening destination database: %w", err)
	}
	defer func() { err = errors.Join(err, dst.Close()) }()

	counts, err := src.CopyTo(dst)
	if err != nil {
		return fmt.Errorf("copy failed, nothing was written: %w", err)
	}

	mismatch := false
	fmt.Printf("%-22s %10s %10s\n", "TABLE", "SQLITE", "POSTGRES")
	for _, c := range counts {
		mark := ""
		if c.Source != c.Dest {
			mark, mismatch = "  <-- MISMATCH", true
		}
		fmt.Printf("%-22s %10d %10d%s\n", c.Table, c.Source, c.Dest, mark)
	}
	if mismatch {
		return errors.New("row counts differ; do not switch QuickFeed to PostgreSQL")
	}
	return nil
}

// copySQLiteFile copies the database file and, if present, its write-ahead
// log (-wal), which may hold changes not yet written to the main file.
func copySQLiteFile(from, to string) error {
	if err := copyFile(from, to); err != nil {
		return err
	}
	if _, err := os.Stat(from + "-wal"); err == nil {
		return copyFile(from+"-wal", to+"-wal")
	}
	return nil
}

func copyFile(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }() // read-only; errors on close do not matter
	out, err := os.Create(to)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		return errors.Join(err, out.Close())
	}
	return out.Close()
}
