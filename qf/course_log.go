package qf

import (
	"math"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultCourseLogInterval = time.Hour
	defaultCourseLogLimit    = 2000
	maxCourseLogLimit        = 5000
)

// Interval returns req's time bounds, clamped to [minFrom, maxTo].
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
