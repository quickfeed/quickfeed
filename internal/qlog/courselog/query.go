package courselog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"iter"
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
// The result lists every repository with an entry in req's time range,
// including those excluded by req's repository and level filters, so the
// client can offer them all as filter options. If more entries match than
// req's limit, the result is marked truncated and holds the oldest matches,
// or the newest if req.Newest is set.
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
	for day := range r.days() {
		span := r.span(day)
		path := s.path(org, day)
		if err := sc.file(path, span); err != nil {
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

// readRange holds the file positions and timestamp bounds for a query.
// If excludeEnd is set, the entry ending at end is left out.
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
	sp := span{day: day, start: 0, end: qf.EndOfLogFile}
	if day.Equal(r.start.Day()) {
		sp.start = r.start.Position()
	}
	if day.Equal(r.end.Day()) {
		sp.end, sp.exclusive = r.end.Position(), r.excludeEnd
	}
	return sp
}

// span holds the day and byte range of one log file read by a query.
// If exclusive is set, the record ending at end is left out.
type span struct {
	day        time.Time
	start, end int64
	exclusive  bool
}

// oldestRetainedDate returns the UTC midnight of the oldest date
// cleanupCourseDir guarantees to keep.
func oldestRetainedDate(now time.Time) time.Time {
	cutoff := now.Add(-Retention).UTC()
	midnight := qf.LogDay(cutoff)
	if midnight.Before(cutoff) {
		return midnight.AddDate(0, 0, 1)
	}
	return midnight
}

// days yields each UTC day from r's start through its end, inclusive.
func (r *readRange) days() iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		for day := r.start.Day(); !day.After(r.end.Day()); day = day.AddDate(0, 0, 1) {
			if !yield(day) {
				return
			}
		}
	}
}

// scan holds a query's filters and entries collected across date files.
type scan struct {
	from, to   time.Time
	repository string
	level      qf.CourseLogEntry_Level
	kept       *entryLimit
	repos      map[string]bool
}

// file reads sp's part of path, adding matching entries to kept and all
// repositories with entries in the time range to repos.
// A missing file means the course had no activity that day.
func (sc *scan) file(path string, sp span) error {
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

	// Keep one line pending until the next is scanned. Only the final line
	// may be ignored if a partial write left it malformed.
	var pending string
	var pendingEnd int64
	havePending := false
	for scanner.Scan() {
		if havePending {
			if err := sc.apply(pending, sp, pendingEnd, false); err != nil {
				return err
			}
		}
		pending, pendingEnd, havePending = scanner.Text(), offset, true
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if havePending {
		return sc.apply(pending, sp, pendingEnd, true)
	}
	return nil
}

// apply decodes line into kept and repos, setting its cursor to end in sp's day.
// A malformed final line is ignored because a write may have stopped midway.
func (sc *scan) apply(line string, sp span, end int64, final bool) error {
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
	entry.Cursor = qf.NewLogCursor(sp.day, end)
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

// atRecordBoundary returns an error wrapping qf.ErrInvalidLogCursor if offset
// is neither zero nor immediately after a newline in f.
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

// parseLevel maps slog.JSONHandler's level text to CourseLogEntry_Level.
// Unrecognized levels map to INFO.
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

// entryLimit retains up to limit entries, keeping the newest if newest is set
// and the oldest otherwise.
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

// ordered returns the retained entries in insertion order and reports whether
// any were discarded.
func (l *entryLimit) ordered() ([]*qf.CourseLogEntry, bool) {
	if l.total <= l.limit || l.start == 0 {
		return l.buf, l.total > l.limit
	}
	ordered := make([]*qf.CourseLogEntry, 0, len(l.buf))
	ordered = append(ordered, l.buf[l.start:]...)
	ordered = append(ordered, l.buf[:l.start]...)
	return ordered, true
}
