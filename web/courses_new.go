package web

import (
	"context"
	"fmt"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"

	"github.com/quickfeed/quickfeed/scm"
)

const (
	private = true
	public  = !private
)

// RepoPaths maps from QuickFeed repository path names to a boolean indicating
// whether or not the repository should be create as public or private.
var RepoPaths = map[string]bool{
	qf.InfoRepo:       public,
	qf.AssignmentRepo: private,
	qf.TestsRepo:      private,
}

// createCourse creates a new course for the directory specified in the request
// and creates the repositories for the course. Requires that the directory
// does not contain the QuickFeed repositories that will be created.
func (s *QuickFeedService) createCourse(ctx context.Context, sc scm.SCM, request *qf.Course) (*qf.Course, error) {
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
	orgOptions := &scm.OrganizationOptions{
		Path:              org.GetName(),
		DefaultPermission: scm.OrgNone,
		RepoPermissions:   false,
	}
	if err = sc.UpdateOrganization(ctx, orgOptions); err != nil {
		s.logger.Debugf("createCourse: failed to update permissions for GitHub organization %s: %s", orgOptions.Path, err)
	}

	// create a push hook on organization level
	hookOptions := &scm.CreateHookOptions{
		URL:          auth.GetEventsURL(s.bh.BaseURL),
		Secret:       s.bh.Secret,
		Organization: org.Name,
	}

	err = sc.CreateHook(ctx, hookOptions)
	if err != nil {
		s.logger.Debugf("createCourse: failed to create organization hook for %s: %s", org.GetName(), err)
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

		dbRepo := qf.Repository{
			OrganizationID: org.ID,
			RepositoryID:   repo.ID,
			HTMLURL:        repo.HTMLURL,
			RepoType:       qf.RepoType(path),
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
	opt := &scm.NewTeamOptions{
		Organization: org.Name,
		TeamName:     scm.TeachersTeam,
		Users:        []string{courseCreator.GetLogin()},
	}
	if _, err = sc.CreateTeam(ctx, opt); err != nil {
		s.logger.Debugf("createCourse: failed to create teachers team: %s", err)
		return nil, err
	}
	// create student team without any members
	studOpt := &scm.NewTeamOptions{Organization: org.Name, TeamName: scm.StudentsTeam}
	if _, err = sc.CreateTeam(ctx, studOpt); err != nil {
		s.logger.Debugf("createCourse: failed to create students team: %s", err)
		return nil, err
	}

	// add student repo for the course creator
	scmRepo, err := createStudentRepo(ctx, sc, org, qf.StudentRepoName(courseCreator.GetLogin()), courseCreator.GetLogin())
	if err != nil {
		return nil, err
	}
	repoQuery := &qf.Repository{
		OrganizationID: org.GetID(),
		RepositoryID:   scmRepo.ID,
		UserID:         courseCreator.ID,
		HTMLURL:        scmRepo.HTMLURL,
		RepoType:       qf.Repository_USER,
	}
	if err := s.db.CreateRepository(repoQuery); err != nil {
		return nil, err
	}

	request.OrganizationName = org.GetName()
	if err := s.db.CreateCourse(request.GetCourseCreatorID(), request); err != nil {
		s.logger.Debugf("createCourse: failed to create database record for course %s: %s", request.Name, err)
		return nil, err
	}
	return request, nil
}

// isDirty returns true if the list of provided repositories contains
// any of the repositories that QuickFeed wants to create.
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
