package web

import (
	"context"
	"errors"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

var (
	// ErrContextCanceled indicates that method failed because of scm interaction that took longer than expected
	// and not because of some application error
	ErrContextCanceled = errors.New("context canceled because the github interaction took too long. Please try again later")
)

// InitSCMs creates and saves SCM clients for each course without an active SCM client.
func (q *QuickFeedService) InitSCMs(ctx context.Context) error {
	courses, err := q.db.GetCourses()
	if err != nil {
		return err
	}
	for _, course := range courses {
		_, err := q.getSCM(ctx, course.GetOrganizationName())
		if err != nil {
			return err
		}
	}
	return nil
}

// GetSCM returns an SCM client for the course organization.
func (q *QuickFeedService) getSCM(ctx context.Context, organization string) (scm.SCM, error) {
	return q.scmMgr.GetOrCreateSCM(ctx, q.logger, organization)
}

// getSCMForCourse returns an SCM client for the course organization.
func (q *QuickFeedService) getSCMForCourse(ctx context.Context, courseID uint64) (scm.SCM, error) {
	course, err := q.db.GetCourse(courseID, false)
	if err != nil {
		return nil, err
	}
	return q.getSCM(ctx, course.OrganizationName)
}

// getCredsForUserSCM returns the given user's personal access token.
func (q *QuickFeedService) getCredsForUserSCM(user *qf.User) (string, error) {
	refreshToken, err := user.GetRefreshToken(env.ScmProvider())
	if err != nil {
		return "", err
	}
	// Exchange a refresh token for an access token.
	token, err := q.scmMgr.ExchangeToken(refreshToken)
	if err != nil {
		return "", err
	}
	// Save user's refresh token in the database.
	remoteIdentity := user.GetRemoteIDFor(env.ScmProvider())
	// TODO(meling) rename UpdateAccessToken() database method to UpdateRefreshToken()
	// TODO(meling) later: move RefreshToken and ScmRemoteID directly into User type (requires updating the User proto message)
	// TODO(meling) rename RemoteIdentity.AccessToken to RemoteIdentity.RefreshToken
	remoteIdentity.AccessToken = token.RefreshToken
	if err := q.db.UpdateAccessToken(remoteIdentity); err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

// createRepoAndTeam invokes the SCM to create a repository and team for the
// specified course (represented with organization ID). The SCM team name
// is also used as the group name and repository path. The provided user names represent the SCM group members.
// This function performs several sequential queries and updates on the SCM.
// Ideally, we should provide corresponding rollbacks, but that is not supported yet.
func createRepoAndTeam(ctx context.Context, sc scm.SCM, course *qf.Course, group *qf.Group) (*qf.Repository, *scm.Team, error) {
	if course.GetOrganizationName() == "" {
		org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{ID: course.GetOrganizationID()})
		if err != nil {
			return nil, nil, fmt.Errorf("createRepoAndTeam: organization not found: %w", err)
		}
		course.OrganizationName = org.GetName()
	}
	org := &qf.Organization{ID: course.GetOrganizationID(), Name: course.GetOrganizationName()}
	repo, err := sc.CreateRepository(ctx, &scm.CreateRepositoryOptions{
		Organization: org.Name,
		Path:         group.GetName(),
		Private:      true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("createRepoAndTeam: failed to create repo: %w", err)
	}

	team, err := sc.CreateTeam(ctx, &scm.NewTeamOptions{
		Organization: org.Name,
		TeamName:     group.GetName(),
		Users:        group.UserNames(),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("createRepoAndTeam: failed to create team: %w", err)
	}

	err = sc.AddTeamRepo(ctx, &scm.AddTeamRepoOptions{
		TeamID:         team.ID,
		OrganizationID: course.GetOrganizationID(),
		Owner:          repo.Owner,
		Repo:           repo.Path,
		Permission:     scm.RepoPush,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("createRepoAndTeam: failed to add team to repo: %w", err)
	}

	groupRepo := &qf.Repository{
		OrganizationID: course.GetOrganizationID(),
		RepositoryID:   repo.ID,
		GroupID:        group.GetID(),
		HTMLURL:        repo.HTMLURL,
		RepoType:       qf.Repository_GROUP,
	}
	return groupRepo, team, nil
}

// deletes group repository and team
func deleteGroupRepoAndTeam(ctx context.Context, sc scm.SCM, repositoryID, teamID, orgID uint64) error {
	if err := sc.DeleteRepository(ctx, &scm.RepositoryOptions{ID: repositoryID}); err != nil {
		return fmt.Errorf("deleteGroupRepoAndTeam: failed to delete repository: %w", err)
	}
	if err := sc.DeleteTeam(ctx, &scm.TeamOptions{TeamID: teamID, OrganizationID: orgID}); err != nil {
		return fmt.Errorf("deleteGroupRepoAndTeam: failed to delete team: %w", err)
	}
	return nil
}

func updateGroupTeam(ctx context.Context, sc scm.SCM, group *qf.Group, orgID uint64) error {
	opt := &scm.UpdateTeamOptions{
		TeamID:         group.TeamID,
		OrganizationID: orgID,
		Users:          group.UserNames(),
	}
	return sc.UpdateTeamMembers(ctx, opt)
}

// isEmpty ensured that all of the provided repositories are empty
func isEmpty(ctx context.Context, sc scm.SCM, repos []*qf.Repository) error {
	for _, r := range repos {
		if !sc.RepositoryIsEmpty(ctx, &scm.RepositoryOptions{ID: r.GetRepositoryID()}) {
			return fmt.Errorf("repository is not empty")
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
	errStruct, ok := err.(scm.ErrFailedSCM)
	if ok {
		return ok, connect.NewError(connect.CodeNotFound, errors.New(errStruct.Message))
	}
	return ok, nil
}
