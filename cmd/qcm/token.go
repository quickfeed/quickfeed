package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// tokenEnv is the environment variable that may hold the GitHub access token.
const tokenEnv = "GITHUB_ACCESS_TOKEN"

// tokenHelp explains how to give qcm a GitHub access token. It is printed
// when none could be found, since the fix is a one-time setup step that a
// teacher should not have to look up.
const tokenHelp = `qcm needs a GitHub access token to reach the course's repositories. Either

  - sign in with the GitHub CLI: gh auth login
    qcm then uses the token that 'gh auth token' prints; or
  - set ` + tokenEnv + `, or pass -token, to a personal access token created at
    https://github.com/settings/tokens with the 'repo' scope, or to a
    fine-grained token with read access to the course organization's repositories.

In both cases the account must be a member of the course organization, and an
organization that enforces SAML single sign-on must have authorized the token.`

// ghRunner runs the GitHub CLI with the given arguments and returns its
// standard output. Tests substitute it to stand in for an installed, signed-in,
// or missing CLI.
type ghRunner func(args ...string) ([]byte, error)

// runGh runs the installed GitHub CLI.
func runGh(args ...string) ([]byte, error) {
	return exec.Command("gh", args...).Output()
}

// gh is the GitHub CLI runner used to look up a token; see resolveToken.
var gh ghRunner = runGh

// resolveToken returns the GitHub access token qcm should use. An explicit
// token, from -token or the environment, always wins, so that a teacher can
// act as a different account than the GitHub CLI is signed in with. Otherwise
// the token of the GitHub CLI's login is used, which is the setup most teachers
// already have. When neither is available the error explains both options.
func resolveToken(explicit string, gh ghRunner) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	token, err := ghAuthToken(gh)
	if err != nil {
		return "", fmt.Errorf("no GitHub access token found: %w\n\n%s", err, tokenHelp)
	}
	return token, nil
}

// ghAuthToken returns the token the GitHub CLI holds for github.com, or an
// error saying why there is none: the CLI is not installed, or it is not signed
// in. The CLI is asked rather than its configuration files read, since it may
// keep the token in the operating system's keyring.
func ghAuthToken(gh ghRunner) (string, error) {
	out, err := gh("auth", "token", "--hostname", "github.com")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", errors.New("the GitHub CLI (gh) is not installed")
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			// The CLI's own explanation, typically that no account is signed in.
			return "", fmt.Errorf("gh auth token: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("gh auth token: %w", err)
	}
	token := strings.TrimSpace(string(out))
	if token == "" {
		return "", errors.New("gh auth token printed no token")
	}
	return token, nil
}
