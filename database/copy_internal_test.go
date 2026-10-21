package database

import (
	"database/sql"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// TestCopyTo copies the anonymized test database (doc/dev/test.db) and checks
// that every row in every table arrives unchanged.
//
// The destination is a new SQLite file, or, if $QUICKFEED_TEST_DATABASE_URL
// is set, a new PostgreSQL database on that server:
//
//	QUICKFEED_TEST_DATABASE_URL='postgres://postgres@localhost:5432/postgres?sslmode=disable' \
//	  go test ./database -run TestCopyTo
func TestCopyTo(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	srcPath := filepath.Join(t.TempDir(), "qf.db")
	copyTestFile(t, filepath.Join("..", "doc", "dev", "test.db"), srcPath)
	src, err := NewGormDB(srcPath, logger)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeDB(t, src) })

	dst, err := NewGormDB(testDestination(t), logger)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeDB(t, dst) })

	counts, err := src.CopyTo(dst)
	if err != nil {
		t.Fatal(err)
	}
	var total int64
	for _, c := range counts {
		if c.Source != c.Dest {
			t.Errorf("table %s: %d rows in source, %d in destination", c.Table, c.Source, c.Dest)
		}
		total += c.Source
	}
	if total == 0 {
		t.Fatal("test database is empty; nothing was tested")
	}

	// Compare the content of every row.
	tables, err := src.tables()
	if err != nil {
		t.Fatal(err)
	}
	for _, tbl := range tables {
		want, got := readAll(t, src.conn, tbl), readAll(t, dst.conn, tbl)
		if diff := cmp.Diff(want, got, protocmp.Transform(), microseconds); diff != "" {
			t.Errorf("table %s differs after copy (-source +destination):\n%s", tbl.name, diff)
		}
	}

	// Copying again must be refused, since the destination is no longer empty.
	if _, err := src.CopyTo(dst); err == nil || !strings.Contains(err.Error(), "not empty") {
		t.Errorf("second CopyTo: got error %v, want 'not empty'", err)
	}

	// New rows must get new IDs (the ID sequences must be past the copied IDs).
	newCourse := &qf.Course{Code: "COPY101", Year: 2099, ScmOrganizationID: 987654321, ScmOrganizationName: "copy-test"}
	if err := dst.CreateCourse(1, newCourse); err != nil {
		t.Errorf("creating a course after the copy: %v", err)
	}
}

// microseconds compares timestamps to the microsecond, the precision of PostgreSQL.
var microseconds = protocmp.FilterMessage(&timestamppb.Timestamp{},
	cmp.Comparer(func(a, b protocmp.Message) bool {
		micros := func(m protocmp.Message) (int64, int32) {
			secs, _ := m["seconds"].(int64)
			nanos, _ := m["nanos"].(int32)
			return secs, nanos / 1000
		}
		as, an := micros(a)
		bs, bn := micros(b)
		return as == bs && an == bn
	}),
)

// readAll returns all rows of the table, in primary key order.
func readAll(t *testing.T, conn *gorm.DB, tbl table) any {
	t.Helper()
	stmt := &gorm.Statement{DB: conn}
	order := "1, 2" // tables without a primary key: order by the first columns
	if tbl.model != nil {
		if err := stmt.Parse(tbl.model); err != nil {
			t.Fatal(err)
		}
		if pk := stmt.Schema.PrimaryFieldDBNames; len(pk) > 0 {
			order = strings.Join(pk, ", ")
		}
	}
	query := conn.Table(tbl.name).Order(order)
	if tbl.model == nil {
		var rows []map[string]any
		if err := query.Find(&rows).Error; err != nil {
			t.Fatal(err)
		}
		// SQLite and PostgreSQL return integers as different Go types.
		for _, row := range rows {
			for k, v := range row {
				row[k] = reflect.ValueOf(v).Convert(reflect.TypeFor[int64]()).Interface()
			}
		}
		return rows
	}
	rows := reflect.New(reflect.SliceOf(reflect.TypeOf(tbl.model)))
	if err := query.Model(tbl.model).Find(rows.Interface()).Error; err != nil {
		t.Fatal(err)
	}
	return rows.Elem().Interface()
}

// testDestination returns a new, empty destination database for the copy.
func testDestination(t *testing.T) string {
	t.Helper()
	adminURL := os.Getenv("QUICKFEED_TEST_DATABASE_URL")
	if adminURL == "" {
		return filepath.Join(t.TempDir(), "copy.db")
	}
	admin, err := sql.Open("pgx", adminURL)
	if err != nil {
		t.Fatal(err)
	}
	const name = "qftest_copy"
	if _, err := admin.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)"); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)"); err != nil {
			t.Error(err)
		}
		if err := admin.Close(); err != nil {
			t.Error(err)
		}
	})
	u, err := url.Parse(adminURL)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name
	return u.String()
}

func closeDB(t *testing.T, db *GormDB) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Error(err)
	}
}

func copyTestFile(t *testing.T, from, to string) {
	t.Helper()
	data, err := os.ReadFile(from)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(to, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
