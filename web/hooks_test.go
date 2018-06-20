package web_test

import (
	"context"
	"net/http"
	"testing"

	webhooks "gopkg.in/go-playground/webhooks.v3"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/models"
	"github.com/autograde/aguis/web"
	"gopkg.in/go-playground/webhooks.v3/github"
)

type mockRunner struct {
	runs []*ci.Job
}

func (m *mockRunner) Run(_ context.Context, job *ci.Job) (string, error) {
	m.runs = append(m.runs, job)
	return "", nil
}

func TestGithubHook(t *testing.T) {
	// This test have to be rewritten, since a lot of information is required to run a CI Build
	// Both the Assignments, CloneURL, Repository information and language type is needed, so
	// this is more like a integration test then a unit test right now.
	t.Skip("Test must be rewritten; see comment in code")
	db, cleanup := setup(t)
	defer cleanup()

	var user models.User
	if err := db.CreateUserFromRemoteIdentity(
		&user,
		&models.RemoteIdentity{
			Provider:    "github",
			RemoteID:    0,
			AccessToken: "",
		},
	); err != nil {
		t.Fatal(err)
	}

	runner := &mockRunner{}
	hook := web.GithubHook(nullLogger(), db, runner, "buildscripts")

	var h http.Header = make(map[string][]string)
	h.Set("X-Github-Event", string(github.PushEvent))
	hook(github.PushPayload{}, webhooks.Header(h))

	if len(runner.runs) != 1 {
		t.Fatalf("have %d runs want %d", len(runner.runs), 1)
	}

	const goImage = "golang:1.8.3"
	if runner.runs[0].Image != goImage {
		t.Errorf("have image %#v want %#v", runner.runs[0].Image, goImage)
	}
}
