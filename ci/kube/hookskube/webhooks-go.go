package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"net/http"

	ghclient "github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/go-playground/webhooks.v3/github"
	webhooks "gopkg.in/go-playground/webhooks.v3"
)

const (
	// Secret given to github. Used for verifying the incoming objects.
	personalAccessTokenKey = "GITHUB_PERSONAL_TOKEN"
	// Personal Access Token created in github that allows us to make
	// calls into github.
	webhookSecretKey = "WEBHOOK_SECRET"
)

// GithubHandler holds necessary objects for communicating with the Github.
type GithubHandler struct {
	client *ghclient.Client
	ctx    context.Context
}

// HandlePushRequest is invoked whenever a push is registred
func (handler *GithubHandler) HandlePushRequest(payload interface{}, header webhooks.Header) {
		h := http.Header(header)
		event := github.Event(h.Get("X-GitHub-Event"))

		switch event {
		case github.PushEvent:
			p := payload.(github.PushPayload)
			log.Println("There is some push recieved!", p.Repository.Name)
			
		default:
			log.Println("default case sw")
		}
	}

func main() {
	flag.Parse()
	log.Print("gitwebhook sample started.")
	personalAccessToken := os.Getenv(personalAccessTokenKey)
	secretToken := os.Getenv(webhookSecretKey)

	// Set up the auth for being able to talk to Github. It's
	// odd that you have to also pass context around for the
	// calls even after giving it to client. But, whatever.
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: personalAccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := ghclient.NewClient(tc)

	h := &GithubHandler{
		client: client,
		ctx:    ctx,
	}

	hook := github.New(&github.Config{Secret: secretToken})
	hook.RegisterEvents(h.HandlePushRequest,github.PushEvent)
	
	err := webhooks.Run(hook, ":8080", "/")
	if err != nil {
		fmt.Println("Failed to run the webhook")
	}
}