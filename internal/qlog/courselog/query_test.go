package courselog

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog/label"
)

// writeEntry logs one record directly through the Handler, bypassing
// slog.Logger so the record's timestamp is exactly at rather than whatever
// time.Now() happens to be when the test runs. It also points the store's
// clock at at, so the entry lands in the date file a real write at that
// moment would have used.
func writeEntry(t *testing.T, store *Store, org string, at time.Time, level slog.Level, msg string, attrs ...slog.Attr) {
	t.Helper()
	store.now = func() time.Time { return at }
	h := NewHandler(store).WithAttrs([]slog.Attr{
		slog.Uint64(label.CourseID, 1),
		slog.String(label.CourseCode, "DAT520"),
		slog.String(label.CourseLog, org),
	})
	r := slog.NewRecord(at, level, msg, 0)
	r.AddAttrs(attrs...)
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
}

func TestQueryEmptyLog(t *testing.T) {
	store, _ := newTestStore(t)
	now := time.Now().UTC()
	entries, repos, truncated, err := store.Query("never-logged", Query{
		From: now.Add(-time.Hour), To: now, Limit: 100,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(entries) != 0 || len(repos) != 0 || truncated {
		t.Errorf("Query() = (%v, %v, %v), want empty result for a course with no log", entries, repos, truncated)
	}
}

func TestQueryFiltersByTime(t *testing.T) {
	store, _ := newTestStore(t)
	base := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	writeEntry(t, store, "dat520-2026", base, slog.LevelInfo, "in range")
	writeEntry(t, store, "dat520-2026", base.Add(-2*time.Hour), slog.LevelInfo, "before range")
	writeEntry(t, store, "dat520-2026", base.Add(2*time.Hour), slog.LevelInfo, "after range")

	entries, _, _, err := store.Query("dat520-2026", Query{
		From: base.Add(-time.Minute), To: base.Add(time.Minute), Limit: 100,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(entries) != 1 || entries[0].Message != "in range" {
		t.Errorf("Query() entries = %v, want only the in-range record", entries)
	}
}

func TestQueryFiltersByLevel(t *testing.T) {
	store, _ := newTestStore(t)
	base := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	writeEntry(t, store, "dat520-2026", base, slog.LevelDebug, "debug record")
	writeEntry(t, store, "dat520-2026", base.Add(time.Minute), slog.LevelError, "error record")

	entries, _, _, err := store.Query("dat520-2026", Query{
		From: base.Add(-time.Hour), To: base.Add(time.Hour), Level: slog.LevelWarn, Limit: 100,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(entries) != 1 || entries[0].Message != "error record" {
		t.Errorf("Query() entries = %v, want only the record at or above the minimum level", entries)
	}
}

// TestQueryRepositoryFilterKeepsFullRepositoryList guards the design choice
// that the repository dropdown a teacher sees should not shrink to just the
// repository they already selected: the returned repository list reflects
// the time window, not the repository filter.
func TestQueryRepositoryFilterKeepsFullRepositoryList(t *testing.T) {
	store, _ := newTestStore(t)
	base := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	writeEntry(t, store, "dat520-2026", base, slog.LevelInfo, "repo a", slog.String(label.Repository, "repo-a"))
	writeEntry(t, store, "dat520-2026", base.Add(time.Minute), slog.LevelInfo, "repo b", slog.String(label.Repository, "repo-b"))

	entries, repos, _, err := store.Query("dat520-2026", Query{
		From: base.Add(-time.Hour), To: base.Add(time.Hour), Repository: "repo-a", Limit: 100,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(entries) != 1 || entries[0].Repository != "repo-a" {
		t.Errorf("Query() entries = %v, want only repo-a's record", entries)
	}
	want := []string{"repo-a", "repo-b"}
	if len(repos) != len(want) || repos[0] != want[0] || repos[1] != want[1] {
		t.Errorf("Query() repositories = %v, want %v regardless of the repository filter", repos, want)
	}
}

func TestQueryLimitKeepsNewestInChronologicalOrder(t *testing.T) {
	store, _ := newTestStore(t)
	base := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	for i := range 5 {
		writeEntry(t, store, "dat520-2026", base.Add(time.Duration(i)*time.Minute), slog.LevelInfo, "record")
	}

	entries, _, truncated, err := store.Query("dat520-2026", Query{
		From: base.Add(-time.Hour), To: base.Add(time.Hour), Limit: 3,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if !truncated {
		t.Error("truncated = false, want true when the match count exceeds Limit")
	}
	if len(entries) != 3 {
		t.Fatalf("len(entries) = %d, want 3", len(entries))
	}
	for i, want := range []time.Time{base.Add(2 * time.Minute), base.Add(3 * time.Minute), base.Add(4 * time.Minute)} {
		if !entries[i].Time.Equal(want) {
			t.Errorf("entries[%d].Time = %v, want %v (the newest 3, oldest first)", i, entries[i].Time, want)
		}
	}
}

func TestQueryIgnoresMalformedFinalLine(t *testing.T) {
	store, _ := newTestStore(t)
	base := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	writeEntry(t, store, "dat520-2026", base, slog.LevelInfo, "valid record")

	path := filepath.Join(store.dir, "dat520-2026", "2026-03-10.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`{"time":"2026-03-10T12:01:00Z","level":"INFO","msg":"cut off b`); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	entries, _, _, err := store.Query("dat520-2026", Query{
		From: base.Add(-time.Hour), To: base.Add(time.Hour), Limit: 100,
	})
	if err != nil {
		t.Fatalf("Query() error = %v, want the malformed final line ignored", err)
	}
	if len(entries) != 1 || entries[0].Message != "valid record" {
		t.Errorf("Query() entries = %v, want only the well-formed record", entries)
	}
}

func TestQuerySurfacesMalformedNonFinalLine(t *testing.T) {
	store, _ := newTestStore(t)
	base := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	writeEntry(t, store, "dat520-2026", base, slog.LevelInfo, "first record")

	path := filepath.Join(store.dir, "dat520-2026", "2026-03-10.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("not json at all\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	writeEntry(t, store, "dat520-2026", base.Add(time.Minute), slog.LevelInfo, "third record")

	_, _, _, err = store.Query("dat520-2026", Query{
		From: base.Add(-time.Hour), To: base.Add(time.Hour), Limit: 100,
	})
	if err == nil {
		t.Fatal("Query() error = nil, want a failure for a malformed line that is not the file's last")
	}
}

func TestQuerySpansMultipleDays(t *testing.T) {
	store, _ := newTestStore(t)
	day1 := time.Date(2026, 3, 10, 23, 0, 0, 0, time.UTC)
	day2 := day1.Add(2 * time.Hour) // 2026-03-11 UTC
	writeEntry(t, store, "dat520-2026", day1, slog.LevelInfo, "day one")
	writeEntry(t, store, "dat520-2026", day2, slog.LevelInfo, "day two")

	entries, _, _, err := store.Query("dat520-2026", Query{
		From: day1.Add(-time.Hour), To: day2.Add(time.Hour), Limit: 100,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(entries) != 2 || entries[0].Message != "day one" || entries[1].Message != "day two" {
		t.Errorf("Query() entries = %v, want both days' records in order", entries)
	}
}

func TestQueryDecodesDedicatedAndGenericAttributes(t *testing.T) {
	store, _ := newTestStore(t)
	base := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	writeEntry(t, store, "dat520-2026", base, slog.LevelInfo, "docker build output",
		slog.String(label.Repository, "student-repo"),
		slog.String(label.RepositoryType, "USER"),
		slog.Bool(label.Truncated, true),
		slog.String(label.Assignment, "lab1"),
		slog.Uint64(label.SubmissionID, 7),
	)

	entries, _, _, err := store.Query("dat520-2026", Query{From: base.Add(-time.Minute), To: base.Add(time.Minute), Limit: 10})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	e := entries[0]
	if e.Repository != "student-repo" || e.RepositoryType != "USER" || !e.Truncated {
		t.Errorf("entry = %+v, want dedicated fields decoded from their labels", e)
	}
	if e.Fields[label.Assignment] != "lab1" {
		t.Errorf("Fields[%q] = %q, want %q", label.Assignment, e.Fields[label.Assignment], "lab1")
	}
	if e.Fields[label.SubmissionID] != "7" {
		t.Errorf("Fields[%q] = %q, want %q", label.SubmissionID, e.Fields[label.SubmissionID], "7")
	}
	if _, ok := e.Fields[label.CourseID]; ok {
		t.Errorf("Fields carries %q, want it excluded: a query is already scoped to one course", label.CourseID)
	}
	if _, ok := e.Fields[label.CourseCode]; ok {
		t.Errorf("Fields carries %q, want it excluded: a query is already scoped to one course", label.CourseCode)
	}
}
