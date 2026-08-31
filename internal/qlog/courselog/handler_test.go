package courselog

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog/label"
)

func TestHandlerDropsRecordsWithoutMarker(t *testing.T) {
	store, _ := newTestStore(t)
	logger := slog.New(NewHandler(store))
	logger.Info("no course scope")
	logger.With(label.CourseID, uint64(1)).Info("course id but no marker")

	entries, err := os.ReadDir(store.dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("store directory has entries %v, want none: unmarked records must be dropped", entries)
	}
}

func TestHandlerRoutesMarkedRecords(t *testing.T) {
	store, _ := newTestStore(t)
	logger := slog.New(NewHandler(store))
	scoped := logger.With(label.CourseID, uint64(42), label.CourseCode, "DAT520", label.CourseLog, "dat520-2026")
	scoped.Info("hello", "extra", "value")

	entry := readSingleEntry(t, store, "dat520-2026")
	if entry["msg"] != "hello" {
		t.Errorf("msg = %v, want %q", entry["msg"], "hello")
	}
	if entry["extra"] != "value" {
		t.Errorf("extra = %v, want %q", entry["extra"], "value")
	}
	if entry[label.CourseID] != float64(42) {
		t.Errorf("course_id = %v, want 42", entry[label.CourseID])
	}
	if entry[label.CourseCode] != "DAT520" {
		t.Errorf("course_code = %v, want DAT520", entry[label.CourseCode])
	}
	if _, ok := entry[label.CourseLog]; ok {
		t.Errorf("course log entry carries the internal %q marker, want it stripped", label.CourseLog)
	}
}

// TestHandlerPendingAttrsReplayedOnceMarkerAppears guards that an attribute
// attached before the course marker, such as rpc_method on every request, is
// not lost once a later scope in the same chain marks it for the course log.
func TestHandlerPendingAttrsReplayedOnceMarkerAppears(t *testing.T) {
	store, _ := newTestStore(t)
	logger := slog.New(NewHandler(store))
	scoped := logger.With("rpc_method", "/qf.QuickFeedService/UpdateAssignments").
		With(label.CourseID, uint64(1), label.CourseCode, "DAT520", label.CourseLog, "dat520-2026")
	scoped.Info("scoped record")

	entry := readSingleEntry(t, store, "dat520-2026")
	if entry["rpc_method"] != "/qf.QuickFeedService/UpdateAssignments" {
		t.Errorf("rpc_method = %v, want the attribute attached before the marker", entry["rpc_method"])
	}
}

func TestHandlerTruncatesOversizedFields(t *testing.T) {
	store, _ := newTestStore(t)
	logger := slog.New(NewHandler(store)).With(label.CourseID, uint64(1), label.CourseCode, "DAT520", label.CourseLog, "dat520-2026")
	big := strings.Repeat("x", maxFieldBytes+1)
	logger.Info("test output", "output", big)

	entry := readSingleEntry(t, store, "dat520-2026")
	output, ok := entry["output"].(string)
	if !ok {
		t.Fatalf("output = %v (%T), want string", entry["output"], entry["output"])
	}
	if len(output) != maxFieldBytes {
		t.Errorf("len(output) = %d, want %d", len(output), maxFieldBytes)
	}
	if entry[label.Truncated] != true {
		t.Errorf("truncated = %v, want true", entry[label.Truncated])
	}
}

// readSingleEntry reads and decodes the single JSONL record expected in
// org's log file for the current UTC date.
func readSingleEntry(t *testing.T, store *Store, org string) map[string]any {
	t.Helper()
	date := time.Now().UTC().Format(dateLayout)
	data, err := os.ReadFile(filepath.Join(store.dir, org, date+".jsonl"))
	if err != nil {
		t.Fatalf("reading course log: %v", err)
	}
	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(data), &entry); err != nil {
		t.Fatalf("unmarshal course log line %q: %v", data, err)
	}
	return entry
}
