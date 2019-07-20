package web

import (
	"context"
	"fmt"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
)

// createRepoAndTeam creates the user or group repo in course on the provided SCM.
// This function performs several sequential queries and updates on the SCM.
// Ideally, we should provide corresponding rollbacks, but that is not supported yet.
func createRepoAndTeam(ctx context.Context, sc scm.SCM, course *pb.Course, path, teamName string, userNames []string) (*scm.Repository, *scm.Team, error) {
	org, err := sc.GetOrganization(ctx, course.OrganizationID)
	if err != nil {
		return nil, nil, fmt.Errorf("createRepoAndTeam: organization not found: %w", err)
	}

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
