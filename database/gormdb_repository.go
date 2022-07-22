package database

import "github.com/quickfeed/quickfeed/qf"

// CreateRepository creates a new repository record.
func (db *GormDB) CreateRepository(repo *qf.Repository) error {
	if repo.OrganizationID == 0 || repo.RepositoryID == 0 {
		// both organization and repository must be non-zero
		return ErrCreateRepo
	}
	switch {
	case repo.UserID > 0:
		// check that user exists before creating repo in database
		if err := db.conn.First(&qf.User{}, repo.UserID).Error; err != nil {
			return err
		}
	case repo.GroupID > 0:
		// check that group exists before creating repo in database
		if err := db.conn.First(&qf.Group{}, repo.GroupID).Error; err != nil {
			return err
		}
	case !repo.RepoType.IsCourseRepo():
		// both user and group unset, then repository type must be an QuickFeed repo type
		return ErrCreateRepo
	}

	return db.conn.Create(repo).Error
}

// GetRepositories returns all repositories satisfying the given query.
func (db *GormDB) GetRepositories(query *qf.Repository) ([]*qf.Repository, error) {
	var repos []*qf.Repository
	if err := db.conn.Find(&repos, query).Error; err != nil {
		return nil, err
	}
	return repos, nil
}

// DeleteRepository deletes repository for the given remote provider's ID.
func (db *GormDB) DeleteRepository(remoteID uint64) error {
	return db.conn.Delete(&qf.Repository{}, &qf.Repository{RepositoryID: remoteID}).Error
}

// GetRepositoriesWithIssues gets repositories with issues
func (db *GormDB) GetRepositoriesWithIssues(query *qf.Repository) ([]*qf.Repository, error) {
	var repos []*qf.Repository
	if err := db.conn.Preload("Issues").Find(&repos, query).Error; err != nil {
		return nil, err
	}
	return repos, nil
}
