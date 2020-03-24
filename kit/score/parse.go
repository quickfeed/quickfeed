package score

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// hiddenSecret is used to replace the global secret when parsing.
var hiddenSecret = "hidden"

// ErrScoreNotFound is returned if the parse string did not contain a
// JSON score string.
var ErrScoreNotFound = errors.New("score not found in string")

// Parse returns a score object for the provided JSON string s
// which contains secret.
func Parse(s, secret string) (*Score, error) {
	if strings.Contains(s, secret) {
		var sc Score
		err := json.Unmarshal([]byte(s), &sc)
		fmt.Println("here what the sc is :", &sc)
		if err == nil {
			if sc.Secret == secret {
				sc.Secret = hiddenSecret // overwrite secret
			}
			return &sc, nil
		}
		if strings.Contains(err.Error(), secret) {
			// this is probably not necessary, but to be safe
			return nil, errors.New("error suppressed to avoid revealing secret")
		}
		return nil, err
	}
	return nil, ErrScoreNotFound
}

// HasPrefix returns true if the provided string s has a parsable prefix string.
func HasPrefix(s string) bool {
	prefixes := []string{
		`{"Secret":`,
		`{"TestName":`,
		`{"Score":`,
		`{"MaxScore":`,
		`{"Weight":`,
	}
	trimmed := strings.TrimSpace(s)
	for _, prefix := range prefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	return false
}
