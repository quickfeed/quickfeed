package web

import (
	"context"
	"fmt"

	"github.com/quickfeed/quickfeed/qf"

	"github.com/quickfeed/quickfeed/scm"
)

const (
	private = true
	public  = !private
)

// RepoPaths maps from QuickFeed repository path names to a boolean indicating
// whether or not the repository should be create as public or private.
var RepoPaths = map[string]bool{
	qf.InfoRepo:        public,
	qf.AssignmentsRepo: private,
	qf.TestsRepo:       private,
}

// createCourse creates a new course for the directory specified in the request
// and creates the repositories for the course. Requires that the directory
// does not contain the QuickFeed repositories that will be created.
func (s *QuickFeedService) createCourse(ctx context.Context, sc scm.SCM, request *qf.Course) (*qf.Course, error) {
	courseCreator, err := s.db.GetUser(request.GetCourseCreatorID())
	if err != nil {
		return nil, fmt.Errorf("createCourse: failed to get course creator record from database: %w", err)
	}
	repos, err := sc.CreateCourse(ctx, &scm.NewCourseOptions{
		CourseCreator:  courseCreator.Login,
		OrganizationID: request.OrganizationID,
	})
	if err != nil {
		s.logger.Debugf("createCourse: failed to create course repositories or teams: %w", err)
		return nil, err
	}

	for _, repo := range repos {
		dbRepo := qf.Repository{
			OrganizationID: request.OrganizationID,
			RepositoryID:   repo.ID,
			HTMLURL:        repo.HTMLURL,
			RepoType:       qf.RepoType(repo.Path),
		}
		if dbRepo.IsUserRepo() {
			dbRepo.UserID = courseCreator.ID
		}
		if err := s.db.CreateRepository(&dbRepo); err != nil {
			s.logger.Debugf("createCourse: failed to create database record for repository %s: %s", repo.Path, err)
			return nil, err
		}
	}

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
