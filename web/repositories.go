package web

import (
	"context"
	"fmt"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"gorm.io/gorm"
)

func (s *QuickFeedService) getRepo(course *qf.Course, id uint64, repoType qf.Repository_Type) (*qf.Repository, error) {
	query := &qf.Repository{
		OrganizationID: course.GetOrganizationID(),
		RepoType:       repoType,
	}
	switch repoType {
	case qf.Repository_USER:
		query.UserID = id
	case qf.Repository_GROUP:
		query.GroupID = id
	}
	repos, err := s.db.GetRepositories(query)
	if err != nil {
		return nil, err
	}
	if len(repos) < 1 {
		return nil, fmt.Errorf("no %s repositories found for %s id %d: %w", repoType, course.GetCode(), id, gorm.ErrRecordNotFound)
	}
	return repos[0], nil
}

// isEmptyRepo returns nil if all repositories for the given course and student or group are empty,
// returns an error otherwise.
func (s *QuickFeedService) isEmptyRepo(ctx context.Context, sc scm.SCM, request *qf.RepositoryRequest) error {
	course, err := s.db.GetCourse(request.GetCourseID(), false)
	if err != nil {
		return err
	}
	repos, err := s.db.GetRepositories(&qf.Repository{
		OrganizationID: course.GetOrganizationID(),
		UserID:         request.GetUserID(),
		GroupID:        request.GetGroupID(),
	})
	if err != nil {
		return err
	}
	if len(repos) < 1 {
		return fmt.Errorf("no repositories found")
	}
	return isEmpty(ctx, sc, repos)
}
