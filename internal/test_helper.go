package internal

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/log"
)

// TestDB returns a test database and close function.
// This function should only be used as a test helper.
func TestDB(t *testing.T) (database.Database, func()) {
	t.Helper()

	f, err := ioutil.TempFile(t.TempDir(), "test.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	db, err := database.NewGormDB(f.Name(), log.Zap(true))
	if err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	return db, func() {
		if err := os.Remove(f.Name()); err != nil {
			t.Error(err)
		}
	}
}
