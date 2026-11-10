package env

import (
	"errors"
	"os"
)

// DispatcherURL returns the URL of the test dispatcher from QUICKFEED_DISPATCHER_URL.
// An empty URL, the default, means that tests run in local Docker containers.
func DispatcherURL() string {
	return os.Getenv("QUICKFEED_DISPATCHER_URL")
}

// DispatcherToken returns the RPC token for the test dispatcher from QUICKFEED_DISPATCHER_TOKEN.
// It returns an error if the token is not set.
func DispatcherToken() (string, error) {
	token := os.Getenv("QUICKFEED_DISPATCHER_TOKEN")
	if token == "" {
		return "", errors.New("missing QUICKFEED_DISPATCHER_TOKEN for test dispatcher")
	}
	return token, nil
}
