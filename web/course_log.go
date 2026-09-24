package web

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/courselog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
)

// GetCourseLog returns the course's teacher-visible log for the requested
// range, repository, and minimum level. From defaults to an hour before To,
// To to now, and Limit to 2000; Store.Query clamps all three to their
// server-enforced maximums and reports an over-limit result as Truncated.
// A cursor the server did not hand out is rejected as InvalidArgument.
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

	entries, repositories, truncated, err := s.courseLogs.Query(course.GetScmOrganizationName(), in, nil)
	if errors.Is(err, courselog.ErrInvalidCursor) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid course log cursor"))
	}
	if err != nil {
		logger.Error("failed to read course log", label.Error, err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("reading course log"))
	}

	return &qf.CourseLog{
		Entries:      entries,
		Repositories: repositories,
		Truncated:    truncated,
	}, nil
}
