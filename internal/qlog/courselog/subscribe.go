package courselog

import (
	"errors"
	"io/fs"
	"os"
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
// before entries are dropped. The channel holds one more slot, reserved for a
// gap report.
const subscriptionBuffer = 512

// Subscription delivers a course's log entries as they are written. Close it
// when done; an open subscription holds a buffer and costs the writer a decode
// per record.
type Subscription struct {
	store *Store
	org   string
	start *qf.LogCursor
	ch    chan *qf.CourseLogEntry

	// dropped counts entries discarded because ch was full since the last
	// report of a gap. It is only touched by publish, under the store's mutex.
	dropped int
}

// Start returns the cursor at which the subscription began; entries after it
// arrive on the channel.
func (sub *Subscription) Start() *qf.LogCursor { return sub.start }

// C returns the channel entries arrive on. It is closed when the subscription
// is closed, or when the store shuts down.
func (sub *Subscription) C() <-chan *qf.CourseLogEntry { return sub.ch }

// Close stops delivery and releases the subscription's buffer. It is safe to
// call more than once.
func (sub *Subscription) Close() {
	sub.store.unsubscribe(sub)
}

// Subscribe returns a Subscription delivering org's entries written after
// its Start. A backlog read with Query bounded by Start therefore holds
// exactly the entries the subscription does not.
func (s *Store) Subscribe(org string) *Subscription {
	sub := &Subscription{
		store: s,
		org:   sanitize(org),
		ch:    make(chan *qf.CourseLogEntry, subscriptionBuffer+1),
	}
	// write holds cf.mu from writing a record until publishing it, so with
	// it held, each record is either before the end read here or delivered
	// to the subscription registered here.
	cf := s.courseFileFor(sub.org)
	cf.mu.Lock()
	defer cf.mu.Unlock()
	sub.start = s.end(sub.org, cf)

	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	if s.closed {
		// No one is left to close the channel, so close it now.
		close(sub.ch)
		return sub
	}
	s.subs[sub.org] = append(s.subs[sub.org], sub)
	return sub
}

// end returns the cursor just past org's last record, which may be in a file
// from before a restart. It returns nil if the file's size cannot be read,
// leaving the backlog unbounded. Callers must hold cf.mu.
func (s *Store) end(org string, cf *courseFile) *qf.LogCursor {
	if end := cf.end(); end != nil {
		return end
	}
	today := qf.LogDay(s.now())
	info, err := os.Stat(s.path(org, today))
	switch {
	case err == nil:
		return qf.NewLogCursor(today, info.Size())
	case errors.Is(err, fs.ErrNotExist):
		return qf.NewLogCursor(today, 0)
	default:
		s.reportError(org, "reading course log file size", err)
		return nil
	}
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

// closeSubscriptions closes every open subscription, and marks the store
// closed so that a later Subscribe returns a closed subscription.
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

// hasSubscribers reports whether any subscription is open for org.
func (s *Store) hasSubscribers(org string) bool {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	return len(s.subs[org]) > 0
}

// publish delivers the records in p, already written to org's file for day at
// offset start, to org's subscribers. It never blocks and never fails: a
// subscriber that cannot keep up loses entries and is told so, and a record
// that will not decode is reported to the operator and skipped. Logging must
// not be held up, nor fail, because someone is watching.
func (s *Store) publish(org string, day time.Time, start int64, p []byte) {
	if !s.hasSubscribers(org) {
		return
	}
	// A Write may carry several records; give each its own cursor.
	pos := start
	for line := range strings.SplitAfterSeq(string(p), "\n") {
		pos += int64(len(line))
		if strings.TrimSpace(line) == "" {
			continue
		}
		entry, err := decodeEntry(line)
		if err != nil {
			s.reportError(org, "decoding course log record for subscribers", err)
			continue
		}
		entry.Cursor = qf.NewLogCursor(day, pos)
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

// send queues entry, or drops it and queues a gap report if the subscriber is
// too far behind. Callers must hold the store's subsMu.
func (sub *Subscription) send(entry *qf.CourseLogEntry) {
	if sub.dropped > 0 {
		sub.reportDropped()
	}
	// Sends happen under subsMu and the reader only takes entries out, so
	// queuing below the reserved slot never blocks.
	if len(sub.ch) < cap(sub.ch)-1 {
		sub.ch <- entry
		return
	}
	sub.dropped++
	sub.reportDropped()
}

// reportDropped queues a gap report in the reserved slot and resets the
// dropped count. If an earlier report is still unread, the count carries over
// to the next report.
func (sub *Subscription) reportDropped() {
	select {
	case sub.ch <- droppedEntry(sub.dropped):
		sub.dropped = 0
	default:
	}
}

// Dropped reports whether entry is a subscription's gap report rather than a
// record from the course log.
func Dropped(entry *qf.CourseLogEntry) bool {
	_, ok := entry.GetFields()[droppedLabel]
	return ok
}

// droppedEntry returns a gap report for dropped entries.
func droppedEntry(dropped int) *qf.CourseLogEntry {
	return &qf.CourseLogEntry{
		Time:    timestamppb.Now(),
		Level:   qf.CourseLogEntry_WARN,
		Message: "course log stream fell behind; entries were not delivered",
		Fields:  map[string]string{droppedLabel: strconv.Itoa(dropped)},
	}
}
