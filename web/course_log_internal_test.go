package web

import (
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog/courselog"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// A teacher watching one repository, or only errors, must still be told when
// the stream fell behind: the report says the view is incomplete, so hiding it
// leaves a hole the teacher cannot see.
func TestDeliverableKeepsGapReportPastFilters(t *testing.T) {
	// courselog builds this entry itself and keeps the attribute naming it
	// private; the check below fails loudly if the two ever drift apart.
	gap := &qf.CourseLogEntry{
		Level:   qf.CourseLogEntry_WARN,
		Message: "course log stream fell behind; entries were not delivered",
		Fields:  map[string]string{"dropped": "3"},
	}
	if !courselog.Dropped(gap) {
		t.Fatal("courselog.Dropped() = false for this package's idea of a gap report; the attribute naming one has moved")
	}
	record := &qf.CourseLogEntry{
		Level:      qf.CourseLogEntry_WARN,
		Message:    "an ordinary record",
		Repository: "student-b",
	}

	tests := map[string]struct {
		entry *qf.CourseLogEntry
		in    *qf.CourseLogRequest
		want  bool
	}{
		"gap report, no filters":          {gap, &qf.CourseLogRequest{}, true},
		"gap report, other repository":    {gap, &qf.CourseLogRequest{Repository: "student-a"}, true},
		"gap report, level above its own": {gap, &qf.CourseLogRequest{Level: qf.CourseLogEntry_ERROR}, true},
		"record, no filters":              {record, &qf.CourseLogRequest{}, true},
		"record, its own repository":      {record, &qf.CourseLogRequest{Repository: "student-b"}, true},
		"record, other repository":        {record, &qf.CourseLogRequest{Repository: "student-a"}, false},
		"record, level above its own":     {record, &qf.CourseLogRequest{Level: qf.CourseLogEntry_ERROR}, false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := deliverable(tt.entry, tt.in); got != tt.want {
				t.Errorf("deliverable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// A live entry was written after the backlog's upper bound, but From can be
// set without To: a window beginning after the subscription did leaves a
// stretch the backlog query excluded, which the tail must not send in its
// place.
func TestDeliverableAppliesTheRequestedLowerBound(t *testing.T) {
	from := time.Date(2026, time.September, 22, 12, 0, 0, 0, time.UTC)
	at := func(d time.Duration) *qf.CourseLogEntry {
		return &qf.CourseLogEntry{
			Time:    timestamppb.New(from.Add(d)),
			Level:   qf.CourseLogEntry_INFO,
			Message: "a record",
		}
	}
	in := &qf.CourseLogRequest{From: qf.TimePosition(from)}

	if deliverable(at(-time.Second), in) {
		t.Error("deliverable(entry before From) = true, want false")
	}
	if !deliverable(at(0), in) {
		t.Error("deliverable(entry at From) = false, want true: the interval is inclusive at both ends")
	}
	if !deliverable(at(time.Second), in) {
		t.Error("deliverable(entry after From) = false, want true")
	}
	// Without From there is no lower bound to apply; Query's own default is
	// always in the past, so it can never withhold a live entry.
	if !deliverable(at(-time.Hour), &qf.CourseLogRequest{}) {
		t.Error("deliverable(entry, request without From) = false, want true")
	}
	// Every live entry was written after a From cursor, whatever its stamp.
	if !deliverable(at(-time.Hour), &qf.CourseLogRequest{From: qf.CursorPosition(qf.NewLogCursor(from, 42))}) {
		t.Error("deliverable(entry, request with a From cursor) = false, want true")
	}
	// A gap report says the view is incomplete, which an interval must not be
	// able to hide any more than a filter may.
	gap := at(-time.Hour)
	gap.Fields = map[string]string{"dropped": "3"}
	if !courselog.Dropped(gap) {
		t.Fatal("courselog.Dropped() = false for this package's idea of a gap report; the attribute naming one has moved")
	}
	if !deliverable(gap, in) {
		t.Error("deliverable(gap report before From) = false, want true")
	}
}
