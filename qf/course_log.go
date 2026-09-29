package qf

import (
	"errors"
	"fmt"
	"math"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultCourseLogInterval = time.Hour
	defaultCourseLogLimit    = 2000
	maxCourseLogLimit        = 5000
)

// EndOfLogFile is a LogCursor offset past the end of any date file.
const EndOfLogFile = math.MaxInt64

// ErrInvalidLogCursor reports a cursor that the course log did not hand out.
var ErrInvalidLogCursor = errors.New("invalid course log cursor")

// Interval returns req's time bounds, kept within [minFrom, maxTo].
// To defaults to maxTo, and From to an hour before To.
// A From cursor yields minFrom, and a To cursor yields maxTo.
func (req *CourseLogRequest) Interval(minFrom, maxTo time.Time) (from, to time.Time) {
	to = maxTo
	if t := req.GetTo().GetTime(); t != nil {
		to = t.AsTime()
	}
	switch f := req.GetFrom(); {
	case f.GetTime() != nil:
		from = f.GetTime().AsTime()
	case f.GetCursor() != nil:
		from = minFrom
	default:
		from = to.Add(-defaultCourseLogInterval)
	}
	if to.After(maxTo) {
		to = maxTo
	}
	if from.Before(minFrom) {
		from = minFrom
	}
	return from, to
}

// FollowLog reports whether req asks for new entries as they are logged, which
// it does unless To is set.
func (req *CourseLogRequest) FollowLog() bool {
	return req.GetTo() == nil
}

// CheckCursors returns an error wrapping ErrInvalidLogCursor if req's From or
// To cursor is invalid or lies beyond now's date file, or if its From cursor
// lies beyond until.
func (req *CourseLogRequest) CheckCursors(until *LogCursor, now time.Time) error {
	after, before := req.GetFrom().GetCursor(), req.GetTo().GetCursor()
	for _, c := range []*LogCursor{after, before} {
		if c == nil {
			continue
		}
		if !c.IsValid() {
			return fmt.Errorf("%w: %v", ErrInvalidLogCursor, c)
		}
		if c.Day().After(LogDay(now)) {
			return beyondTheEnd(c)
		}
	}
	if after != nil && until != nil && after.Beyond(until) {
		return beyondTheEnd(after)
	}
	return nil
}

func beyondTheEnd(c *LogCursor) error {
	return fmt.Errorf("%w: offset %d of %s is beyond the end of the log", ErrInvalidLogCursor, c.GetOffset(), c.Day().Format(time.DateOnly))
}

// Start returns where reading req's range begins: at req's From cursor, or
// at the start of from's date file if that is later.
func (req *CourseLogRequest) Start(from time.Time) *LogCursor {
	start := NewLogCursor(from, 0)
	if c := req.GetFrom().GetCursor(); c != nil && c.Beyond(start) {
		return c
	}
	return start
}

// End returns where reading req's range ends: the earliest of req's To
// cursor, until, and the end of the date file that can hold a record stamped
// at to. It also reports whether the entry ending there is left out, which
// only a To cursor does. A record stamped just before midnight may be written
// to the next day's file, so the range reaches one day past to, unless that
// is beyond now.
func (req *CourseLogRequest) End(to, now time.Time, until *LogCursor) (*LogCursor, bool) {
	last := to.Add(24 * time.Hour)
	if last.After(now) {
		last = now
	}
	end := NewLogCursor(last, EndOfLogFile)
	if until != nil && end.Beyond(until) {
		end = until
	}
	if c := req.GetTo().GetCursor(); c != nil && !c.Beyond(end) {
		return c, true
	}
	return end, false
}

// EffectiveLimit returns req's requested entry limit, defaulting to 2000 and
// capped at 5000.
func (req *CourseLogRequest) EffectiveLimit() int {
	switch requested := req.GetLimit(); {
	case requested == 0:
		return defaultCourseLogLimit
	case requested > maxCourseLogLimit:
		return maxCourseLogLimit
	default:
		return int(requested)
	}
}

// HasMoreNewer reports whether entries newer than l's last one match but
// were left out to stay within the limit. A truncated backlog keeps its
// oldest matches unless newest is set, so it then ends before the newest
// match: with a limit of 2000 and 5000 matches from 09:00 to 10:00, it holds
// 09:00 to 09:20 and leaves out 09:20 to 10:00.
func (l *CourseLog) HasMoreNewer(newest bool) bool {
	return l.GetTruncated() && !newest
}

// InInterval reports whether e's timestamp falls within [from, to].
func (e *CourseLogEntry) InInterval(from, to time.Time) bool {
	t := e.GetTime().AsTime()
	return !t.Before(from) && !t.After(to)
}

// Matches reports whether e falls within [from, to], is at or above level,
// and, when repository is given, was recorded against it.
func (e *CourseLogEntry) Matches(from, to time.Time, repository string, level CourseLogEntry_Level) bool {
	return e.InInterval(from, to) && e.MatchesFilters(repository, level)
}

// MatchesFilters reports whether e is at or above level and, if repository
// is non-empty, belongs to repository.
func (e *CourseLogEntry) MatchesFilters(repository string, level CourseLogEntry_Level) bool {
	if e.GetLevel() < level {
		return false
	}
	return repository == "" || e.GetRepository() == repository
}

// NewLogCursor returns the cursor at offset in the date file of t's day.
func NewLogCursor(t time.Time, offset int64) *LogCursor {
	return &LogCursor{Date: timestamppb.New(LogDay(t)), Offset: uint64(offset)}
}

// TimePosition returns a LogPosition for t.
func TimePosition(t time.Time) *LogPosition {
	return &LogPosition{Position: &LogPosition_Time{Time: timestamppb.New(t)}}
}

// CursorPosition returns a LogPosition for c.
func CursorPosition(c *LogCursor) *LogPosition {
	return &LogPosition{Position: &LogPosition_Cursor{Cursor: c}}
}

// IsValid reports whether p holds a valid time or a valid cursor.
func (p *LogPosition) IsValid() bool {
	switch pos := p.GetPosition().(type) {
	case *LogPosition_Time:
		return pos.Time.IsValid()
	case *LogPosition_Cursor:
		return pos.Cursor.IsValid()
	default:
		return false
	}
}

// LogDay returns the UTC midnight of t's day, which names t's date file.
func LogDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// IsValid reports whether c's date is a UTC midnight and its offset fits in
// an int64.
func (c *LogCursor) IsValid() bool {
	date := c.GetDate()
	if !date.IsValid() || c.GetOffset() > math.MaxInt64 {
		return false
	}
	day := date.AsTime()
	return day.Equal(LogDay(day))
}

// Day returns the day of c's date file.
func (c *LogCursor) Day() time.Time {
	return c.GetDate().AsTime()
}

// Position returns c's offset as an int64, which is exact if c is valid.
func (c *LogCursor) Position() int64 {
	return int64(c.GetOffset())
}

// Beyond reports whether c comes after other in the log.
func (c *LogCursor) Beyond(other *LogCursor) bool {
	if day, otherDay := c.Day(), other.Day(); !day.Equal(otherDay) {
		return day.After(otherDay)
	}
	return c.GetOffset() > other.GetOffset()
}
