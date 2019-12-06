package web

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
)

var (
	repoNames = fmt.Sprintf("(%s, %s, %s, %s)",
		pb.InfoRepo, pb.AssignmentRepo, pb.TestsRepo, pb.SolutionsRepo)

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
// specified namespace (typically the course name), the path of the repository
// (typically the name of the student with a '-labs' suffix or the group name).
// The team name is the student name or group name, whereas the user names are
// the members of the team. For single student repositories, the user names are
// typically just the one student's user name.
// This function performs several sequential queries and updates on the SCM.
// Ideally, we should provide corresponding rollbacks, but that is not supported yet.
func createRepoAndTeam(ctx context.Context, sc scm.SCM, org *pb.Organization, path, teamName string, userNames []string) (*scm.Repository, *scm.Team, error) {
	repo, err := sc.CreateRepository(ctx, &scm.CreateRepositoryOptions{
		Organization: org,
		Path:         path,
		Private:      true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("createRepoAndTeam: failed to create repo: %w", err)
	}

	team, err := sc.CreateTeam(ctx, &scm.TeamOptions{
		Organization: org,
		TeamName:     teamName,
		Users:        userNames,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("createRepoAndTeam: failed to create team: %w", err)
	}

	err = sc.AddTeamRepo(ctx, &scm.AddTeamRepoOptions{
		TeamID:     team.ID,
		Owner:      repo.Owner,
		Repo:       repo.Path,
		Permission: scm.RepoPush,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("createRepoAndTeam: failed to add team to repo: %w", err)
	}
	return repo, team, nil
}

// deletes group repository and team
func deleteGroupRepoAndTeam(ctx context.Context, sc scm.SCM, repositoryID uint64, teamID uint64) error {

	if err := sc.DeleteRepository(ctx, &scm.RepositoryOptions{ID: repositoryID}); err != nil {
		return fmt.Errorf("deleteGroupRepoAndTeam: failed to delete repository: %w", err)
	}

	if err := sc.DeleteTeam(ctx, &scm.TeamOptions{TeamID: teamID}); err != nil {
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
func addUserToStudentsTeam(ctx context.Context, sc scm.SCM, org *pb.Organization, userName string) error {
	opt := &scm.TeamMembershipOptions{
		Organization: org,
		TeamSlug:     scm.StudentsTeam,
		Username:     userName,
		Role:         scm.TeamMember,
	}
	if err := sc.AddTeamMember(ctx, opt); err != nil {
		return err
	}
	return nil
}

// add user to the organization's "teachers" team, and remove user from "students" team.
func promoteUserToTeachersTeam(ctx context.Context, sc scm.SCM, org *pb.Organization, userName string) error {
	studentsTeam := &scm.TeamMembershipOptions{
		Organization: org,
		Username:     userName,
		TeamSlug:     scm.StudentsTeam,
	}
	if err := sc.RemoveTeamMember(ctx, studentsTeam); err != nil {
		return err
	}

	teachersTeam := &scm.TeamMembershipOptions{
		Organization: org,
		Username:     userName,
		TeamSlug:     scm.TeachersTeam,
		Role:         scm.TeamMaintainer,
	}
	if err := sc.AddTeamMember(ctx, teachersTeam); err != nil {
		return err
	}
	return nil
}

func updateGroupTeam(ctx context.Context, sc scm.SCM, org *pb.Organization, group *pb.Group) error {
	opt := &scm.TeamOptions{
		Organization: org,
		TeamName:     group.Name,
		TeamID:       group.TeamID,
		Users:        group.UserNames(),
	}
	return sc.UpdateTeamMembers(ctx, opt)
}

func removeUserFromCourse(ctx context.Context, sc scm.SCM, login string, repo *pb.Repository) error {

	org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{
		ID: repo.GetOrganizationID(),
	})
	if err != nil {
		return err
	}
	opt := &scm.OrgMembershipOptions{
		Organization: org,
		Username:     login,
	}

	if err := sc.RemoveMember(ctx, opt); err != nil {
		return err
	}

	return sc.DeleteRepository(ctx, &scm.RepositoryOptions{ID: repo.GetRepositoryID()})
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
	return ctx.Err() == context.Canceled
}
