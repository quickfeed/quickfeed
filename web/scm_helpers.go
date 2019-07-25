package web

import (
	"context"
	"fmt"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
)

//
const (
	teachersTeam = "allteachers"
	studentsTeam = "allstudents"
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

	team, err := sc.CreateTeam(ctx, &scm.CreateTeamOptions{
		Organization: org,
		TeamName:     teamName,
		Users:        userNames,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("createRepoAndTeam: failed to create team: %w", err)
	}

	err = sc.AddTeamRepo(ctx, &scm.AddTeamRepoOptions{
		TeamID: team.ID,
		Owner:  repo.Owner,
		Repo:   repo.Path,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("createRepoAndTeam: failed to add team to repo: %w", err)
	}
	return repo, team, nil
}

// add user to the organization's "students" team.
func addUserToStudentsTeam(ctx context.Context, sc scm.SCM, org *pb.Organization, userName string) error {
	opt := &scm.TeamMembershipOptions{
		Organization: org,
		TeamSlug:     studentsTeam,
		Username:     userName,
		Role:         "member",
	}
	if err := sc.AddTeamMember(ctx, opt); err != nil {
		return fmt.Errorf("addUserToStudentsTeam: failed to add '%s' to students team: %w", userName, err)
	}
	return nil
}

// add user to the organization's "teachers" team, and remove user from "students" team.
func promoteUserToTeachersTeam(ctx context.Context, sc scm.SCM, org *pb.Organization, userName string) error {
	studentsTeam := &scm.TeamMembershipOptions{
		Organization: org,
		Username:     userName,
		TeamSlug:     studentsTeam,
	}
	if err := sc.RemoveTeamMember(ctx, studentsTeam); err != nil {
		return fmt.Errorf("promoteUserToTeachersTeam: failed to remove '%s' from students team: %w", userName, err)
	}

	teachersTeam := &scm.TeamMembershipOptions{
		Organization: org,
		Username:     userName,
		TeamSlug:     teachersTeam,
		Role:         "maintainer",
	}
	if err := sc.AddTeamMember(ctx, teachersTeam); err != nil {
		return fmt.Errorf("promoteUserToTeachersTeam: failed to add '%s' to teachers team: %w", userName, err)
	}
	return nil
}
