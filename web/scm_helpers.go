package web

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/scm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	repoNames = fmt.Sprintf("(%s, %s, %s)",
		pb.InfoRepo, pb.AssignmentRepo, pb.TestsRepo)

	// ErrAlreadyExists indicates that one or more Autograder repositories
	// already exists for the directory (or GitHub organization).
	ErrAlreadyExists = errors.New("course repositories already exist for that organization: " + repoNames)
	// ErrFreePlan indicates that payment plan for given organization does not allow provate
	// repositories and must be upgraded
	ErrFreePlan = errors.New("organization does not allow creation of private repositories")
	// ErrContextCanceled indicates that method failed because of scm interaction that took longer than expected
	// and not because of some application error
	ErrContextCanceled = "context canceled because the github interaction took too long. Please try again later"
	// FreeOrgPlan indicates that organization's payment plan does not allow creation of private repositories
	FreeOrgPlan = "free"
)

// createRepoAndTeam invokes the SCM to create a repository and team for the
// specified course (represented with organization ID). The SCM team name
// is also used as the group name and repository path. The provided user names represent the SCM group members.
// This function performs several sequential queries and updates on the SCM.
// Ideally, we should provide corresponding rollbacks, but that is not supported yet.
func createRepoAndTeam(ctx context.Context, sc scm.SCM, course *pb.Course, group *pb.Group) (*pb.Repository, *scm.Team, error) {
	if course.GetOrganizationPath() == "" {
		org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{ID: course.GetOrganizationID()})
		if err != nil {
			return nil, nil, fmt.Errorf("createRepoAndTeam: organization not found: %w", err)
		}
		course.OrganizationPath = org.GetPath()
	}
	org := &pb.Organization{ID: course.GetOrganizationID(), Path: course.GetOrganizationPath()}
	repo, err := sc.CreateRepository(ctx, &scm.CreateRepositoryOptions{
		Organization: org,
		Path:         group.GetName(),
		Private:      true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("createRepoAndTeam: failed to create repo: %w", err)
	}

	team, err := sc.CreateTeam(ctx, &scm.NewTeamOptions{
		Organization: org.Path,
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

	groupRepo := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		RepositoryID:   repo.ID,
		GroupID:        group.GetID(),
		HTMLURL:        repo.WebURL,
		RepoType:       pb.Repository_GROUP,
	}
	return groupRepo, team, nil
}

// deletes group repository and team
func deleteGroupRepoAndTeam(ctx context.Context, sc scm.SCM, repositoryID uint64, teamID, orgID uint64) error {
	if err := sc.DeleteRepository(ctx, &scm.RepositoryOptions{ID: repositoryID}); err != nil {
		return fmt.Errorf("deleteGroupRepoAndTeam: failed to delete repository: %w", err)
	}

	if err := sc.DeleteTeam(ctx, &scm.TeamOptions{TeamID: teamID, OrganizationID: orgID}); err != nil {
		return fmt.Errorf("deleteGroupRepoAndTeam: failed to delete team: %w", err)
	}
	return nil
}

// creates {username}-labs repository and provides pull/push access to it for the given student
func createStudentRepo(ctx context.Context, sc scm.SCM, org *pb.Organization, path string, student string) (*scm.Repository, error) {
	// we have to check that repository for given user has not already been created on github
	// if repo is found, it is safe to reuse it
	repo, err := sc.GetRepository(ctx, &scm.RepositoryOptions{
		Path:  path,
		Owner: org.GetPath(),
	})
	if err != nil {
		fmt.Println("createStudentRepo: repo not found (as expected). Error: ", err.Error())
	}

	// if no github repository found, create it
	if repo == nil {
		repo, err = sc.CreateRepository(ctx, &scm.CreateRepositoryOptions{
			Organization: org,
			Path:         path,
			Private:      true,
		})
		if err != nil {
			return nil, fmt.Errorf("createStudentRepo: failed to create repo: %w", err)
		}
	}

	// add push access to student repo
	if err = sc.UpdateRepoAccess(ctx, &scm.Repository{Owner: repo.Owner, Path: repo.Path}, student, scm.RepoPush); err != nil {
		return nil, err
	}
	return repo, nil
}

// add user to the organization's "students" team.
func addUserToStudentsTeam(ctx context.Context, sc scm.SCM, organizationPath string, userName string) error {
	opt := &scm.TeamMembershipOptions{
		Organization: organizationPath,
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
func promoteUserToTeachersTeam(ctx context.Context, sc scm.SCM, organizationPath string, userName string) error {
	studentsTeam := &scm.TeamMembershipOptions{
		Organization: organizationPath,
		Username:     userName,
		TeamName:     scm.StudentsTeam,
	}
	if err := sc.RemoveTeamMember(ctx, studentsTeam); err != nil {
		return err
	}

	teachersTeam := &scm.TeamMembershipOptions{
		Organization: organizationPath,
		Username:     userName,
		TeamName:     scm.TeachersTeam,
		Role:         scm.TeamMaintainer,
	}
	if err := sc.AddTeamMember(ctx, teachersTeam); err != nil {
		return err
	}
	return nil
}

func updateReposAndTeams(ctx context.Context, sc scm.SCM, course *pb.Course, login string, state pb.Enrollment_UserStatus) (*scm.Repository, error) {
	org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{ID: course.OrganizationID})
	if err != nil {
		return nil, err
	}

	switch state {
	case pb.Enrollment_STUDENT:
		// give access to course-info and assignments repositories
		if err := grantAccessToCourseRepos(ctx, sc, org.GetPath(), login); err != nil {
			return nil, err
		}

		// add student to the organization's "students" team
		if err = addUserToStudentsTeam(ctx, sc, org.GetPath(), login); err != nil {
			return nil, err
		}

		return createStudentRepo(ctx, sc, org, pb.StudentRepoName(login), login)

	case pb.Enrollment_TEACHER:
		// if teacher, promote to owner, remove from students team, add to teachers team
		orgUpdate := &scm.OrgMembershipOptions{
			Organization: org.Path,
			Username:     login,
			Role:         scm.OrgOwner,
		}
		// when promoting to teacher, promote to organization owner as well
		if err = sc.UpdateOrgMembership(ctx, orgUpdate); err != nil {
			return nil, fmt.Errorf("UpdateReposAndTeams: failed to update org membership for %s: %w", login, err)
		}
		err = promoteUserToTeachersTeam(ctx, sc, org.GetPath(), login)
	}
	return nil, err
}

func grantAccessToCourseRepos(ctx context.Context, sc scm.SCM, org, login string) error {
	commonRepos := []string{pb.InfoRepo, pb.AssignmentRepo}

	for _, repoType := range commonRepos {
		if err := sc.UpdateRepoAccess(ctx, &scm.Repository{Owner: org, Path: repoType}, login, scm.RepoPull); err != nil {
			return fmt.Errorf("updateReposAndTeams: failed to update repo access to repo %s for user %s: %w ", repoType, login, err)
		}
	}
	return nil
}

func updateGroupTeam(ctx context.Context, sc scm.SCM, group *pb.Group, orgID uint64) error {
	opt := &scm.UpdateTeamOptions{
		TeamID:         group.TeamID,
		OrganizationID: orgID,
		Users:          group.UserNames(),
	}
	return sc.UpdateTeamMembers(ctx, opt)
}

// remove user from the organization, delete user repository
func removeUserFromCourse(ctx context.Context, sc scm.SCM, login string, repo *pb.Repository) error {
	org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{
		ID: repo.GetOrganizationID(),
	})
	if err != nil {
		return err
	}

	opt := &scm.OrgMembershipOptions{
		Organization: org.Path,
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
func isEmpty(ctx context.Context, sc scm.SCM, repos []*pb.Repository) error {
	for _, r := range repos {
		if !sc.RepositoryIsEmpty(ctx, &scm.RepositoryOptions{ID: r.GetRepositoryID()}) {
			return fmt.Errorf("repository is not empty")
		}
	}
	return nil
}

// contextCanceled returns true if the context has been canceled.
// It is a recurring cause of unexplainable method failures when
// creating a course, approving, changing status of, or deleting
// a course enrollment or group
func contextCanceled(ctx context.Context) bool {
	// debugging context related errors
	if ctx.Err() != nil {
		fmt.Println("Context error: ", ctx.Err().Error())
	}
	return ctx.Err() == context.Canceled
}

// Returns true and formatted error if error type is SCM error
// designed to be shown to user
func parseSCMError(err error) (bool, error) {
	errStruct, ok := err.(scm.ErrFailedSCM)
	if ok {
		return ok, status.Errorf(codes.NotFound, errStruct.Message)
	}
	return ok, nil
}
