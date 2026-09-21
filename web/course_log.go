package web

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/courselog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/proto"
)

// CourseLogStream streams the course log entries matching the request's
// interval, repository, and minimum level. The first message holds the
// backlog and the repositories available as filter options. From defaults to
// 24 hours ago and Limit to 2000; a backlog exceeding Limit is marked Truncated.
//
// If To is set, the stream ends after the first message. Otherwise, it stays
// open and sends each new entry as it is logged.
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
	// A closed interval gets no new entries; send the backlog and end the stream.
	backlog, err := s.courseLog(ctx, org, in)
	if err != nil {
		return err
	}
	return st.Send(backlog)
}

// followCourseLog sends the backlog up to now, followed by each entry logged
// afterwards, until the client disconnects or the store shuts down.
func (s *QuickFeedService) followCourseLog(ctx context.Context, st *connect.ServerStream[qf.CourseLog], org string, in *qf.CourseLogRequest) error {
	// Subscribe before querying, so entries logged during the query are
	// received on the subscription rather than lost.
	sub := s.courseLogs.Subscribe(org)
	defer sub.Close()

	// Entries after sub.Since() arrive on the subscription.
	backlog, err := s.courseLog(ctx, org, in.WithTo(sub.Since()))
	if err != nil {
		return err
	}
	if err := st.Send(backlog); err != nil {
		return err
	}
	return tailCourseLog(ctx, st, sub, in, backlog)
}

// courseLog returns the stored entries matching in as a single message.
func (s *QuickFeedService) courseLog(ctx context.Context, org string, in *qf.CourseLogRequest) (*qf.CourseLog, error) {
	entries, repositories, truncated, err := s.courseLogs.Query(org, in)
	if err != nil {
		qlog.FromContext(ctx).Error("failed to read course log", label.Error, err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("reading course log"))
	}
	return &qf.CourseLog{Entries: entries, Repositories: repositories, Truncated: truncated}, nil
}

// tailCourseLog sends each entry received on sub that matches in, one per
// message, skipping entries already sent in the backlog. A message lists the
// entry's repository only if the client has not been sent it before.
func tailCourseLog(ctx context.Context, st *connect.ServerStream[qf.CourseLog], sub *courselog.Subscription, in *qf.CourseLogRequest, backlog *qf.CourseLog) error {
	seen := newSeenRepositories(backlog.GetRepositories())
	sent := newHandoff(backlog.GetEntries(), sub.Since())

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
			if !deliverable(entry, in) || sent.duplicate(entry) {
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

// deliverable reports whether a live entry matches in's filters and From.
// Gap reports are always delivered: they signal missing entries and have no
// repository or level to filter on.
//
// From is checked because the subscription begins at sub.Since(), which may
// precede From; the backlog query applies From, but the subscription does not.
func deliverable(entry *qf.CourseLogEntry, in *qf.CourseLogRequest) bool {
	if courselog.Dropped(entry) {
		return true
	}
	if from := in.GetFrom(); from != nil && entry.GetTime().AsTime().Before(from.AsTime()) {
		return false
	}
	return entry.MatchesFilters(in.GetRepository(), in.GetLevel())
}

// handoff detects live entries that were already sent in the backlog.
//
// An entry is timestamped before it is written, so one timestamped before the
// subscription began may be written after it and appear both in the backlog
// and on the subscription. Only entries timestamped at or before the
// subscription start can be duplicates; later entries skip the lookup. The
// whole backlog is tracked, since the delay between timestamping and writing
// an entry is unbounded.
type handoff struct {
	since   time.Time
	pending map[digest]int
}

// digest is the SHA-256 hash of an entry's encoding, which bounds the memory
// handoff needs to 32 bytes per backlog entry.
type digest [sha256.Size]byte

func newHandoff(backlog []*qf.CourseLogEntry, since time.Time) *handoff {
	h := &handoff{since: since, pending: make(map[digest]int, len(backlog))}
	for _, entry := range backlog {
		if key, ok := entryDigest(entry); ok {
			h.pending[key]++
		}
	}
	return h
}

// duplicate reports whether entry was sent in the backlog. Each backlog entry
// matches at most once, so an identical entry logged again is still delivered.
func (h *handoff) duplicate(entry *qf.CourseLogEntry) bool {
	if len(h.pending) == 0 || entry.GetTime().AsTime().After(h.since) {
		return false
	}
	key, ok := entryDigest(entry)
	if !ok || h.pending[key] == 0 {
		return false
	}
	h.pending[key]--
	if h.pending[key] == 0 {
		delete(h.pending, key)
	}
	return true
}

// entryDigest hashes entry's deterministic encoding; determinism is needed
// because Fields is a map. It returns false if entry cannot be encoded, in
// which case entry is never treated as a duplicate.
func entryDigest(entry *qf.CourseLogEntry) (digest, bool) {
	encoded, err := proto.MarshalOptions{Deterministic: true}.Marshal(entry)
	if err != nil {
		return digest{}, false
	}
	return sha256.Sum256(encoded), true
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
