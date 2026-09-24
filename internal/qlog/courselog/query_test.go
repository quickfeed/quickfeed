package courselog

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
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

// req builds a CourseLogRequest for Query, with From/To always given so
// Interval's defaulting never kicks in.
func req(from, to time.Time, opts ...func(*qf.CourseLogRequest)) *qf.CourseLogRequest {
	r := &qf.CourseLogRequest{From: qf.TimePosition(from), To: qf.TimePosition(to), Limit: 100}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func withLimit(limit uint32) func(*qf.CourseLogRequest) {
	return func(r *qf.CourseLogRequest) { r.Limit = limit }
}

func withLevel(level qf.CourseLogEntry_Level) func(*qf.CourseLogRequest) {
	return func(r *qf.CourseLogRequest) { r.Level = level }
}

func withRepository(repository string) func(*qf.CourseLogRequest) {
	return func(r *qf.CourseLogRequest) { r.Repository = repository }
}

func withNewest() func(*qf.CourseLogRequest) {
	return func(r *qf.CourseLogRequest) { r.Newest = true }
}

func TestQueryEmptyLog(t *testing.T) {
	store, _ := newTestStore(t)
	now := time.Now().UTC()
	entries, repos, truncated, err := store.Query("never-logged", req(now.Add(-time.Hour), now), nil)
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

	entries, _, _, err := store.Query("dat520-2026", req(base.Add(-time.Minute), base.Add(time.Minute)), nil)
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

	entries, _, _, err := store.Query("dat520-2026", req(base.Add(-time.Hour), base.Add(time.Hour), withLevel(qf.CourseLogEntry_WARN)), nil)
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

	entries, repos, _, err := store.Query("dat520-2026", req(base.Add(-time.Hour), base.Add(time.Hour), withRepository("repo-a")), nil)
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

func TestQueryLimit(t *testing.T) {
	store, _ := newTestStore(t)
	base := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	for i := range 5 {
		writeEntry(t, store, "dat520-2026", base.Add(time.Duration(i)*time.Minute), slog.LevelInfo, fmt.Sprintf("record %d", i))
	}

	tests := map[string]struct {
		opts          []func(*qf.CourseLogRequest)
		want          []string
		wantTruncated bool
	}{
		"oldest by default": {opts: nil, want: []string{"record 0", "record 1", "record 2"}, wantTruncated: true},
		"newest":            {opts: []func(*qf.CourseLogRequest){withNewest()}, want: []string{"record 2", "record 3", "record 4"}, wantTruncated: true},
		"exactly the limit": {opts: []func(*qf.CourseLogRequest){withLimit(5)}, want: []string{"record 0", "record 1", "record 2", "record 3", "record 4"}},
		"newest, exactly the limit": {
			opts: []func(*qf.CourseLogRequest){withLimit(5), withNewest()},
			want: []string{"record 0", "record 1", "record 2", "record 3", "record 4"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			opts := append([]func(*qf.CourseLogRequest){withLimit(3)}, tt.opts...)
			entries, _, truncated, err := store.Query("dat520-2026", req(base.Add(-time.Hour), base.Add(time.Hour), opts...), nil)
			if err != nil {
				t.Fatalf("Query() error = %v", err)
			}
			if got := messages(entries); !slices.Equal(got, tt.want) {
				t.Errorf("Query() = %v, want %v, in the order written", got, tt.want)
			}
			if truncated != tt.wantTruncated {
				t.Errorf("truncated = %v, want %v", truncated, tt.wantTruncated)
			}
		})
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

	entries, _, _, err := store.Query("dat520-2026", req(base.Add(-time.Hour), base.Add(time.Hour)), nil)
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

	_, _, _, err = store.Query("dat520-2026", req(base.Add(-time.Hour), base.Add(time.Hour)), nil)
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

	entries, _, _, err := store.Query("dat520-2026", req(day1.Add(-time.Hour), day2.Add(time.Hour)), nil)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(entries) != 2 || entries[0].Message != "day one" || entries[1].Message != "day two" {
		t.Errorf("Query() entries = %v, want both days' records in order", entries)
	}
}

// TestQueryClampsToClockAndRetention guards Query as the one place that
// bounds a request's interval: a far-future To, or a From far beyond
// Retention, must not make it walk one date file per day across that whole
// span, regardless of what a caller (such as the RPC handler) asks for.
func TestQueryClampsToClockAndRetention(t *testing.T) {
	store, _ := newTestStore(t)
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }
	writeEntry(t, store, "dat520-2026", now, slog.LevelInfo, "recent record")

	farFuture := time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
	farPast := now.Add(-10 * Retention)
	entries, _, _, err := store.Query("dat520-2026", req(farPast, farFuture), nil)
	if err != nil {
		t.Fatalf("Query() error = %v, want the far-future/far-past interval clamped rather than walked in full", err)
	}
	if len(entries) != 1 || entries[0].Message != "recent record" {
		t.Errorf("Query() entries = %v, want only the one written record", entries)
	}
}

// TestQueryEmptyWhenIntervalInvertedAfterClamp guards that clamping To down
// to now can legitimately invert the interval (an explicit From already
// after the clamped To), which must yield an empty result rather than an
// error or a backward walk.
func TestQueryEmptyWhenIntervalInvertedAfterClamp(t *testing.T) {
	store, _ := newTestStore(t)
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }

	farFuture := time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
	entries, repos, truncated, err := store.Query("dat520-2026", req(now.Add(time.Minute), farFuture), nil)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(entries) != 0 || len(repos) != 0 || truncated {
		t.Errorf("Query() = (%v, %v, %v), want empty once To is clamped to now and From is still after it", entries, repos, truncated)
	}
}

// TestQueryFromClampedToRetainedFileBoundary guards that From is clamped to
// the UTC midnight of the oldest date cleanup still guarantees to keep, not
// to the raw now-Retention instant: cleanupCourseDir removes a date file once
// its UTC midnight falls before that instant, so clamping to the instant
// itself could report a window wider than what the on-disk files can answer
// for.
func TestQueryFromClampedToRetainedFileBoundary(t *testing.T) {
	store, _ := newTestStore(t)
	// A "now" with a non-zero time-of-day, so now-Retention does not itself
	// land on a UTC midnight, exercising the boundary's round-up.
	now := time.Date(2026, 3, 10, 15, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }

	oldestKept := now.Add(-Retention).AddDate(0, 0, 1)
	oldestKept = time.Date(oldestKept.Year(), oldestKept.Month(), oldestKept.Day(), 0, 0, 0, 0, time.UTC)
	// One record just before the boundary (in a file cleanup has already
	// removed) and one just after (in the oldest file cleanup still keeps).
	writeEntry(t, store, "dat520-2026", oldestKept.Add(-time.Minute), slog.LevelInfo, "expired")
	writeEntry(t, store, "dat520-2026", oldestKept.Add(time.Minute), slog.LevelInfo, "retained")
	store.now = func() time.Time { return now } // writeEntry moved the clock; restore it for Query

	entries, _, _, err := store.Query("dat520-2026", req(now.Add(-2*Retention), now), nil)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(entries) != 1 || entries[0].Message != "retained" {
		t.Errorf("Query() entries = %v, want only the record inside the retained-file boundary", entries)
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

	entries, _, _, err := store.Query("dat520-2026", req(base.Add(-time.Minute), base.Add(time.Minute), withLimit(10)), nil)
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

// withAfter bounds the request's From at after's entry, if after is not nil.
func withAfter(after *qf.LogCursor) func(*qf.CourseLogRequest) {
	return func(r *qf.CourseLogRequest) {
		if after != nil {
			r.From = qf.CursorPosition(after)
		}
	}
}

// withBefore bounds the request's To at before's entry, if before is not nil.
func withBefore(before *qf.LogCursor) func(*qf.CourseLogRequest) {
	return func(r *qf.CourseLogRequest) {
		if before != nil {
			r.To = qf.CursorPosition(before)
		}
	}
}

// messages returns the message of each entry, in order.
func messages(entries []*qf.CourseLogEntry) []string {
	msgs := make([]string, len(entries))
	for i, e := range entries {
		msgs[i] = e.GetMessage()
	}
	return msgs
}

// cursors returns the cursor each entry carries, in order.
func cursors(t *testing.T, entries []*qf.CourseLogEntry) []*qf.LogCursor {
	t.Helper()
	cs := make([]*qf.LogCursor, len(entries))
	for i, e := range entries {
		cs[i] = cursorOf(t, e)
	}
	return cs
}

// Each entry's cursor must be the byte offset just past its line, since that
// is where reading on from it has to begin.
func TestQueryEntriesCarryTheirPosition(t *testing.T) {
	store, _ := newTestStore(t)
	base := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	for i := range 3 {
		writeEntry(t, store, testOrg, base.Add(time.Duration(i)*time.Minute), slog.LevelInfo, fmt.Sprintf("record %d", i))
	}

	entries, _, _, err := store.Query(testOrg, req(base.Add(-time.Hour), base.Add(time.Hour)), nil)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	content, err := os.ReadFile(store.path(testOrg, base))
	if err != nil {
		t.Fatal(err)
	}
	var end int64
	lines := strings.SplitAfter(strings.TrimSuffix(string(content), "\n"), "\n")
	if len(entries) != len(lines) {
		t.Fatalf("len(entries) = %d, want %d", len(entries), len(lines))
	}
	for i, line := range lines {
		end += int64(len(line))
		if i == len(lines)-1 {
			end++ // the final newline trimmed above
		}
		if got, want := cursorOf(t, entries[i]), qf.NewLogCursor(base, end); !proto.Equal(got, want) {
			t.Errorf("entries[%d] cursor = %v, want %v", i, got, want)
		}
	}
}

func TestQueryBoundedByPosition(t *testing.T) {
	store, _ := newTestStore(t)
	base := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	for i := range 5 {
		writeEntry(t, store, testOrg, base.Add(time.Duration(i)*time.Minute), slog.LevelInfo, fmt.Sprintf("record %d", i))
	}
	all, _, _, err := store.Query(testOrg, req(base.Add(-time.Hour), base.Add(time.Hour)), nil)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	c := cursors(t, all)
	startOfDay := qf.NewLogCursor(base, 0)

	tests := map[string]struct {
		after, before, until *qf.LogCursor
		want                 []string
	}{
		"unbounded":                 {want: []string{"record 0", "record 1", "record 2", "record 3", "record 4"}},
		"after an entry":            {after: c[1], want: []string{"record 2", "record 3", "record 4"}},
		"until an entry":            {until: c[2], want: []string{"record 0", "record 1", "record 2"}},
		"before an entry":           {before: c[2], want: []string{"record 0", "record 1"}},
		"between two entries":       {after: c[1], until: c[3], want: []string{"record 2", "record 3"}},
		"after one, before another": {after: c[1], before: c[3], want: []string{"record 2"}},
		"after the newest":          {after: c[4], want: nil},
		"before the oldest":         {before: c[0], want: nil},
		"after equal to until":      {after: c[2], until: c[2], want: nil},
		"after equal to before":     {after: c[2], before: c[2], want: nil},
		"before tighter than until": {before: c[2], until: c[3], want: []string{"record 0", "record 1"}},
		"until tighter than before": {before: c[3], until: c[1], want: []string{"record 0", "record 1"}},
		"before equal to until":     {before: c[2], until: c[2], want: []string{"record 0", "record 1"}},
		"after the start of a day":  {after: startOfDay, want: []string{"record 0", "record 1", "record 2", "record 3", "record 4"}},
		"until the start of a day":  {until: startOfDay, want: nil},
		"before the start of a day": {before: startOfDay, want: nil},
		"until an earlier day":      {until: qf.NewLogCursor(base.AddDate(0, 0, -1), 10), want: nil},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			entries, _, _, err := store.Query(testOrg, req(base.Add(-time.Hour), base.Add(time.Hour), withAfter(tt.after), withBefore(tt.before)), tt.until)
			if err != nil {
				t.Fatalf("Query() error = %v", err)
			}
			if got := messages(entries); !slices.Equal(got, tt.want) {
				t.Errorf("Query() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Over the limit, the newest entries before a cursor are the page that
// continues a view back from its oldest entry, and the oldest after a cursor
// the page that continues it forward from its newest.
func TestQueryPagesByCursor(t *testing.T) {
	store, _ := newTestStore(t)
	base := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	for i := range 6 {
		writeEntry(t, store, testOrg, base.Add(time.Duration(i)*time.Minute), slog.LevelInfo, fmt.Sprintf("record %d", i))
	}
	all, _, _, err := store.Query(testOrg, req(base.Add(-time.Hour), base.Add(time.Hour)), nil)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	c := cursors(t, all)

	tests := map[string]struct {
		opts          []func(*qf.CourseLogRequest)
		want          []string
		wantTruncated bool
	}{
		"back from the oldest on screen": {
			opts:          []func(*qf.CourseLogRequest){withBefore(c[4]), withNewest()},
			want:          []string{"record 2", "record 3"},
			wantTruncated: true,
		},
		"forward from the newest on screen": {
			opts:          []func(*qf.CourseLogRequest){withAfter(c[1])},
			want:          []string{"record 2", "record 3"},
			wantTruncated: true,
		},
		"forward to the end": {
			opts: []func(*qf.CourseLogRequest){withAfter(c[3])},
			want: []string{"record 4", "record 5"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			opts := append([]func(*qf.CourseLogRequest){withLimit(2)}, tt.opts...)
			entries, _, truncated, err := store.Query(testOrg, req(base.Add(-time.Hour), base.Add(time.Hour), opts...), nil)
			if err != nil {
				t.Fatalf("Query() error = %v", err)
			}
			if got := messages(entries); !slices.Equal(got, tt.want) {
				t.Errorf("Query() = %v, want %v", got, tt.want)
			}
			if truncated != tt.wantTruncated {
				t.Errorf("truncated = %v, want %v", truncated, tt.wantTruncated)
			}
		})
	}
}

// A cursor bounds one end of the range; a time bounding the other still
// applies.
func TestQueryCursorAndTimeBoundsBothApply(t *testing.T) {
	store, _ := newTestStore(t)
	base := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	writeEntry(t, store, testOrg, base, slog.LevelInfo, "first")
	writeEntry(t, store, testOrg, base.Add(time.Hour), slog.LevelInfo, "outside the interval")
	writeEntry(t, store, testOrg, base.Add(time.Minute), slog.LevelInfo, "inside the interval")
	first, _, _, err := store.Query(testOrg, req(base, base), nil)
	if err != nil || len(first) != 1 {
		t.Fatalf("Query() = %v, %v, want the first record", first, err)
	}

	entries, _, _, err := store.Query(testOrg, req(base, base.Add(30*time.Minute), withAfter(cursorOf(t, first[0]))), nil)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if got, want := messages(entries), []string{"inside the interval"}; !slices.Equal(got, want) {
		t.Errorf("Query() = %v, want %v", got, want)
	}
}

// A cursor stays valid across a restart: the files are only appended to, and
// a new Store appends where the old one left off.
func TestQueryAfterSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	// Within retention, or the second store's startup sweep removes the file.
	today := time.Now().UTC()
	base := time.Date(today.Year(), today.Month(), today.Day(), 12, 0, 0, 0, time.UTC)
	open := func() *Store {
		store, err := NewStore(dir, slog.New(slog.DiscardHandler))
		if err != nil {
			t.Fatalf("NewStore() error = %v", err)
		}
		t.Cleanup(func() { _ = store.Close() })
		return store
	}

	before := open()
	writeEntry(t, before, testOrg, base, slog.LevelInfo, "before the restart")
	seen, _, _, err := before.Query(testOrg, req(base.Add(-time.Hour), base.Add(time.Hour)), nil)
	if err != nil || len(seen) != 1 {
		t.Fatalf("Query() = %v, %v, want one record", seen, err)
	}
	if err := before.Close(); err != nil {
		t.Fatal(err)
	}

	after := open()
	writeEntry(t, after, testOrg, base.Add(time.Minute), slog.LevelInfo, "after the restart")
	entries, _, _, err := after.Query(testOrg, req(base.Add(-time.Hour), base.Add(time.Hour), withAfter(cursorOf(t, seen[0]))), nil)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if got, want := messages(entries), []string{"after the restart"}; !slices.Equal(got, want) {
		t.Fatalf("Query() = %v, want %v", got, want)
	}
	if got, want := cursorOf(t, entries[0]).Position(), fileSize(t, after.path(testOrg, base)); got != want {
		t.Errorf("cursor offset = %d, want %d: the new store must append where the old one left off", got, want)
	}
}

// A cursor in one date file bounds the files around it: everything in a later
// file is after it, and nothing in an earlier one is.
func TestQueryPositionAcrossDateRollover(t *testing.T) {
	store, _ := newTestStore(t)
	day1 := time.Date(2026, 3, 10, 23, 58, 0, 0, time.UTC)
	writeEntry(t, store, testOrg, day1, slog.LevelInfo, "day one, first")
	writeEntry(t, store, testOrg, day1.Add(time.Minute), slog.LevelInfo, "day one, second")
	writeEntry(t, store, testOrg, day1.Add(2*time.Minute), slog.LevelInfo, "day two")
	interval := req(day1.Add(-time.Hour), day1.Add(time.Hour))

	all, _, _, err := store.Query(testOrg, interval, nil)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	c := cursors(t, all)
	firstDay := qf.LogDay(day1)
	nextDay := firstDay.AddDate(0, 0, 1)
	if !c[1].Day().Equal(firstDay) || !c[2].Day().Equal(nextDay) {
		t.Fatalf("cursors = %v, want the third record in the next day's file", c)
	}
	// A new day's file starts from nothing, so its first record's position
	// is its own length, not a continuation of the previous day's.
	if want := fileSize(t, store.path(testOrg, nextDay)); c[2].Position() != want {
		t.Errorf("day-two cursor = %v, want offset %d, the size of that day's file", c[2], want)
	}

	tests := map[string]struct {
		after, until *qf.LogCursor
		want         []string
	}{
		"after day one's first": {c[0], nil, []string{"day one, second", "day two"}},
		"after day one's last":  {c[1], nil, []string{"day two"}},
		"until day one's last":  {nil, c[1], []string{"day one, first", "day one, second"}},
		"across the rollover":   {c[0], c[2], []string{"day one, second", "day two"}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			bounded := req(day1.Add(-time.Hour), day1.Add(time.Hour), withAfter(tt.after))
			entries, _, _, err := store.Query(testOrg, bounded, tt.until)
			if err != nil {
				t.Fatalf("Query() error = %v", err)
			}
			if got := messages(entries); !slices.Equal(got, tt.want) {
				t.Errorf("Query() = %v, want %v", got, tt.want)
			}
		})
	}
	for name, tt := range map[string]struct {
		before *qf.LogCursor
		want   []string
	}{
		"before day two's first": {c[2], []string{"day one, first", "day one, second"}},
		"before day one's last":  {c[1], []string{"day one, first"}},
	} {
		t.Run(name, func(t *testing.T) {
			entries, _, _, err := store.Query(testOrg, req(day1.Add(-time.Hour), day1.Add(time.Hour), withBefore(tt.before)), nil)
			if err != nil {
				t.Fatalf("Query() error = %v", err)
			}
			if got := messages(entries); !slices.Equal(got, tt.want) {
				t.Errorf("Query() = %v, want %v", got, tt.want)
			}
		})
	}
}

// A cursor the store did not hand out must be rejected as such, not read on
// from: an offset inside a record would otherwise surface as a malformed line
// and fail the query as if the log were corrupt.
func TestQueryRejectsInvalidCursor(t *testing.T) {
	store, _ := newTestStore(t)
	base := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	writeEntry(t, store, testOrg, base, slog.LevelInfo, "a record")
	size := fileSize(t, store.path(testOrg, base))
	at := qf.NewLogCursor
	yesterday, tomorrow := base.AddDate(0, 0, -1), base.AddDate(0, 0, 1)

	tests := map[string]struct {
		cursor  *qf.LogCursor
		until   *qf.LogCursor
		wantErr bool
	}{
		// The validation interceptor rejects these first, but Query is not
		// always called behind it.
		"no date":                {cursor: &qf.LogCursor{Offset: 3}, wantErr: true},
		"a date off midnight":    {cursor: &qf.LogCursor{Date: timestamppb.New(base)}, wantErr: true},
		"inside a record":        {cursor: at(base, 3), wantErr: true},
		"beyond the end":         {cursor: at(base, size+10), wantErr: true},
		"at the end":             {cursor: at(base, size)},
		"a day with no log file": {cursor: at(yesterday, 42)},
		"a day yet to come":      {cursor: at(tomorrow, 0), wantErr: true},
	}
	for name, tt := range tests {
		for side, bound := range map[string]func(*qf.LogCursor) func(*qf.CourseLogRequest){"from": withAfter, "to": withBefore} {
			t.Run(side+" "+name, func(t *testing.T) {
				r := req(base.Add(-48*time.Hour), base.Add(time.Hour), bound(tt.cursor))
				_, _, _, err := store.Query(testOrg, r, nil)
				if tt.wantErr != errors.Is(err, ErrInvalidCursor) {
					t.Errorf("Query(%s: %v) error = %v, want ErrInvalidCursor: %v", side, tt.cursor, err, tt.wantErr)
				}
				if !tt.wantErr && err != nil {
					t.Errorf("Query(%s: %v) error = %v, want nil", side, tt.cursor, err)
				}
			})
		}
	}

	// Until is where a subscription began, so a From cursor beyond it was not
	// handed out by the time the query ran.
	for name, tt := range map[string]struct {
		after, until *qf.LogCursor
		wantErr      bool
	}{
		"at until":           {after: at(base, size), until: at(base, size)},
		"beyond until":       {after: at(base, size), until: at(base, size-1), wantErr: true},
		"a day beyond until": {after: at(base, 0), until: at(yesterday, 42), wantErr: true},
	} {
		t.Run(name, func(t *testing.T) {
			_, _, _, err := store.Query(testOrg, req(base.Add(-48*time.Hour), base.Add(time.Hour), withAfter(tt.after)), tt.until)
			if tt.wantErr != errors.Is(err, ErrInvalidCursor) {
				t.Errorf("Query(From: %v, until %v) error = %v, want ErrInvalidCursor: %v", tt.after, tt.until, err, tt.wantErr)
			}
		})
	}
}

// A record is filed under the date it is written, not the one it is stamped
// with, so a record stamped just before midnight can land in the next day's
// file. A query whose interval ends before midnight must still find it.
func TestQueryFindsRecordFiledAfterMidnight(t *testing.T) {
	store, _ := newTestStore(t)
	stamped := time.Date(2026, 3, 10, 23, 59, 59, 999_999_999, time.UTC)
	written := stamped.Add(2 * time.Nanosecond)
	h := NewHandler(store).WithAttrs([]slog.Attr{slog.String(label.CourseLog, testOrg)})
	store.now = func() time.Time { return written }
	if err := h.Handle(context.Background(), slog.NewRecord(stamped, slog.LevelInfo, "late", 0)); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if _, err := os.Stat(store.path(testOrg, written)); err != nil {
		t.Fatalf("record not filed under the day it was written: %v", err)
	}

	for name, now := range map[string]time.Time{
		"the next day is today": written.Add(time.Hour),
		"the next day is past":  written.Add(72 * time.Hour),
	} {
		t.Run(name, func(t *testing.T) {
			store.now = func() time.Time { return now }
			entries, _, _, err := store.Query(testOrg, req(stamped.Add(-time.Hour), stamped), nil)
			if err != nil {
				t.Fatalf("Query() error = %v", err)
			}
			if got, want := messages(entries), []string{"late"}; !slices.Equal(got, want) {
				t.Errorf("Query() = %v, want %v", got, want)
			}
		})
	}
}
