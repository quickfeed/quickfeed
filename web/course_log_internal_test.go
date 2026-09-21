package web

import (
	"fmt"
	"strconv"
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

// An entry logged before a subscription began but written after it is picked
// up by both the bounded backlog query and the live tail, and must be sent
// once. Remembering only a fixed number of the backlog's newest entries left a
// blind spot, since a record is timestamped before its handler takes the
// course file's lock and nothing bounds how far behind its timestamp it may be
// written.
func TestHandoffSuppressesDuplicatesAnywhereInTheBacklog(t *testing.T) {
	since := time.Date(2026, time.September, 21, 12, 0, 0, 0, time.UTC)
	record := func(i int) *qf.CourseLogEntry {
		return &qf.CourseLogEntry{
			Time:       timestamppb.New(since.Add(time.Duration(i-1000) * time.Millisecond)),
			Level:      qf.CourseLogEntry_INFO,
			Message:    fmt.Sprintf("record %d", i),
			Repository: "student-a",
			// A map field only hashes the same way every time if the encoding
			// is deterministic, which is the point of several keys here.
			Fields: map[string]string{"i": strconv.Itoa(i), "run": "1", "branch": "main"},
		}
	}
	backlog := make([]*qf.CourseLogEntry, 200)
	for i := range backlog {
		backlog[i] = record(i)
	}
	// The same record logged twice is in the backlog twice.
	backlog = append(backlog, record(0))

	sent := newHandoff(backlog, since)
	// The backlog's oldest entry is as much a duplicate as its newest.
	if !sent.duplicate(record(len(backlog) - 2)) {
		t.Error("duplicate(newest backlog entry) = false, want true")
	}
	if !sent.duplicate(record(0)) {
		t.Error("duplicate(oldest backlog entry) = false, want true")
	}
	// Two copies were sent, so two are suppressed.
	if !sent.duplicate(record(0)) {
		t.Error("duplicate(entry the backlog held twice, second time) = false, want true")
	}
	// The third copy is a record the client has not been sent.
	if sent.duplicate(record(0)) {
		t.Error("duplicate(entry the backlog held twice, third time) = true, want false")
	}
	// A record the backlog never held is not a duplicate.
	if sent.duplicate(record(len(backlog) + 1)) {
		t.Error("duplicate(entry not in the backlog) = true, want false")
	}
	// Nor is one logged after the subscription began, whatever it says: the
	// backlog is bounded by since, so it cannot have carried this one.
	later := record(0)
	later.Time = timestamppb.New(since.Add(time.Millisecond))
	if sent.duplicate(later) {
		t.Error("duplicate(entry logged after the subscription began) = true, want false")
	}
}

// A live entry is newer than the backlog's upper bound, but From can be set
// without To: a window beginning after the subscription did leaves a stretch
// the backlog query excluded, which the tail must not send in its place.
func TestDeliverableAppliesTheRequestedLowerBound(t *testing.T) {
	from := time.Date(2026, time.September, 22, 12, 0, 0, 0, time.UTC)
	at := func(d time.Duration) *qf.CourseLogEntry {
		return &qf.CourseLogEntry{
			Time:    timestamppb.New(from.Add(d)),
			Level:   qf.CourseLogEntry_INFO,
			Message: "a record",
		}
	}
	in := &qf.CourseLogRequest{From: timestamppb.New(from)}

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
