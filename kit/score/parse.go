package score

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/quickfeed/quickfeed/kit/internal/test"
)

var (
	// ErrScoreNotFound is returned if the parsed string did not contain a JSON score string.
	ErrScoreNotFound    = errors.New("score not found in string")
	ErrScoreInterval    = errors.New("score must be in the interval [0, MaxScore]")
	ErrMaxScore         = errors.New("max score must be greater than 0")
	ErrWeight           = errors.New("weight must be greater than 0")
	ErrEmptyTestName    = errors.New("test name must be specified")
	ErrSecret           = errors.New("secret field must match expected secret")
	ErrSuppressedSecret = errors.New("error suppressed to avoid revealing secret")
)

// parse returns a score object for the provided JSON string s
// which contains secret.
func parse(s, secret string) (*Score, error) {
	if strings.Contains(s, secret) {
		var sc Score
		err := json.Unmarshal([]byte(s), &sc)
		if err == nil {
			if err = sc.isValid(secret); err != nil {
				return nil, err
			}
			return &sc, nil
		}
		if strings.Contains(err.Error(), secret) {
			// this is probably not necessary, but to be safe
			return nil, ErrSuppressedSecret
		}
		return nil, err
	}
	return nil, ErrScoreNotFound
}

// isValid returns an error if the score object is invalid.
// Otherwise, nil is returned.
// If the given secret matches the score's secret value,
// the Secret field is redacted with the empty string "".
func (sc *Score) isValid(secret string) error {
	tName := sc.GetTestName()
	if tName == "" {
		return test.ErrMsg("", ErrEmptyTestName.Error())
	}
	if sc.GetMaxScore() <= 0 {
		return test.ErrMsg(tName, ErrMaxScore.Error())
	}
	if sc.GetWeight() <= 0 {
		return test.ErrMsg(tName, ErrWeight.Error())
	}
	if sc.GetScore() < 0 || sc.Score > sc.MaxScore {
		return test.ErrMsg(tName, ErrScoreInterval.Error())
	}
	if sc.GetSecret() != secret {
		return test.ErrMsg(tName, ErrSecret.Error())
	}
	sc.Secret = "" // redact the secret session key
	return nil
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
