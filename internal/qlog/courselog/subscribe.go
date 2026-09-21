package courselog

import (
	"strconv"
	"strings"
	"time"

	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// droppedLabel names the attribute carrying how many entries a subscription
// lost. It is local to this package: nothing else logs or queries it.
const droppedLabel = "dropped"

// subscriptionBuffer is how many entries a subscriber may fall behind by
// before entries are dropped. A teacher's browser reads far slower than a
// course can log, so the buffer absorbs a burst, such as a push that rebuilds
// every submission, without either blocking the logger or losing the burst.
// The channel holds one slot beyond this, kept free so that a report of what
// was dropped always has somewhere to go.
const subscriptionBuffer = 512

// Subscription delivers a course's log entries as they are written. Close it
// when done; an open subscription holds a buffer and costs the writer a decode
// per record.
type Subscription struct {
	store *Store
	org   string
	since time.Time
	ch    chan *qf.CourseLogEntry

	// dropped counts entries discarded because ch was full since the last
	// report of a gap. It is only touched by publish, under the store's mutex.
	dropped int
}

// Since reports when the subscription began. Entries logged before it belong
// to the backlog a caller reads with Query; entries logged after it arrive on
// the channel.
func (sub *Subscription) Since() time.Time { return sub.since }

// C returns the channel entries arrive on. It is closed when the subscription
// is closed, or when the store shuts down.
func (sub *Subscription) C() <-chan *qf.CourseLogEntry { return sub.ch }

// Close stops delivery and releases the subscription's buffer. It is safe to
// call more than once.
func (sub *Subscription) Close() {
	sub.store.unsubscribe(sub)
}

// Subscribe returns a Subscription delivering org's entries as they are
// written. Subscribe before reading the backlog: an entry written between the
// two is then delivered rather than missed.
func (s *Store) Subscribe(org string) *Subscription {
	sub := &Subscription{
		store: s,
		org:   sanitize(org),
		ch:    make(chan *qf.CourseLogEntry, subscriptionBuffer+1),
	}
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	// since is read under the same lock publish takes, so a record written
	// while this runs is either newer than since, and so delivered here, or
	// older, and so within the backlog a caller then reads. Taken before the
	// lock, a record could fall between the two and reach neither.
	sub.since = s.now()
	if s.closed {
		// The store is shut down, so nothing will be written or delivered
		// again: hand back a subscription that is already over rather than
		// one whose channel no one is left to close.
		close(sub.ch)
		return sub
	}
	s.subs[sub.org] = append(s.subs[sub.org], sub)
	return sub
}

func (s *Store) unsubscribe(sub *Subscription) {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	subs := s.subs[sub.org]
	for i, other := range subs {
		if other != sub {
			continue
		}
		s.subs[sub.org] = append(subs[:i], subs[i+1:]...)
		if len(s.subs[sub.org]) == 0 {
			delete(s.subs, sub.org)
		}
		close(sub.ch)
		return
	}
}

// closeSubscriptions closes every open subscription, ending the streams that
// read from them, and marks the store closed so a later Subscribe ends at once
// rather than waiting on a channel nothing will close.
func (s *Store) closeSubscriptions() {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	s.closed = true
	for org, subs := range s.subs {
		for _, sub := range subs {
			close(sub.ch)
		}
		delete(s.subs, org)
	}
}

// hasSubscribers reports whether any subscription is open for org. The write
// path checks this before decoding, so logging costs nothing extra while no
// teacher is watching.
func (s *Store) hasSubscribers(org string) bool {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	return len(s.subs[org]) > 0
}

// publish delivers the records in p, already written to org's file, to org's
// subscribers. It never blocks and never fails: a subscriber that cannot keep
// up loses entries and is told so, and a record that will not decode is
// reported to the operator and skipped. Logging must not be held up, nor fail,
// because someone is watching.
func (s *Store) publish(org string, p []byte) {
	if !s.hasSubscribers(org) {
		return
	}
	// slog.JSONHandler writes one complete record per call, but a Write
	// carrying several lines would otherwise decode as one malformed record.
	for line := range strings.SplitSeq(string(p), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		entry, err := decodeEntry(line)
		if err != nil {
			s.reportError(org, "decoding course log record for subscribers", err)
			continue
		}
		s.deliver(org, entry)
	}
}

func (s *Store) deliver(org string, entry *qf.CourseLogEntry) {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	for _, sub := range s.subs[org] {
		sub.send(entry)
	}
}

// send queues entry, dropping it if the subscriber is too far behind. A
// subscriber that has dropped entries is told so, so a teacher knows the view
// has a hole in it rather than silently missing records. Callers must hold the
// store's subsMu.
func (sub *Subscription) send(entry *qf.CourseLogEntry) {
	if sub.dropped > 0 {
		sub.reportDropped()
	}
	// Only the reserved slot is left once the buffer holds subscriptionBuffer
	// entries. Sends happen under subsMu and the reader only takes entries
	// out, so a length seen below the reserve cannot have grown by the time
	// we send: queuing here never blocks.
	if len(sub.ch) < cap(sub.ch)-1 {
		sub.ch <- entry
		return
	}
	sub.dropped++
	sub.reportDropped()
}

// reportDropped queues a gap report in the slot kept free for it, and resets
// the count it carries. Reporting at the moment of the drop is what makes the
// loss visible at all: a burst that fills the buffer may be the last thing the
// course logs, and a report deferred to the next record would then never be
// sent, leaving the subscriber to drain the buffer and see nothing amiss.
//
// The slot is free unless an earlier report is still unread. Drops while one
// is outstanding therefore keep accumulating and ride on the next report, so a
// report counts the entries lost since the one before it rather than every
// entry lost so far. A tail of drops behind an unread report is not counted at
// all, but the report the subscriber does receive already says the view is
// incomplete, which is what it has to act on.
func (sub *Subscription) reportDropped() {
	select {
	case sub.ch <- droppedEntry(sub.dropped):
		sub.dropped = 0
	default:
	}
}

// Dropped reports whether entry is a subscription's own report of a gap in
// delivery rather than a record read from the course log. It says the view is
// incomplete, so a stream must deliver it whatever the caller asked to see;
// nothing written through the sink carries this attribute.
func Dropped(entry *qf.CourseLogEntry) bool {
	_, ok := entry.GetFields()[droppedLabel]
	return ok
}

// droppedEntry reports a gap in a subscription's delivery as a log entry of
// its own, so it reaches the teacher through the same path as any other.
func droppedEntry(dropped int) *qf.CourseLogEntry {
	return &qf.CourseLogEntry{
		Time:    timestamppb.Now(),
		Level:   qf.CourseLogEntry_WARN,
		Message: "course log stream fell behind; entries were not delivered",
		Fields:  map[string]string{droppedLabel: strconv.Itoa(dropped)},
	}
}
