package hooks

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

const (
	// maxConcurrentSyncForks is the maximum number of concurrent fork sync operations.
	maxConcurrentSyncForks = 10
	// syncForkDelay is the delay between starting each sync operation to avoid GitHub abuse detection.
	syncForkDelay = 100 * time.Millisecond
	// maxSyncRetries is the maximum number of retries for rate-limited requests.
	maxSyncRetries = 3
	// initialRetryDelay is the initial delay before retrying a rate-limited request.
	initialRetryDelay = time.Second
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

	wh.logger.Infof("Syncing %d student repositories for course %s", len(studentRepos), course.GetName())
	start := time.Now()

	// counting semaphore: limit concurrent sync operations to avoid rate limiting
	sem := make(chan struct{}, maxConcurrentSyncForks)
	errCnt := int32(0)
	var wg sync.WaitGroup
	wg.Add(len(studentRepos))

	for i, repo := range studentRepos {
		// Stagger the start of goroutines to avoid burst requests
		if i > 0 {
			time.Sleep(syncForkDelay)
		}
		go func(r *qf.Repository) {
			defer wg.Done()
			sem <- struct{}{}        // acquire semaphore
			defer func() { <-sem }() // release semaphore

			err := wh.syncForkWithRetry(ctx, scmClient, course.GetScmOrganizationName(), r.Name(), branch)
			if err != nil {
				atomic.AddInt32(&errCnt, 1)
				wh.logger.Warnf("Failed to sync repository %s: %v", r.Name(), err)
			}
		}(repo)
	}

	wg.Wait()
	close(sem)

	duration := time.Since(start)
	if errCnt > 0 {
		wh.logger.Warnf("Synced student repositories for course %s in %v with %d errors", course.GetName(), duration, errCnt)
	} else {
		wh.logger.Infof("Successfully synced %d student repositories for course %s in %v", len(studentRepos), course.GetName(), duration)
	}
}

// syncForkWithRetry attempts to sync a fork with exponential backoff retry on rate limit errors.
func (wh GitHubWebHook) syncForkWithRetry(ctx context.Context, scmClient scm.SCM, org, repo, branch string) error {
	var lastErr error
	retryDelay := initialRetryDelay

	for attempt := 0; attempt <= maxSyncRetries; attempt++ {
		if attempt > 0 {
			wh.logger.Debugf("Retrying sync for %s (attempt %d/%d) after %v", repo, attempt, maxSyncRetries, retryDelay)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
			}
			retryDelay *= 2 // exponential backoff
		}

		err := scmClient.SyncFork(ctx, &scm.SyncForkOptions{
			Organization: org,
			Repository:   repo,
			Branch:       branch,
		})
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if this is a rate limit error that we should retry
		var rateLimitErr *github.RateLimitError
		var abuseErr *github.AbuseRateLimitError
		if errors.As(err, &rateLimitErr) {
			// Use the reset time from the rate limit error if available
			if rateLimitErr.Rate.Reset.After(time.Now()) {
				retryDelay = time.Until(rateLimitErr.Rate.Reset.Time) + time.Second
			}
			wh.logger.Warnf("Rate limited while syncing %s, will retry in %v", repo, retryDelay)
			continue
		}
		if errors.As(err, &abuseErr) {
			// Use the retry-after duration if provided
			if abuseErr.RetryAfter != nil {
				retryDelay = *abuseErr.RetryAfter
			}
			wh.logger.Warnf("Abuse rate limit hit while syncing %s, will retry in %v", repo, retryDelay)
			continue
		}

		// Non-rate-limit error, don't retry
		return err
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}
