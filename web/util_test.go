package web_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/autograde/aguis/database"
)

func setup(t *testing.T) (*database.GormDB, func()) {
	t.Helper()
	const (
		driver = "sqlite3"
		prefix = "testdb"
	)

	f, err := ioutil.TempFile(os.TempDir(), prefix)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	db, err := database.NewGormDB(driver, f.Name(), database.NewGormLogger())
	if err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	return db, func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
		if err := os.Remove(f.Name()); err != nil {
			t.Error(err)
		}
	}
}

func assertCode(t *testing.T, haveCode, wantCode int) {
	t.Helper()
	if haveCode != wantCode {
		t.Errorf("have status code %d want %d", haveCode, wantCode)
	}
}
