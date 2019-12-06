package web

import (
	"context"
	"fmt"

	"github.com/autograde/aguis/web/auth"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
)

const (
	private = true
	public  = !private
)

// RepoPaths maps from Autograder repository path names to a boolean indicating
// whether or not the repository should be create as public or private.
var RepoPaths = map[string]bool{
	pb.InfoRepo:       public,
	pb.AssignmentRepo: private,
	pb.TestsRepo:      private,
	pb.SolutionsRepo:  private,
}

// createCourse creates a new course for the directory specified in the request
// and creates the repositories for the course. Requires that the directory
// does not contain the Autograder repositories that will be created.
func (s *AutograderService) createCourse(ctx context.Context, sc scm.SCM, request *pb.Course) (*pb.Course, error) {
	org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{ID: request.OrganizationID})
	if err != nil {
		return nil, err
	}
	if org.GetPaymentPlan() == FreeOrgPlan {
		return nil, ErrFreePlan
	}
	repos, err := sc.GetRepositories(ctx, org)
	if err != nil {
		return nil, err
	}
	if isDirty(repos) {
		return nil, ErrAlreadyExists
	}

	// set default repository access level for all students to "none"
	// will not affect organization owners (teachers)
	orgOptions := &scm.CreateOrgOptions{
		Path:              org.GetPath(),
		DefaultPermission: scm.OrgNone,
	}
	if err = sc.UpdateOrganization(ctx, orgOptions); err != nil {
		s.logger.Debugf("createCourse: failed to update permissions for GitHub organization %s: %s", orgOptions.Path, err)
	}

	// create a push hook on organization level
	hookOptions := &scm.OrgHookOptions{
		URL:          auth.GetEventsURL(s.bh.BaseURL, request.Provider),
		Secret:       s.bh.Secret,
		Organization: org,
	}

	err = sc.CreateOrgHook(ctx, hookOptions)
	if err != nil {
		s.logger.Debugf("createCourse: failed to create organization hook for %s: %s", org.GetPath(), err)
	}

	// create course repos and webhooks for each repo
	for path, private := range RepoPaths {

		repoOptions := &scm.CreateRepositoryOptions{
			Path:         path,
			Organization: org,
			Private:      private,
		}
		repo, err := sc.CreateRepository(ctx, repoOptions)
		if err != nil {
			return nil, err
		}

		dbRepo := pb.Repository{
			OrganizationID: org.ID,
			RepositoryID:   repo.ID,
			HTMLURL:        repo.WebURL,
			RepoType:       pb.RepoType(path),
		}
		if err := s.db.CreateRepository(&dbRepo); err != nil {
			s.logger.Debugf("createCourse: failed to create database record for repository %s: %s", path, err)
			return nil, err
		}
	}

	// add course creator to teacher team
	courseCreator, err := s.db.GetUser(request.GetCourseCreatorID())
	if err != nil {
		return nil, fmt.Errorf("createCourse: failed to get course creator record from database: %w", err)
	}
	// create teacher team with course creator
	opt := &scm.TeamOptions{
		Organization: org,
		TeamName:     scm.TeachersTeam,
		Users:        []string{courseCreator.GetLogin()},
	}
	if _, err = sc.CreateTeam(ctx, opt); err != nil {
		s.logger.Debugf("createCourse: failed to create teachers team: %s", err)
		return nil, err
	}
	// create student team without any members
	studOpt := &scm.TeamOptions{Organization: org, TeamName: scm.StudentsTeam}
	if _, err = sc.CreateTeam(ctx, studOpt); err != nil {
		s.logger.Debugf("createCourse: failed to create students team: %s", err)
		return nil, err
	}

	if err := s.db.CreateCourse(request.GetCourseCreatorID(), request); err != nil {
		s.logger.Debugf("createCourse: failed to create database record for course %s: %s", request.Name, err)
		return nil, err
	}
	return request, nil
}

// isDirty returns true if the list of provided repositories contains
// any of the repositories that Autograder wants to create.
func isDirty(repos []*scm.Repository) bool {
	if len(repos) == 0 {
		return false
	}
	for _, repo := range repos {
		if _, exists := RepoPaths[repo.Path]; exists {
			return true
		}
	}
	return false
}
