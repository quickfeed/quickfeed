package courselog

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/proto"
)

const testOrg = "dat520-2026"

// courseLogger returns a logger whose records the sink routes to org.
func courseLogger(store *Store, org string) *slog.Logger {
	return slog.New(NewHandler(store)).With(label.CourseLog, org)
}

// receive returns the next entry on sub, failing if none arrives promptly.
func receive(t *testing.T, sub *Subscription) *qf.CourseLogEntry {
	t.Helper()
	select {
	case entry, ok := <-sub.C():
		if !ok {
			t.Fatal("subscription closed, want an entry")
		}
		return entry
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for an entry")
		return nil
	}
}

func TestSubscriptionReceivesLoggedEntries(t *testing.T) {
	store, _ := newTestStore(t)
	sub := store.Subscribe(testOrg)
	defer sub.Close()

	logger := courseLogger(store, testOrg)
	logger.Info("cloning submission repository", label.Repository, "meling-labs")

	entry := receive(t, sub)
	if entry.GetMessage() != "cloning submission repository" {
		t.Errorf("Message = %q, want %q", entry.GetMessage(), "cloning submission repository")
	}
	if entry.GetLevel() != qf.CourseLogEntry_INFO {
		t.Errorf("Level = %v, want INFO", entry.GetLevel())
	}
	if entry.GetRepository() != "meling-labs" {
		t.Errorf("Repository = %q, want %q", entry.GetRepository(), "meling-labs")
	}
	if entry.GetTime() == nil {
		t.Error("Time is nil, want the record's timestamp")
	}
}

func TestSubscriptionIsolatedPerCourse(t *testing.T) {
	store, _ := newTestStore(t)
	sub := store.Subscribe(testOrg)
	defer sub.Close()

	courseLogger(store, "dat320-2026").Info("other course")
	courseLogger(store, testOrg).Info("this course")

	if got := receive(t, sub).GetMessage(); got != "this course" {
		t.Errorf("Message = %q, want %q: another course's records must not be delivered", got, "this course")
	}
}

// A course log query is scoped to a single course, so an org name that
// sanitizes to the same path element must reach the same subscribers as the
// writer does.
func TestSubscriptionSanitizesOrganization(t *testing.T) {
	store, _ := newTestStore(t)
	sub := store.Subscribe("dat520/2026")
	defer sub.Close()

	courseLogger(store, "dat520_2026").Info("sanitized to the same course")
	if got := receive(t, sub).GetMessage(); got != "sanitized to the same course" {
		t.Errorf("Message = %q, want the record written under the sanitized name", got)
	}
}

// Start is the end of the course's log when the subscription began, whether
// the file is open in this process or was left on disk by an earlier one; a
// record logged afterwards is positioned after it.
func TestSubscriptionStart(t *testing.T) {
	store, _ := newTestStore(t)
	today := qf.LogDay(time.Now())

	sub := store.Subscribe(testOrg)
	if got, want := sub.Start(), qf.NewLogCursor(today, 0); !proto.Equal(got, want) {
		t.Errorf("Start() = %+v with no file yet, want %+v", got, want)
	}
	sub.Close()

	// A file left by an earlier process, which the next write appends to.
	written := "{\"msg\":\"before a restart\"}\n"
	if err := os.MkdirAll(filepath.Join(store.dir, testOrg), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.path(testOrg, today), []byte(written), 0o600); err != nil {
		t.Fatal(err)
	}
	sub = store.Subscribe(testOrg)
	defer sub.Close()
	if got, want := sub.Start(), qf.NewLogCursor(today, int64(len(written))); !proto.Equal(got, want) {
		t.Errorf("Start() = %+v with a file on disk, want %+v", got, want)
	}

	courseLogger(store, testOrg).Info("after subscribing")
	entry := receive(t, sub)
	got := cursorOf(t, entry)
	if !got.Day().Equal(today) || !got.Beyond(sub.Start()) {
		t.Errorf("entry cursor = %v, want after Start() %v", got, sub.Start())
	}
	if size := fileSize(t, store.path(testOrg, today)); got.Position() != size {
		t.Errorf("entry cursor offset = %d, want %d, just past the record", got.Position(), size)
	}
	// With the file now open, Start is read from it rather than from disk.
	again := store.Subscribe(testOrg)
	defer again.Close()
	if !proto.Equal(again.Start(), got) {
		t.Errorf("Start() = %+v with the file open, want %+v, the end of the last record", again.Start(), got)
	}
}

// A Write can carry several records, and each must be positioned just past
// its own newline rather than at the end of the Write.
func TestPublishPositionsEachRecordInAWrite(t *testing.T) {
	store, _ := newTestStore(t)
	sub := store.Subscribe(testOrg)
	defer sub.Close()

	first := "{\"msg\":\"first\"}\n"
	second := "{\"msg\":\"second, a little longer\"}\n"
	if _, err := store.Writer(testOrg).Write([]byte(first + second)); err != nil {
		t.Fatal(err)
	}
	for _, want := range []int64{int64(len(first)), int64(len(first) + len(second))} {
		if got := cursorOf(t, receive(t, sub)).Position(); got != want {
			t.Errorf("cursor offset = %d, want %d", got, want)
		}
	}
}

func TestSubscriptionCloseEndsDelivery(t *testing.T) {
	store, _ := newTestStore(t)
	sub := store.Subscribe(testOrg)
	sub.Close()

	if _, ok := <-sub.C(); ok {
		t.Error("channel delivered after Close, want it closed")
	}
	// Closing twice must not panic on an already-closed channel.
	sub.Close()
	// Logging after the last subscriber left must still write the file.
	courseLogger(store, testOrg).Info("nobody is watching")
}

func TestStoreCloseEndsSubscriptions(t *testing.T) {
	store, _ := newTestStore(t)
	sub := store.Subscribe(testOrg)
	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	select {
	case _, ok := <-sub.C():
		if ok {
			t.Error("channel delivered after the store closed, want it closed")
		}
	case <-time.After(time.Second):
		t.Error("subscription still open after the store closed; its stream would never end")
	}
	// newTestStore's cleanup closes the store a second time.
	if err := store.Close(); err != nil {
		t.Errorf("second Close() error = %v, want nil", err)
	}
}

// A subscriber that never reads must lose entries rather than block the
// logger, and must be told that it did without waiting for another record to
// be logged: the burst that filled the buffer may be the last thing the course
// logs, and a report deferred until then would never arrive.
func TestSubscriptionDropsWhenFull(t *testing.T) {
	store, _ := newTestStore(t)
	sub := store.Subscribe(testOrg)
	defer sub.Close()

	logger := courseLogger(store, testOrg)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := range subscriptionBuffer + 10 {
			logger.Info("record", "i", i)
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("logging blocked on a full subscription buffer")
	}

	// Nothing further is logged: draining what is already buffered has to be
	// enough to find the gap report, or the loss is silent.
	var dropped *qf.CourseLogEntry
	for dropped == nil {
		select {
		case entry := <-sub.C():
			if Dropped(entry) {
				dropped = entry
			}
		default:
			t.Fatal("buffer drained without a dropped-entry report; the teacher would not know the view has a hole")
		}
	}
	if !strings.Contains(dropped.GetMessage(), "fell behind") {
		t.Errorf("Message = %q, want it to say the stream fell behind", dropped.GetMessage())
	}
	if dropped.GetLevel() != qf.CourseLogEntry_WARN {
		t.Errorf("Level = %v, want WARN", dropped.GetLevel())
	}
}

// Logging must cost nothing extra while nobody is watching.
func TestPublishSkippedWithoutSubscribers(t *testing.T) {
	store, operator := newTestStore(t)
	if _, err := store.Writer(testOrg).Write([]byte("not json at all\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if got := operator.String(); got != "" {
		t.Errorf("operator log = %q, want empty: undecodable records must not be decoded without subscribers", got)
	}
}

// A record that will not decode must be reported and skipped, never returned
// to the logger as a write failure.
func TestPublishReportsUndecodableRecord(t *testing.T) {
	store, operator := newTestStore(t)
	sub := store.Subscribe(testOrg)
	defer sub.Close()

	if _, err := store.Writer(testOrg).Write([]byte("not json at all\n")); err != nil {
		t.Errorf("Write() error = %v, want nil: a subscriber must not make logging fail", err)
	}
	if !strings.Contains(operator.String(), "decoding course log record") {
		t.Errorf("operator log = %q, want it to report the undecodable record", operator.String())
	}
	select {
	case entry := <-sub.C():
		t.Errorf("delivered %v, want nothing for an undecodable record", entry)
	default:
	}
}

// A gap report must be recognizable as the subscription's own, so a stream can
// deliver it past filters that would otherwise hide the fact that the view is
// incomplete.
func TestDroppedRecognizesOnlyTheGapReport(t *testing.T) {
	if !Dropped(droppedEntry(3)) {
		t.Error("Dropped(droppedEntry(3)) = false, want true")
	}
	store, _ := newTestStore(t)
	sub := store.Subscribe(testOrg)
	defer sub.Close()
	courseLogger(store, testOrg).Info("an ordinary record", label.Repository, "meling-labs")
	if entry := receive(t, sub); Dropped(entry) {
		t.Error("Dropped(<logged record>) = true, want false")
	}
}

// Subscribe must not hand back a subscription whose channel nothing is left to
// close: a stream opened while the store shuts down would wait forever.
func TestSubscribeAfterCloseEndsAtOnce(t *testing.T) {
	store, _ := newTestStore(t)
	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	sub := store.Subscribe(testOrg)
	defer sub.Close() // must be a no-op rather than a second close
	select {
	case _, ok := <-sub.C():
		if ok {
			t.Error("received an entry from a closed store, want the channel closed")
		}
	case <-time.After(time.Second):
		t.Fatal("subscription taken after Close never ended")
	}
}

// Every record is in exactly one of the backlog bounded by the subscription's
// Start and the subscription itself, however writes and subscribing interleave:
// in neither, it is lost for good; in both, the teacher sees it twice.
func TestSubscribeHandoffIsExact(t *testing.T) {
	store, _ := newTestStore(t)
	logger := courseLogger(store, testOrg)

	const writers, perWriter = 4, 100
	var wg sync.WaitGroup
	started := make(chan struct{})
	for w := range writers {
		wg.Go(func() {
			for i := range perWriter {
				if w == 0 && i == perWriter/4 {
					close(started)
				}
				logger.Info("record", "id", fmt.Sprintf("%d-%d", w, i))
			}
		})
	}
	// Subscribe while the writers are busy, so the handoff falls between
	// records being written rather than before or after all of them.
	<-started
	sub := store.Subscribe(testOrg)
	defer sub.Close()
	wg.Wait()

	backlog, _, _, err := store.Query(testOrg, &qf.CourseLogRequest{Limit: 5000}, sub.Start())
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	seen := make(map[string]int)
	for _, entry := range backlog {
		if c := cursorOf(t, entry); c.Beyond(sub.Start()) {
			t.Errorf("backlog entry at %+v is past Start() %+v", c, sub.Start())
		}
		seen[entry.GetFields()["id"]]++
	}
	for len(sub.C()) > 0 {
		entry := <-sub.C()
		if Dropped(entry) {
			t.Fatal("subscription dropped entries; the buffer is meant to hold them all")
		}
		if c := cursorOf(t, entry); !c.Beyond(sub.Start()) {
			t.Errorf("live entry at %+v is at or before Start() %+v", c, sub.Start())
		}
		seen[entry.GetFields()["id"]]++
	}
	for w := range writers {
		for i := range perWriter {
			if id := fmt.Sprintf("%d-%d", w, i); seen[id] != 1 {
				t.Errorf("record %s seen %d times, want exactly once", id, seen[id])
			}
		}
	}
}

// cursorOf returns the cursor entry carries, failing if it has no valid one.
func cursorOf(t *testing.T, entry *qf.CourseLogEntry) *qf.LogCursor {
	t.Helper()
	c := entry.GetCursor()
	if !c.IsValid() {
		t.Fatalf("entry %q has cursor %v, want a valid one", entry.GetMessage(), c)
	}
	return c
}

func fileSize(t *testing.T, path string) int64 {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Size()
}
