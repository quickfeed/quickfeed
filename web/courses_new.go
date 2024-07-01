package web

import (
	"context"
	"fmt"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

// createCourse creates a new course for the directory specified in the request
// and creates the repositories for the course. Requires that the directory
// does not contain the QuickFeed repositories that will be created.
func (s *QuickFeedService) createCourse(ctx context.Context, sc scm.SCM, request *qf.Course) (*qf.Course, error) {
	courseCreator, err := s.db.GetUser(request.GetCourseCreatorID())
	if err != nil {
		return nil, fmt.Errorf("failed to get course creator record from database: %w", err)
	}
	repos, err := sc.CreateCourse(ctx, &scm.CourseOptions{
		CourseCreator:  courseCreator.Login,
		OrganizationID: request.ScmOrganizationID,
	})
	if err != nil {
		return nil, err
	}
	for _, repo := range repos {
		dbRepo := qf.Repository{
			ScmOrganizationID: request.ScmOrganizationID,
			ScmRepositoryID:   repo.ID,
			HTMLURL:           repo.HTMLURL,
			RepoType:          qf.RepoType(repo.Path),
		}
		if dbRepo.IsUserRepo() {
			dbRepo.UserID = courseCreator.ID
		}
		if err := s.db.CreateRepository(&dbRepo); err != nil {
			return nil, fmt.Errorf("failed to create database record for repository %s: %w", repo.Path, err)
		}
	}

	if err := s.db.CreateCourse(request.GetCourseCreatorID(), request); err != nil {
		return nil, fmt.Errorf("failed to create database record for course %s: %w", request.Name, err)
	}
	return request, nil
}
