package hooks

import (
	"context"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
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
	logger := qlog.FromContext(ctx)
	repos, err := wh.db.GetRepositories(&qf.Repository{ScmOrganizationID: course.GetScmOrganizationID()})
	if err != nil {
		logger.Error("failed to get course repositories", label.Error, err)
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
		logger.Debug("no student repositories to synchronize")
		return
	}

	logger.Info("synchronizing student repositories", "count", len(studentRepos))
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
			logger.Warn("failed to synchronize repository", label.Repository, repo.Name(), label.Error, err)
		}
	}

	duration := time.Since(start)
	logger.Info("synchronized student repositories", label.Duration, duration, "count", len(studentRepos)-errCnt, "errors", errCnt)
}
