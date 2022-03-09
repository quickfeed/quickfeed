package database

import (
	"fmt"

	pb "github.com/autograde/quickfeed/ag"
	"gorm.io/gorm"
)

// CreateRepository creates a new repository record.
func (db *GormDB) CreateRepository(repo *pb.Repository) error {
	if repo.OrganizationID == 0 || repo.RepositoryID == 0 {
		// both organization and repository must be non-zero
		return ErrCreateRepo
	}

	switch {
	case repo.UserID > 0:
		// check that user exists before creating repo in database
		if err := db.conn.First(&pb.User{}, repo.UserID).Error; err != nil {
			return err
		}
	case repo.GroupID > 0:
		// check that group exists before creating repo in database
		if err := db.conn.First(&pb.Group{}, repo.GroupID).Error; err != nil {
			return err
		}
	case !repo.RepoType.IsCourseRepo():
		// both user and group unset, then repository type must be an QuickFeed repo type
		return ErrCreateRepo
	}

	return db.conn.Create(repo).Error
}

// GetRepository returns the repository satisfying the given query.
// If more than one repository satisfies the query, an error is returned.
func (db *GormDB) GetRepository(query *pb.Repository) (*pb.Repository, error) {
	var repos []*pb.Repository
	if err := db.conn.Find(&repos, query).Error; err != nil {
		return nil, err
	}
	if len(repos) > 1 {
		return nil, fmt.Errorf("ambiguous query: found %d repositories", len(repos))
	} else if len(repos) == 0 {
		// no repositories found
		return nil, gorm.ErrRecordNotFound
	}
	return repos[0], nil
}

// DeleteRepository deletes repository for the given remote provider's ID.
func (db *GormDB) DeleteRepository(remoteID uint64) error {
	return db.conn.Delete(&pb.Repository{}, &pb.Repository{RepositoryID: remoteID}).Error
}
