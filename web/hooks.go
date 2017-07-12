package web

import (
	"fmt"

	webhooks "gopkg.in/go-playground/webhooks.v3"
)

// GithubHook handles events from GitHub.
func GithubHook(payload interface{}, header webhooks.Header) {
	fmt.Println(payload)
	fmt.Println(header)
}

// GitlabHook handles events from GitLab.
func GitlabHook(payload interface{}, header webhooks.Header) {
	fmt.Println(payload)
	fmt.Println(header)
}
