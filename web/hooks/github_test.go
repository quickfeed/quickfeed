package hooks_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/steinfletcher/apitest"
)

const secret = "secret"

func TestHandlePush(t *testing.T) {
	wh := NewMockWebHook(qtest.Logger(t), secret)

	pushPayload := qlog.IndentJson(pushEvent)
	signature := hMAC([]byte(pushPayload), []byte(secret))

	tests := []struct {
		name       string
		payload    string
		signature  string
		wantStatus int
	}{
		{
			name:       "valid push event",
			payload:    pushPayload,
			signature:  signature,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid signature",
			payload:    pushPayload,
			signature:  "invalid",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid payload",
			payload:    pushPayload + "invalid",
			signature:  hMAC([]byte(pushPayload+"invalid"), []byte(secret)),
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apitest.New().
				HandlerFunc(wh.Handle()).
				Post(auth.Hook).
				Headers(map[string]string{
					"Content-Type":    "application/json",
					"X-Github-Event":  "push",
					"X-Hub-Signature": "sha256=" + tt.signature,
				}).
				Body(tt.payload).
				Expect(t).
				Status(tt.wantStatus).
				End()
		})
	}
}

func TestConcurrentHandlePush(t *testing.T) {
	const concurrentPushEvents = 1000
	wh := NewMockWebHook(qtest.Logger(t), secret)
	handlerFunc := wh.Handle()

	var wg sync.WaitGroup
	wg.Add(concurrentPushEvents)
	for i := 0; i < concurrentPushEvents; i++ {
		i := i
		go func() {
			myPushEvent := &github.PushEvent{
				Repo: &github.PushEventRepository{
					Name: github.String(fmt.Sprintf("repo-%02d", i)),
				},
			}
			pushPayload := qlog.IndentJson(&myPushEvent)
			signature := hMAC([]byte(pushPayload), []byte(secret))

			apitest.New().
				HandlerFunc(handlerFunc).
				Post(auth.Hook).
				Headers(map[string]string{
					"Content-Type":    "application/json",
					"X-Github-Event":  "push",
					"X-Hub-Signature": "sha256=" + signature,
				}).
				Body(pushPayload).
				Expect(t).
				Status(http.StatusOK).
				End()

			wg.Done()
		}()
	}
	wg.Wait()
	wh.wg.Wait()
	// All goroutines were executed
	if wh.totalCnt != concurrentPushEvents {
		t.Errorf("totalCnt = %d, want %d", wh.totalCnt, concurrentPushEvents)
	}
	// All goroutines should have completed.
	if wh.currentConcurrencyCnt != 0 {
		t.Errorf("currentConcurrencyCnt = %d, want 0", wh.currentConcurrencyCnt)
	}
}

var pushEvent = &github.PushEvent{
	Ref: github.String("refs/heads/master"),
	Repo: &github.PushEventRepository{
		ID:            github.Int64(1),
		Name:          github.String("meling-labs"),
		FullName:      github.String("qf104-2022/meling-labs"),
		DefaultBranch: github.String("master"),
	},
	Sender: &github.User{
		Login: github.String("meling"),
	},
	HeadCommit: &github.HeadCommit{
		ID:       github.String("c5b97d5ae6c19d5c5df71a34c7fbeeda2479ccbc"),
		Message:  github.String("Add a README.md"),
		Added:    []string{"lab1/README.md"},
		Removed:  []string{},
		Modified: []string{"lab2/README.md"},
	},
	Commits: []*github.HeadCommit{
		{
			ID:       github.String("c5b97d5ae6c19d5c5df71a34c7fbeeda2479ccbc"),
			Message:  github.String("Add a README.md"),
			Added:    []string{"lab1/README.md"},
			Removed:  []string{},
			Modified: []string{"lab2/README.md"},
		},
	},
}

// hMAC returns the HMAC signature for a message provided the secret key and hashFunc.
func hMAC(message, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}
