package web

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/courselog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/proto"
)

// handoffWindow is how many of the backlog's newest entries are remembered to
// suppress a duplicate from the live tail. An entry logged just before the
// subscription began but written just after it can appear in both; a handful
// of entries is more than the gap between the two can hold.
const handoffWindow = 50

// CourseLogStream streams the course's teacher-visible log for the requested
// interval, repository, and minimum level. The first message holds the backlog
// and the repositories to offer as filter options; From defaults to the last
// 24 hours and Limit to 2000, both clamped to their server-enforced maximums,
// with an over-limit backlog reported as Truncated.
//
// A request with To set is answered with that one message and the stream ends.
// A request that leaves To unset keeps the stream open, sending each later
// entry as it is logged, so a teacher watching a push sees it happen.
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

	if in.GetTo() != nil {
		backlog, err := s.courseLog(ctx, org, in)
		if err != nil {
			return err
		}
		return st.Send(backlog)
	}

	// Subscribe before reading the backlog, so an entry logged while the
	// backlog is being read is delivered rather than missed.
	sub := s.courseLogs.Subscribe(org)
	defer sub.Close()

	backlog, err := s.courseLog(ctx, org, in.WithTo(sub.Since()))
	if err != nil {
		return err
	}
	if err := st.Send(backlog); err != nil {
		return err
	}
	return s.tailCourseLog(ctx, st, sub, in, backlog)
}

// courseLog reads the stored entries matching in, as one stream message.
func (s *QuickFeedService) courseLog(ctx context.Context, org string, in *qf.CourseLogRequest) (*qf.CourseLog, error) {
	entries, repositories, truncated, err := s.courseLogs.Query(org, in)
	if err != nil {
		qlog.FromContext(ctx).Error("failed to read course log", label.Error, err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("reading course log"))
	}
	return &qf.CourseLog{Entries: entries, Repositories: repositories, Truncated: truncated}, nil
}

// tailCourseLog forwards entries as they are logged until the client
// disconnects or the store shuts down. Each message carries one entry, and
// names its repository only when the client has not been offered it yet.
func (*QuickFeedService) tailCourseLog(ctx context.Context, st *connect.ServerStream[qf.CourseLog], sub *courselog.Subscription, in *qf.CourseLogRequest, backlog *qf.CourseLog) error {
	seen := newSeenRepositories(backlog.GetRepositories())
	recent := backlog.GetEntries()
	if len(recent) > handoffWindow {
		recent = recent[len(recent)-handoffWindow:]
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case entry, ok := <-sub.C():
			if !ok {
				// The store is shutting down; end the stream cleanly so the
				// client reconnects rather than reporting a failure.
				return nil
			}
			if !deliverable(entry, in) {
				continue
			}
			if duplicateOf(entry, recent, sub.Since()) {
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

// deliverable reports whether a live entry belongs in a stream opened with
// in's filters. The stream's own report that it fell behind is always
// delivered: it says entries are missing, which is not something a repository
// or level filter may hide, and it carries neither of those to be judged on.
func deliverable(entry *qf.CourseLogEntry, in *qf.CourseLogRequest) bool {
	return courselog.Dropped(entry) || entry.MatchesFilters(in.GetRepository(), in.GetLevel())
}

// duplicateOf reports whether entry was already in the backlog. Only an entry
// timestamped at or before the subscription began can be, so a live tail costs
// one time comparison per entry once the handoff is past.
func duplicateOf(entry *qf.CourseLogEntry, recent []*qf.CourseLogEntry, since time.Time) bool {
	if entry.GetTime().AsTime().After(since) {
		return false
	}
	for _, seen := range recent {
		if proto.Equal(entry, seen) {
			return true
		}
	}
	return false
}

// seenRepositories tracks the repositories a client has been offered as filter
// options, so a message names one only when it is new.
type seenRepositories map[string]bool

func newSeenRepositories(repositories []string) seenRepositories {
	seen := make(seenRepositories, len(repositories))
	for _, repo := range repositories {
		seen[repo] = true
	}
	return seen
}

func (s seenRepositories) add(repository string) []string {
	if repository == "" || s[repository] {
		return nil
	}
	s[repository] = true
	return []string{repository}
}
