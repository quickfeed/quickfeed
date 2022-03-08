package database

import (
	pb "github.com/autograde/quickfeed/ag"
)

/// Repositories ///

// CreateRepository creates a new repository record.
func (db *GormDB) CreateRepository(repo *pb.Repository) error {
	if repo.OrganizationID == 0 || repo.RepositoryID == 0 {
		// both organization and repository must be non-zero
		return ErrCreateRepo
	}

	switch {
	case repo.UserID > 0:
		// check that user exists before creating repo in database
		err := db.conn.First(&pb.User{}, repo.UserID).Error
		if err != nil {
			return err
		}
	case repo.GroupID > 0:
		// check that group exists before creating repo in database
		err := db.conn.First(&pb.Group{}, repo.GroupID).Error
		if err != nil {
			return err
		}
	case !repo.RepoType.IsCourseRepo():
		// both user and group unset, then repository type must be an QuickFeed repo type
		return ErrCreateRepo
	}

	return db.conn.Create(repo).Error
}

// GetRepositoryByRemoteID fetches repository by provider's ID.
func (db *GormDB) GetRepositoryByRemoteID(remoteID uint64) (*pb.Repository, error) {
	var repo pb.Repository
	if err := db.conn.First(&repo, &pb.Repository{RepositoryID: remoteID}).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}

// GetRepositories fetches all repositories satisfying the given query.
func (db *GormDB) GetRepositories(query *pb.Repository) ([]*pb.Repository, error) {
	var repos []*pb.Repository
	if err := db.conn.Find(&repos, query).Error; err != nil {
		return nil, err
	}
	return repos, nil
}

// DeleteRepositoryByRemoteID deletes repository by provider's ID
func (db *GormDB) DeleteRepositoryByRemoteID(rid uint64) error {
	repo, err := db.GetRepositoryByRemoteID(rid)
	if err != nil {
		return err
	}
	return db.conn.Delete(repo).Error
}
