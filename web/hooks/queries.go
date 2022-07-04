package hooks

import (
	"fmt"

	pb "github.com/quickfeed/quickfeed/ag"
)

func (wh GitHubWebHook) getRepository(repoID int64) (*pb.Repository, error) {
	repos, err := wh.db.GetRepositories(&pb.Repository{RepositoryID: uint64(repoID)})
	if err != nil {
		return nil, fmt.Errorf("failed to get repository by remote ID %d: %w", repoID, err)
	}
	if len(repos) != 1 {
		return nil, fmt.Errorf("unknown repository: %d", repoID)
	}
	return repos[0], nil
}

func (wh GitHubWebHook) getRepositoryWithIssues(repoID int64) (*pb.Repository, error) {
	repos, err := wh.db.GetRepositoriesWithIssues(&pb.Repository{RepositoryID: uint64(repoID)})
	if err != nil {
		return nil, fmt.Errorf("failed to get repository by remote ID %d: %w", repoID, err)
	}
	if len(repos) != 1 {
		return nil, fmt.Errorf("unknown repository: %d", repoID)
	}
	return repos[0], nil
}

func (wh GitHubWebHook) getTask(taskID uint64) (*pb.Task, error) {
	tasks, err := wh.db.GetTasks(&pb.Task{ID: taskID})
	if err != nil {
		return nil, fmt.Errorf("failed to get task by ID %d: %w", taskID, err)
	}
	if len(tasks) != 1 {
		return nil, fmt.Errorf("unknown task: %d", taskID)
	}
	return tasks[0], nil
}
