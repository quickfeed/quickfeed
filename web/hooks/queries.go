package hooks

import (
	"fmt"

	"github.com/quickfeed/quickfeed/qf"
)

func (wh GitHubWebHook) getRepository(repoID int64) (*qf.Repository, error) {
	repos, err := wh.db.GetRepositories(&qf.Repository{ScmRepositoryID: uint64(repoID)})
	if err != nil {
		return nil, fmt.Errorf("failed to get repository by remote ID %d: %w", repoID, err)
	}
	if len(repos) != 1 {
		return nil, fmt.Errorf("unknown repository: %d", repoID)
	}
	return repos[0], nil
}

func (wh GitHubWebHook) getRepositoryWithIssues(repoID int64) (*qf.Repository, error) {
	repos, err := wh.db.GetRepositoriesWithIssues(&qf.Repository{ScmRepositoryID: uint64(repoID)})
	if err != nil {
		return nil, fmt.Errorf("failed to get repository by remote ID %d: %w", repoID, err)
	}
	if len(repos) != 1 {
		return nil, fmt.Errorf("unknown repository: %d", repoID)
	}
	return repos[0], nil
}

func (wh GitHubWebHook) getTask(taskID uint64) (*qf.Task, error) {
	tasks, err := wh.db.GetTasks(&qf.Task{ID: taskID})
	if err != nil {
		return nil, fmt.Errorf("failed to get task by ID %d: %w", taskID, err)
	}
	if len(tasks) != 1 {
		return nil, fmt.Errorf("unknown task: %d", taskID)
	}
	return tasks[0], nil
}
