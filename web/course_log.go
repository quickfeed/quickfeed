package web

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/courselog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultCourseLogInterval = 24 * time.Hour
	defaultCourseLogLimit    = 2000
	maxCourseLogLimit        = 5000
)

// GetCourseLog returns the course's teacher-visible log for the requested
// interval, repository, and minimum level. From/To default to the last 24
// hours, and Limit to 2000; all three are clamped to their server-enforced
// maximums rather than rejected, matching how a result that still exceeds
// Limit is reported with Truncated instead of an error.
func (s *QuickFeedService) GetCourseLog(ctx context.Context, in *qf.CourseLogRequest) (*qf.CourseLog, error) {
	logger := qlog.FromContext(ctx)
	logger.Debug("fetching course log", label.Repository, in.GetRepository())
	if s.courseLogs == nil {
		logger.Error("course log store not configured")
		return nil, connect.NewError(connect.CodeInternal, errors.New("reading course log"))
	}
	course, err := s.db.GetCourse(in.GetCourseID())
	if err != nil {
		logger.Error("failed to get course", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("course not found"))
	}

	from, to := courseLogInterval(in)
	entries, repositories, truncated, err := s.courseLogs.Query(course.GetScmOrganizationName(), courselog.Query{
		From:       from,
		To:         to,
		Repository: in.GetRepository(),
		Level:      slogLevel(in.GetLevel()),
		Limit:      courseLogLimit(in.GetLimit()),
	})
	if err != nil {
		logger.Error("failed to read course log", label.Error, err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("reading course log"))
	}

	return &qf.CourseLog{
		Entries:      toCourseLogEntries(entries),
		Repositories: repositories,
		Truncated:    truncated,
	}, nil
}

// courseLogInterval applies in's defaults (the last 24 hours) and clamps From
// to the retention window, rather than rejecting a request for an interval
// wider than what the store can ever hold.
func courseLogInterval(in *qf.CourseLogRequest) (from, to time.Time) {
	to = time.Now()
	if t := in.GetTo(); t != nil {
		to = t.AsTime()
	}
	from = to.Add(-defaultCourseLogInterval)
	if f := in.GetFrom(); f != nil {
		from = f.AsTime()
	}
	if cutoff := time.Now().Add(-courselog.Retention); from.Before(cutoff) {
		from = cutoff
	}
	return from, to
}

func courseLogLimit(requested uint32) int {
	switch {
	case requested == 0:
		return defaultCourseLogLimit
	case requested > maxCourseLogLimit:
		return maxCourseLogLimit
	default:
		return int(requested)
	}
}

func slogLevel(level qf.CourseLogEntry_Level) slog.Level {
	switch level {
	case qf.CourseLogEntry_INFO:
		return slog.LevelInfo
	case qf.CourseLogEntry_WARN:
		return slog.LevelWarn
	case qf.CourseLogEntry_ERROR:
		return slog.LevelError
	default: // qf.CourseLogEntry_DEBUG
		return slog.LevelDebug
	}
}

func courseLogEntryLevel(level slog.Level) qf.CourseLogEntry_Level {
	switch {
	case level >= slog.LevelError:
		return qf.CourseLogEntry_ERROR
	case level >= slog.LevelWarn:
		return qf.CourseLogEntry_WARN
	case level >= slog.LevelInfo:
		return qf.CourseLogEntry_INFO
	default:
		return qf.CourseLogEntry_DEBUG
	}
}

func toCourseLogEntries(entries []courselog.Entry) []*qf.CourseLogEntry {
	out := make([]*qf.CourseLogEntry, len(entries))
	for i, e := range entries {
		out[i] = &qf.CourseLogEntry{
			Time:           timestamppb.New(e.Time),
			Level:          courseLogEntryLevel(e.Level),
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
