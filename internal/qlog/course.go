package qlog

import (
	"context"
	"log/slog"

	"github.com/quickfeed/quickfeed/internal/qlog/label"
)

// Course is the subset of qf.Course that WithCourse and WithCourseLog need.
// Declaring it here, rather than importing qf, keeps qlog free of a
// dependency on the database model package.
type Course interface {
	GetID() uint64
	GetCode() string
	GetScmOrganizationName() string
}

// CourseAttrs returns the attributes WithCourse attaches, for the rare caller
// that already holds an enriched logger it only wants to extend, rather than
// a context to enrich.
func CourseAttrs(course Course) []any {
	return []any{
		label.CourseID, course.GetID(),
		label.CourseCode, course.GetCode(),
		label.CourseLog, course.GetScmOrganizationName(),
	}
}

// WithCourse returns a context and logger scoped to course, in addition to
// attrs. Records logged through the returned logger are copied to the
// course's teacher-visible log, so course must come from the database, never
// from request data: nothing else can mark a scope for a course's log.
func WithCourse(ctx context.Context, course Course, attrs ...any) (context.Context, *slog.Logger) {
	return WithLogger(ctx, append(CourseAttrs(course), attrs...)...)
}

// WithCourseLog is WithCourse without CourseID, for the handful of RPC call
// paths where the logging interceptors already attached CourseID to the
// request logger from the caller's claims; see enrichRequestLogger. It still
// attaches CourseCode, which is not part of that RPC scope, and the course
// log marker.
func WithCourseLog(ctx context.Context, course Course, attrs ...any) (context.Context, *slog.Logger) {
	scoped := append([]any{
		label.CourseCode, course.GetCode(),
		label.CourseLog, course.GetScmOrganizationName(),
	}, attrs...)
	return WithLogger(ctx, scoped...)
}
