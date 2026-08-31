package courselog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Query returns org's entries matching req in chronological order, every
// repository with an entry timestamped within req's resolved interval
// regardless of req's repository or level filter (so a repository filter
// never removes its own options), and whether the match count exceeded req's
// resolved limit.
//
// req's interval and limit default and clamp the same way GetCourseLog
// documents: To is clamped to now and From to oldestRetainedDate. An interval
// left inverted after clamping returns an empty result.
//
// A malformed final line, characteristic of a partial write, is ignored; any
// other read failure is returned. A course with no activity in range, or no
// log at all, returns an empty result rather than an error.
func (s *Store) Query(org string, req *qf.CourseLogRequest) ([]*qf.CourseLogEntry, []string, bool, error) {
	now := s.now()
	from, to := req.Interval(oldestRetainedDate(now), now)
	if from.After(to) {
		return nil, nil, false, nil
	}

	dir := filepath.Join(s.dir, sanitize(org))
	ring := newEntryRing(req.EffectiveLimit())
	repos := make(map[string]bool)
	repository, level := req.GetRepository(), req.GetLevel()

	for _, date := range datesBetween(from, to) {
		path := filepath.Join(dir, date+".jsonl")
		if err := scanFile(path, from, to, repository, level, ring, repos); err != nil {
			return nil, nil, false, fmt.Errorf("reading course log %s: %w", path, err)
		}
	}

	entries, truncated := ring.ordered()
	repositories := make([]string, 0, len(repos))
	for repo := range repos {
		repositories = append(repositories, repo)
	}
	slices.Sort(repositories)
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

// scanFile decodes path line by line, adding entries matching [from, to],
// repository, and level to ring, and every repository with an entry
// timestamped within [from, to] to repos. A missing file means the course had
// no activity that day, not a failure.
func scanFile(path string, from, to time.Time, repository string, level qf.CourseLogEntry_Level, ring *entryRing, repos map[string]bool) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	// apply decodes line and folds it into ring/repos. final marks the file's
	// true last line: only there is a decode error tolerated, on the theory
	// that the process was killed mid-write, rather than returned as a
	// failure of the whole query.
	apply := func(line string, final bool) error {
		entry, err := decodeEntry(line)
		if err != nil {
			if final {
				return nil
			}
			return err
		}
		if entry.GetRepository() != "" && entry.InInterval(from, to) {
			repos[entry.GetRepository()] = true
		}
		if entry.Matches(from, to, repository, level) {
			ring.add(entry)
		}
		return nil
	}

	scanner := bufio.NewScanner(f)
	// A record can carry several fields independently truncated to 64 KiB
	// each (see maxFieldBytes), so grow well past both a single field and
	// bufio.Scanner's 64 KiB default cap; a line still over this is corrupt
	// rather than merely large, and fails the query as any other read error.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	// pending trails the scanner by one line, so a line is only ever applied
	// once Scan() has confirmed whether a further line follows it; that is
	// what lets the final line alone get the partial-write tolerance above,
	// without first holding the whole file (tens to hundreds of MiB for a
	// busy course) in memory to find out which line that was.
	var pending string
	havePending := false
	for scanner.Scan() {
		if havePending {
			if err := apply(pending, false); err != nil {
				return err
			}
		}
		pending, havePending = scanner.Text(), true
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if havePending {
		return apply(pending, true)
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

// entryRing keeps the newest limit entries added to it, in the order added.
type entryRing struct {
	limit int
	buf   []*qf.CourseLogEntry
	start int
	total int
}

func newEntryRing(limit int) *entryRing {
	if limit < 1 {
		limit = 1
	}
	return &entryRing{limit: limit, buf: make([]*qf.CourseLogEntry, 0, limit)}
}

func (r *entryRing) add(e *qf.CourseLogEntry) {
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
func (r *entryRing) ordered() ([]*qf.CourseLogEntry, bool) {
	if r.total <= r.limit {
		return r.buf, false
	}
	ordered := make([]*qf.CourseLogEntry, 0, len(r.buf))
	ordered = append(ordered, r.buf[r.start:]...)
	ordered = append(ordered, r.buf[:r.start]...)
	return ordered, true
}
