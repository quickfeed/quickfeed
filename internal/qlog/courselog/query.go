package courselog

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ErrInvalidCursor reports a cursor that the store did not hand out.
var ErrInvalidCursor = errors.New("invalid course log cursor")

// Query returns org's entries matching req in the order they were written,
// every repository with an entry timestamped within req's resolved interval
// regardless of req's repository or level filter (so a repository filter
// never removes its own options), and whether the match count exceeded req's
// resolved limit. Over the limit, req's Newest keeps the newest matches;
// otherwise the oldest are kept. Each entry carries its cursor.
//
// req's interval and limit default and clamp the same way CourseLogStream
// documents: To is clamped to now and From to oldestRetainedDate. An interval
// left inverted after clamping returns an empty result.
//
// From and To may be cursors, which leave out the entry they came from. If
// until is non-nil, the result ends at until, inclusive. A cursor the store
// did not hand out fails with an error wrapping ErrInvalidCursor.
//
// A malformed final line, characteristic of a partial write, is ignored; any
// other read failure is returned. A course with no activity in range, or no
// log at all, returns an empty result rather than an error.
func (s *Store) Query(org string, req *qf.CourseLogRequest, until *qf.LogCursor) ([]*qf.CourseLogEntry, []string, bool, error) {
	after, before := req.GetFrom().GetCursor(), req.GetTo().GetCursor()
	now := s.now()
	if err := checkCursors(after, before, until, now); err != nil {
		return nil, nil, false, err
	}
	from, to := req.Interval(oldestRetainedDate(now), now)
	if from.After(to) {
		return nil, nil, false, nil
	}
	last := lastFiled(to, now)
	if before != nil && before.Day().Before(last) {
		last = before.Day()
	}
	upper, exclusive := until, false
	if before != nil && (upper == nil || !before.Beyond(upper)) {
		upper, exclusive = before, true
	}

	org = sanitize(org)
	sc := &scan{
		from:       from,
		to:         to,
		repository: req.GetRepository(),
		level:      req.GetLevel(),
		kept:       newEntryLimit(req.EffectiveLimit(), req.GetNewest()),
		repos:      make(map[string]bool),
	}
	for _, day := range daysBetween(from, last) {
		sp, ok := spanOf(day, after, upper, exclusive)
		if !ok {
			continue
		}
		path := s.path(org, day)
		if err := sc.file(path, day, sp); err != nil {
			return nil, nil, false, fmt.Errorf("reading course log %s: %w", path, err)
		}
	}

	entries, truncated := sc.kept.ordered()
	repositories := make([]string, 0, len(sc.repos))
	for repo := range sc.repos {
		repositories = append(repositories, repo)
	}
	slices.Sort(repositories)
	return entries, repositories, truncated, nil
}

// checkCursors returns an error wrapping ErrInvalidCursor if after or before
// is invalid or lies beyond today, or if after lies beyond until.
func checkCursors(after, before, until *qf.LogCursor, now time.Time) error {
	for _, c := range []*qf.LogCursor{after, before} {
		if c == nil {
			continue
		}
		if !c.IsValid() {
			return fmt.Errorf("%w: %v", ErrInvalidCursor, c)
		}
		if c.Day().After(qf.LogDay(now)) {
			return beyondTheEnd(c)
		}
	}
	if after != nil && until != nil && after.Beyond(until) {
		return beyondTheEnd(after)
	}
	return nil
}

func beyondTheEnd(c *qf.LogCursor) error {
	return fmt.Errorf("%w: offset %d of %s is beyond the end of the log", ErrInvalidCursor, c.GetOffset(), c.Day().Format(dateLayout))
}

// lastFiled returns the latest instant whose date file can hold a record
// stamped at or before to. A record stamped just before midnight may be
// written to the next day's file.
func lastFiled(to, now time.Time) time.Time {
	if next := to.Add(24 * time.Hour); next.Before(now) {
		return next
	}
	return now
}

// span is the byte range of one date file a query reads: from start, and up
// to end unless end is negative. If exclusive is set, the record ending at end
// is left out, which is how a cursor bound leaves out the entry it came from.
type span struct {
	start, end int64
	exclusive  bool
}

// spanOf returns the part of day's file after after and up to upper, or false
// if that part is empty. A nil cursor bounds nothing.
func spanOf(day time.Time, after, upper *qf.LogCursor, exclusive bool) (span, bool) {
	sp := span{start: 0, end: -1}
	if after != nil {
		switch {
		case day.Before(after.Day()):
			return span{}, false
		case day.Equal(after.Day()):
			sp.start = after.Position()
		}
	}
	if upper != nil {
		switch {
		case day.After(upper.Day()):
			return span{}, false
		case day.Equal(upper.Day()):
			sp.end, sp.exclusive = upper.Position(), exclusive
		}
	}
	if sp.end >= 0 && sp.end <= sp.start {
		return span{}, false
	}
	return sp, true
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
	if sp.end < 0 {
		return f, nil
	}
	return io.LimitReader(f, sp.end-sp.start), nil
}

// atRecordBoundary returns an error wrapping ErrInvalidCursor unless offset
// is zero or just past a newline in f.
func atRecordBoundary(f *os.File, offset int64) error {
	if offset == 0 {
		return nil
	}
	var last [1]byte
	if _, err := f.ReadAt(last[:], offset-1); err != nil || last[0] != '\n' {
		return fmt.Errorf("%w: offset %d is not at a record boundary", ErrInvalidCursor, offset)
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
