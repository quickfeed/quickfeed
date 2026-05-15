package database

import (
	"context"
	"database/sql"

	"github.com/quickfeed/quickfeed/qf"
)

// CreateRepository creates a new repository record.
func (db *BunDB) CreateRepository(repo *qf.Repository) error {
	if repo.GetScmOrganizationID() == 0 || repo.GetScmRepositoryID() == 0 {
		return ErrCreateRepo
	}
	ctx := context.Background()
	switch {
	case repo.GetUserID() > 0:
		exists, err := db.conn.NewSelect().Model((*qf.User)(nil)).
			Where("id = ?", repo.GetUserID()).Exists(ctx)
		if err != nil {
			return err
		}
		if !exists {
			return sql.ErrNoRows
		}
	case repo.GetGroupID() > 0:
		exists, err := db.conn.NewSelect().Model((*qf.Group)(nil)).
			Where("id = ?", repo.GetGroupID()).Exists(ctx)
		if err != nil {
			return err
		}
		if !exists {
			return sql.ErrNoRows
		}
	case !repo.GetRepoType().IsCourseRepo():
		return ErrCreateRepo
	}
	_, err := db.conn.NewInsert().Model(repo).Exec(ctx)
	return err
}

// GetRepositories returns all repositories satisfying the given query.
func (db *BunDB) GetRepositories(query *qf.Repository) ([]*qf.Repository, error) {
	ctx := context.Background()
	var repos []*qf.Repository
	q := db.conn.NewSelect().Model(&repos)
	if query.GetScmOrganizationID() > 0 {
		q = q.Where("scm_organization_id = ?", query.GetScmOrganizationID())
	}
	if query.GetScmRepositoryID() > 0 {
		q = q.Where("scm_repository_id = ?", query.GetScmRepositoryID())
	}
	if query.GetUserID() > 0 {
		q = q.Where("user_id = ?", query.GetUserID())
	}
	if query.GetGroupID() > 0 {
		q = q.Where("group_id = ?", query.GetGroupID())
	}
	if query.GetRepoType() != 0 {
		q = q.Where("repo_type = ?", query.GetRepoType())
	}
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	return repos, nil
}

// DeleteRepository deletes the repository for the given remote provider's repository ID.
func (db *BunDB) DeleteRepository(scmRepositoryID uint64) error {
	ctx := context.Background()
	_, err := db.conn.NewDelete().Model((*qf.Repository)(nil)).
		Where("scm_repository_id = ?", scmRepositoryID).
		Exec(ctx)
	return err
}

// GetRepositoriesWithIssues returns repositories with their associated issues.
func (db *BunDB) GetRepositoriesWithIssues(query *qf.Repository) ([]*qf.Repository, error) {
	ctx := context.Background()
	var repos []*qf.Repository
	q := db.conn.NewSelect().Model(&repos).Relation("Issues")
	if query.GetScmOrganizationID() > 0 {
		q = q.Where("repository.scm_organization_id = ?", query.GetScmOrganizationID())
	}
	if query.GetScmRepositoryID() > 0 {
		q = q.Where("repository.scm_repository_id = ?", query.GetScmRepositoryID())
	}
	if query.GetUserID() > 0 {
		q = q.Where("repository.user_id = ?", query.GetUserID())
	}
	if query.GetGroupID() > 0 {
		q = q.Where("repository.group_id = ?", query.GetGroupID())
	}
	if query.GetRepoType() != 0 {
		q = q.Where("repository.repo_type = ?", query.GetRepoType())
	}
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	return repos, nil
}
