// Package courselog stores structured log records scoped to a course, so
// teachers can review webhook processing, CI and Docker output, and course
// operations without operator access to the shared process log.
//
// Store holds one newline-delimited JSON file per course per UTC day, under
// <dir>/<organization>/<date>.jsonl. A single Store instance serves every
// course for the lifetime of the process: courses are registered lazily, on
// their first write, so a course created after the server started requires
// no restart before it starts logging.
//
// Every entry read back carries a cursor, its position in the course's log,
// which is what a reader resumes from: Query returns only entries after a
// cursor, and a Subscription says at which position its delivery begins.
package courselog

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
)

// Retention is how long a course's date files are kept before being removed.
// Store.Query clamps a request's interval to this window.
const Retention = 14 * 24 * time.Hour

const dateLayout = "2006-01-02"

// cleanupInterval is how often the retention sweep runs after startup.
const cleanupInterval = 24 * time.Hour

// Store manages the on-disk course log files under a single root directory.
// One Store instance serves every course for the lifetime of the process;
// courses are registered lazily, on their first write, so a course created
// after the process started requires no restart.
type Store struct {
	dir      string
	operator *slog.Logger
	now      func() time.Time // overridden in tests to exercise date rollover without waiting for it

	mu      sync.Mutex // guards courses; each course's own file is guarded by its courseFile
	courses map[string]*courseFile

	// subsMu guards subs and closed. It is the innermost of the store's locks:
	// publish takes it while holding a courseFile's lock, so that subscribers
	// see a course's records in the order they were written, and Subscribe
	// does too, so that a subscription begins exactly where the file ends.
	// Nothing takes mu or a courseFile lock while holding it.
	subsMu sync.Mutex
	subs   map[string][]*Subscription
	closed bool

	closeOnce sync.Once
	stop      chan struct{}
	stopped   chan struct{}
}

// courseFile holds the currently open file for one course, the UTC day it
// covers, and its size. Access is serialized by mu, so a course's records are
// written in order even under concurrent logging, and each is given the
// position it was written at.
type courseFile struct {
	mu   sync.Mutex
	file *os.File
	day  time.Time
	// size is where the next write lands: the file is opened for appending,
	// so its offset is always its size, and only this process writes to it.
	size int64
}

// end returns the position just past the last record written to cf, or nil if
// no file is open. Callers must hold cf.mu.
func (cf *courseFile) end() *qf.LogCursor {
	if cf.file == nil {
		return nil
	}
	return qf.NewLogCursor(cf.day, cf.size)
}

// NewStore creates dir if it does not already exist, removes any date files
// already past retention, and starts a daily cleanup loop. Call Close to stop
// the loop and release open file handles.
func NewStore(dir string, operator *slog.Logger) (*Store, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("creating course log directory: %w", err)
	}
	s := &Store{
		dir:      dir,
		operator: operator,
		now:      time.Now,
		courses:  make(map[string]*courseFile),
		subs:     make(map[string][]*Subscription),
		stop:     make(chan struct{}),
		stopped:  make(chan struct{}),
	}
	s.cleanupExpired()
	go s.cleanupLoop()
	return s, nil
}

func (s *Store) cleanupLoop() {
	defer close(s.stopped)
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.cleanupExpired()
		case <-s.stop:
			return
		}
	}
}

// Close stops the retention loop, ends every open subscription, and closes
// every open course file. Calling it more than once is a no-op, so a store
// handed to both a server's cleanup and a test's is safe to close twice.
func (s *Store) Close() error {
	var errs []error
	s.closeOnce.Do(func() {
		close(s.stop)
		<-s.stopped
		s.closeSubscriptions()

		s.mu.Lock()
		defer s.mu.Unlock()
		for _, cf := range s.courses {
			cf.mu.Lock()
			if cf.file != nil {
				if err := cf.file.Close(); err != nil {
					errs = append(errs, err)
				}
				cf.file = nil
			}
			cf.mu.Unlock()
		}
	})
	return errors.Join(errs...)
}

// Writer returns an io.Writer that appends to org's current date file,
// creating the course's directory and file on first use. org is sanitized
// before use as a path element, so it cannot escape the store's directory.
func (s *Store) Writer(org string) io.Writer {
	return &courseWriter{store: s, org: sanitize(org)}
}

type courseWriter struct {
	store *Store
	org   string
}

func (w *courseWriter) Write(p []byte) (int, error) {
	return w.store.write(w.org, p)
}

func (s *Store) write(org string, p []byte) (int, error) {
	cf := s.courseFileFor(org)
	cf.mu.Lock()
	defer cf.mu.Unlock()

	today := qf.LogDay(s.now())
	if cf.file == nil || !cf.day.Equal(today) {
		if cf.file != nil {
			_ = cf.file.Close()
		}
		cf.file = nil
		f, size, err := s.openFile(org, today)
		if err != nil {
			s.reportError(org, "opening course log file", err)
			return 0, err
		}
		cf.file, cf.day, cf.size = f, today, size
	}
	start := cf.size
	n, err := cf.file.Write(p)
	// Counted even on failure: a short write still moved the end of the file,
	// and a later record's position has to say where it really is.
	cf.size += int64(n)
	if err != nil {
		s.reportError(org, "writing course log record", err)
		return n, err
	}
	// Still holding cf.mu, so subscribers see this course's records in the
	// same order the file has them, and Subscribe cannot read the file's end
	// between this record being written and it being delivered.
	s.publish(org, cf.day, start, p)
	return n, nil
}

// courseFileFor returns org's file state, registering the course on its
// first appearance. This is the only place a course is added to the store:
// no course needs to be known when the Store is constructed, and none needs
// a restart to start logging after it is created.
func (s *Store) courseFileFor(org string) *courseFile {
	s.mu.Lock()
	defer s.mu.Unlock()
	cf, ok := s.courses[org]
	if !ok {
		cf = &courseFile{}
		s.courses[org] = cf
	}
	return cf
}

// openFile opens org's file for day for appending, and returns its size, which
// is where the first record written to it will begin.
func (s *Store) openFile(org string, day time.Time) (*os.File, int64, error) {
	dir := filepath.Join(s.dir, org)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, 0, err
	}
	f, err := os.OpenFile(s.path(org, day), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, 0, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, 0, err
	}
	return f, info.Size(), nil
}

// path returns the file holding org's records for day, a UTC midnight. org
// must already be sanitized.
func (s *Store) path(org string, day time.Time) string {
	return filepath.Join(s.dir, org, day.Format(dateLayout)+".jsonl")
}

func (s *Store) reportError(org, action string, err error) {
	s.operator.Error("course log store: "+action, label.Organization, org, label.Error, err)
}

// cleanupExpired removes date files older than the retention window, across
// every course directory.
func (s *Store) cleanupExpired() {
	cutoff := s.now().UTC().Add(-Retention)
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		s.operator.Error("course log store: listing course directories", label.Error, err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			s.cleanupCourseDir(entry.Name(), cutoff)
		}
	}
}

func (s *Store) cleanupCourseDir(org string, cutoff time.Time) {
	dir := filepath.Join(s.dir, org)
	files, err := os.ReadDir(dir)
	if err != nil {
		s.reportError(org, "listing course log files", err)
		return
	}
	for _, f := range files {
		date, ok := strings.CutSuffix(f.Name(), ".jsonl")
		if !ok {
			continue
		}
		t, err := time.Parse(dateLayout, date)
		if err != nil {
			// Not a date file this store wrote; leave it alone.
			continue
		}
		if t.Before(cutoff) {
			if err := os.Remove(filepath.Join(dir, f.Name())); err != nil {
				s.reportError(org, "removing expired course log file", err)
			}
		}
	}
}

// sanitize restricts org to characters safe as a single path element,
// replacing anything else with '_', so a course's SCM organization name can
// never escape the course log directory.
func sanitize(org string) string {
	var b strings.Builder
	for _, r := range org {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	switch sanitized := b.String(); sanitized {
	case "", ".", "..":
		return "_"
	default:
		return sanitized
	}
}
