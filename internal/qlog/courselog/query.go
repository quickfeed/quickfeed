package courselog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"slices"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Query returns org's course log entries that match req, in the order they
// were written, each carrying its cursor.
//
// The result lists every repository with an entry in req's time range, also
// those that req's repository and level filters leave out, so the client can
// offer them all as filter options. If more entries match than req's limit,
// the result is marked truncated and holds the oldest matches, or the newest
// if req.Newest is set.
//
// The time range ends no later than now and starts no earlier than the
// oldest date the store retains. A From or To cursor leaves out the entry it
// names. If until is non-nil, the result ends at until, inclusive.
//
// A cursor the store did not hand out fails with an error wrapping
// qf.ErrInvalidLogCursor. A malformed final line, as left by a partial
// write, is skipped.
func (s *Store) Query(org string, req *qf.CourseLogRequest, until *qf.LogCursor) (*qf.CourseLog, error) {
	now := s.now()
	if err := req.CheckCursors(until, now); err != nil {
		return nil, err
	}
	r := newReadRange(req, until, now)
	if r.empty() {
		return &qf.CourseLog{}, nil
	}

	sc := &scan{
		from:       r.from,
		to:         r.to,
		repository: req.GetRepository(),
		level:      req.GetLevel(),
		kept:       newEntryLimit(req.EffectiveLimit(), req.GetNewest()),
		repos:      make(map[string]bool),
	}
	org = sanitize(org)
	for _, day := range daysBetween(r.start.Day(), r.end.Day()) {
		path := s.path(org, day)
		if err := sc.file(path, day, r.span(day)); err != nil {
			return nil, fmt.Errorf("reading course log %s: %w", path, err)
		}
	}
	entries, truncated := sc.kept.ordered()
	return &qf.CourseLog{
		Entries:      entries,
		Repositories: slices.Sorted(maps.Keys(sc.repos)),
		Truncated:    truncated,
	}, nil
}

// readRange is the part of the log a query reads, from start up to end, and
// the time range its entries must fall within. If excludeEnd is set, the
// entry ending at end is left out.
type readRange struct {
	from, to   time.Time
	start, end *qf.LogCursor
	excludeEnd bool
}

// newReadRange returns the part of the log to read for req, ending at until
// if it is non-nil.
func newReadRange(req *qf.CourseLogRequest, until *qf.LogCursor, now time.Time) *readRange {
	from, to := req.Interval(oldestRetainedDate(now), now)
	end, excludeEnd := req.End(to, now, until)
	return &readRange{from: from, to: to, start: req.Start(from), end: end, excludeEnd: excludeEnd}
}

// empty reports whether r holds no entries.
func (r *readRange) empty() bool {
	return r.from.After(r.to) || !r.end.Beyond(r.start)
}

// span returns the part of day's file that r covers.
func (r *readRange) span(day time.Time) span {
	sp := span{start: 0, end: qf.EndOfLogFile}
	if day.Equal(r.start.Day()) {
		sp.start = r.start.Position()
	}
	if day.Equal(r.end.Day()) {
		sp.end, sp.exclusive = r.end.Position(), r.excludeEnd
	}
	return sp
}

// span is the byte range of one date file a query reads: from start up to
// end. If exclusive is set, the record ending at end is left out.
type span struct {
	start, end int64
	exclusive  bool
}

// oldestRetainedDate returns the UTC midnight of the oldest date
// cleanupCourseDir still guarantees to keep, as of now.
func oldestRetainedDate(now time.Time) time.Time {
	cutoff := now.Add(-Retention).UTC()
	midnight := qf.LogDay(cutoff)
	if midnight.Before(cutoff) {
		return midnight.AddDate(0, 0, 1)
	}
	return midnight
}

// daysBetween returns the UTC midnights of the days from from's through to's,
// inclusive.
func daysBetween(from, to time.Time) []time.Time {
	var days []time.Time
	for d := qf.LogDay(from); !d.After(to); d = d.AddDate(0, 0, 1) {
		days = append(days, d)
	}
	return days
}

// scan holds a query's filters and what it has collected so far, across the
// date files it reads.
type scan struct {
	from, to   time.Time
	repository string
	level      qf.CourseLogEntry_Level
	kept       *entryLimit
	repos      map[string]bool
}

// file decodes the part of path that sp covers line by line, adding entries
// matching the query's interval, repository, and level to kept, and
// every repository with an entry timestamped within the interval to repos.
// A missing file means the course had no activity that day, not a failure.
func (sc *scan) file(path string, day time.Time, sp span) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	r, err := sp.reader(f)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(r)
	// A record can carry several fields independently truncated to 64 KiB
	// each (see maxFieldBytes), so grow well past both a single field and
	// bufio.Scanner's 64 KiB default cap; a line still over this is corrupt
	// rather than merely large, and fails the query as any other read error.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	// Counting what the scanner consumes, newline included, is what gives
	// each line its position; the text it returns has the newline cut off.
	offset := sp.start
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		advance, token, err := bufio.ScanLines(data, atEOF)
		offset += int64(advance)
		return advance, token, err
	})

	// pending trails the scanner by one line, so a line is only ever applied
	// once Scan() has confirmed whether a further line follows it; that is
	// what lets the final line alone get apply's partial-write tolerance,
	// without first holding the whole file (tens to hundreds of MiB for a
	// busy course) in memory to find out which line that was.
	var pending string
	var pendingEnd int64
	havePending := false
	for scanner.Scan() {
		if havePending {
			if err := sc.apply(pending, day, sp, pendingEnd, false); err != nil {
				return err
			}
		}
		pending, pendingEnd, havePending = scanner.Text(), offset, true
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if havePending {
		return sc.apply(pending, day, sp, pendingEnd, true)
	}
	return nil
}

// apply decodes line and folds it into kept and repos, giving the entry the
// cursor end in day's file. A decode error is tolerated only if final is set,
// on the theory that the process was killed mid-write.
func (sc *scan) apply(line string, day time.Time, sp span, end int64, final bool) error {
	if sp.exclusive && end == sp.end {
		return nil
	}
	entry, err := decodeEntry(line)
	if err != nil {
		if final {
			return nil
		}
		return err
	}
	entry.Cursor = qf.NewLogCursor(day, end)
	if entry.GetRepository() != "" && entry.InInterval(sc.from, sc.to) {
		sc.repos[entry.GetRepository()] = true
	}
	if entry.Matches(sc.from, sc.to, sc.repository, sc.level) {
		sc.kept.add(entry)
	}
	return nil
}

// reader returns a reader of the part of f that sp covers.
func (sp span) reader(f *os.File) (io.Reader, error) {
	if err := atRecordBoundary(f, sp.start); err != nil {
		return nil, err
	}
	if sp.exclusive {
		if err := atRecordBoundary(f, sp.end); err != nil {
			return nil, err
		}
	}
	if _, err := f.Seek(sp.start, io.SeekStart); err != nil {
		return nil, err
	}
	return io.LimitReader(f, sp.end-sp.start), nil
}

// atRecordBoundary returns an error wrapping qf.ErrInvalidLogCursor unless offset
// is zero or just past a newline in f.
func atRecordBoundary(f *os.File, offset int64) error {
	if offset == 0 {
		return nil
	}
	var last [1]byte
	if _, err := f.ReadAt(last[:], offset-1); err != nil || last[0] != '\n' {
		return fmt.Errorf("%w: offset %d is not at a record boundary", qf.ErrInvalidLogCursor, offset)
	}
	return nil
}

// decodeEntry parses one JSONL record written by Handler. slog.JSONHandler's
// standard keys (time, level, msg, source) and the labels Handler promotes to
// dedicated CourseLogEntry fields are recognized by name; everything else
// becomes a Fields entry.
func decodeEntry(line string) (*qf.CourseLogEntry, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return nil, err
	}

	entry := &qf.CourseLogEntry{Fields: make(map[string]string, len(raw))}
	var t time.Time
	for key, value := range raw {
		var err error
		switch key {
		case slog.TimeKey:
			err = json.Unmarshal(value, &t)
		case slog.LevelKey:
			var level string
			if err = json.Unmarshal(value, &level); err == nil {
				entry.Level = parseLevel(level)
			}
		case slog.MessageKey:
			err = json.Unmarshal(value, &entry.Message)
		case slog.SourceKey:
			entry.Source, err = decodeSource(value)
		case label.Repository:
			err = json.Unmarshal(value, &entry.Repository)
		case label.RepositoryType:
			err = json.Unmarshal(value, &entry.RepositoryType)
		case label.Truncated:
			err = json.Unmarshal(value, &entry.Truncated)
		case label.CourseID, label.CourseCode:
			// Redundant: a query is always scoped to a single, known course.
		default:
			entry.Fields[key] = stringify(value)
		}
		if err != nil {
			return nil, fmt.Errorf("decoding %q: %w", key, err)
		}
	}
	entry.Time = timestamppb.New(t)
	return entry, nil
}

func decodeSource(raw json.RawMessage) (string, error) {
	var source struct {
		File string `json:"file"`
		Line int    `json:"line"`
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		return "", err
	}
	if source.File == "" {
		return "", nil
	}
	return fmt.Sprintf("%s:%d", source.File, source.Line), nil
}

// parseLevel maps the level text slog.JSONHandler writes to the coarser
// CourseLogEntry_Level; anything unrecognized reads as INFO.
func parseLevel(s string) qf.CourseLogEntry_Level {
	switch s {
	case "DEBUG":
		return qf.CourseLogEntry_DEBUG
	case "WARN":
		return qf.CourseLogEntry_WARN
	case "ERROR":
		return qf.CourseLogEntry_ERROR
	default:
		return qf.CourseLogEntry_INFO
	}
}

// stringify renders raw as the text an entry's Fields value should carry: a
// string value unquoted, anything else (numbers, booleans, objects) as its
// compact JSON text.
func stringify(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

// entryLimit keeps limit of the entries added to it, in the order added: the
// newest ones if newest is set, or else the oldest.
type entryLimit struct {
	limit  int
	newest bool
	buf    []*qf.CourseLogEntry
	start  int
	total  int
}

func newEntryLimit(limit int, newest bool) *entryLimit {
	if limit < 1 {
		limit = 1
	}
	return &entryLimit{limit: limit, newest: newest, buf: make([]*qf.CourseLogEntry, 0, limit)}
}

func (l *entryLimit) add(e *qf.CourseLogEntry) {
	l.total++
	switch {
	case len(l.buf) < l.limit:
		l.buf = append(l.buf, e)
	case l.newest:
		// A ring: the oldest kept entry makes room for the new one.
		l.buf[l.start] = e
		l.start = (l.start + 1) % l.limit
	}
}

// ordered returns the kept entries in the order added, and whether any entry
// was left out to stay within limit.
func (l *entryLimit) ordered() ([]*qf.CourseLogEntry, bool) {
	if l.total <= l.limit || l.start == 0 {
		return l.buf, l.total > l.limit
	}
	ordered := make([]*qf.CourseLogEntry, 0, len(l.buf))
	ordered = append(ordered, l.buf[l.start:]...)
	ordered = append(ordered, l.buf[:l.start]...)
	return ordered, true
}
