package score

import (
	"encoding/json"
	"errors"
	"strings"
)

// ErrScoreNotFound is returned if the parse string did not contain a
// JSON score string.
var ErrScoreNotFound = errors.New("score not found in string")

const (
	ErrScoreInterval = "Score must be in the interval [0, MaxScore]"
	ErrMaxScore      = "MaxScore must be greater than 0"
	ErrWeight        = "Weight must be greater than 0"
	ErrEmptyTestName = "TestName must be specified"
	ErrSecret        = "Secret field must match expected secret"
)

// hiddenSecret is used to replace the global secret when parsing.
const hiddenSecret = "hidden"

// Parse returns a score object for the provided JSON string s
// which contains secret.
func Parse(s, secret string) (*Score, error) {
	if strings.Contains(s, secret) {
		var sc Score
		err := json.Unmarshal([]byte(s), &sc)
		if err == nil {
			if err = sc.IsValid(secret); err != nil {
				return nil, err
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

// IsValid returns an error if the score object is invalid.
// Otherwise, nil is returned.
// If the given secret matches the score's secret value,
// the Secret field is overwritten with the string "hidden".
func (sc *Score) IsValid(secret string) error {
	tName := sc.GetTestName()
	if tName == "" {
		return errMsg("", ErrEmptyTestName)
	}
	if sc.MaxScore <= 0 {
		return errMsg(tName, ErrMaxScore)
	}
	if sc.Weight <= 0 {
		return errMsg(tName, ErrWeight)
	}
	if sc.Score < 0 || sc.Score > sc.MaxScore {
		return errMsg(tName, ErrScoreInterval)
	}
	if sc.Secret != secret {
		return errMsg(tName, ErrSecret)
	}
	sc.Secret = hiddenSecret // overwrite secret
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
