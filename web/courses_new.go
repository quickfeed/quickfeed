package web

import (
	"context"
	"errors"
	"fmt"

	"github.com/autograde/aguis/web/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

var (
	repoNames = fmt.Sprintf("(%s, %s, %s, %s)",
		pb.InfoRepo, pb.AssignmentRepo, pb.TestsRepo, pb.SolutionsRepo)

	// ErrAlreadyExists indicates that one or more Autograder repositories
	// already exists for the directory (or GitHub organization).
	ErrAlreadyExists = errors.New("repositories already exists in SCM " + repoNames)
)

// createCourse creates a new course for the directory specified in the request
// and creates the repositories for the course. Requires that the directory
// does not contain the Autograder repositories that will be created.
func (s *AutograderService) createCourse(ctx context.Context, request *pb.Course, sc scm.SCM) (*pb.Course, error) {
	org, err := sc.GetOrganization(ctx, request.OrganizationID)
	if err != nil {
		return nil, err
	}
	repos, err := sc.GetRepositories(ctx, org)
	if err != nil {
		return nil, err
	}
	if isDirty(repos) {
		return nil, ErrAlreadyExists
	}

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

		hookOptions := &scm.CreateHookOptions{
			URL:        auth.GetEventsURL(s.bh.BaseURL, request.Provider),
			Secret:     s.bh.Secret,
			Repository: repo,
		}
		if err := sc.CreateHook(ctx, hookOptions); err != nil {
			return nil, err
		}

		dbRepo := pb.Repository{
			OrganizationID: org.ID,
			RepositoryID:   repo.ID,
			HTMLURL:        repo.WebURL,
			RepoType:       pb.RepoType(path),
		}
		if err := s.db.CreateRepository(&dbRepo); err != nil {
			return nil, err
		}
	}

	// we want to add course creator to teacher team
	courseCreator, err := s.db.GetUser(request.GetCourseCreatorID())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal database error")
	}
	// create two teams on course organization: one for all students and one for all teachers
	opt := &scm.CreateTeamOptions{
		Organization: org,
		TeamName:     "teachers",
		Users:        []string{courseCreator.GetLogin()},
	}
	if _, err = sc.CreateTeam(ctx, opt); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create teacher team")
	}
	if _, err = sc.CreateTeam(ctx, &scm.CreateTeamOptions{Organization: org, TeamName: "students"}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create student team")
	}

	// then add course to the database
	if err := s.db.CreateCourse(request.GetCourseCreatorID(), request); err != nil {
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
