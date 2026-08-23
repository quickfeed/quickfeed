package qf

import "time"

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
