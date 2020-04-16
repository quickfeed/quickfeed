package ag

import "time"

const (
	layout = "2006-01-02T15:04:05"
)

// DurationUntilDeadline returns the duration since the deadline.
func (m Assignment) DurationUntilDeadline(now time.Time) time.Duration {
	deadline, err := time.Parse(layout, m.GetDeadline())
	if err != nil {
	}
	return now.Sub(deadline)
}
