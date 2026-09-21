package courselog

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func TestSubscriptionSinceIsBeforeLaterRecords(t *testing.T) {
	store, _ := newTestStore(t)
	before := time.Now()
	sub := store.Subscribe(testOrg)
	defer sub.Close()

	if sub.Since().Before(before) {
		t.Errorf("Since() = %v, want at or after %v", sub.Since(), before)
	}
	courseLogger(store, testOrg).Info("after subscribing")
	if got := receive(t, sub).GetTime().AsTime(); got.Before(sub.Since()) {
		t.Errorf("entry time %v is before Since() %v; the backlog would miss it too", got, sub.Since())
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

// A record written while a subscription is being registered must reach the
// subscriber, or fall within the backlog its Since bounds: never neither.
//
// The clock hook stands in for the instant Subscribe reads it: it lets a
// writer loose and returns the earlier time, so the record is timestamped
// after Since. Reading the clock under the same lock publish takes is what
// holds that writer until the subscription is registered; read before the
// lock, the record reaches neither path and is lost.
func TestSubscribeHandoffLosesNoRecord(t *testing.T) {
	store, _ := newTestStore(t)
	logger := courseLogger(store, testOrg)

	writing := make(chan struct{})
	go func() {
		<-writing
		logger.Info("written while subscribing")
	}()
	subscribing := true
	store.now = func() time.Time {
		at := time.Now()
		if subscribing {
			// Only Subscribe's own call; the write below sees this false,
			// having been released by the send that precedes it.
			subscribing = false
			writing <- struct{}{}
			// Give the writer every chance to publish before the
			// subscription is registered: the window the race opens.
			time.Sleep(20 * time.Millisecond)
		}
		return at
	}

	sub := store.Subscribe(testOrg)
	defer sub.Close()
	since := sub.Since()

	// The record is either newer than Since, and so delivered, or not, and so
	// within the backlog Since bounds. It must not be missing from both.
	select {
	case entry := <-sub.C():
		if got := entry.GetMessage(); got != "written while subscribing" {
			t.Errorf("delivered Message = %q, want %q", got, "written while subscribing")
		}
		return
	case <-time.After(time.Second):
	}
	entries, _, _, err := store.Query(testOrg, &qf.CourseLogRequest{To: timestamppb.New(since)})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("record was neither delivered nor in the backlog: it fell between the two")
	}
}
