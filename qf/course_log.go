package qf

import (
	"log/slog"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog/courselog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultCourseLogInterval = 24 * time.Hour
	defaultCourseLogLimit    = 2000
	maxCourseLogLimit        = 5000
)

// Interval returns req's query bounds, defaulting to the last 24 hours ending
// now when From/To are not given. From/To are otherwise returned unclamped;
// courselog.Store.Query clamps them to their server-enforced maximums.
func (req *CourseLogRequest) Interval() (from, to time.Time) {
	to = time.Now()
	if t := req.GetTo(); t != nil {
		to = t.AsTime()
	}
	from = to.Add(-defaultCourseLogInterval)
	if f := req.GetFrom(); f != nil {
		from = f.AsTime()
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

// SlogLevel converts level to its slog.Level equivalent.
func (level CourseLogEntry_Level) SlogLevel() slog.Level {
	switch level {
	case CourseLogEntry_INFO:
		return slog.LevelInfo
	case CourseLogEntry_WARN:
		return slog.LevelWarn
	case CourseLogEntry_ERROR:
		return slog.LevelError
	default: // CourseLogEntry_DEBUG
		return slog.LevelDebug
	}
}

// CourseLogLevelFromSlog converts level to the coarser CourseLogEntry_Level.
func CourseLogLevelFromSlog(level slog.Level) CourseLogEntry_Level {
	switch {
	case level >= slog.LevelError:
		return CourseLogEntry_ERROR
	case level >= slog.LevelWarn:
		return CourseLogEntry_WARN
	case level >= slog.LevelInfo:
		return CourseLogEntry_INFO
	default:
		return CourseLogEntry_DEBUG
	}
}

// CourseLogEntriesFrom converts entries, in order, to their proto representation.
func CourseLogEntriesFrom(entries []courselog.Entry) []*CourseLogEntry {
	out := make([]*CourseLogEntry, len(entries))
	for i, e := range entries {
		out[i] = &CourseLogEntry{
			Time:           timestamppb.New(e.Time),
			Level:          CourseLogLevelFromSlog(e.Level),
			Message:        e.Message,
			Source:         e.Source,
			Repository:     e.Repository,
			RepositoryType: e.RepositoryType,
			Fields:         e.Fields,
			Truncated:      e.Truncated,
		}
	}
	return out
}
