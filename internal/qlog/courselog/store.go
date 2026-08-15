// Package courselog stores structured log records scoped to a course, so
// teachers can review webhook processing, CI and Docker output, and course
// operations without operator access to the shared process log.
//
// Store holds one newline-delimited JSON file per course per UTC day, under
// <dir>/<organization>/<date>.jsonl. A single Store instance serves every
// course for the lifetime of the process: courses are registered lazily, on
// their first write, so a course created after the server started requires
// no restart before it starts logging.
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
)

// retention is how long a course's date files are kept before being removed.
const retention = 14 * 24 * time.Hour

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

	stop    chan struct{}
	stopped chan struct{}
}

// courseFile holds the currently open file for one course and the UTC date
// it covers. Access is serialized by mu, so a course's records are written in
// order even under concurrent logging.
type courseFile struct {
	mu   sync.Mutex
	file *os.File
	date string
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

// Close stops the retention loop and closes every open course file.
func (s *Store) Close() error {
	close(s.stop)
	<-s.stopped

	s.mu.Lock()
	defer s.mu.Unlock()
	var errs []error
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

	today := s.now().UTC().Format(dateLayout)
	if cf.file == nil || cf.date != today {
		if cf.file != nil {
			_ = cf.file.Close()
		}
		f, err := s.openFile(org, today)
		if err != nil {
			s.reportError(org, "opening course log file", err)
			return 0, err
		}
		cf.file = f
		cf.date = today
	}
	n, err := cf.file.Write(p)
	if err != nil {
		s.reportError(org, "writing course log record", err)
	}
	return n, err
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

func (s *Store) openFile(org, date string) (*os.File, error) {
	dir := filepath.Join(s.dir, org)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, date+".jsonl")
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
}

func (s *Store) reportError(org, action string, err error) {
	s.operator.Error("course log store: "+action, label.Organization, org, label.Error, err)
}

// cleanupExpired removes date files older than the retention window, across
// every course directory.
func (s *Store) cleanupExpired() {
	cutoff := s.now().UTC().Add(-retention)
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
