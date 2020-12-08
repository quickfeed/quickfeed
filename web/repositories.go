package web

import (
	"fmt"

	pb "github.com/autograde/quickfeed/ag"
)

func (s *AutograderService) getUserRepo(course *pb.Course, userID uint64) (*pb.Repository, error) {
	repoQuery := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		UserID:         userID,
		RepoType:       pb.Repository_USER,
	}
	repos, err := s.db.GetRepositories(repoQuery)
	if err != nil || len(repos) < 1 {
		return nil, fmt.Errorf("could not find user repository for user: %d, course: %s: %w", userID, course.GetCode(), err)
	}
	return repos[0], nil
}

func (s *AutograderService) getGroupRepo(course *pb.Course, groupID uint64) (*pb.Repository, error) {
	repoQuery := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		GroupID:        groupID,
		RepoType:       pb.Repository_GROUP,
	}
	repos, err := s.db.GetRepositories(repoQuery)
	if err != nil || len(repos) < 1 {
		return nil, fmt.Errorf("could not find group repository for group: %d, course: %s: %w", groupID, course.GetCode(), err)
	}
	return repos[0], nil
}
