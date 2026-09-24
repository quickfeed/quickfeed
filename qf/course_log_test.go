package qf_test

import (
	"math"
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// farPast and farFuture never bind Interval's [minFrom, maxTo] clamp in the
// tests below, isolating the defaulting behavior under test.
var (
	farPast   = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	farFuture = time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)
)

func TestCourseLogRequestIntervalDefaults(t *testing.T) {
	before := time.Now()
	from, to := (&qf.CourseLogRequest{}).Interval(farPast, time.Now())
	after := time.Now()

	if to.Before(before) || to.After(after) {
		t.Errorf("to = %v, want within [%v, %v]", to, before, after)
	}
	if got, want := to.Sub(from), time.Hour; got != want {
		t.Errorf("to.Sub(from) = %v, want the default interval %v", got, want)
	}
}

// TestCourseLogRequestIntervalHonorsExplicitBounds guards that Interval
// passes explicit From/To through unclamped when they fall within
// [minFrom, maxTo].
func TestCourseLogRequestIntervalHonorsExplicitBounds(t *testing.T) {
	to := time.Now().Add(-time.Hour)
	from := to.Add(-time.Hour)
	gotFrom, gotTo := (&qf.CourseLogRequest{
		From: qf.TimePosition(from),
		To:   qf.TimePosition(to),
	}).Interval(farPast, farFuture)
	if !gotFrom.Equal(from) || !gotTo.Equal(to) {
		t.Errorf("Interval() = (%v, %v), want (%v, %v)", gotFrom, gotTo, from, to)
	}
}

// TestCourseLogRequestIntervalClampsToBounds guards that an explicit From
// before minFrom, or a To after maxTo (including one defaulted from a To
// beyond maxTo), is clamped rather than passed through.
func TestCourseLogRequestIntervalClampsToBounds(t *testing.T) {
	minFrom := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	maxTo := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

	from, to := (&qf.CourseLogRequest{
		From: qf.TimePosition(minFrom.Add(-time.Hour)),
		To:   qf.TimePosition(maxTo.Add(time.Hour)),
	}).Interval(minFrom, maxTo)
	if !from.Equal(minFrom) || !to.Equal(maxTo) {
		t.Errorf("Interval() = (%v, %v), want (%v, %v)", from, to, minFrom, maxTo)
	}
}

// TestCourseLogRequestIntervalLeavesCursorSidesOpen guards that a cursor bound
// sets no time bound on its side, since the cursor alone decides where the
// range begins or ends; the other side's time still applies.
func TestCourseLogRequestIntervalLeavesCursorSidesOpen(t *testing.T) {
	minFrom := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	maxTo := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	at := maxTo.Add(-time.Hour)
	cursor := qf.CursorPosition(qf.NewLogCursor(at, 42))

	tests := map[string]struct {
		req              *qf.CourseLogRequest
		wantFrom, wantTo time.Time
	}{
		"from a cursor":            {&qf.CourseLogRequest{From: cursor}, minFrom, maxTo},
		"from a cursor, to a time": {&qf.CourseLogRequest{From: cursor, To: qf.TimePosition(at)}, minFrom, at},
		"to a cursor":              {&qf.CourseLogRequest{To: cursor}, maxTo.Add(-time.Hour), maxTo},
		"from a time, to a cursor": {&qf.CourseLogRequest{From: qf.TimePosition(at), To: cursor}, at, maxTo},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			from, to := tt.req.Interval(minFrom, maxTo)
			if !from.Equal(tt.wantFrom) || !to.Equal(tt.wantTo) {
				t.Errorf("Interval() = (%v, %v), want (%v, %v)", from, to, tt.wantFrom, tt.wantTo)
			}
		})
	}
}

func TestCourseLogRequestEffectiveLimit(t *testing.T) {
	const maxCourseLogLimit = 5000
	tests := []struct {
		requested uint32
		want      int
	}{
		{requested: 0, want: 2000},
		{requested: 10, want: 10},
		{requested: maxCourseLogLimit, want: maxCourseLogLimit},
		{requested: maxCourseLogLimit + 1000, want: maxCourseLogLimit},
	}
	for _, test := range tests {
		req := &qf.CourseLogRequest{Limit: test.requested}
		if got := req.EffectiveLimit(); got != test.want {
			t.Errorf("EffectiveLimit(%d) = %d, want %d", test.requested, got, test.want)
		}
	}
}

func TestCourseLogEntryInInterval(t *testing.T) {
	base := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		at   time.Time
		want bool
	}{
		{name: "before", at: base.Add(-time.Minute), want: false},
		{name: "at from", at: base, want: true},
		{name: "inside", at: base.Add(30 * time.Second), want: true},
		{name: "at to", at: base.Add(time.Minute), want: true},
		{name: "after", at: base.Add(2 * time.Minute), want: false},
	}
	for _, test := range tests {
		e := &qf.CourseLogEntry{Time: timestamppb.New(test.at)}
		if got := e.InInterval(base, base.Add(time.Minute)); got != test.want {
			t.Errorf("%s: InInterval() = %v, want %v", test.name, got, test.want)
		}
	}
}

func TestCourseLogEntryMatches(t *testing.T) {
	base := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	from, to := base, base.Add(time.Hour)

	tests := []struct {
		name       string
		e          *qf.CourseLogEntry
		repository string
		want       bool
	}{
		{
			name: "outside interval",
			e:    &qf.CourseLogEntry{Time: timestamppb.New(base.Add(-time.Minute)), Level: qf.CourseLogEntry_INFO},
			want: false,
		},
		{
			name: "below minimum level",
			e:    &qf.CourseLogEntry{Time: timestamppb.New(base), Level: qf.CourseLogEntry_DEBUG},
			want: false,
		},
		{
			name:       "wrong repository",
			e:          &qf.CourseLogEntry{Time: timestamppb.New(base), Level: qf.CourseLogEntry_INFO, Repository: "repo-b"},
			repository: "repo-a",
			want:       false,
		},
		{
			name:       "matches with repository filter",
			e:          &qf.CourseLogEntry{Time: timestamppb.New(base), Level: qf.CourseLogEntry_WARN, Repository: "repo-a"},
			repository: "repo-a",
			want:       true,
		},
		{
			name: "matches with no repository filter",
			e:    &qf.CourseLogEntry{Time: timestamppb.New(base), Level: qf.CourseLogEntry_INFO, Repository: "repo-b"},
			want: true,
		},
	}
	for _, test := range tests {
		if got := test.e.Matches(from, to, test.repository, qf.CourseLogEntry_INFO); got != test.want {
			t.Errorf("%s: Matches() = %v, want %v", test.name, got, test.want)
		}
	}
}

func TestNewLogCursor(t *testing.T) {
	day := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	oslo := time.FixedZone("CET", 60*60)
	tests := map[string]struct {
		at      time.Time
		wantDay time.Time
	}{
		"midnight":                  {at: day, wantDay: day},
		"during the day":            {at: day.Add(12 * time.Hour), wantDay: day},
		"just before midnight":      {at: day.Add(24*time.Hour - time.Nanosecond), wantDay: day},
		"the day in UTC, not local": {at: time.Date(2026, 3, 11, 0, 30, 0, 0, oslo), wantDay: day},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := qf.NewLogCursor(tt.at, 42)
			if !c.Day().Equal(tt.wantDay) || c.Position() != 42 {
				t.Errorf("NewLogCursor(%v, 42) = (%v, %d), want (%v, 42)", tt.at, c.Day(), c.Position(), tt.wantDay)
			}
			if !c.IsValid() {
				t.Errorf("NewLogCursor(%v, 42).IsValid() = false, want true", tt.at)
			}
		})
	}
}

func TestLogCursorIsValid(t *testing.T) {
	midnight := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	tests := map[string]struct {
		c    *qf.LogCursor
		want bool
	}{
		"start of a date file":   {c: &qf.LogCursor{Date: timestamppb.New(midnight)}, want: true},
		"within a date file":     {c: &qf.LogCursor{Date: timestamppb.New(midnight), Offset: 1234}, want: true},
		"largest file position":  {c: &qf.LogCursor{Date: timestamppb.New(midnight), Offset: math.MaxInt64}, want: true},
		"nil":                    {c: nil, want: false},
		"no date":                {c: &qf.LogCursor{Offset: 1234}, want: false},
		"date not at midnight":   {c: &qf.LogCursor{Date: timestamppb.New(midnight.Add(time.Hour))}, want: false},
		"a nanosecond past":      {c: &qf.LogCursor{Date: timestamppb.New(midnight.Add(time.Nanosecond))}, want: false},
		"a local midnight":       {c: &qf.LogCursor{Date: timestamppb.New(time.Date(2026, 3, 10, 0, 0, 0, 0, time.FixedZone("CET", 60*60)))}, want: false},
		"not a timestamp":        {c: &qf.LogCursor{Date: &timestamppb.Timestamp{Seconds: midnight.Unix(), Nanos: -1}}, want: false},
		"beyond a file position": {c: &qf.LogCursor{Date: timestamppb.New(midnight), Offset: math.MaxInt64 + 1}, want: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.c.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogCursorBeyond(t *testing.T) {
	day := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	at := qf.NewLogCursor
	tests := map[string]struct {
		c, other *qf.LogCursor
		want     bool
	}{
		"further into the same file": {c: at(day, 20), other: at(day, 10), want: true},
		"earlier in the same file":   {c: at(day, 10), other: at(day, 20), want: false},
		"the same position":          {c: at(day, 10), other: at(day, 10), want: false},
		// A new day's file starts from nothing, so its offsets are no smaller
		// or larger than the previous day's in any sense that orders them.
		"a later day, smaller offset":   {c: at(day.AddDate(0, 0, 1), 0), other: at(day, 1000), want: true},
		"an earlier day, larger offset": {c: at(day.AddDate(0, 0, -1), 1000), other: at(day, 0), want: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.c.Beyond(tt.other); got != tt.want {
				t.Errorf("Beyond() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogPositionIsValid(t *testing.T) {
	now := time.Now()
	tests := map[string]struct {
		p    *qf.LogPosition
		want bool
	}{
		"a time":               {p: qf.TimePosition(now), want: true},
		"a cursor":             {p: qf.CursorPosition(qf.NewLogCursor(now, 42)), want: true},
		"nil":                  {p: nil, want: false},
		"neither":              {p: &qf.LogPosition{}, want: false},
		"an invalid time":      {p: &qf.LogPosition{Position: &qf.LogPosition_Time{Time: &timestamppb.Timestamp{Nanos: -1}}}, want: false},
		"an invalid cursor":    {p: qf.CursorPosition(&qf.LogCursor{Offset: 42}), want: false},
		"a cursor that is nil": {p: qf.CursorPosition(nil), want: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.p.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCourseLogRequestIsValid(t *testing.T) {
	now := time.Now()
	hourAgo := now.Add(-time.Hour)
	earlier, later := qf.NewLogCursor(now, 10), qf.NewLogCursor(now, 20)
	pos := qf.CursorPosition
	tests := map[string]struct {
		req  *qf.CourseLogRequest
		want bool
	}{
		"course only":               {req: &qf.CourseLogRequest{CourseID: 1}, want: true},
		"no course":                 {req: &qf.CourseLogRequest{}, want: false},
		"from before to":            {req: &qf.CourseLogRequest{CourseID: 1, From: qf.TimePosition(hourAgo), To: qf.TimePosition(now)}, want: true},
		"from after to":             {req: &qf.CourseLogRequest{CourseID: 1, From: qf.TimePosition(now), To: qf.TimePosition(hourAgo)}, want: false},
		"a valid from cursor":       {req: &qf.CourseLogRequest{CourseID: 1, From: pos(earlier)}, want: true},
		"a valid to cursor":         {req: &qf.CourseLogRequest{CourseID: 1, To: pos(earlier)}, want: true},
		"a cursor with no date":     {req: &qf.CourseLogRequest{CourseID: 1, From: pos(&qf.LogCursor{Offset: 42})}, want: false},
		"a cursor off midnight":     {req: &qf.CourseLogRequest{CourseID: 1, To: pos(&qf.LogCursor{Date: timestamppb.New(qf.LogDay(now).Add(time.Hour))})}, want: false},
		"a position naming neither": {req: &qf.CourseLogRequest{CourseID: 1, From: &qf.LogPosition{}}, want: false},
		"from cursor before to":     {req: &qf.CourseLogRequest{CourseID: 1, From: pos(earlier), To: pos(later)}, want: true},
		"from cursor at to":         {req: &qf.CourseLogRequest{CourseID: 1, From: pos(earlier), To: pos(earlier)}, want: true},
		"from cursor beyond to":     {req: &qf.CourseLogRequest{CourseID: 1, From: pos(later), To: pos(earlier)}, want: false},
		// A cursor and a time cannot be compared without reading the log.
		"a cursor and a time": {req: &qf.CourseLogRequest{CourseID: 1, From: pos(later), To: qf.TimePosition(hourAgo)}, want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.req.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
