package hooks

import (
	"context"
	"time"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

const (
	// maxSyncRetries is the maximum number of retries for rate-limited requests.
	maxSyncRetries = 3
)

// syncStudentRepos syncs all student repositories (forks of assignments repo) with the upstream
// assignments repository. This is called when a push event is received for the assignments repo.
func (wh GitHubWebHook) syncStudentRepos(ctx context.Context, scmClient scm.SCM, course *qf.Course, branch string) {
	repos, err := wh.db.GetRepositories(&qf.Repository{ScmOrganizationID: course.GetScmOrganizationID()})
	if err != nil {
		wh.logger.Errorf("Failed to get repositories for course %s: %v", course.GetName(), err)
		return
	}

	// Filter for student repos only
	var studentRepos []*qf.Repository
	for _, repo := range repos {
		if repo.IsStudentRepo() {
			studentRepos = append(studentRepos, repo)
		}
	}
	if len(studentRepos) == 0 {
		wh.logger.Debugf("No student repositories to sync for course %s", course.GetName())
		return
	}

	wh.logger.Infof("Synchronizing %d student repositories for course %s", len(studentRepos), course.GetName())
	start := time.Now()
	errCnt := 0
	for _, repo := range studentRepos {
		err := scmClient.SyncFork(ctx, &scm.SyncForkOptions{
			Organization: course.GetScmOrganizationName(),
			Repository:   repo.Name(),
			Branch:       branch,
			MaxRetries:   maxSyncRetries,
		})
		if err != nil {
			errCnt++
			wh.logger.Warnf("Failed to sync repository %s: %v", repo.Name(), err)
		}
	}

	duration := time.Since(start)
	wh.logger.Infof("Synchronized %d student repositories for course %s in %v (%d errors)",
		len(studentRepos)-errCnt, course.GetName(), duration, errCnt)
}
