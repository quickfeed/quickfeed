package hooks_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/hooks"
	"github.com/steinfletcher/apitest"
)

func TestHandlePush(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	user := qtest.CreateFakeUser(t, db, 1)
	repo := &qf.Repository{
		ID:             1,
		RepositoryID:   1,
		OrganizationID: 1,
		UserID:         user.ID,
		RepoType:       qf.Repository_USER,
		HTMLURL:        "https://github.com/qf104-2022/meling-labs",
	}
	if err := db.CreateRepository(repo); err != nil {
		t.Fatal(err)
	}
	course := &qf.Course{
		Name:             "QuickFeed Course 4",
		Code:             "QF104",
		Year:             2022,
		Tag:              "Spring",
		Provider:         "github",
		OrganizationID:   1,
		OrganizationName: "qf104-2022",
	}
	qtest.CreateCourse(t, db, user, course)

	lab1 := &qf.Assignment{
		CourseID:         course.ID,
		Name:             "lab1",
		RunScriptContent: "Script for assignment 1",
		Deadline:         "12.12.2021",
		AutoApprove:      false,
		Order:            1,
		IsGroupLab:       false,
	}

	lab2 := &qf.Assignment{
		CourseID:         course.ID,
		Name:             "lab2",
		RunScriptContent: "Script for assignment 1",
		Deadline:         "12.01.2022",
		AutoApprove:      false,
		Order:            2,
		IsGroupLab:       false,
	}

	for _, a := range []*qf.Assignment{lab1, lab2} {
		if err := db.CreateAssignment(a); err != nil {
			t.Fatal(err)
		}
	}

	const secret = "secret"
	env.SetFakeProvider(t)
	wh := hooks.NewGitHubWebHook(qtest.Logger(t), db, scm.NewSCMManager(nil), &ci.Local{}, secret)

	pushPayload := qlog.IndentJson(pushEvent)
	signature := hMAC([]byte(pushPayload), []byte(secret))

	apitest.New().
		// Debug().
		HandlerFunc(wh.Handle()).
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

	// Currently, we need to wait here for wh.handlePush() to finish, since it is running in a goroutine.
	// TODO(meling) find a more robust way to wait for the goroutine to finish.
	time.Sleep(2000 * time.Millisecond)
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
