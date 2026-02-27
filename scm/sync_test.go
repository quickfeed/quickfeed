package scm

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/internal/qtest"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func githubResponse(status int, body string, headers map[string]string) *http.Response {
	h := make(http.Header)
	for k, v := range headers {
		h.Set(k, v)
	}
	var b io.ReadCloser = http.NoBody
	if body != "" {
		b = io.NopCloser(strings.NewReader(body))
	}
	return &http.Response{
		StatusCode: status,
		Body:       b,
		Header:     h,
	}
}

func TestSyncForkWithRetry(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			calls := 0
			s := NewGithubUserClient(qtest.Logger(t), "token")
			s.client = github.NewClient(&http.Client{
				Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
					calls++
					return githubResponse(http.StatusOK, "", nil), nil
				}),
			})

			err := s.SyncFork(t.Context(), &SyncForkOptions{
				Organization: "org",
				Repository:   "repo",
				Branch:       "master",
				MaxRetries:   3,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if calls != 1 {
				t.Errorf("expected 1 call, got %d", calls)
			}
		})
	})

	t.Run("RateLimitRetrySuccess", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			calls := 0
			s := NewGithubUserClient(qtest.Logger(t), "token")
			s.client = github.NewClient(&http.Client{
				Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
					calls++
					if calls == 1 {
						return githubResponse(http.StatusForbidden, `{"message": "rate limit exceeded"}`, map[string]string{
							"X-RateLimit-Limit":     "60",
							"X-RateLimit-Remaining": "0",
							"X-RateLimit-Reset":     fmt.Sprint(time.Now().Add(3 * time.Second).Unix()),
						}), nil
					}
					return githubResponse(http.StatusOK, "", nil), nil
				}),
			})

			start := time.Now()
			err := s.SyncFork(t.Context(), &SyncForkOptions{
				Organization: "org",
				Repository:   "repo",
				Branch:       "master",
				MaxRetries:   3,
			})
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if calls != 2 {
				t.Errorf("expected 2 calls, got %d", calls)
			}
			if duration < 3*time.Second {
				t.Errorf("expected retry delay of at least 3s, got %v", duration)
			}
		})
	})

	t.Run("AbuseRateLimitRetrySuccess", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			calls := 0
			s := NewGithubUserClient(qtest.Logger(t), "token")
			s.client = github.NewClient(&http.Client{
				Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
					calls++
					if calls == 1 {
						return githubResponse(http.StatusForbidden, `{"message": "You have exceeded a secondary rate limit.", "documentation_url": "https://docs.github.com/en/rest/overview/resources-in-the-rest-api#secondary-rate-limits"}`, map[string]string{
							"Content-Type": "application/json",
							"Retry-After":  "1",
						}), nil
					}
					return githubResponse(http.StatusOK, "", nil), nil
				}),
			})

			start := time.Now()
			err := s.SyncFork(t.Context(), &SyncForkOptions{
				Organization: "org",
				Repository:   "repo",
				Branch:       "master",
				MaxRetries:   3,
			})
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if calls != 2 {
				t.Errorf("expected 2 calls, got %d", calls)
			}
			if duration < time.Second {
				t.Errorf("expected retry delay of at least 1s, got %v", duration)
			}
		})
	})

	t.Run("NonRetryableError", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			calls := 0
			s := NewGithubUserClient(qtest.Logger(t), "token")
			s.client = github.NewClient(&http.Client{
				Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
					calls++
					return githubResponse(http.StatusInternalServerError, `{"message": "permanent error"}`, nil), nil
				}),
			})

			err := s.SyncFork(t.Context(), &SyncForkOptions{
				Organization: "org",
				Repository:   "repo",
				Branch:       "master",
				MaxRetries:   3,
			})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if calls != 1 {
				t.Errorf("expected 1 call, got %d", calls)
			}
		})
	})

	t.Run("MaxRetriesExceeded", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			calls := 0
			s := NewGithubUserClient(qtest.Logger(t), "token")
			s.client = github.NewClient(&http.Client{
				Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
					calls++
					return githubResponse(http.StatusForbidden, `{"message": "rate limit exceeded"}`, map[string]string{
						"X-RateLimit-Remaining": "0",
					}), nil
				}),
			})

			maxRetries := 2
			err := s.SyncFork(t.Context(), &SyncForkOptions{
				Organization: "org",
				Repository:   "repo",
				Branch:       "master",
				MaxRetries:   maxRetries,
			})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if calls != maxRetries {
				t.Errorf("expected %d calls, got %d", maxRetries, calls)
			}
		})
	})
}
