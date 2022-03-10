package web

import (
	"context"
	"fmt"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/scm"
	"gorm.io/gorm"
)

func (s *AutograderService) getRepo(course *pb.Course, id uint64, repoType pb.Repository_Type) (*pb.Repository, error) {
	query := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		RepoType:       repoType,
	}
	switch repoType {
	case pb.Repository_USER:
		query.UserID = id
	case pb.Repository_GROUP:
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

// getRepositoryURL returns URL of a course repository of the given type.
func (s *AutograderService) getRepositoryURL(currentUser *pb.User, courseID uint64, repoType pb.Repository_Type) (string, error) {
	course, err := s.db.GetCourse(courseID, false)
	if err != nil {
		return "", err
	}
	userRepoQuery := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		RepoType:       repoType,
	}

	switch repoType {
	case pb.Repository_USER:
		userRepoQuery.UserID = currentUser.GetID()
	case pb.Repository_GROUP:
		enrol, err := s.db.GetEnrollmentByCourseAndUser(courseID, currentUser.GetID())
		if err != nil {
			return "", err
		}
		if enrol.GetGroupID() > 0 {
			userRepoQuery.GroupID = enrol.GroupID
		}
	}

	repos, err := s.db.GetRepositories(userRepoQuery)
	if err != nil {
		return "", err
	}
	if len(repos) != 1 {
		return "", fmt.Errorf("found %d repositories for query %+v", len(repos), userRepoQuery)
	}
	return repos[0].HTMLURL, nil
}

// isEmptyRepo returns nil if all repositories for the given course and student or group are empty,
// returns an error otherwise.
func (s *AutograderService) isEmptyRepo(ctx context.Context, sc scm.SCM, request *pb.RepositoryRequest) error {
	course, err := s.db.GetCourse(request.GetCourseID(), false)
	if err != nil {
		return err
	}
	repos, err := s.db.GetRepositories(&pb.Repository{
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
