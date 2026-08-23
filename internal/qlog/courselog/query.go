package courselog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog/label"
)

// Entry is one decoded course log record. Fields holds every attribute not
// already promoted to a dedicated field, keyed by its label, with values
// stringified as their JSON text (so a number reads as "42", a string
// unquoted, and anything else as compact JSON).
type Entry struct {
	Time           time.Time
	Level          slog.Level
	Message        string
	Source         string // repository-relative file:line
	Repository     string
	RepositoryType string
	Truncated      bool
	Fields         map[string]string
}

// Query selects entries from a course's log.
type Query struct {
	From, To   time.Time
	Repository string // exact match; zero value selects every repository
	Level      slog.Level
	Limit      int // must be positive; the caller is responsible for defaults and maximums
}

// Query returns org's entries matching q in chronological order, every
// repository with an entry timestamped within [q.From, q.To] regardless of
// q.Repository or q.Level (so a repository filter never removes its own
// options), and whether the match count exceeded q.Limit.
//
// q.To is clamped to now, and q.From to oldestRetainedDate. An interval left
// inverted after clamping returns an empty result.
//
// A malformed final line, characteristic of a partial write, is ignored; any
// other read failure is returned. A course with no activity in range, or no
// log at all, returns an empty result rather than an error.
func (s *Store) Query(org string, q Query) ([]Entry, []string, bool, error) {
	now := s.now()
	if q.To.After(now) {
		q.To = now
	}
	if cutoff := oldestRetainedDate(now); q.From.Before(cutoff) {
		q.From = cutoff
	}
	if q.From.After(q.To) {
		return nil, nil, false, nil
	}

	dir := filepath.Join(s.dir, sanitize(org))
	ring := newEntryRing(q.Limit)
	repos := make(map[string]bool)

	for _, date := range datesBetween(q.From, q.To) {
		path := filepath.Join(dir, date+".jsonl")
		if err := scanFile(path, q, ring, repos); err != nil {
			return nil, nil, false, fmt.Errorf("reading course log %s: %w", path, err)
		}
	}

	entries, truncated := ring.ordered()
	repositories := make([]string, 0, len(repos))
	for repo := range repos {
		repositories = append(repositories, repo)
	}
	sort.Strings(repositories)
	return entries, repositories, truncated, nil
}

// oldestRetainedDate returns the UTC midnight of the oldest date
// cleanupCourseDir still guarantees to keep, as of now.
func oldestRetainedDate(now time.Time) time.Time {
	cutoff := now.Add(-Retention).UTC()
	midnight := time.Date(cutoff.Year(), cutoff.Month(), cutoff.Day(), 0, 0, 0, 0, time.UTC)
	if midnight.Before(cutoff) {
		return midnight.AddDate(0, 0, 1)
	}
	return midnight
}

// datesBetween returns the UTC calendar dates, inclusive, that could hold
// records timestamped within [from, to].
func datesBetween(from, to time.Time) []string {
	from, to = from.UTC(), to.UTC()
	first := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	var dates []string
	for d := first; !d.After(to); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format(dateLayout))
	}
	return dates
}

// scanFile decodes path line by line, adding matching entries to ring and
// every in-range repository to repos. A missing file means the course had no
// activity that day, not a failure.
func scanFile(path string, q Query, ring *entryRing, repos map[string]bool) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	// A line may hold a message or attribute truncated up to 64 KiB, plus its
	// surrounding JSON, so grow well past bufio.Scanner's 64 KiB default cap.
	scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	for i, line := range lines {
		entry, err := decodeEntry(line)
		if err != nil {
			if i == len(lines)-1 {
				// The process may have been killed mid-write; the last line
				// of the most recent file is the only one this can affect.
				return nil
			}
			return err
		}
		if entry.Repository != "" && !entry.Time.Before(q.From) && !entry.Time.After(q.To) {
			repos[entry.Repository] = true
		}
		if matches(entry, q) {
			ring.add(entry)
		}
	}
	return nil
}

func matches(e Entry, q Query) bool {
	if e.Time.Before(q.From) || e.Time.After(q.To) {
		return false
	}
	if e.Level < q.Level {
		return false
	}
	if q.Repository != "" && e.Repository != q.Repository {
		return false
	}
	return true
}

// decodeEntry parses one JSONL record written by Handler. slog.JSONHandler's
// standard keys (time, level, msg, source) and the labels Handler promotes to
// dedicated Entry fields are recognized by name; everything else becomes a
// Fields entry.
func decodeEntry(line string) (Entry, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return Entry{}, err
	}

	entry := Entry{Fields: make(map[string]string, len(raw))}
	for key, value := range raw {
		var err error
		switch key {
		case slog.TimeKey:
			err = json.Unmarshal(value, &entry.Time)
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
			return Entry{}, fmt.Errorf("decoding %q: %w", key, err)
		}
	}
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

func parseLevel(s string) slog.Level {
	switch s {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// stringify renders raw as the text an Entry.Fields value should carry: a
// string value unquoted, anything else (numbers, booleans, objects) as its
// compact JSON text.
func stringify(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

// entryRing keeps the newest limit entries added to it, in the order added.
type entryRing struct {
	limit int
	buf   []Entry
	start int
	total int
}

func newEntryRing(limit int) *entryRing {
	if limit < 1 {
		limit = 1
	}
	return &entryRing{limit: limit, buf: make([]Entry, 0, limit)}
}

func (r *entryRing) add(e Entry) {
	r.total++
	if len(r.buf) < r.limit {
		r.buf = append(r.buf, e)
		return
	}
	r.buf[r.start] = e
	r.start = (r.start + 1) % r.limit
}

// ordered returns the retained entries in chronological order, and whether
// any entry was evicted to stay within limit.
func (r *entryRing) ordered() ([]Entry, bool) {
	if r.total <= r.limit {
		return r.buf, false
	}
	ordered := make([]Entry, 0, len(r.buf))
	ordered = append(ordered, r.buf[r.start:]...)
	ordered = append(ordered, r.buf[:r.start]...)
	return ordered, true
}
