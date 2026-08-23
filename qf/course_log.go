package qf

import "time"

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
