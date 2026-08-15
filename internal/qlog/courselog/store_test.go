package courselog

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// newTestStore returns a Store rooted at a temporary directory, along with
// the buffer its operator logger writes to, and registers its cleanup.
func newTestStore(t *testing.T) (*Store, *bytes.Buffer) {
	t.Helper()
	var operatorOutput bytes.Buffer
	operator := slog.New(slog.NewTextHandler(&operatorOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))
	store, err := NewStore(t.TempDir(), operator)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	return store, &operatorOutput
}

func TestStoreCreatesDirectoryAndFileLazily(t *testing.T) {
	store, _ := newTestStore(t)
	if _, err := store.Writer("dat520-2026").Write([]byte("line1\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	date := time.Now().UTC().Format(dateLayout)
	path := filepath.Join(store.dir, "dat520-2026", date+".jsonl")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%s) error = %v", path, err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("file mode = %v, want 0600", perm)
	}
	dirInfo, err := os.Stat(filepath.Join(store.dir, "dat520-2026"))
	if err != nil {
		t.Fatalf("Stat(course dir) error = %v", err)
	}
	if perm := dirInfo.Mode().Perm(); perm != 0o750 {
		t.Errorf("directory mode = %v, want 0750", perm)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "line1\n" {
		t.Errorf("file content = %q, want %q", got, "line1\n")
	}
}

// TestStoreRegistersCourseCreatedAfterConstruction guards the property that
// makes a course log usable without a server restart: a course the Store has
// never seen still gets a directory and file on its very first write.
func TestStoreRegistersCourseCreatedAfterConstruction(t *testing.T) {
	store, _ := newTestStore(t)
	if _, err := store.Writer("brand-new-2026").Write([]byte("hello\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(store.dir, "brand-new-2026")); err != nil {
		t.Errorf("course directory not created for a course unknown at startup: %v", err)
	}
}

func TestStorePerCourseIsolation(t *testing.T) {
	store, _ := newTestStore(t)
	if _, err := store.Writer("course-a").Write([]byte("a\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Writer("course-b").Write([]byte("b\n")); err != nil {
		t.Fatal(err)
	}

	date := time.Now().UTC().Format(dateLayout)
	gotA, err := os.ReadFile(filepath.Join(store.dir, "course-a", date+".jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	gotB, err := os.ReadFile(filepath.Join(store.dir, "course-b", date+".jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if string(gotA) != "a\n" || string(gotB) != "b\n" {
		t.Errorf("course-a = %q, course-b = %q, want isolated per-course contents", gotA, gotB)
	}
}

func TestStoreConcurrentWrites(t *testing.T) {
	store, _ := newTestStore(t)
	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		go func() {
			defer wg.Done()
			if _, err := fmt.Fprintf(store.Writer("dat520-2026"), "line-%d\n", i); err != nil {
				t.Errorf("Write() error = %v", err)
			}
		}()
	}
	wg.Wait()

	date := time.Now().UTC().Format(dateLayout)
	got, err := os.ReadFile(filepath.Join(store.dir, "dat520-2026", date+".jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(got), "\n"), "\n")
	if len(lines) != n {
		t.Errorf("got %d lines, want %d; concurrent writes may have interleaved or been lost", len(lines), n)
	}
}

func TestStoreDateRollover(t *testing.T) {
	store, _ := newTestStore(t)
	day1 := time.Date(2026, 1, 1, 23, 59, 0, 0, time.UTC)
	store.now = func() time.Time { return day1 }
	if _, err := store.Writer("dat520-2026").Write([]byte("day1\n")); err != nil {
		t.Fatal(err)
	}

	day2 := day1.Add(2 * time.Minute) // crosses into 2026-01-02 UTC
	store.now = func() time.Time { return day2 }
	if _, err := store.Writer("dat520-2026").Write([]byte("day2\n")); err != nil {
		t.Fatal(err)
	}

	file1, err := os.ReadFile(filepath.Join(store.dir, "dat520-2026", "2026-01-01.jsonl"))
	if err != nil {
		t.Fatalf("day-1 file: %v", err)
	}
	file2, err := os.ReadFile(filepath.Join(store.dir, "dat520-2026", "2026-01-02.jsonl"))
	if err != nil {
		t.Fatalf("day-2 file: %v", err)
	}
	if string(file1) != "day1\n" || string(file2) != "day2\n" {
		t.Errorf("file1 = %q, file2 = %q, want separate contents per UTC day", file1, file2)
	}
}

func TestStoreRetentionCleanup(t *testing.T) {
	store, _ := newTestStore(t)
	courseDir := filepath.Join(store.dir, "dat520-2026")
	if err := os.MkdirAll(courseDir, 0o750); err != nil {
		t.Fatal(err)
	}
	expired := filepath.Join(courseDir, "2000-01-01.jsonl")
	if err := os.WriteFile(expired, []byte("old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	recent := filepath.Join(courseDir, time.Now().UTC().Format(dateLayout)+".jsonl")
	if err := os.WriteFile(recent, []byte("new\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	notADateFile := filepath.Join(courseDir, "not-a-date.jsonl")
	if err := os.WriteFile(notADateFile, []byte("keep\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	store.cleanupExpired()

	if _, err := os.Stat(expired); !os.IsNotExist(err) {
		t.Errorf("Stat(expired) error = %v, want the expired file removed", err)
	}
	if _, err := os.Stat(recent); err != nil {
		t.Errorf("recent file was removed: %v", err)
	}
	if _, err := os.Stat(notADateFile); err != nil {
		t.Errorf("non-date file was removed: %v", err)
	}
}

func TestStoreCloseReleasesHandles(t *testing.T) {
	var operatorOutput bytes.Buffer
	operator := slog.New(slog.NewTextHandler(&operatorOutput, &slog.HandlerOptions{Level: slog.LevelDebug}))
	store, err := NewStore(t.TempDir(), operator)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if _, err := store.Writer("dat520-2026").Write([]byte("line\n")); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if cf := store.courses["dat520-2026"]; cf.file != nil {
		t.Errorf("file handle not released after Close()")
	}
}

// TestStoreErrorsReachOperatorWithoutRecursion guards that a write failure is
// reported through the plain operator logger, not one carrying the course
// marker, which would otherwise recurse back into the handler that calls it.
func TestStoreErrorsReachOperatorWithoutRecursion(t *testing.T) {
	store, operatorOutput := newTestStore(t)
	const org = "dat520-2026"
	// A regular file where the course's directory should go makes MkdirAll
	// fail regardless of the test's privileges, unlike a permission check.
	blocked := filepath.Join(store.dir, org)
	if err := os.WriteFile(blocked, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := store.Writer(org).Write([]byte("line\n")); err == nil {
		t.Fatal("Write() error = nil, want failure creating the course directory")
	}

	got := operatorOutput.String()
	if !strings.Contains(got, "course log store") {
		t.Errorf("operator log = %q, does not report the store failure", got)
	}
	if !strings.Contains(got, org) {
		t.Errorf("operator log = %q, does not identify the course", got)
	}
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"dat520-2026", "dat520-2026"},
		{"", "_"},
		{".", "_"},
		{"..", "_"},
		{"../../etc", ".._.._etc"},
		{"a/b", "a_b"},
	}
	for _, test := range tests {
		if got := sanitize(test.in); got != test.want {
			t.Errorf("sanitize(%q) = %q, want %q", test.in, got, test.want)
		}
	}
}
