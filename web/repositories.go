package web

import (
	"context"
	"fmt"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/scm"
)

func (s *AutograderService) getUserRepo(course *pb.Course, userID uint64) (*pb.Repository, error) {
	repo, err := s.db.GetRepository(&pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		UserID:         userID,
		RepoType:       pb.Repository_USER,
	})
	if err != nil {
		return nil, fmt.Errorf("could not find %s repository for user: %d: %w", course.GetCode(), userID, err)
	}
	return repo, nil
}

func (s *AutograderService) getGroupRepo(course *pb.Course, groupID uint64) (*pb.Repository, error) {
	repo, err := s.db.GetRepository(&pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		GroupID:        groupID,
		RepoType:       pb.Repository_GROUP,
	})
	if err != nil {
		return nil, fmt.Errorf("could not find %s repository for group: %d: %w", course.GetCode(), groupID, err)
	}
	return repo, nil
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

	repo, err := s.db.GetRepository(userRepoQuery)
	if err != nil {
		return "", err
	}
	return repo.GetHTMLURL(), nil
}

// isEmptyRepo returns nil if all repositories for the given course and student or group are empty,
// returns an error otherwise.
func (s *AutograderService) isEmptyRepo(ctx context.Context, sc scm.SCM, request *pb.RepositoryRequest) error {
	course, err := s.db.GetCourse(request.GetCourseID(), false)
	if err != nil {
		return err
	}
	repo, err := s.db.GetRepository(&pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		UserID:         request.GetUserID(),
		GroupID:        request.GetGroupID(),
	})
	if err != nil {
		return err
	}
	if !sc.RepositoryIsEmpty(ctx, &scm.RepositoryOptions{ID: repo.GetRepositoryID()}) {
		return fmt.Errorf("repository is not empty")
	}
	return nil
	// return isEmpty(ctx, sc, repos)
}
