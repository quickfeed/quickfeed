package web

import (
	"fmt"

	"github.com/quickfeed/quickfeed/qf"
	"gorm.io/gorm"
)

func (s *QuickFeedService) getRepo(course *qf.Course, id uint64, repoType qf.Repository_Type) (*qf.Repository, error) {
	query := &qf.Repository{
		ScmOrganizationID: course.GetScmOrganizationID(),
		RepoType:          repoType,
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

func repoTypes(enrollment *qf.Enrollment) []qf.Repository_Type {
	repositories := []qf.Repository_Type{
		qf.Repository_INFO,
		qf.Repository_ASSIGNMENTS,
		qf.Repository_USER,
	}
	if enrollment.IsTeacher() {
		repositories = append(repositories, qf.Repository_TESTS)
	}
	if enrollment.GroupID > 0 {
		repositories = append(repositories, qf.Repository_GROUP)
	}
	return repositories
}
