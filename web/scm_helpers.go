package web

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

// ErrContextCanceled indicates that method failed because of scm interaction that took longer than expected
// and not because of some application error
var ErrContextCanceled = errors.New("context canceled because the github interaction took too long. Please try again later")

// GetSCM returns an SCM client for the course organization.
func (q *QuickFeedService) getSCM(ctx context.Context, organization string) (scm.SCM, error) {
	return q.scmMgr.GetOrCreateSCM(ctx, q.logger, organization)
}

// getSCMForCourse returns an SCM client for the course organization.
func (q *QuickFeedService) getSCMForCourse(ctx context.Context, courseID uint64) (scm.SCM, error) {
	course, err := q.db.GetCourse(courseID)
	if err != nil {
		return nil, err
	}
	return q.getSCM(ctx, course.GetScmOrganizationName())
}

// createRepo invokes the SCM to create a repository for the
// specified course (represented with organization ID). The group name
// is used as the repository path. The provided user names represent the SCM group members.
// This function performs several sequential queries and updates on the SCM.
// Ideally, we should provide corresponding rollbacks, but that is not supported yet.
func createRepo(ctx context.Context, sc scm.SCM, course *qf.Course, group *qf.Group) (*qf.Repository, error) {
	opt := &scm.GroupOptions{
		Organization: course.GetScmOrganizationName(),
		GroupName:    group.GetName(),
		Users:        group.UserNames(),
	}
	repo, err := sc.CreateGroup(ctx, opt)
	if err != nil {
		return nil, err
	}
	groupRepo := &qf.Repository{
		ScmOrganizationID: course.GetScmOrganizationID(),
		ScmRepositoryID:   repo.ID,
		GroupID:           group.GetID(),
		HTMLURL:           repo.HTMLURL,
		RepoType:          qf.Repository_GROUP,
	}
	return groupRepo, nil
}

func updateGroupMembers(ctx context.Context, sc scm.SCM, group *qf.Group, orgName string) error {
	opt := &scm.GroupOptions{
		GroupName:    group.GetName(),
		Organization: orgName,
		Users:        group.UserNames(),
	}
	return sc.UpdateGroupMembers(ctx, opt)
}

// isEmpty ensured that all of the provided repositories are empty
func isEmpty(ctx context.Context, sc scm.SCM, repos []*qf.Repository) error {
	for _, r := range repos {
		if !sc.RepositoryIsEmpty(ctx, &scm.RepositoryOptions{ID: r.GetScmRepositoryID()}) {
			return fmt.Errorf("repository %s is not empty", r.Name())
		}
	}
	return nil
}

// ctxErr returns a context error. There could be two reasons
// for a context error: exceeded deadline or canceled context.
// Canceled context is a recurring cause of unexplainable
// method failures when creating a course, approving, changing
// status of, or deleting a course enrollment or group.
func ctxErr(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return connect.NewError(connect.CodeCanceled, ctx.Err())
	case context.DeadlineExceeded:
		return connect.NewError(connect.CodeDeadlineExceeded, ctx.Err())
	}
	return nil
}

// Returns true and formatted error if error type is SCM error
// designed to be shown to user
func parseSCMError(err error) (bool, error) {
	var scmErr *scm.SCMError
	if errors.As(err, &scmErr) {
		userErr := scmErr.UserError()
		if errors.Is(err, scm.ErrAlreadyExists) {
			return true, connect.NewError(connect.CodeAlreadyExists, userErr)
		}
		if errors.Is(err, scm.ErrNotOwner) {
			return true, connect.NewError(connect.CodePermissionDenied, userErr)
		}
		return true, connect.NewError(connect.CodeNotFound, userErr)
	}
	return false, nil
}
