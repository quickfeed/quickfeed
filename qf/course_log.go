package qf

import (
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultCourseLogInterval = 24 * time.Hour
	defaultCourseLogLimit    = 2000
	maxCourseLogLimit        = 5000
)

// Interval returns req's query bounds, clamped to [minFrom, maxTo]. From/To
// default to the last 24 hours ending at maxTo when not given.
func (req *CourseLogRequest) Interval(minFrom, maxTo time.Time) (from, to time.Time) {
	to = maxTo
	if t := req.GetTo(); t != nil {
		to = t.AsTime()
	}
	from = to.Add(-defaultCourseLogInterval)
	if f := req.GetFrom(); f != nil {
		from = f.AsTime()
	}
	if to.After(maxTo) {
		to = maxTo
	}
	if from.Before(minFrom) {
		from = minFrom
	}
	return from, to
}

// WithTo returns a copy of req bounded above by to, leaving req untouched.
func (req *CourseLogRequest) WithTo(to time.Time) *CourseLogRequest {
	bounded := proto.CloneOf(req)
	bounded.To = timestamppb.New(to)
	return bounded
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

// MatchesFilters reports whether e is at or above level and, when repository
// is given, was recorded against it. It leaves the interval out, for a live
// tail where every entry is newer than the query's upper bound by definition.
func (e *CourseLogEntry) MatchesFilters(repository string, level CourseLogEntry_Level) bool {
	if e.GetLevel() < level {
		return false
	}
	return repository == "" || e.GetRepository() == repository
}
