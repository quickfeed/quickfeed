package hooks

import (
	"fmt"

	"github.com/quickfeed/quickfeed/qf"
)

func (wh GitHubWebHook) getRepository(repoID int64) (*qf.Repository, error) {
	repos, err := wh.db.GetRepositories(&qf.Repository{ScmRepositoryID: uint64(repoID)})
	if err != nil {
		return nil, fmt.Errorf("getting repository by remote ID %d: %w", repoID, err)
	}
	if len(repos) != 1 {
		return nil, fmt.Errorf("unknown repository: %d", repoID)
	}
	return repos[0], nil
}
