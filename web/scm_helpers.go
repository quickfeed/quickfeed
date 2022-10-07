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
	repoNames = fmt.Sprintf("(%s, %s, %s)",
		qf.InfoRepo, qf.AssignmentRepo, qf.TestsRepo)

	// ErrAlreadyExists indicates that one or more QuickFeed repositories
	// already exists for the directory (or GitHub organization).
	ErrAlreadyExists = errors.New("course repositories already exist for that organization: " + repoNames)
	// ErrFreePlan indicates that payment plan for given organization does not allow private
	// repositories and must be upgraded
	ErrFreePlan = errors.New("organization does not allow creation of private repositories")
	// ErrContextCanceled indicates that method failed because of scm interaction that took longer than expected
	// and not because of some application error
	ErrContextCanceled = errors.New("context canceled because the github interaction took too long. Please try again later")
	// FreeOrgPlan indicates that organization's payment plan does not allow creation of private repositories
	FreeOrgPlan = "free"
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

// getSCMForUser returns an SCM client based on the user's personal access token.
func (q *QuickFeedService) getSCMForUser(user *qf.User) (scm.SCMInvite, error) {
	refreshToken, err := user.GetRefreshToken(env.ScmProvider())
	if err != nil {
		return nil, err
	}
	// Exchange a refresh token for an access token.
	token, err := q.scmMgr.ExchangeToken(refreshToken)
	if err != nil {
		return nil, err
	}
	// Save user's refresh token in the database.
	remoteIdentity := user.GetRemoteIDFor(env.ScmProvider())
	// TODO(meling) rename UpdateAccessToken() database method to UpdateRefreshToken()
	// TODO(meling) later: move RefreshToken and ScmRemoteID directly into User type (requires updating the User proto message)
	// TODO(meling) rename RemoteIdentity.AccessToken to RemoteIdentity.RefreshToken
	remoteIdentity.AccessToken = token.RefreshToken
	if err := q.db.UpdateAccessToken(remoteIdentity); err != nil {
		return nil, err
	}
	return scm.NewInviteOnlySCMClient(token.AccessToken), nil
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

// creates {username}-labs repository and provides pull/push access to it for the given student
func createStudentRepo(ctx context.Context, sc scm.SCM, org *qf.Organization, path string, student string) (*scm.Repository, error) {
	// create repo, or return existing repo if it already exists
	// if repo is found, it is safe to reuse it
	repo, err := sc.CreateRepository(ctx, &scm.CreateRepositoryOptions{
		Organization: org.Name,
		Path:         path,
		Private:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("createStudentRepo: failed to create repo: %w", err)
	}

	// add push access to student repo
	if err = sc.UpdateRepoAccess(ctx, repo, student, scm.RepoPush); err != nil {
		return nil, fmt.Errorf("createStudentRepo: failed to update repo push access: %w", err)
	}
	return repo, nil
}

// add user to the organization's "students" team.
func addUserToStudentsTeam(ctx context.Context, sc scm.SCM, organizationName string, userName string) error {
	opt := &scm.TeamMembershipOptions{
		Organization: organizationName,
		TeamName:     scm.StudentsTeam,
		Username:     userName,
		Role:         scm.TeamMember,
	}
	if err := sc.AddTeamMember(ctx, opt); err != nil {
		return err
	}
	return nil
}

// add user to the organization's "teachers" team, and remove user from "students" team.
func promoteUserToTeachersTeam(ctx context.Context, sc scm.SCM, organizationName string, userName string) error {
	studentsTeam := &scm.TeamMembershipOptions{
		Organization: organizationName,
		Username:     userName,
		TeamName:     scm.StudentsTeam,
	}
	if err := sc.RemoveTeamMember(ctx, studentsTeam); err != nil {
		return err
	}

	teachersTeam := &scm.TeamMembershipOptions{
		Organization: organizationName,
		Username:     userName,
		TeamName:     scm.TeachersTeam,
		Role:         scm.TeamMaintainer,
	}
	if err := sc.AddTeamMember(ctx, teachersTeam); err != nil {
		return err
	}
	return nil
}

func updateReposAndTeams(ctx context.Context, sc scm.SCM, course *qf.Course, login string, state qf.Enrollment_UserStatus) (*scm.Repository, error) {
	org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{ID: course.OrganizationID})
	if err != nil {
		return nil, err
	}

	switch state {
	case qf.Enrollment_STUDENT:
		// give access to the course's info and assignments repositories
		if err := grantAccessToCourseRepos(ctx, sc, org.GetName(), login); err != nil {
			return nil, err
		}

		// add student to the organization's "students" team
		if err = addUserToStudentsTeam(ctx, sc, org.GetName(), login); err != nil {
			return nil, err
		}

		return createStudentRepo(ctx, sc, org, qf.StudentRepoName(login), login)

	case qf.Enrollment_TEACHER:
		// if teacher, promote to owner, remove from students team, add to teachers team
		orgUpdate := &scm.OrgMembershipOptions{
			Organization: org.Name,
			Username:     login,
			Role:         scm.OrgOwner,
		}
		// when promoting to teacher, promote to organization owner as well
		if err = sc.UpdateOrgMembership(ctx, orgUpdate); err != nil {
			return nil, fmt.Errorf("UpdateReposAndTeams: failed to update org membership for %s: %w", login, err)
		}
		err = promoteUserToTeachersTeam(ctx, sc, org.GetName(), login)
	}
	return nil, err
}

func grantAccessToCourseRepos(ctx context.Context, sc scm.SCM, org, login string) error {
	commonRepos := []string{qf.InfoRepo, qf.AssignmentRepo}

	for _, repoType := range commonRepos {
		if err := sc.UpdateRepoAccess(ctx, &scm.Repository{Owner: org, Path: repoType}, login, scm.RepoPull); err != nil {
			return fmt.Errorf("updateReposAndTeams: failed to update repo access to repo %s for user %s: %w ", repoType, login, err)
		}
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

// remove user from the organization, delete user repository
func removeUserFromCourse(ctx context.Context, sc scm.SCM, login string, repo *qf.Repository) error {
	org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{
		ID: repo.GetOrganizationID(),
	})
	if err != nil {
		return err
	}

	opt := &scm.OrgMembershipOptions{
		Organization: org.Name,
		Username:     login,
	}
	if err := sc.RemoveMember(ctx, opt); err != nil {
		return err
	}
	return sc.DeleteRepository(ctx, &scm.RepositoryOptions{ID: repo.GetRepositoryID()})
}

// remove user from teachers team, set organization status from owner to regular member
func revokeTeacherStatus(ctx context.Context, sc scm.SCM, org, userName string) error {
	teamOpts := &scm.TeamMembershipOptions{
		Organization: org,
		TeamName:     scm.TeachersTeam,
		Username:     userName,
	}

	if err := sc.RemoveTeamMember(ctx, teamOpts); err != nil {
		return err
	}

	teamOpts.TeamName = scm.StudentsTeam
	if err := sc.AddTeamMember(ctx, teamOpts); err != nil {
		return err
	}

	return sc.UpdateOrgMembership(ctx, &scm.OrgMembershipOptions{
		Organization: org,
		Username:     userName,
		Role:         scm.OrgMember,
	})
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
