package web

import (
	"testing"

	"github.com/quickfeed/quickfeed/internal/qlog/courselog"
	"github.com/quickfeed/quickfeed/qf"
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
