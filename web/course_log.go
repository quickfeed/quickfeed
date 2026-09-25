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

// CourseLogStream streams the course log entries matching the request's
// range, repository, and minimum level. The first message holds the backlog
// and the repositories available as filter options. From defaults to an hour
// ago and Limit to 2000; a backlog exceeding Limit is marked Truncated and
// holds the oldest entries, or the newest if Newest is set. A cursor the
// server did not hand out is rejected with InvalidArgument.
//
// If To is set, the stream ends after the first message. Otherwise, it stays
// open and sends each new entry as it is logged, unless the backlog was
// truncated before the present.
func (s *QuickFeedService) CourseLogStream(ctx context.Context, in *qf.CourseLogRequest, st *connect.ServerStream[qf.CourseLog]) error {
	logger := qlog.FromContext(ctx)
	logger.Debug("streaming course log", label.Repository, in.GetRepository())
	if s.courseLogs == nil {
		logger.Error("course log store not configured")
		return connect.NewError(connect.CodeInternal, errors.New("reading course log"))
	}
	course, err := s.db.GetCourse(in.GetCourseID())
	if err != nil {
		logger.Error("failed to get course", label.Error, err)
		return connect.NewError(connect.CodeNotFound, errors.New("course not found"))
	}
	org := course.GetScmOrganizationName()

	// If To is not set, the stream stays open and sends each new entry as it is logged.
	if in.GetTo() == nil {
		return s.followCourseLog(ctx, st, org, in)
	}
	// A closed range gets no new entries; send the backlog and end the stream.
	backlog, err := s.courseLog(ctx, org, in, nil)
	if err != nil {
		return err
	}
	return st.Send(backlog)
}

// followCourseLog sends the backlog up to now, followed by each entry logged
// afterwards, until the client disconnects or the store shuts down. A backlog
// cut short before the present ends the stream instead.
func (s *QuickFeedService) followCourseLog(ctx context.Context, st *connect.ServerStream[qf.CourseLog], org string, in *qf.CourseLogRequest) error {
	// Subscribe before querying, so entries logged during the query are
	// received on the subscription rather than lost.
	sub := s.courseLogs.Subscribe(org)
	defer sub.Close()

	// Entries after sub.Start() arrive on the subscription.
	backlog, err := s.courseLog(ctx, org, in, sub.Start())
	if err != nil {
		return err
	}
	if err := st.Send(backlog); err != nil {
		return err
	}
	// Following now would skip the entries the limit left out; the client
	// continues from the backlog's newest entry instead.
	if backlog.GetTruncated() && !in.GetNewest() {
		return nil
	}
	return tailCourseLog(ctx, st, sub, in, backlog)
}

// courseLog returns the stored entries matching in as a single message,
// ending at until if it is non-nil.
func (s *QuickFeedService) courseLog(ctx context.Context, org string, in *qf.CourseLogRequest, until *qf.LogCursor) (*qf.CourseLog, error) {
	entries, repositories, truncated, err := s.courseLogs.Query(org, in, until)
	if errors.Is(err, courselog.ErrInvalidCursor) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid course log cursor"))
	}
	if err != nil {
		qlog.FromContext(ctx).Error("failed to read course log", label.Error, err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("reading course log"))
	}
	return &qf.CourseLog{Entries: entries, Repositories: repositories, Truncated: truncated}, nil
}

// tailCourseLog sends each entry received on sub that matches in, one per
// message. A message lists the entry's repository only if the client has not
// been sent it before.
func tailCourseLog(ctx context.Context, st *connect.ServerStream[qf.CourseLog], sub *courselog.Subscription, in *qf.CourseLogRequest, backlog *qf.CourseLog) error {
	seen := newSeenRepositories(backlog.GetRepositories())

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case entry, ok := <-sub.C():
			if !ok {
				// The store is shutting down; end without an error so the
				// client reconnects.
				return nil
			}
			if !deliverable(entry, in) {
				continue
			}
			if err := st.Send(&qf.CourseLog{
				Entries:      []*qf.CourseLogEntry{entry},
				Repositories: seen.add(entry.GetRepository()),
			}); err != nil {
				return err
			}
		}
	}
}

// deliverable reports whether a live entry matches in's filters and From time.
// Gap reports are always delivered: they signal missing entries and have no
// repository or level to filter on.
func deliverable(entry *qf.CourseLogEntry, in *qf.CourseLogRequest) bool {
	if courselog.Dropped(entry) {
		return true
	}
	if from := in.GetFrom().GetTime(); from != nil && entry.GetTime().AsTime().Before(from.AsTime()) {
		return false
	}
	return entry.MatchesFilters(in.GetRepository(), in.GetLevel())
}

// seenRepositories records the repositories sent to the client as filter
// options.
type seenRepositories map[string]bool

func newSeenRepositories(repositories []string) seenRepositories {
	seen := make(seenRepositories, len(repositories))
	for _, repo := range repositories {
		seen[repo] = true
	}
	return seen
}

// add records repository and returns it as a one-element slice if it is new,
// or nil otherwise.
func (s seenRepositories) add(repository string) []string {
	if repository == "" || s[repository] {
		return nil
	}
	s[repository] = true
	return []string{repository}
}
